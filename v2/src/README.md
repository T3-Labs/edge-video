# 💻 Source Code - Edge Video V2

## 📁 File Organization

| File | Purpose | Lines | Status |
|------|---------|-------|--------|
| **main.go** | Main entry point, initialization, stats monitor | ~200 | ✅ Active |
| **camera_stream.go** | FFmpeg capture, Latest Frame Policy, Circuit Breaker integration | ~400 | ✅ Active |
| **circuit_breaker.go** | Circuit Breaker implementation with exponential backoff | ~390 | ✅ Active (V2.2) |
| **publisher.go** | RabbitMQ AMQP publisher with auto-reconnect | ~300 | ✅ Active |
| **config.go** | YAML configuration loader | ~100 | ✅ Active |
| **profiling.go** | Performance profiling + System metrics (CPU/RAM) | ~200 | ✅ Active (V2.2) |
| **pool.go** | Local buffer pooling per camera | ~50 | ✅ Active |
| **camera.go** | Legacy camera interface (deprecated) | ~100 | ⚠️ Deprecated |

**Total**: ~1,740 lines of production Go code

---

## 🏗️ Architecture Overview

### Dual-Goroutine per Camera

```
┌─────────────────────────────────────────────┐
│         CameraStream (per camera)           │
├─────────────────────────────────────────────┤
│                                             │
│  ┌──────────┐      ┌────────────┐          │
│  │ FFmpeg   │      │ frameChan  │          │
│  │ Reader   ├─────>│  (buf=5)   ├─────┐    │
│  └──────────┘      └────────────┘     │    │
│  (readFrames)                         │    │
│                                       ▼    │
│                              ┌──────────────┐│
│                              │ publishLoop  ││
│                              │ (Latest Frame)│
│                              └──────┬───────┘│
│                                     │        │
└─────────────────────────────────────┼────────┘
                                      │
                                      ▼
                              ┌──────────────┐
                              │  Publisher   │
                              │  (AMQP)      │
                              └──────┬───────┘
                                     │
                                     ▼
                              ┌──────────┐
                              │RabbitMQ  │
                              └──────────┘
```

### Key Components

1. **Camera Capture** (`camera_stream.go`)
   - FFmpeg process management
   - JPEG frame detection (0xFFD8...0xFFD9)
   - Local buffer pooling (10 buffers per camera)
   - Latest Frame Policy (sync guarantee)

2. **Circuit Breaker** (`circuit_breaker.go`)
   - States: CLOSED → OPEN → HALF_OPEN
   - Exponential backoff: 5s → 10s → 20s → 40s → max 5min
   - Per-camera isolation
   - Auto-recovery detection

3. **Publisher** (`publisher.go`)
   - Auto-reconnect with exponential backoff
   - Connection monitoring
   - Graceful degradation
   - Per-camera AMQP channel

4. **Profiling** (`profiling.go`)
   - Publishing latency tracking
   - Memory/GC statistics
   - System metrics (CPU/RAM via gopsutil)
   - Circuit Breaker state tracking

---

## 🔧 Building

### Standard Build
```bash
cd v2
go build -o bin/edge-video-v2.exe ./src
```

### Optimized Build (Production)
```bash
cd v2
go build -ldflags="-s -w" -o bin/edge-video-v2.exe ./src
```

Flags explanation:
- `-s`: Strip symbol table (smaller binary)
- `-w`: Strip DWARF debug info (smaller binary)

### Cross-Compilation

**Linux**:
```bash
GOOS=linux GOARCH=amd64 go build -o bin/edge-video-v2-linux ./src
```

**macOS**:
```bash
GOOS=darwin GOARCH=amd64 go build -o bin/edge-video-v2-darwin ./src
```

**ARM (Raspberry Pi)**:
```bash
GOOS=linux GOARCH=arm GOARM=7 go build -o bin/edge-video-v2-rpi ./src
```

---

## 🧪 Testing

### Unit Tests
```bash
cd v2/src
go test -v ./...
```

### Benchmarks
```bash
cd v2/src
go test -bench=. -benchmem
```

### Code Coverage
```bash
cd v2/src
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## 📊 Code Metrics

### Complexity by File

| File | Lines | Functions | Complexity | Maintainability |
|------|-------|-----------|------------|-----------------|
| camera_stream.go | ~400 | 12 | Medium | High |
| circuit_breaker.go | ~390 | 15 | Medium | High |
| publisher.go | ~300 | 10 | Low | High |
| profiling.go | ~200 | 8 | Low | High |
| main.go | ~200 | 5 | Low | High |
| config.go | ~100 | 3 | Low | High |
| pool.go | ~50 | 4 | Low | High |

**Total Cyclomatic Complexity**: Low (all files < 15 per function)

### Dependencies

```
github.com/rabbitmq/amqp091-go v1.10.0    # RabbitMQ AMQP client
github.com/shirou/gopsutil/v3 v3.24.5     # System metrics (CPU/RAM)
gopkg.in/yaml.v3 v3.0.1                   # YAML parsing
```

**Total external dependencies**: 3 (minimal footprint)

---

## 🔍 Code Guidelines

### Style
- **Formatting**: `gofmt` enforced
- **Linting**: `golangci-lint` recommended
- **Imports**: Standard library first, then external
- **Comments**: GoDoc style for all exported functions

### Naming Conventions
- **Structs**: PascalCase (`CameraStream`, `Publisher`)
- **Functions**: camelCase (`readFrames`, `publishLoop`)
- **Constants**: SCREAMING_SNAKE_CASE for private, PascalCase for exported
- **Variables**: camelCase, descriptive names

### Error Handling
```go
// GOOD: Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to connect camera %s: %w", id, err)
}

// BAD: Swallow errors
if err != nil {
    log.Println(err)
}
```

### Concurrency
- **Mutexes**: Use `sync.RWMutex` when read-heavy
- **Channels**: Buffered for async operations, unbuffered for sync
- **Goroutines**: Always have cleanup mechanism (`context.Context`)

---

## 🐛 Debugging

### Enable Debug Logs
```go
// In main.go, add:
log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
```

### Race Detector
```bash
go build -race -o bin/edge-video-v2-race ./src
./bin/edge-video-v2-race
```

### Memory Profiling
```bash
go build -o bin/edge-video-v2-prof ./src
./bin/edge-video-v2-prof &
go tool pprof http://localhost:6060/debug/pprof/heap
```

### CPU Profiling
```bash
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
```

---

## 📝 Adding New Features

### Example: Adding a New Camera Protocol

1. **Define interface** in `camera_stream.go`:
```go
type CameraProtocol interface {
    Start() error
    ReadFrame() ([]byte, error)
    Stop() error
}
```

2. **Implement protocol**:
```go
type WebRTCCamera struct {
    // fields
}

func (w *WebRTCCamera) Start() error {
    // implementation
}
```

3. **Integrate with CameraStream**:
```go
// In NewCameraStream, detect protocol
if strings.HasPrefix(url, "webrtc://") {
    c.protocol = &WebRTCCamera{...}
}
```

4. **Test thoroughly**:
```bash
go test -v ./src -run TestWebRTCCamera
```

---

## 🔐 Security Considerations

### Sensitive Data
- ✅ AMQP credentials loaded from config (not hardcoded)
- ✅ Camera credentials in URL (supports URL encoding)
- ⚠️ Config file should have restricted permissions (0600)

### Input Validation
- ✅ FPS validated (1-60 range)
- ✅ Quality validated (2-31 range)
- ✅ URL format validated
- ⚠️ Add sanitization for camera IDs (prevent path traversal)

### Dependencies
- ✅ All dependencies from trusted sources (official Go modules)
- ✅ Minimal dependency tree (3 external packages)
- ⚠️ Run `go mod verify` regularly

---

## 📚 Additional Resources

- **Go Documentation**: https://golang.org/doc/
- **AMQP 0.9.1 Spec**: https://www.rabbitmq.com/amqp-0-9-1-reference.html
- **FFmpeg Documentation**: https://ffmpeg.org/documentation.html
- **gopsutil**: https://github.com/shirou/gopsutil

---

**Maintained By**: Edge Video V2 Team
**Last Updated**: 2024-12-05
**Go Version**: 1.21+
