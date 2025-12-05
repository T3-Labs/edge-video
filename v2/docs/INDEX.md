# 📚 Edge Video V2 - Documentation Index

## 📖 Core Documentation

### Getting Started
- **[README.md](../README.md)** - Main documentation, architecture, and quick start
- **[TESTING_CHECKLIST.md](TESTING_CHECKLIST.md)** - Testing procedures and validation checklist

### Architecture & Design
- **[BUG_FIX_FRAME_CONTAMINATION.md](BUG_FIX_FRAME_CONTAMINATION.md)** - Critical bug fix documentation (V2.1)
- **[DIAGNOSTICO_JPEG.md](DIAGNOSTICO_JPEG.md)** - JPEG compression diagnostics
- **[ROADMAP_ENTERPRISE.md](ROADMAP_ENTERPRISE.md)** - Enterprise features roadmap

### Release Notes & Changelogs
- **[CHANGELOG_V2.2.md](CHANGELOG_V2.2.md)** - V2.2 release notes (Circuit Breaker & System Metrics)
- **[RELEASE_NOTES_V2.1.md](RELEASE_NOTES_V2.1.md)** - V2.1 release notes (Frame Cross-Contamination fix)
- **[TEST_ALL_CAMERAS_README.md](TEST_ALL_CAMERAS_README.md)** - Multi-camera testing guide

---

## 📂 Project Structure

```
v2/
├── README.md                    # Main documentation
├── config.yaml                  # Configuration file
├── go.mod / go.sum             # Go dependencies
│
├── bin/                        # Compiled binaries
│   └── edge-video-v2.exe
│
├── docs/                       # All documentation
│   ├── INDEX.md                # This file
│   ├── BUG_FIX_FRAME_CONTAMINATION.md
│   ├── CHANGELOG_V2.2.md
│   ├── DIAGNOSTICO_JPEG.md
│   ├── RELEASE_NOTES_V2.1.md
│   ├── ROADMAP_ENTERPRISE.md
│   ├── TEST_ALL_CAMERAS_README.md
│   └── TESTING_CHECKLIST.md
│
├── examples/                   # Example scripts and viewers
│   └── viewer_cam1_sync.py     # Python viewer for testing
│
├── logs/                       # Runtime logs
│   └── test_output.log
│
├── scripts/                    # Utility scripts
│   └── test_all_cameras.bat   # Multi-camera test script
│
└── src/                        # Source code (Go)
    ├── main.go                 # Main entry point
    ├── camera_stream.go        # Camera capture + Latest Frame Policy
    ├── circuit_breaker.go      # Circuit Breaker implementation
    ├── publisher.go            # RabbitMQ publisher with auto-reconnect
    ├── config.go              # Configuration loader
    ├── profiling.go           # Performance profiling + System metrics
    └── pool.go                # Local buffer pooling per camera
```

---

## 🔍 Quick Navigation

### By Topic

**Setup & Configuration**:
- [README.md](../README.md) → Section "🚀 Início Rápido"
- [config.yaml](../config.yaml) → Configuration file with comments

**Performance**:
- [README.md](../README.md) → Section "📊 Métricas de Performance"
- [DIAGNOSTICO_JPEG.md](DIAGNOSTICO_JPEG.md) → JPEG compression analysis

**Bug Fixes**:
- [BUG_FIX_FRAME_CONTAMINATION.md](BUG_FIX_FRAME_CONTAMINATION.md) → V2.1 critical fix
- [CHANGELOG_V2.2.md](CHANGELOG_V2.2.md) → V2.2 new features

**Testing**:
- [TESTING_CHECKLIST.md](TESTING_CHECKLIST.md) → Full testing procedures
- [TEST_ALL_CAMERAS_README.md](TEST_ALL_CAMERAS_README.md) → Multi-camera testing

**Enterprise Features**:
- [ROADMAP_ENTERPRISE.md](ROADMAP_ENTERPRISE.md) → Future enterprise features
- [CHANGELOG_V2.2.md](CHANGELOG_V2.2.md) → Circuit Breaker & System Metrics (V2.2)

---

## 📊 Version History

| Version | Date | Description | Documentation |
|---------|------|-------------|---------------|
| **V2.2** | 2024-12-05 | Circuit Breaker & System Metrics | [CHANGELOG_V2.2.md](CHANGELOG_V2.2.md) |
| **V2.1** | 2024-12-05 | Frame Cross-Contamination Fix | [RELEASE_NOTES_V2.1.md](RELEASE_NOTES_V2.1.md) |
| **V2.0** | 2024-11-27 | Production-Ready Release | [README.md](../README.md) |

---

## 🎯 Recommended Reading Order

### For New Users:
1. **[README.md](../README.md)** - Overview, architecture, and quick start
2. **[config.yaml](../config.yaml)** - Review configuration options
3. **[TESTING_CHECKLIST.md](TESTING_CHECKLIST.md)** - Run initial tests

### For Developers:
1. **[README.md](../README.md)** - Architecture section
2. **[BUG_FIX_FRAME_CONTAMINATION.md](BUG_FIX_FRAME_CONTAMINATION.md)** - Critical design decision
3. **[CHANGELOG_V2.2.md](CHANGELOG_V2.2.md)** - Latest features implementation
4. Source code files in `src/` directory

### For Operations:
1. **[README.md](../README.md)** - Setup and configuration
2. **[ROADMAP_ENTERPRISE.md](ROADMAP_ENTERPRISE.md)** - Enterprise features
3. **[TEST_ALL_CAMERAS_README.md](TEST_ALL_CAMERAS_README.md)** - Multi-camera testing

---

## 📞 Support

For questions and improvements, see:
- **Source code**: Comments in `.go` files are comprehensive
- **Issues**: Open GitHub issue with detailed description
- **Documentation**: All docs are in this `docs/` folder

---

**Last Updated**: 2024-12-05
**Maintained By**: Edge Video V2 Team
