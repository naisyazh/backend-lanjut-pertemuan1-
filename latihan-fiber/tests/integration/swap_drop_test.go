//go:build integration

package integration_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"latihan-fiber/app/model"
)

func (a api) swap(t *testing.T, student model.User, fromCourseID, toCourseID int) (int, string) {
	t.Helper()
	token, err := a.jwt.GenerateAccessToken(model.AuthUser{UserID: student.ID, Username: student.Username, Role: "user"})
	if err != nil {
		t.Errorf("membuat token: %v", err)
		return 0, ""
	}
	body, _ := json.Marshal(model.SwapRequest{FromCourseID: fromCourseID, ToCourseID: toCourseID})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/enrollments/swap", strings.NewReader(string(body)))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.app.Test(req, -1)
	if err != nil {
		t.Errorf("request swap: %v", err)
		return 0, ""
	}
	defer resp.Body.Close()
	var respBody struct {
		Code string `json:"code"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&respBody)
	return resp.StatusCode, respBody.Code
}

func (a api) drop(t *testing.T, student model.User, courseID int) int {
	t.Helper()
	token, err := a.jwt.GenerateAccessToken(model.AuthUser{UserID: student.ID, Username: student.Username, Role: "user"})
	if err != nil {
		t.Errorf("membuat token: %v", err)
		return 0
	}
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/courses/"+strconv.Itoa(courseID)+"/enroll", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.app.Test(req, -1)
	if err != nil {
		t.Errorf("request drop: %v", err)
		return 0
	}
	defer resp.Body.Close()
	return resp.StatusCode
}

// TestEnrollEndpoint_Swap_TargetCourseFull menguji skenario D2.1:
// Mahasiswa sudah terdaftar di kelas asal. Kelas tujuan sudah penuh.
// Mahasiswa mencoba swap ke kelas tujuan -> 409 COURSE_FULL.
// Mahasiswa TETAP terdaftar di kelas asal, enrolled_count kedua kelas tidak berubah,
// dan tidak ada event enrollment.swapped di outbox.
func TestEnrollEndpoint_Swap_TargetCourseFull(t *testing.T) {
	a := newAPI(t)
	student := seedStudent(t)
	otherStudent := seedStudent(t)

	// Kelas asal (kuota 10)
	fromCourse := seedCourse(t, 3, 10)
	// Kelas tujuan (kuota 1, akan dipenuhi oleh otherStudent)
	toCourse := seedCourse(t, 3, 1)

	// Daftarkan student ke fromCourse
	if status, _ := a.enroll(t, student, fromCourse.ID); status != http.StatusCreated {
		t.Fatalf("daftar kelas asal: status = %d, seharusnya 201", status)
	}

	// Daftarkan otherStudent ke toCourse hingga penuh (kuota 1/1)
	if status, _ := a.enroll(t, otherStudent, toCourse.ID); status != http.StatusCreated {
		t.Fatalf("daftar kelas tujuan: status = %d, seharusnya 201", status)
	}

	// Hitung enrolled_count sebelum swap
	fromBefore := countRows(t, `SELECT enrolled_count FROM courses WHERE id = $1`, fromCourse.ID)
	toBefore := countRows(t, `SELECT enrolled_count FROM courses WHERE id = $1`, toCourse.ID)

	// Coba swap ke toCourse yang sudah penuh
	status, code := a.swap(t, student, fromCourse.ID, toCourse.ID)
	if status != http.StatusConflict || code != "COURSE_FULL" {
		t.Fatalf("swap kelas penuh: dapat status %d code %s, seharusnya 409 COURSE_FULL", status, code)
	}

	// Pastikan mahasiswa TETAP terdaftar di kelas asal
	enrolledFrom := countRows(t, `SELECT COUNT(*) FROM enrollments WHERE user_id = $1 AND course_id = $2`, student.ID, fromCourse.ID)
	if enrolledFrom != 1 {
		t.Errorf("mahasiswa harus tetap terdaftar di kelas asal, got %d", enrolledFrom)
	}

	// Pastikan mahasiswa TIDAK terdaftar di kelas tujuan
	enrolledTo := countRows(t, `SELECT COUNT(*) FROM enrollments WHERE user_id = $1 AND course_id = $2`, student.ID, toCourse.ID)
	if enrolledTo != 0 {
		t.Errorf("mahasiswa tidak boleh terdaftar di kelas tujuan, got %d", enrolledTo)
	}

	// Pastikan enrolled_count kedua kelas tidak berubah
	fromAfter := countRows(t, `SELECT enrolled_count FROM courses WHERE id = $1`, fromCourse.ID)
	toAfter := countRows(t, `SELECT enrolled_count FROM courses WHERE id = $1`, toCourse.ID)
	if fromBefore != fromAfter || toBefore != toAfter {
		t.Errorf("enrolled_count berubah: from (%d -> %d), to (%d -> %d)", fromBefore, fromAfter, toBefore, toAfter)
	}

	// Pastikan TIDAK ada event enrollment.swapped di outbox
	swappedEvents := countRows(t, `SELECT COUNT(*) FROM outbox WHERE (payload->>'user_id')::INT = $1 AND event_type = $2`,
		student.ID, model.EventEnrollmentSwapped)
	if swappedEvents != 0 {
		t.Errorf("outbox tidak boleh berisi event enrollment.swapped, got %d", swappedEvents)
	}
}

// TestEnrollEndpoint_Drop_Success menguji skenario D2.2:
// Mahasiswa keluar dari kelas (DELETE /courses/:id/enroll).
// Response 204, baris enrollment hilang, enrolled_count berkurang 1,
// dan outbox berisi tepat satu event enrollment.dropped.
func TestEnrollEndpoint_Drop_Success(t *testing.T) {
	a := newAPI(t)
	student := seedStudent(t)
	course := seedCourse(t, 3, 10)

	// Daftarkan student terlebih dahulu
	if status, _ := a.enroll(t, student, course.ID); status != http.StatusCreated {
		t.Fatalf("pendaftaran awal: status = %d, seharusnya 201", status)
	}

	countBefore := countRows(t, `SELECT enrolled_count FROM courses WHERE id = $1`, course.ID)
	if countBefore != 1 {
		t.Fatalf("enrolled_count sebelum drop = %d, seharusnya 1", countBefore)
	}

	// Drop kelas
	status := a.drop(t, student, course.ID)
	if status != http.StatusNoContent {
		t.Fatalf("drop enrollment: status = %d, seharusnya 204", status)
	}

	// Pastikan baris enrollment hilang
	enrolled := countRows(t, `SELECT COUNT(*) FROM enrollments WHERE user_id = $1 AND course_id = $2`, student.ID, course.ID)
	if enrolled != 0 {
		t.Errorf("baris enrollment harus 0 setelah drop, got %d", enrolled)
	}

	// Pastikan enrolled_count berkurang satu
	countAfter := countRows(t, `SELECT enrolled_count FROM courses WHERE id = $1`, course.ID)
	if countAfter != countBefore-1 {
		t.Errorf("enrolled_count setelah drop = %d, seharusnya %d", countAfter, countBefore-1)
	}

	// Pastikan outbox berisi tepat satu event enrollment.dropped
	droppedEvents := countRows(t, `SELECT COUNT(*) FROM outbox WHERE (payload->>'user_id')::INT = $1 AND event_type = $2`,
		student.ID, model.EventEnrollmentDropped)
	if droppedEvents != 1 {
		t.Errorf("event enrollment.dropped = %d, seharusnya 1", droppedEvents)
	}
}
