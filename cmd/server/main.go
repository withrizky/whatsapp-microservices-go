package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"whatsapp_microservices/internal/model"
	"whatsapp_microservices/internal/waha"
	"whatsapp_microservices/internal/worker"
)

func main() {
	godotenv.Load()

	// 1. Setup Komponen
	wahaClient := waha.NewClient(os.Getenv("WAHA_URL"), os.Getenv("WAHA_API_KEY"))

	// Buat Dispatcher (50 Worker, Kapasitas Antrean 10.000)
	// Tanpa RabbitMQ, antrean disimpan di RAM (Buffered Channel)
	dispatcher := worker.NewDispatcher(50, 10000, wahaClient, os.Getenv("WAHA_SESSIONS"))

	// Jalankan Worker di background
	dispatcher.Run()

	// 2. Setup Gin Server
	r := gin.New()
	r.Use(gin.Recovery())

	r.POST("/send", func(c *gin.Context) {
		var req model.WaPayload
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Payload invalid"})
			return
		}

		// Non-blocking send ke channel
		// Jika antrean penuh (10.000), ini akan memblokir sebentar atau bisa kita handle errornya
		select {
		case dispatcher.JobQueue <- req:
			c.JSON(202, gin.H{"status": "queued", "message": "Pesan masuk antrean"})
		default:
			// Opsional: Handle jika antrean penuh banget
			c.JSON(503, gin.H{"error": "Server sibuk, antrean penuh"})
		}
	})

	// 3. Konfigurasi HTTP Server dengan Graceful Shutdown
	srv := &http.Server{
		Addr:    ":" + os.Getenv("PORT"),
		Handler: r,
	}

	// Jalankan server di goroutine terpisah
	go func() {
		log.Printf("Server berjalan di port %s", os.Getenv("PORT"))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen: %s\n", err)
		}
	}()

	// 4. Graceful Shutdown Logic
	// Menunggu sinyal kill (Ctrl+C atau docker stop)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Timeout 5 detik untuk server HTTP mati
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	// Stop workers (selesaikan job yang tersisa di "tangan" worker)
	log.Println("Menunggu worker menyelesaikan tugas...")
	dispatcher.Stop()

	log.Println("Server exiting")
}
