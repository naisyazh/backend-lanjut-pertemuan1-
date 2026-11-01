package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"

	"latihan-fiber/app/model"
	"latihan-fiber/config"
	"latihan-fiber/database"
)

func main() {
	outboxID := flag.Int64("outbox", 0, "ID baris outbox yang akan di-enqueue ulang")
	copies := flag.Int("copies", 1, "Jumlah salinan task yang dikirim ke Redis")
	flag.Parse()

	if *outboxID <= 0 {
		log.Fatal("argumen -outbox wajib diisi dengan ID valid (> 0)")
	}

	config.LoadEnv()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx)
	if err != nil {
		log.Fatalf("gagal koneksi ke database: %v", err)
	}
	defer pool.Close()

	var eventType string
	var payload []byte
	err = pool.QueryRow(ctx, "SELECT event_type, payload FROM outbox WHERE id = $1", *outboxID).Scan(&eventType, &payload)
	if err != nil {
		log.Fatalf("gagal mengambil outbox ID %d: %v", *outboxID, err)
	}

	job := model.NotificationJob{
		OutboxID:  *outboxID,
		EventType: eventType,
		Payload:   payload,
	}
	jobBytes, err := json.Marshal(job)
	if err != nil {
		log.Fatalf("gagal marshal job: %v", err)
	}

	client := asynq.NewClient(config.RedisOpt())
	defer client.Close()

	for i := 1; i <= *copies; i++ {
		task := asynq.NewTask(
			model.TaskSendNotification,
			jobBytes,
			asynq.Queue(model.QueueNotifications),
			asynq.MaxRetry(5),
			asynq.Timeout(30*time.Second),
			asynq.Retention(24*time.Hour),
		)
		info, err := client.EnqueueContext(ctx, task)
		if err != nil {
			log.Fatalf("gagal enqueue salinan ke-%d: %v", i, err)
		}
		fmt.Printf("Salinan %d/%d terkirim ke Redis (TaskID: %s)\n", i, *copies, info.ID)
	}

	fmt.Printf("Sukses: %d salinan task untuk outbox ID %d berhasil dimasukkan ke Redis queue.\n", *copies, *outboxID)
}
