package camera

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"

	"github.com/T3-Labs/edge-video/pkg/logger"
)

type PersistentCapture struct {
	cameraID   string
	rtspURL    string
	quality    int
	fps        int
	bufferSize int

	mu         sync.RWMutex
	cmd        *exec.Cmd
	stdout     io.ReadCloser
	stderr     io.ReadCloser
	running    bool
	restarting bool // Flag para evitar restarts simultâneos

	ctx    context.Context
	cancel context.CancelFunc

	frameBuffer chan []byte
	errorCount  int64
	lastRestart time.Time
	lastFrameNS atomic.Int64

	readCtx    context.Context    // Context para a goroutine readFrames
	readCancel context.CancelFunc // Cancel para a goroutine readFrames
}

func NewPersistentCapture(ctx context.Context, cameraID, rtspURL string, quality int, fps int, bufferSize int) *PersistentCapture {
	ctx, cancel := context.WithCancel(ctx)

	if bufferSize <= 0 {
		bufferSize = 50
	}

	pc := &PersistentCapture{
		cameraID:    cameraID,
		rtspURL:     rtspURL,
		quality:     quality,
		fps:         fps,
		bufferSize:  bufferSize,
		ctx:         ctx,
		cancel:      cancel,
		frameBuffer: make(chan []byte, bufferSize),
		lastRestart: time.Now(),
	}
	pc.lastFrameNS.Store(time.Now().UnixNano())
	return pc
}

func (pc *PersistentCapture) Start() error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if pc.running {
		return fmt.Errorf("captura já está rodando")
	}

	err := pc.startFFmpeg()
	if err != nil {
		return err
	}

	// Cria context específico para readFrames
	pc.readCtx, pc.readCancel = context.WithCancel(pc.ctx)

	go pc.readFrames()
	go pc.monitorHealth()

	pc.running = true
	logger.Log.Infow("Captura persistente iniciada",
		"camera_id", pc.cameraID,
		"quality", pc.quality)

	return nil
}

func (pc *PersistentCapture) startFFmpeg() error {
	pc.cmd = exec.CommandContext(
		pc.ctx,
		"ffmpeg",
		"-loglevel", "fatal",  // Menos verboso
		"-rtsp_transport", "tcp",
		"-fflags", "+genpts+discardcorrupt",  // Descarta frames corruptos
		"-flags", "low_delay",
		"-err_detect", "ignore_err",  // Ignora erros HEVC
		"-i", pc.rtspURL,
		"-vf", fmt.Sprintf("fps=%d", pc.fps),
		"-f", "image2pipe",
		"-vcodec", "mjpeg",
		"-q:v", fmt.Sprintf("%d", pc.quality),
		"-",
	)

	var err error
	pc.stdout, err = pc.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("erro ao criar stdout pipe: %w", err)
	}

	pc.stderr, err = pc.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("erro ao criar stderr pipe: %w", err)
	}

	err = pc.cmd.Start()
	if err != nil {
		return fmt.Errorf("erro ao iniciar FFmpeg: %w", err)
	}

	go pc.logErrors()

	return nil
}

func (pc *PersistentCapture) readFrames() {
	reader := bufio.NewReader(pc.stdout)
	frameBuffer := bytes.NewBuffer(make([]byte, 0, 512*1024))

	jpegSOI := []byte{0xFF, 0xD8}
	jpegEOI := []byte{0xFF, 0xD9}

	for {
		select {
		case <-pc.readCtx.Done():
			// Context cancelado, sair silenciosamente
			return
		default:
		}

		b, err := reader.ReadByte()
		if err != nil {
			// Verifica se o context foi cancelado antes de reportar erro
			select {
			case <-pc.readCtx.Done():
				// Context cancelado durante restart, sair silenciosamente
				return
			default:
			}

			if err == io.EOF {
				pc.handleError("EOF no stream FFmpeg")
				return
			}
			pc.handleError(fmt.Sprintf("erro ao ler byte: %v", err))
			return
		}

		frameBuffer.WriteByte(b)

		if frameBuffer.Len() >= 2 {
			tail := frameBuffer.Bytes()[frameBuffer.Len()-2:]
			if bytes.Equal(tail, jpegEOI) {
				size := frameBuffer.Len()
				frameData := getFrameBuffer(size)
				copy(frameData, frameBuffer.Bytes())
				frameData = frameData[:size]

				if bytes.HasPrefix(frameData, jpegSOI) {
					select {
					case pc.frameBuffer <- frameData:
					default:
						logger.Log.Warnw("Frame buffer cheio, descartando frame",
							"camera_id", pc.cameraID)
						releaseFrameBuffer(frameData)
					}
					pc.markFrameReceived()
				}

				frameBuffer.Reset()
			}
		}
	}
}

func (pc *PersistentCapture) logErrors() {
	scanner := bufio.NewScanner(pc.stderr)
	for scanner.Scan() {
		line := scanner.Text()
		logger.Log.Warnw("FFmpeg stderr",
			"camera_id", pc.cameraID,
			"message", line)
	}
}

func (pc *PersistentCapture) monitorHealth() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-pc.ctx.Done():
			return

		case <-ticker.C:
			if pc.timeSinceLastFrame() > 30*time.Second {
				logger.Log.Warnw("Nenhum frame recebido há 30s, reiniciando captura",
					"camera_id", pc.cameraID)
				pc.Restart()
			}
		}
	}
}

func (pc *PersistentCapture) handleError(msg string) {
	logger.Log.Errorw("Erro na captura persistente",
		"camera_id", pc.cameraID,
		"error", msg)

	pc.errorCount++

	if pc.errorCount > 5 && time.Since(pc.lastRestart) < time.Minute {
		logger.Log.Errorw("Muitos erros em pouco tempo, aguardando antes de reiniciar",
			"camera_id", pc.cameraID,
			"error_count", pc.errorCount)
		time.Sleep(10 * time.Second)
	}

	pc.Restart()
}

func (pc *PersistentCapture) Restart() error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	// Evita restarts simultâneos
	if pc.restarting {
		logger.Log.Debugw("Restart já em andamento, ignorando",
			"camera_id", pc.cameraID)
		return nil
	}

	pc.restarting = true
	defer func() { pc.restarting = false }()

	// Cancela a goroutine readFrames atual
	if pc.readCancel != nil {
		pc.readCancel()
	}

	// Mata o processo FFmpeg
	if pc.cmd != nil && pc.cmd.Process != nil {
		_ = pc.cmd.Process.Kill()
		_ = pc.cmd.Wait()
	}

	// Fecha os pipes antigos
	if pc.stdout != nil {
		_ = pc.stdout.Close()
	}
	if pc.stderr != nil {
		_ = pc.stderr.Close()
	}

	// Aguarda um pouco antes de reiniciar
	time.Sleep(time.Second)

	// Reinicia o FFmpeg
	err := pc.startFFmpeg()
	if err != nil {
		logger.Log.Errorw("Erro ao reiniciar FFmpeg",
			"camera_id", pc.cameraID,
			"error", err)
		return err
	}

	// Cria novo context para a nova goroutine readFrames
	pc.readCtx, pc.readCancel = context.WithCancel(pc.ctx)

	pc.lastRestart = time.Now()
	pc.errorCount = 0

	// Inicia nova goroutine readFrames
	go pc.readFrames()

	logger.Log.Infow("Captura reiniciada",
		"camera_id", pc.cameraID)

	return nil
}

func (pc *PersistentCapture) GetFrame() ([]byte, bool) {
	select {
	case frame, ok := <-pc.frameBuffer:
		if !ok {
			return nil, false
		}
		return frame, true
	case <-time.After(5 * time.Second):
		return nil, false
	}
}

func (pc *PersistentCapture) GetFrameNonBlocking() ([]byte, bool) {
	select {
	case frame, ok := <-pc.frameBuffer:
		if !ok {
			return nil, false
		}
		return frame, true
	default:
		return nil, false
	}
}

func (pc *PersistentCapture) GetFrameWithTimeout(timeout time.Duration) ([]byte, bool) {
	ctx, cancel := context.WithTimeout(pc.ctx, timeout)
	defer cancel()

	select {
	case frame, ok := <-pc.frameBuffer:
		if !ok {
			return nil, false
		}
		return frame, true
	case <-ctx.Done():
		return nil, false
	}
}

func (pc *PersistentCapture) Stop() {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if !pc.running {
		return
	}

	pc.cancel()

	// Cancela a goroutine readFrames
	if pc.readCancel != nil {
		pc.readCancel()
	}

	if pc.cmd != nil && pc.cmd.Process != nil {
		_ = pc.cmd.Process.Kill()
		_ = pc.cmd.Wait()
	}

	// Fecha os pipes
	if pc.stdout != nil {
		_ = pc.stdout.Close()
	}
	if pc.stderr != nil {
		_ = pc.stderr.Close()
	}

	close(pc.frameBuffer)
	pc.running = false

	logger.Log.Infow("Captura persistente parada",
		"camera_id", pc.cameraID)
}

func (pc *PersistentCapture) IsRunning() bool {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.running
}

func (pc *PersistentCapture) markFrameReceived() {
	pc.lastFrameNS.Store(time.Now().UnixNano())
}

func (pc *PersistentCapture) timeSinceLastFrame() time.Duration {
	last := pc.lastFrameNS.Load()
	if last == 0 {
		return time.Since(pc.lastRestart)
	}
	return time.Since(time.Unix(0, last))
}
