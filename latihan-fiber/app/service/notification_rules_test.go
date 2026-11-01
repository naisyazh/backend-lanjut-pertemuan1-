package service

import (
	"testing"

	"latihan-fiber/app/model"
)

func TestBuildMessage(t *testing.T) {
	user := model.User{
		ID:       10,
		Username: "mhsuji",
		Email:    "mhsuji@student.unair.ac.id",
	}

	event := model.EnrollmentEvent{
		UserID:         10,
		CourseID:       2,
		CourseCode:     "IF102",
		CourseName:     "Basis Data Lanjut",
		FromCourseCode: "IF101",
		FromCourseName: "Basis Data",
	}

	t.Run("created", func(t *testing.T) {
		msg, err := BuildMessage(model.EventEnrollmentCreated, event, user)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if msg.Subject != "KRS: berhasil terdaftar di IF102 Basis Data Lanjut" {
			t.Errorf("subject mismatch: got %q", msg.Subject)
		}
		if msg.To != user.Email {
			t.Errorf("to mismatch: got %q", msg.To)
		}
	})

	t.Run("dropped", func(t *testing.T) {
		msg, err := BuildMessage(model.EventEnrollmentDropped, event, user)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if msg.Subject != "KRS: keluar dari IF102 Basis Data Lanjut" {
			t.Errorf("subject mismatch: got %q", msg.Subject)
		}
	})

	t.Run("swapped", func(t *testing.T) {
		msg, err := BuildMessage(model.EventEnrollmentSwapped, event, user)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if msg.Subject != "KRS: pindah dari IF101 Basis Data ke IF102 Basis Data Lanjut" {
			t.Errorf("subject mismatch: got %q", msg.Subject)
		}
	})

	t.Run("unknown event", func(t *testing.T) {
		_, err := BuildMessage("unknown", event, user)
		if err == nil {
			t.Fatal("expected error for unknown event, got nil")
		}
	})
}

func TestBuildQuotaChangedMessage(t *testing.T) {
	user := model.User{
		ID:       10,
		Username: "mhsuji",
		Email:    "mhsuji@student.unair.ac.id",
	}

	event := model.QuotaChangedEvent{
		CourseID:   2,
		CourseCode: "IF102",
		CourseName: "Basis Data Lanjut",
		NewQuota:   40,
		OldQuota:   30,
	}

	msg := BuildQuotaChangedMessage(event, user)
	expectedSubject := "KRS: kuota IF102 Basis Data Lanjut berubah menjadi 40"
	if msg.Subject != expectedSubject {
		t.Errorf("subject mismatch: want %q, got %q", expectedSubject, msg.Subject)
	}
	if msg.To != user.Email {
		t.Errorf("to mismatch: want %q, got %q", user.Email, msg.To)
	}
}
