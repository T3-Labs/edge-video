package main

import (
	"log"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

// ProfileStats rastreia estatísticas de performance
type ProfileStats struct {
	// FFmpeg
	ffmpegReadTime   atomic.Int64 // nanoseconds
	ffmpegReadCount  atomic.Uint64

	// Frame processing
	frameDecodeTime  atomic.Int64
	frameDecodeCount atomic.Uint64

	// Publishing
	publishTime      atomic.Int64
	publishCount     atomic.Uint64

	// Memory
	lastGCPause      atomic.Uint64 // microseconds
	gcCount          atomic.Uint32

	// Goroutines
	goroutineCount   atomic.Int32

	// Circuit Breaker
	circuitBreakerOpen atomic.Uint32

	// Sistema (CPU e RAM)
	cpuPercent    atomic.Uint64 // Multiplicado por 100 (ex: 45.67% = 4567)
	ramUsedMB     atomic.Uint64
	ramTotalMB    atomic.Uint64
	ramPercentage atomic.Uint64 // Multiplicado por 100

	// Memory Controller
	memoryControllerLevel atomic.Int32 // 0=Normal, 1=Warning, 2=Critical, 3=Emergency
	memoryControllerGCCount atomic.Uint32

	// Publisher Confirms
	publishConfirmsAck  atomic.Uint64 // Total de ACKs recebidos
	publishConfirmsNack atomic.Uint64 // Total de NACKs recebidos
}

var globalProfile ProfileStats
var currentProcess *process.Process

// TrackFFmpegRead rastreia tempo de leitura do FFmpeg
func TrackFFmpegRead(duration time.Duration) {
	globalProfile.ffmpegReadTime.Add(int64(duration))
	globalProfile.ffmpegReadCount.Add(1)
}

// TrackFrameDecode rastreia tempo de decode
func TrackFrameDecode(duration time.Duration) {
	globalProfile.frameDecodeTime.Add(int64(duration))
	globalProfile.frameDecodeCount.Add(1)
}

// TrackPublish rastreia tempo de publicação
func TrackPublish(duration time.Duration) {
	globalProfile.publishTime.Add(int64(duration))
	globalProfile.publishCount.Add(1)
}

// TrackCircuitBreaker rastreia estado do circuit breaker
func TrackCircuitBreaker(openCount uint32) {
	globalProfile.circuitBreakerOpen.Store(openCount)
}

// TrackMemoryController rastreia estado do memory controller
func TrackMemoryController(level MemoryLevel, gcCount uint32) {
	globalProfile.memoryControllerLevel.Store(int32(level))
	globalProfile.memoryControllerGCCount.Store(gcCount)
}

// TrackPublishConfirm rastreia confirmações do RabbitMQ (ACK/NACK)
func TrackPublishConfirm(ack bool) {
	if ack {
		globalProfile.publishConfirmsAck.Add(1)
	} else {
		globalProfile.publishConfirmsNack.Add(1)
	}
}

// UpdateMemoryStats atualiza stats de memória
func UpdateMemoryStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	globalProfile.lastGCPause.Store(m.PauseNs[(m.NumGC+255)%256] / 1000) // Convert to microseconds
	globalProfile.gcCount.Store(m.NumGC)
	globalProfile.goroutineCount.Store(int32(runtime.NumGoroutine()))
}

// UpdateSystemStats atualiza stats de CPU e RAM do sistema
func UpdateSystemStats() {
	// CPU do processo
	if currentProcess != nil {
		cpuPct, err := currentProcess.CPUPercent()
		if err == nil {
			// Multiplica por 100 para armazenar como integer (45.67% = 4567)
			globalProfile.cpuPercent.Store(uint64(cpuPct * 100))
		}

		// RAM do processo
		memInfo, err := currentProcess.MemoryInfo()
		if err == nil {
			ramMB := memInfo.RSS / 1024 / 1024
			globalProfile.ramUsedMB.Store(ramMB)
		}
	}

	// RAM total do sistema
	vmem, err := mem.VirtualMemory()
	if err == nil {
		totalMB := vmem.Total / 1024 / 1024
		globalProfile.ramTotalMB.Store(totalMB)
		globalProfile.ramPercentage.Store(uint64(vmem.UsedPercent * 100))
	}
}

// InitSystemStats inicializa tracking do processo
func InitSystemStats() {
	var err error
	pid := int32(os.Getpid())
	currentProcess, err = process.NewProcess(pid)
	if err != nil {
		log.Printf("⚠ Não foi possível inicializar stats de sistema: %v", err)
	}
}

// PrintProfileReport imprime relatório de profiling
func PrintProfileReport() {
	sep := "================================================================"
	log.Println("\n" + sep)
	log.Println("                  PERFORMANCE PROFILE")
	log.Println(sep)

	// FFmpeg stats
	ffmpegReads := globalProfile.ffmpegReadCount.Load()
	if ffmpegReads > 0 {
		avgFFmpeg := time.Duration(globalProfile.ffmpegReadTime.Load() / int64(ffmpegReads))
		log.Printf("🎥 FFmpeg Read:")
		log.Printf("   Avg Time:  %v", avgFFmpeg)
		log.Printf("   Count:     %d", ffmpegReads)
	}

	// Decode stats
	decodes := globalProfile.frameDecodeCount.Load()
	if decodes > 0 {
		avgDecode := time.Duration(globalProfile.frameDecodeTime.Load() / int64(decodes))
		log.Printf("🔧 Frame Decode:")
		log.Printf("   Avg Time:  %v", avgDecode)
		log.Printf("   Count:     %d", decodes)
	}

	// Publish stats
	publishes := globalProfile.publishCount.Load()
	if publishes > 0 {
		avgPublish := time.Duration(globalProfile.publishTime.Load() / int64(publishes))
		log.Printf("📤 Publishing:")
		log.Printf("   Avg Time:  %v", avgPublish)
		log.Printf("   Count:     %d", publishes)

		// Publisher Confirms stats
		acks := globalProfile.publishConfirmsAck.Load()
		nacks := globalProfile.publishConfirmsNack.Load()
		total := acks + nacks

		if total > 0 {
			ackRate := float64(acks) / float64(total) * 100
			log.Printf("   Confirms:  %d ACKs, %d NACKs (%.1f%% sucesso)", acks, nacks, ackRate)

			// Detecta problemas
			if nacks > 0 {
				log.Printf("   ⚠️  %d frames REJEITADOS pelo RabbitMQ!", nacks)
			}
			if total < publishes {
				pending := publishes - total
				log.Printf("   ⏳  %d confirms pendentes", pending)
			}
		}

		if avgPublish > 50*time.Millisecond {
			log.Printf("   ⚠️  GARGALO: Latência de %v é alta!", avgPublish)
		}
	}

	// Memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	log.Printf("💾 Memory (Go Runtime):")
	log.Printf("   Alloc:     %.2f MB", float64(m.Alloc)/(1024*1024))
	log.Printf("   Sys:       %.2f MB", float64(m.Sys)/(1024*1024))
	log.Printf("   GC Count:  %d", m.NumGC)
	log.Printf("   Last GC:   %v µs", globalProfile.lastGCPause.Load())

	// Sistema
	cpuPct := float64(globalProfile.cpuPercent.Load()) / 100.0
	ramUsedMB := globalProfile.ramUsedMB.Load()
	ramTotalMB := globalProfile.ramTotalMB.Load()
	ramPct := float64(globalProfile.ramPercentage.Load()) / 100.0

	log.Printf("🖥️  Sistema (Processo):")
	log.Printf("   CPU Usage: %.2f%%", cpuPct)
	log.Printf("   RAM Usage: %d MB", ramUsedMB)

	log.Printf("🌐 Sistema (Total):")
	log.Printf("   RAM Total: %d MB", ramTotalMB)
	log.Printf("   RAM Used:  %.2f%%", ramPct)

	log.Printf("🔀 Goroutines: %d", runtime.NumGoroutine())

	// Circuit Breaker stats
	cbOpen := globalProfile.circuitBreakerOpen.Load()
	if cbOpen > 0 {
		log.Printf("🔴 Circuit Breakers OPEN: %d", cbOpen)
	}

	// Memory Controller stats
	memLevel := MemoryLevel(globalProfile.memoryControllerLevel.Load())
	memGCCount := globalProfile.memoryControllerGCCount.Load()
	if memGCCount > 0 {
		log.Printf("🧠 Memory Controller:")
		log.Printf("   Level:     %s", memLevel)
		log.Printf("   Manual GC: %d", memGCCount)
	}

	log.Println(sep + "\n")
}

// StartProfileMonitor inicia monitor de profiling
func StartProfileMonitor() {
	go func() {
		ticker := time.NewTicker(5 * time.Second) // A cada 5s atualiza stats
		defer ticker.Stop()

		for range ticker.C {
			UpdateMemoryStats()
			UpdateSystemStats()
		}
	}()
}
