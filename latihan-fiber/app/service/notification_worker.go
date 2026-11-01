package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"

	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
	"latihan-fiber/helper"
)

// NotificationWorker mengerjakan task "notification:send".
//
// Bagi worker, Asynq adalah framework-nya — sama seperti Fiber bagi API.
// Karena itu method di sini menerima *asynq.Task, seperti method service
// lain menerima *fiber.Ctx.
type NotificationWorker struct {
	users         repository.UserRepository
	courses       repository.CourseRepository
	notifications repository.NotificationRepository
	tx            *repository.TxManager
	mailer        helper.Mailer
	logger        *slog.Logger
}

func NewNotificationWorker(
	users repository.UserRepository,
	courses repository.CourseRepository,
	notifications repository.NotificationRepository,
	tx *repository.TxManager,
	mailer helper.Mailer,
	logger *slog.Logger,
) *NotificationWorker {
	return &NotificationWorker{
		users:         users,
		courses:       courses,
		notifications: notifications,
		tx:            tx,
		mailer:        mailer,
		logger:        logger,
	}
}

// HandleSend mengembalikan error untuk meminta Asynq melakukan retry nanti,
// atau error yang membungkus asynq.SkipRetry untuk menyatakan "jangan di-retry,
// task ini tidak mungkin berhasil".
func (w *NotificationWorker) HandleSend(ctx context.Context, t *asynq.Task) error {
	var job model.NotificationJob
	if err := json.Unmarshal(t.Payload(), &job); err != nil {
		return fmt.Errorf("%w: payload task rusak: %v", asynq.SkipRetry, err)
	}

	if job.EventType == model.EventCourseQuotaChanged {
		return w.handleQuotaChanged(ctx, job)
	}

	return w.handleEnrollment(ctx, job)
}

func (w *NotificationWorker) handleEnrollment(ctx context.Context, job model.NotificationJob) error {
	var event model.EnrollmentEvent
	if err := json.Unmarshal(job.Payload, &event); err != nil {
		return fmt.Errorf("%w: payload event %d rusak: %v", asynq.SkipRetry, job.OutboxID, err)
	}

	user, err := w.users.FindByID(ctx, event.UserID)
	if errors.Is(err, repository.ErrNotFound) {
		// Akunnya sudah dihapus. Melakukan retry tidak akan memunculkannya.
		return fmt.Errorf("%w: user %d tidak ditemukan", asynq.SkipRetry, event.UserID)
	}
	if err != nil {
		return err // database sedang bermasalah: layak di-retry
	}

	msg, err := BuildMessage(job.EventType, event, user)
	if err != nil {
		return fmt.Errorf("%w: %v", asynq.SkipRetry, err)
	}

	alreadySent := false
	err = w.tx.WithTxOnce(ctx, func(tx repository.DBTX) error {
		claimed, err := w.notifications.WithTx(tx).Claim(ctx, model.Notification{
			OutboxID: job.OutboxID,
			UserID:   user.ID,
			Channel:  "email",
			Subject:  msg.Subject,
			Body:     msg.Body,
		})
		if err != nil {
			return err
		}
		if !claimed {
			alreadySent = true
			return nil
		}
		return w.mailer.Send(ctx, msg)
	})

	if errors.Is(err, helper.ErrPermanentDelivery) {
		return fmt.Errorf("%w: %v", asynq.SkipRetry, err)
	}
	if err != nil {
		return err
	}

	retry, _ := asynq.GetRetryCount(ctx)
	w.logger.Info("notification_processed",
		slog.Int64("outbox_id", job.OutboxID),
		slog.String("event", job.EventType),
		slog.Int("retry", retry),
		slog.Bool("already_sent", alreadySent))
	return nil
}

func (w *NotificationWorker) handleQuotaChanged(ctx context.Context, job model.NotificationJob) error {
	var event model.QuotaChangedEvent
	if err := json.Unmarshal(job.Payload, &event); err != nil {
		return fmt.Errorf("%w: payload event %d rusak: %v", asynq.SkipRetry, job.OutboxID, err)
	}

	participants, err := w.courses.FindParticipants(ctx, event.CourseID)
	if err != nil {
		return err
	}

	for _, u := range participants {
		msg := BuildQuotaChangedMessage(event, u)
		alreadySent := false

		err = w.tx.WithTxOnce(ctx, func(tx repository.DBTX) error {
			claimed, err := w.notifications.WithTx(tx).Claim(ctx, model.Notification{
				OutboxID: job.OutboxID,
				UserID:   u.ID,
				Channel:  "email",
				Subject:  msg.Subject,
				Body:     msg.Body,
			})
			if err != nil {
				return err
			}
			if !claimed {
				alreadySent = true
				return nil
			}
			return w.mailer.Send(ctx, msg)
		})

		if errors.Is(err, helper.ErrPermanentDelivery) {
			w.logger.Warn("user_notification_skipped_permanent",
				slog.Int64("outbox_id", job.OutboxID),
				slog.Int("user_id", u.ID),
				slog.String("error", err.Error()))
			continue
		}
		if err != nil {
			// Temporary failure: mengembalikan error agar Asynq me-retry task ini.
			// Pada retry selanjutnya, mahasiswa yang claim-nya sudah commit tidak akan dikirimi email lagi.
			return err
		}

		retry, _ := asynq.GetRetryCount(ctx)
		w.logger.Info("notification_processed",
			slog.Int64("outbox_id", job.OutboxID),
			slog.Int("user_id", u.ID),
			slog.String("event", job.EventType),
			slog.Int("retry", retry),
			slog.Bool("already_sent", alreadySent))
	}

	return nil
}
