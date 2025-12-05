package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// CameraStream usa FFmpeg em modo stream contínuo
// VERSÃO CORRIGIDA: Cada câmera tem seus PRÓPRIOS buffers (sem sync.Pool compartilhado)
type CameraStream struct {
	ID       string
	URL      string
	FPS      int
	Quality  int

	publisher      *Publisher
	circuitBreaker *CircuitBreaker // Circuit breaker para proteção contra falhas
	ctx            context.Context
	cancel         context.CancelFunc

	mu                sync.Mutex
	frameCount        uint64    // Frames publicados
	framesReceived    uint64    // Frames recebidos do FFmpeg
	framesDropped     uint64    // Frames descartados (canal cheio)
	lastFrame         time.Time
	lastFrameReceived time.Time // Último frame do FFmpeg
	running           bool
	retrying          bool      // Flag para evitar múltiplas goroutines de retry

	// CORREÇÃO CRÍTICA: Buffers privados da câmera (não compartilhados!)
	bufferPool chan []byte // Pool LOCAL de buffers (não global!)

	frameChan chan []byte
	cmd       *exec.Cmd
}

// NewCameraStream cria câmera com buffers PRIVADOS e circuit breaker
func NewCameraStream(id, url string, fps, quality int, publisher *Publisher, cbConfig CircuitBreakerConfig) *CameraStream {
	ctx, cancel := context.WithCancel(context.Background())

	c := &CameraStream{
		ID:             id,
		URL:            url,
		FPS:            fps,
		Quality:        quality,
		publisher:      publisher,
		circuitBreaker: NewCircuitBreaker(id, cbConfig),
		ctx:            ctx,
		cancel:         cancel,
		frameChan:      make(chan []byte, 5), // Buffer de 5 frames

		// CRÍTICO: Pool LOCAL de buffers (10 buffers dedicados para ESTA câmera)
		bufferPool: make(chan []byte, 10),
	}

	// Pre-aloca 10 buffers DEDICADOS para esta câmera
	for i := 0; i < 10; i++ {
		buf := make([]byte, 2*1024*1024) // 2MB cada
		c.bufferPool <- buf
	}

	return c
}

// getBuffer pega buffer do pool LOCAL da câmera
func (c *CameraStream) getBuffer() []byte {
	select {
	case buf := <-c.bufferPool:
		return buf
	default:
		// Pool vazio, aloca novo (só acontece em caso extremo)
		log.Printf("[%s] AVISO: Pool local vazio, alocando novo buffer", c.ID)
		return make([]byte, 2*1024*1024)
	}
}

// putBuffer devolve buffer ao pool LOCAL da câmera
func (c *CameraStream) putBuffer(buf []byte) {
	select {
	case c.bufferPool <- buf:
		// Buffer devolvido com sucesso
	default:
		// Pool cheio, descarta buffer (GC vai liberar)
		// Isso é normal se alocamos buffers extras
	}
}

// Start inicia
func (c *CameraStream) Start() {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return
	}
	c.running = true
	c.mu.Unlock()

	go c.startFFmpeg()
	go c.publishLoop()
}

// Stop para
func (c *CameraStream) Stop() {
	c.cancel()
	if c.cmd != nil && c.cmd.Process != nil {
		c.cmd.Process.Kill()
	}
}

// startFFmpeg inicia FFmpeg e lê frames (IDÊNTICO ao original)
func (c *CameraStream) startFFmpeg() {
	log.Printf("[%s] Iniciando stream FFmpeg - FPS: %d", c.ID, c.FPS)

	// Detecta protocolo
	isRTMP := strings.HasPrefix(strings.ToLower(c.URL), "rtmp://") ||
	          strings.HasPrefix(strings.ToLower(c.URL), "rtmps://")
	isRTSP := strings.HasPrefix(strings.ToLower(c.URL), "rtsp://") ||
	          strings.HasPrefix(strings.ToLower(c.URL), "rtsps://")

	args := []string{"ffmpeg"}

	if isRTSP {
		log.Printf("[%s] Protocolo: RTSP", c.ID)
		args = append(args,
			"-rtsp_transport", "tcp",
			"-timeout", "5000000",
		)
	} else if isRTMP {
		log.Printf("[%s] Protocolo: RTMP", c.ID)
		args = append(args,
			"-rw_timeout", "5000000",
			"-listen", "0",
		)
	}

	args = append(args,
		"-fflags", "nobuffer+fastseek+flush_packets+discardcorrupt",
		"-flags", "low_delay",
		"-max_delay", "0",
		"-probesize", "32",
		"-analyzeduration", "0",
		"-err_detect", "ignore_err",
		"-i", c.URL,
		"-vf", fmt.Sprintf("fps=%d", c.FPS),
		"-f", "image2pipe",
		"-vcodec", "mjpeg",
		"-q:v", fmt.Sprintf("%d", c.Quality),
		"-pkt_size", "2097152",
		"-max_muxing_queue_size", "1024",
		"-threads", "1",
		"-",
	)

	c.cmd = exec.CommandContext(c.ctx, args[0], args[1:]...)

	stdout, err := c.cmd.StdoutPipe()
	if err != nil {
		log.Printf("[%s] ERRO ao criar pipe: %v", c.ID, err)
		return
	}

	stderr, err := c.cmd.StderrPipe()
	if err != nil {
		log.Printf("[%s] ERRO ao criar stderr pipe: %v", c.ID, err)
		return
	}

	// Inicia FFmpeg (SEM circuit breaker aqui - só monitora stream reads)
	err = c.cmd.Start()
	if err != nil {
		log.Printf("[%s] ERRO ao iniciar FFmpeg: %v", c.ID, err)

		// Registra falha no circuit breaker
		c.circuitBreaker.Execute(func() error {
			return err
		})

		// Agenda retry (circuit breaker controla backoff)
		// Mas apenas se já não há retry em andamento
		c.mu.Lock()
		if !c.retrying {
			c.retrying = true
			c.mu.Unlock()
			go c.retryFFmpegWithBackoff()
		} else {
			c.mu.Unlock()
		}
		return
	}

	// Monitora stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "error") || strings.Contains(line, "Error") ||
			   strings.Contains(line, "fatal") || strings.Contains(line, "Fatal") {
				log.Printf("[%s] FFmpeg ERRO: %s", c.ID, line)
			}
		}
	}()

	log.Printf("[%s] FFmpeg iniciado!", c.ID)

	reader := bufio.NewReaderSize(stdout, 1024*1024)
	c.readFrames(reader)
}

// readFrames lê frames do FFmpeg - VERSÃO CORRIGIDA
func (c *CameraStream) readFrames(reader *bufio.Reader) {
	frameBuffer := bytes.NewBuffer(make([]byte, 0, 512*1024))

	jpegSOI := []byte{0xFF, 0xD8}
	jpegEOI := []byte{0xFF, 0xD9}

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		b, err := reader.ReadByte()
		if err != nil {
			if c.ctx.Err() == nil {
				log.Printf("[%s] ERRO ao ler: %v", c.ID, err)

				// CRÍTICO: Registra falha no circuit breaker
				c.circuitBreaker.Execute(func() error {
					return err
				})

				// SEMPRE tenta reconectar (circuit breaker controla o backoff)
				// Mas apenas se já não há retry em andamento
				c.mu.Lock()
				if !c.retrying {
					c.retrying = true
					c.mu.Unlock()
					go c.retryFFmpegWithBackoff()
				} else {
					c.mu.Unlock()
				}
			}
			return
		}

		frameBuffer.WriteByte(b)

		// Detecta fim do JPEG
		if frameBuffer.Len() >= 2 {
			tail := frameBuffer.Bytes()[frameBuffer.Len()-2:]
			if bytes.Equal(tail, jpegEOI) {
				// CORREÇÃO CRÍTICA: Usa buffer do pool LOCAL (não global!)
				buf := c.getBuffer()
				frameSize := frameBuffer.Len()

				if frameSize > len(buf) {
					log.Printf("[%s] ERRO: Frame %d bytes > buffer %d bytes", c.ID, frameSize, len(buf))
					c.putBuffer(buf)
					frameBuffer.Reset()
					continue
				}

				// CORREÇÃO CRÍTICA: FAZ CÓPIA IMEDIATA para um novo slice
				// NÃO envia o buffer do pool para o channel!
				frameCopy := make([]byte, frameSize)
				copy(frameCopy, frameBuffer.Bytes())

				// DEVOLVE buffer IMEDIATAMENTE ao pool local
				c.putBuffer(buf)

				// Valida JPEG
				if bytes.HasPrefix(frameCopy, jpegSOI) && bytes.HasSuffix(frameCopy, jpegEOI) {
					c.mu.Lock()
					c.framesReceived++
					c.lastFrameReceived = time.Now()
					c.mu.Unlock()

					// Envia CÓPIA para o channel (não o buffer do pool!)
					select {
					case c.frameChan <- frameCopy:
						// Frame enviado
					default:
						// Canal cheio, descarta (GC vai liberar frameCopy)
						c.mu.Lock()
						c.framesDropped++
						c.mu.Unlock()
					}
				}

				frameBuffer.Reset()
			}
		}
	}
}

// publishLoop - VERSÃO CORRIGIDA (muito mais simples!)
func (c *CameraStream) publishLoop() {
	log.Printf("[%s] Iniciando loop de publicação", c.ID)

	interval := time.Second / time.Duration(c.FPS)
	lastPublish := time.Now()

	for {
		select {
		case <-c.ctx.Done():
			log.Printf("[%s] Publicação parada", c.ID)
			return
		default:
		}

		elapsed := time.Since(lastPublish)
		if elapsed < interval {
			time.Sleep(interval - elapsed)
		}

		// Pega frame mais recente
		var frame []byte
		select {
		case frame = <-c.frameChan:
			// Descarta frames antigos
			flushed := 0
			for len(c.frameChan) > 0 {
				frame = <-c.frameChan // Sobrescreve com mais recente
				flushed++
			}
			if flushed > 10 {
				log.Printf("[%s] Latest Frame Policy: descartou %d frames antigos", c.ID, flushed)
			}
		default:
			continue
		}

		lastPublish = time.Now()
		start := time.Now()

		c.mu.Lock()
		c.frameCount++
		frameNum := c.frameCount
		c.lastFrame = start
		c.mu.Unlock()

		// CORREÇÃO CRÍTICA: NÃO precisa mais de frameCopy aqui!
		// O frame JÁ É UMA CÓPIA INDEPENDENTE feita na linha 245

		// Publica ASSÍNCRONA
		go func(cameraID string, frameData []byte, frameNum uint64, start time.Time) {
			err := c.publisher.Publish(cameraID, frameData, start)
			publishDuration := time.Since(start)
			TrackPublish(publishDuration)

			if frameNum%30 == 0 {
				log.Printf("[%s] Frame #%d - Publicação: %v, Tamanho: %d bytes",
					cameraID, frameNum, publishDuration, len(frameData))
			}

			if err != nil {
				log.Printf("[%s] ERRO ao publicar frame #%d: %v", cameraID, frameNum, err)
			}
		}(c.ID, frame, frameNum, start)
	}
}

// Stats retorna estatísticas
func (c *CameraStream) Stats() (uint64, time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.frameCount, c.lastFrame
}

// DetailedStats retorna estatísticas detalhadas
func (c *CameraStream) DetailedStats() (uint64, uint64, uint64, time.Time, time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.frameCount, c.framesReceived, c.framesDropped, c.lastFrame, c.lastFrameReceived
}

// retryFFmpegWithBackoff tenta reconectar FFmpeg respeitando circuit breaker
// IMPORTANTE: Assume que c.retrying já foi setado para true ANTES de chamar esta função
func (c *CameraStream) retryFFmpegWithBackoff() {
	defer func() {
		c.mu.Lock()
		c.retrying = false
		c.mu.Unlock()
	}()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		// Verifica estado do circuit breaker
		stats := c.circuitBreaker.Stats()

		if stats.State == StateOpen {
			// Aguarda backoff
			if stats.TimeUntilRetry > 0 {
				log.Printf("[%s] Circuit breaker OPEN - aguardando %v antes de retry...",
					c.ID, stats.TimeUntilRetry)
				time.Sleep(stats.TimeUntilRetry)
			}
			continue
		}

		// Tenta reconectar
		log.Printf("[%s] Tentando reconectar FFmpeg (estado: %s)...", c.ID, stats.State)
		go c.startFFmpeg()
		return
	}
}

// GetCircuitBreakerStats retorna estatísticas do circuit breaker
func (c *CameraStream) GetCircuitBreakerStats() CircuitBreakerStats {
	return c.circuitBreaker.Stats()
}
