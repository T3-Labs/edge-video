package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
	"sync"
	"time"
)

// Camera representa uma câmera RTSP com captura controlada por ticker
type Camera struct {
	ID       string
	URL      string
	FPS      int
	Quality  int

	publisher *Publisher
	ctx       context.Context
	cancel    context.CancelFunc

	mu           sync.Mutex
	frameCount   uint64
	lastFrame    time.Time
	running      bool
}

// NewCamera cria uma nova instância de câmera
func NewCamera(id, url string, fps, quality int, publisher *Publisher) *Camera {
	ctx, cancel := context.WithCancel(context.Background())

	return &Camera{
		ID:        id,
		URL:       url,
		FPS:       fps,
		Quality:   quality,
		publisher: publisher,
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start inicia a captura de frames
func (c *Camera) Start() {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return
	}
	c.running = true
	c.mu.Unlock()

	go c.captureLoop()
}

// Stop para a captura
func (c *Camera) Stop() {
	c.cancel()
}

// captureLoop é o loop principal de captura
func (c *Camera) captureLoop() {
	log.Printf("[%s] Iniciando captura - FPS: %d, Quality: %d", c.ID, c.FPS, c.Quality)

	interval := time.Second / time.Duration(c.FPS)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			log.Printf("[%s] Captura parada", c.ID)
			return

		case <-ticker.C:
			c.captureAndPublish()
		}
	}
}

// captureAndPublish captura um frame e publica
func (c *Camera) captureAndPublish() {
	start := time.Now()

	// Captura frame único com FFmpeg
	frame, err := c.captureFrame()
	if err != nil {
		log.Printf("[%s] ERRO ao capturar frame: %v", c.ID, err)
		return
	}

	captureDuration := time.Since(start)

	// Incrementa contador
	c.mu.Lock()
	c.frameCount++
	frameNum := c.frameCount
	c.lastFrame = start
	c.mu.Unlock()

	// Publica no RabbitMQ
	err = c.publisher.Publish(c.ID, frame, start)
	if err != nil {
		log.Printf("[%s] ERRO ao publicar frame #%d: %v", c.ID, frameNum, err)
		return
	}

	publishDuration := time.Since(start)

	// Log a cada 30 frames
	if frameNum%30 == 0 {
		log.Printf("[%s] Frame #%d - Captura: %v, Total: %v, Tamanho: %d bytes",
			c.ID, frameNum, captureDuration, publishDuration, len(frame))
	}
}

// captureFrame captura um único frame usando FFmpeg
func (c *Camera) captureFrame() ([]byte, error) {
	// Timeout de 5 segundos para a captura (primeira pode demorar mais)
	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
	defer cancel()

	// Comando FFmpeg otimizado para captura rápida de frame único
	cmd := exec.CommandContext(ctx,
		"ffmpeg",
		"-loglevel", "error",     // Apenas erros
		"-rtsp_transport", "tcp",
		"-i", c.URL,
		"-frames:v", "1",         // Apenas 1 frame
		"-vcodec", "mjpeg",
		"-q:v", fmt.Sprintf("%d", c.Quality),
		"-f", "image2pipe",
		"-",
	)

	// Captura stdout e stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Executa
	err := cmd.Run()
	if err != nil {
		errMsg := stderr.String()
		if errMsg != "" {
			return nil, fmt.Errorf("ffmpeg: %s", errMsg)
		}
		return nil, fmt.Errorf("ffmpeg falhou: %w", err)
	}

	// Valida se tem dados
	frameData := stdout.Bytes()
	if len(frameData) == 0 {
		return nil, fmt.Errorf("frame vazio")
	}

	// Valida se é JPEG válido (começa com FF D8)
	if len(frameData) < 2 || frameData[0] != 0xFF || frameData[1] != 0xD8 {
		return nil, fmt.Errorf("frame não é JPEG válido")
	}

	return frameData, nil
}

// Stats retorna estatísticas da câmera
func (c *Camera) Stats() (uint64, time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.frameCount, c.lastFrame
}
