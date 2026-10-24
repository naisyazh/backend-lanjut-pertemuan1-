// Alat uji beban: membuktikan ada atau tidaknya race condition.
// Folder tools/ bukan bagian aplikasi dan tidak ikut di-deploy.
//
// Cara pakai (server harus sedang berjalan di terminal lain):
//
//	go run ./tools/loadtest -mode crowd    # 60 mahasiswa berebut kelas IF399 (kuota 5)
//	go run ./tools/loadtest -mode credits  # 10 mahasiswa mendaftar 12 kelas sekaligus
//	go run ./tools/loadtest -mode swap     # 20 pasang mahasiswa pindah kelas berlawanan
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ---------- konfigurasi —————————————————————————————————————————————-

const (
	baseURL   = "http://localhost:3002/api/v1"
	jwtSecret = "a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456"
	jwtIssuer = "praktikum-backend"
)

// ID kelas sesuai data seed migration 005 (urutan INSERT).
// Sesuaikan bila ID di database Anda berbeda.
const (
	courseIF101 = 1  // Pemrograman Backend Lanjut
	courseIF102 = 2  // Basis Data Lanjut
	courseIF399 = 13 // Kelas Rebutan, kuota 5
)

// ID 12 kelas untuk uji batas SKS (masing-masing 3 SKS = total 36 > 24).
var twelveCoursesIDs = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}

// ---------- JWT helper ———————————————————————————————————————————————

// makeToken membuat access token untuk user dengan id dan role tertentu.
// Alat ini mencetak token sendiri karena ia memegang JWT_SECRET.
// Itulah sebabnya JWT_SECRET tidak boleh bocor: siapa pun yang
// memegangnya dapat mencetak token atas nama siapa saja.
func makeToken(userID int, username, role string) string {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":      strconv.Itoa(userID),
		"username": username,
		"role":     role,
		"iss":      jwtIssuer,
		"iat":      now.Unix(),
		"exp":      now.Add(2 * time.Hour).Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte(jwtSecret))
	if err != nil {
		fmt.Fprintf(os.Stderr, "gagal membuat token: %v\n", err)
		os.Exit(1)
	}
	return signed
}

// ---------- HTTP helper ——————————————————————————————————————————————

func doPost(url, token string, body any) string {
	var buf *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		buf = bytes.NewBuffer(b)
	} else {
		buf = bytes.NewBuffer(nil)
	}
	req, _ := http.NewRequest(http.MethodPost, url, buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "ERR_NETWORK"
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return parseResult(resp.StatusCode, raw)
}

func doDelete(url, token string) string {
	req, _ := http.NewRequest(http.MethodDelete, url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "ERR_NETWORK"
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return parseResult(resp.StatusCode, raw)
}

func parseResult(status int, raw []byte) string {
	if status == http.StatusCreated || status == http.StatusOK || status == http.StatusNoContent {
		return fmt.Sprintf("%d OK", status)
	}
	var out struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(raw, &out); err != nil || out.Code == "" {
		return fmt.Sprintf("%d", status)
	}
	return fmt.Sprintf("%d %s", status, out.Code)
}

// ---------- fungsi fire —————————————————————————————————————————————-

// fire menjalankan n pekerjaan secara serentak dan menghitung hasilnya.
//
// Setiap goroutine menunggu di depan "gerbang" start yang sama. Gerbang
// baru dibuka setelah SELURUH goroutine siap, sehingga request benar-benar
// berangkat hampir bersamaan — bukan bergiliran sesuai urutan dibuatnya.
func fire(n int, job func(i int) string) map[string]int {
	var wg sync.WaitGroup
	start := make(chan struct{})
	results := make(chan string, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start           // tunggu gerbang dibuka
			results <- job(i) // kirim request dan taruh hasilnya
		}(i)
	}

	close(start) // buka gerbang: semua goroutine berangkat hampir bersamaan
	wg.Wait()    // tunggu SEMUA goroutine selesai
	close(results)

	tally := map[string]int{}
	for r := range results {
		tally[r]++
	}
	return tally
}

// ---------- helpers ——————————————————————————————————————————————————

// getMhsUserIDs mengambil ID ke-60 akun mhs001-mhs060 lewat endpoint /users.
// Admin token digunakan untuk query; user ID dibutuhkan agar JWT valid.
func getMhsUserIDs(count int) []int {
	adminToken := makeToken(1, "admin", "admin")
	ids := make([]int, 0, count)

	// Ambil semua user dengan search prefix "mhs", limit 100
	url := fmt.Sprintf("%s/users?search=mhs&limit=100&sort=id&order=asc", baseURL)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gagal menghubungi server: %v\nPastikan server berjalan.\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Data []struct {
			ID       int    `json:"id"`
			Username string `json:"username"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Fprintf(os.Stderr, "respons tidak bisa diparsing: %s\n", string(body))
		os.Exit(1)
	}
	if len(result.Data) < count {
		fmt.Fprintf(os.Stderr,
			"hanya ditemukan %d akun mhs, butuh %d. Jalankan migration 006 terlebih dahulu.\n",
			len(result.Data), count)
		os.Exit(1)
	}

	for i := 0; i < count; i++ {
		ids = append(ids, result.Data[i].ID)
	}
	return ids
}

// ---------- mode crowd ———————————————————————————————————————————————

// modeCrowd: 60 mahasiswa berebut kelas IF399 (kuota 5) secara serentak.
// Hasil yang benar setelah fix: tepat 5 berhasil, 55 COURSE_FULL.
func modeCrowd(userIDs []int) {
	fmt.Println("60 mahasiswa berebut kelas IF399 secara serentak...")
	t0 := time.Now()
	tally := fire(60, func(i int) string {
		uid := userIDs[i]
		token := makeToken(uid, fmt.Sprintf("mhs%03d", i+1), "user")
		url := fmt.Sprintf("%s/courses/%d/enroll", baseURL, courseIF399)
		return doPost(url, token, nil)
	})
	elapsed := time.Since(t0)

	fmt.Println("hasil request:")
	for k, v := range tally {
		fmt.Printf("  %-35s %d\n", k, v)
	}
	fmt.Printf("waktu total : %v\n\n", elapsed.Round(time.Millisecond))
	fmt.Println("Periksa di database:")
	fmt.Printf("  SELECT code, quota, enrolled_count FROM courses WHERE code='IF399';\n")
	fmt.Printf("  SELECT COUNT(*) FROM enrollments e JOIN courses c ON c.id=e.course_id WHERE c.code='IF399';\n")
}

// ---------- mode credits —————————————————————————————————————————————

// modeCredits: 10 mahasiswa, masing-masing mendaftar ke 12 kelas 3-SKS
// secara serentak (total 36 SKS > batas 24 SKS).
// Hasil yang benar: setiap mahasiswa hanya berhasil mendaftar ke ≤ 8 kelas (≤24 SKS).
func modeCredits(userIDs []int) {
	nUsers := 10
	nCourses := len(twelveCoursesIDs)
	n := nUsers * nCourses // 120 request total

	fmt.Printf("%d mahasiswa, masing-masing mendaftar ke %d kelas 3 SKS secara serentak (batas 24 SKS)...\n",
		nUsers, nCourses)
	t0 := time.Now()
	tally := fire(n, func(i int) string {
		userIdx := i % nUsers
		courseIdx := i / nUsers
		uid := userIDs[userIdx]
		token := makeToken(uid, fmt.Sprintf("mhs%03d", userIdx+1), "user")
		courseID := twelveCoursesIDs[courseIdx]
		url := fmt.Sprintf("%s/courses/%d/enroll", baseURL, courseID)
		return doPost(url, token, nil)
	})
	elapsed := time.Since(t0)

	fmt.Println("hasil request:")
	for k, v := range tally {
		fmt.Printf("  %-35s %d\n", k, v)
	}
	fmt.Printf("waktu total : %v\n\n", elapsed.Round(time.Millisecond))
	fmt.Println("Periksa di database (tidak ada mahasiswa dengan total_credits > 24):")
	fmt.Println("  SELECT u.username, SUM(c.credits) AS total")
	fmt.Println("  FROM enrollments e")
	fmt.Println("  JOIN users u ON u.id=e.user_id")
	fmt.Println("  JOIN courses c ON c.id=e.course_id")
	fmt.Println("  WHERE u.username LIKE 'mhs%'")
	fmt.Println("  GROUP BY u.username HAVING SUM(c.credits) > 24;")
}

// ---------- mode swap ————————————————————————————————————————————————

// modeSwap: 20 pasang mahasiswa (mhs001↔mhs020 dan mhs021↔mhs040)
// pindah kelas berlawanan arah (IF101→IF102 dan IF102→IF101) secara serentak.
// Hasil yang benar: 40 berhasil, 0 deadlock, kedua kelas konsisten.
func modeSwap(userIDs []int) {
	// Persiapan: pastikan mhs001-mhs020 terdaftar di IF101
	// dan mhs021-mhs040 terdaftar di IF102
	fmt.Println("Mempersiapkan data awal untuk mode swap...")
	fmt.Println("(Daftarkan mhs001-mhs020 ke IF101 dan mhs021-mhs040 ke IF102 terlebih dahulu)")
	fmt.Println("Mendaftarkan mhs001-mhs020 ke IF101...")

	for i := 0; i < 20; i++ {
		uid := userIDs[i]
		token := makeToken(uid, fmt.Sprintf("mhs%03d", i+1), "user")
		url := fmt.Sprintf("%s/courses/%d/enroll", baseURL, courseIF101)
		result := doPost(url, token, nil)
		if result != "201 OK" && result != "409 ALREADY_ENROLLED" {
			fmt.Printf("  mhs%03d IF101: %s\n", i+1, result)
		}
	}

	fmt.Println("Mendaftarkan mhs021-mhs040 ke IF102...")
	for i := 20; i < 40; i++ {
		uid := userIDs[i]
		token := makeToken(uid, fmt.Sprintf("mhs%03d", i+1), "user")
		url := fmt.Sprintf("%s/courses/%d/enroll", baseURL, courseIF102)
		result := doPost(url, token, nil)
		if result != "201 OK" && result != "409 ALREADY_ENROLLED" {
			fmt.Printf("  mhs%03d IF102: %s\n", i+1, result)
		}
	}

	fmt.Println("\n20 pasang mahasiswa pindah kelas berlawanan arah secara serentak...")
	t0 := time.Now()

	// mhs001-mhs020 pindah IF101 -> IF102
	// mhs021-mhs040 pindah IF102 -> IF101
	tally := fire(40, func(i int) string {
		var fromID, toID int
		var uid int

		if i < 20 {
			uid = userIDs[i]
			fromID = courseIF101
			toID = courseIF102
		} else {
			uid = userIDs[i]
			fromID = courseIF102
			toID = courseIF101
		}

		token := makeToken(uid, fmt.Sprintf("mhs%03d", i+1), "user")
		url := fmt.Sprintf("%s/enrollments/swap", baseURL)
		body := map[string]int{
			"from_course_id": fromID,
			"to_course_id":   toID,
		}
		return doPost(url, token, body)
	})
	elapsed := time.Since(t0)

	fmt.Println("hasil request:")
	for k, v := range tally {
		fmt.Printf("  %-35s %d\n", k, v)
	}
	fmt.Printf("waktu total : %v\n\n", elapsed.Round(time.Millisecond))
	fmt.Println("Periksa di database:")
	fmt.Println("  SELECT code, enrolled_count,")
	fmt.Println("    (SELECT COUNT(*) FROM enrollments e WHERE e.course_id=c.id) AS baris_enrollments")
	fmt.Println("  FROM courses WHERE code IN ('IF101','IF102');")
}

// ---------- main —————————————————————————————————————————————————————

func main() {
	mode := flag.String("mode", "crowd", "mode uji: crowd | credits | swap")
	flag.Parse()

	fmt.Printf("=== Alat Uji Beban Modul 9 — mode: %s ===\n\n", *mode)

	// Ambil ID user mhs001-mhs060 dari server
	var userCount int
	switch *mode {
	case "swap":
		userCount = 40
	case "credits":
		userCount = 10
	default:
		userCount = 60
	}

	fmt.Printf("Mengambil ID %d akun mhs dari server...\n", userCount)
	userIDs := getMhsUserIDs(userCount)
	fmt.Printf("Berhasil: %d akun ditemukan (ID %d - %d)\n\n", len(userIDs), userIDs[0], userIDs[len(userIDs)-1])

	switch *mode {
	case "crowd":
		modeCrowd(userIDs)
	case "credits":
		modeCredits(userIDs)
	case "swap":
		modeSwap(userIDs)
	default:
		fmt.Fprintf(os.Stderr, "mode tidak dikenal: %s\nGunakan: crowd | credits | swap\n", *mode)
		os.Exit(1)
	}
}
