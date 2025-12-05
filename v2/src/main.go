package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var startTime time.Time

func main() {
	// Marca início do sistema
	startTime = time.Now()

	// Parse flags
	configFile := flag.String("config", "config.yaml", "Arquivo de configuração")
	flag.Parse()

	// Banner
	log.Println("========================================")
	log.Println("  Edge Video V2 - Simple & Reliable")
	log.Println("========================================")

	// Carrega configuração
	config, err := LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("ERRO ao carregar config: %v", err)
	}

	log.Printf("Configuração carregada: %d câmeras, %d FPS, Quality %d",
		len(config.Cameras), config.FPS, config.Quality)

	// Cria e inicia câmeras usando FFmpeg stream
	// CRÍTICO: Cada câmera tem seu PRÓPRIO publisher para evitar race conditions!
	cameras := make([]*CameraStream, 0, len(config.Cameras))
	publishers := make([]*Publisher, 0, len(config.Cameras))

	for _, camCfg := range config.Cameras {
		// Usa exchange e routing_key dedicados da câmera
		// Se não especificados, usa os globais como fallback
		exchange := camCfg.Exchange
		if exchange == "" {
			exchange = config.AMQP.Exchange
		}

		routingKey := camCfg.RoutingKey
		if routingKey == "" {
			routingKey = config.AMQP.RoutingKeyPrefix + camCfg.ID
		}

		// Cria publisher DEDICADO para esta câmera com exchange e routing_key únicos
		publisher, err := NewPublisher(
			config.AMQP.URL,
			exchange,
			routingKey,                 // Passa routing_key COMPLETA ao invés de prefixo
			config.AMQP.PrefetchCount, // QoS: prefetch_count configurável via YAML
		)
		if err != nil {
			log.Fatalf("ERRO ao criar publisher para %s: %v", camCfg.ID, err)
		}
		defer publisher.Close()
		publishers = append(publishers, publisher)

		cam := NewCameraStream(
			camCfg.ID,
			camCfg.URL,
			config.FPS,
			config.Quality,
			publisher,
			config.CircuitBreaker, // Passa config do circuit breaker
		)

		cam.Start()
		cameras = append(cameras, cam)

		log.Printf("[%s] Câmera iniciada | Exchange: %s | RoutingKey: %s", camCfg.ID, exchange, routingKey)
	}

	// Monitor de estatísticas (usa primeiro publisher para contagem geral)
	go statsMonitor(cameras, publishers[0])

	// Inicia profiling monitor
	InitSystemStats() // Inicializa tracking de CPU/RAM
	StartProfileMonitor()

	// Inicializa Memory Controller (se habilitado)
	var memController *MemoryController
	if config.MemoryController.Enabled {
		log.Printf("Memory Controller HABILITADO (max: %d MB)", config.MemoryController.MaxMemoryMB)
		memController = NewMemoryController(config.MemoryController)

		// Registra callback para tracking
		memController.RegisterCallback(MemoryWarning, func(stats MemoryStats) {
			log.Printf("⚠️  Memory WARNING: %.1f%% (%d MB / %d MB)",
				stats.UsagePercent, stats.AllocMB, config.MemoryController.MaxMemoryMB)
		})
		memController.RegisterCallback(MemoryCritical, func(stats MemoryStats) {
			log.Printf("🔴 Memory CRITICAL: %.1f%% (%d MB / %d MB)",
				stats.UsagePercent, stats.AllocMB, config.MemoryController.MaxMemoryMB)
		})
		memController.RegisterCallback(MemoryEmergency, func(stats MemoryStats) {
			log.Printf("💀 Memory EMERGENCY: %.1f%% (%d MB / %d MB)",
				stats.UsagePercent, stats.AllocMB, config.MemoryController.MaxMemoryMB)
		})

		memController.Start()
		defer memController.Stop()

		// Goroutine para atualizar stats de memory controller no profiling
		go func() {
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()

			for range ticker.C {
				stats := memController.GetStats()
				TrackMemoryController(stats.Level, stats.NumGC)
			}
		}()
	} else {
		log.Println("Memory Controller DESABILITADO (pode ser habilitado no config.yaml)")
	}

	// Aguarda sinal de término
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	log.Println("\n✓ Sistema iniciado com sucesso!")
	log.Println("✓ Capturando frames... (Ctrl+C para parar)")

	<-sigChan

	// Shutdown graceful
	log.Println("\n\n🛑 Recebido sinal de término, parando...")

	for _, cam := range cameras {
		cam.Stop()
	}

	time.Sleep(500 * time.Millisecond)

	// RELATÓRIO FINAL DE ESTATÍSTICAS
	printFinalReport(cameras, publishers[0], config.FPS)

	// RELATÓRIO DE PROFILING
	PrintProfileReport()

	log.Println("✓ Sistema encerrado com sucesso")
}

// statsMonitor exibe estatísticas periodicamente
func statsMonitor(cameras []*CameraStream, publisher *Publisher) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		sep := "============================================================"

		log.Println("\n" + sep)
		log.Println("ESTATÍSTICAS")
		log.Println(sep)

		// Stats do publisher
		pubCount, pubErrors := publisher.Stats()
		errorRate := 0.0
		if pubCount > 0 {
			errorRate = float64(pubErrors) / float64(pubCount) * 100
		}

		connStatus := "✓ CONECTADO"
		if !publisher.IsConnected() {
			connStatus = "⚠ DESCONECTADO"
		}

		log.Printf("Publisher: %s - %d publicados, %d erros (%.2f%%)",
			connStatus, pubCount, pubErrors, errorRate)

		// Stats das câmeras
		openCircuits := uint32(0)
		for _, cam := range cameras {
			count, lastFrame := cam.Stats()
			age := time.Since(lastFrame)

			status := "OK"
			if age > 5*time.Second {
				status = "WARN"
			}

			// Circuit breaker status
			cbStats := cam.GetCircuitBreakerStats()
			if cbStats.Enabled && cbStats.State == StateOpen {
				status = "CB_OPEN"
				openCircuits++
			}

			cbInfo := ""
			if cbStats.Enabled {
				cbInfo = fmt.Sprintf(" | CB: %s", cbStats.State)
			}

			log.Printf("[%s] %s - Frames: %d, Último: %v atrás%s",
				cam.ID, status, count, age.Round(time.Second), cbInfo)
		}

		// Atualiza métrica de circuit breakers
		TrackCircuitBreaker(openCircuits)

		log.Println(sep + "\n")
	}
}

// printFinalReport exibe relatório completo ao encerrar
func printFinalReport(cameras []*CameraStream, publisher *Publisher, targetFPS int) {
	uptime := time.Since(startTime)

	sep := "================================================================"
	log.Println("\n" + sep)
	log.Println("                    RELATÓRIO FINAL")
	log.Println(sep)

	// Uptime
	log.Printf("⏱  Uptime Total: %v", uptime.Round(time.Second))
	log.Println("")

	// Stats do Publisher
	pubCount, pubErrors := publisher.Stats()
	errorRate := 0.0
	if pubCount > 0 {
		errorRate = float64(pubErrors) / float64(pubCount) * 100
	}

	log.Println("📤 PUBLISHER (RabbitMQ)")
	log.Printf("   Total Publicado:  %d frames", pubCount)
	log.Printf("   Erros:            %d (%.2f%%)", pubErrors, errorRate)

	// Publisher Confirms
	acks, nacks := publisher.ConfirmStats()
	totalConfirms := acks + nacks
	if totalConfirms > 0 {
		confirmRate := float64(acks) / float64(totalConfirms) * 100
		log.Printf("   Confirms (ACK):   %d (%.2f%%)", acks, confirmRate)
		log.Printf("   Rejeições (NACK): %d (%.2f%%)", nacks, 100-confirmRate)

		// Análise de perda
		if totalConfirms < pubCount {
			pending := pubCount - totalConfirms
			log.Printf("   ⏳ Pendentes:     %d frames", pending)
		}
		if nacks > 0 {
			log.Printf("   ⚠️  ALERTA: %d frames foram REJEITADOS pelo RabbitMQ!", nacks)
		}
		if acks == pubCount && nacks == 0 {
			log.Printf("   ✅ 100%% dos frames CONFIRMADOS pelo RabbitMQ!")
		}
	}

	// Throughput
	if uptime.Seconds() > 0 {
		fps := float64(pubCount) / uptime.Seconds()
		log.Printf("   Throughput:       %.2f frames/s", fps)
	}
	log.Println("")

	// Stats por câmera
	log.Println("📹 CÂMERAS")

	var totalFrames uint64
	var totalBytesEstimated uint64

	for _, cam := range cameras {
		frameCount, framesReceived, framesDropped, lastFrame, lastFrameReceived := cam.DetailedStats()
		totalFrames += frameCount

		// Estima tamanho médio (JPEG quality 5 ≈ 50KB)
		estimatedBytes := frameCount * 50000
		totalBytesEstimated += estimatedBytes

		// FPS médio de PUBLICAÇÃO
		avgFPSPublish := 0.0
		if uptime.Seconds() > 0 {
			avgFPSPublish = float64(frameCount) / uptime.Seconds()
		}

		// FPS médio da CÂMERA (recebido do FFmpeg)
		avgFPSCamera := 0.0
		if uptime.Seconds() > 0 {
			avgFPSCamera = float64(framesReceived) / uptime.Seconds()
		}

		// Efficiency (quanto do target FPS foi atingido)
		efficiency := 0.0
		if targetFPS > 0 {
			efficiency = (avgFPSPublish / float64(targetFPS)) * 100
		}

		// Tempo desde último frame
		lastFrameReceivedAge := time.Since(lastFrameReceived)
		status := "✓"
		if lastFrameReceivedAge > 5*time.Second {
			status = "⚠"
		}
		_ = lastFrame // Evita warning unused

		// Calcula % de frames descartados
		dropRate := 0.0
		if framesReceived > 0 {
			dropRate = (float64(framesDropped) / float64(framesReceived)) * 100
		}

		// Circuit Breaker stats
		cbStats := cam.GetCircuitBreakerStats()

		log.Printf("   %s [%s]", status, cam.ID)
		log.Printf("      Frames da Câmera:   %d (%.2f FPS real)", framesReceived, avgFPSCamera)
		log.Printf("      Frames Publicados:  %d (%.2f FPS)", frameCount, avgFPSPublish)
		log.Printf("      Frames Descartados: %d (%.1f%%)", framesDropped, dropRate)
		log.Printf("      FPS Target:         %d", targetFPS)
		log.Printf("      Eficiência:         %.1f%%", efficiency)
		log.Printf("      Volume Estimado:    %.2f MB", float64(estimatedBytes)/(1024*1024))
		log.Printf("      Último da Câmera:   %v atrás", lastFrameReceivedAge.Round(time.Second))

		// Circuit Breaker info
		if cbStats.Enabled {
			log.Printf("      Circuit Breaker:    %s | Calls: %d (✓%d ✗%d 🚫%d) | Changes: %d",
				cbStats.State, cbStats.TotalCalls, cbStats.TotalSuccesses,
				cbStats.TotalFailures, cbStats.TotalRejected, cbStats.StateChanges)
		} else {
			log.Printf("      Circuit Breaker:    DISABLED")
		}

		log.Println("")
	}

	// Totais gerais
	log.Println("📊 TOTAIS GERAIS")
	log.Printf("   Câmeras Ativas:        %d", len(cameras))
	log.Printf("   Total de Frames:       %d", totalFrames)
	log.Printf("   Volume Total Estimado: %.2f MB", float64(totalBytesEstimated)/(1024*1024))

	if uptime.Seconds() > 0 {
		totalFPS := float64(totalFrames) / uptime.Seconds()
		throughputMBps := (float64(totalBytesEstimated) / uptime.Seconds()) / (1024 * 1024)

		log.Printf("   FPS Total Sistema:     %.2f frames/s", totalFPS)
		log.Printf("   Throughput Total:      %.2f MB/s", throughputMBps)

		// Taxa de sucesso
		successRate := 100.0
		if pubCount > 0 {
			successRate = float64(pubCount-pubErrors) / float64(pubCount) * 100
		}
		log.Printf("   Taxa de Sucesso:       %.2f%%", successRate)
	}

	log.Println(sep)
	log.Println("")
}
