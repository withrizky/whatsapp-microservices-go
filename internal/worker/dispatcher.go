package worker

import (
	"log"
	"strings"
	"sync"
	"sync/atomic"

	"whatsapp_microservices/internal/model"
	"whatsapp_microservices/internal/waha"
)

// Dispatcher mengatur antrean dan pekerja
type Dispatcher struct {
	JobQueue       chan model.WaPayload // Channel sebagai antrean (pengganti RabbitMQ)
	MaxWorkers     int
	WahaClient     *waha.Client
	Sessions       []string
	sessionCounter uint32
	wg             sync.WaitGroup // Untuk menunggu worker selesai saat shutdown
}

func NewDispatcher(maxWorkers int, bufferSize int, wahaClient *waha.Client, sessionStr string) *Dispatcher {
	return &Dispatcher{
		JobQueue:   make(chan model.WaPayload, bufferSize), // Buffered channel
		MaxWorkers: maxWorkers,
		WahaClient: wahaClient,
		Sessions:   strings.Split(sessionStr, ","),
	}
}

// Start menjalankan para pekerja
func (d *Dispatcher) Run() {
	for i := 0; i < d.MaxWorkers; i++ {
		d.wg.Add(1) // Tambah counter waitgroup
		go func(workerID int) {
			defer d.wg.Done() // Kurangi counter saat worker mati

			log.Printf("Worker #%d siap menerima tugas", workerID)

			// Worker akan terus loop mengambil data dari channel
			for job := range d.JobQueue {
				d.processJob(workerID, job)
			}

			log.Printf("Worker #%d berhenti.", workerID)
		}(i)
	}
}

// Stop menutup antrean dan menunggu semua pekerja selesai (Graceful Shutdown)
func (d *Dispatcher) Stop() {
	close(d.JobQueue) // Tutup pintu antrean
	d.wg.Wait()       // Tunggu semua pekerja menyelesaikan tugas terakhir mereka
}

func (d *Dispatcher) processJob(id int, job model.WaPayload) {
	// Round Robin Session Rotation
	sess := d.Sessions[0]
	if len(d.Sessions) > 1 {
		idx := atomic.AddUint32(&d.sessionCounter, 1) % uint32(len(d.Sessions))
		sess = d.Sessions[idx]
	}

	err := d.WahaClient.SendText(job, sess)
	if err != nil {
		log.Printf("[Worker-%d] ❌ Gagal ke %s: %v", id, job.To, err)
		// Opsional: Implementasi Retry Logic bisa ditaruh di sini
	} else {
		log.Printf("[Worker-%d] ✅ Terkirim ke %s [%s]", id, job.To, sess)
	}
}
