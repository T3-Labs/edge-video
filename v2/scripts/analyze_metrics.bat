@echo off
REM Analyze Memory and Metrics - Edge Video V2
REM This script shows detailed memory consumption analysis

echo ================================================
echo   Memory and Metrics Analysis - Edge Video V2
echo ================================================
echo.

cd /d "%~dp0.."

echo Configuration Analysis:
echo ----------------------
echo Cameras configured: 6
echo FPS target: 15
echo JPEG Quality: 5
echo.

echo Expected Memory Calculation:
echo ----------------------------
echo.
echo Per Camera Components:
echo - FFmpeg process: ~50-100 MB per camera
echo - Frame buffer (5 frames): ~0.5-2 MB per camera
echo - AMQP channel: ~5-10 MB per camera
echo - Circuit breaker: ~1 KB per camera (negligible)
echo - Goroutines (2 per camera): ~8-16 KB per camera
echo.

echo 6 Cameras Estimation:
echo - FFmpeg (6x 75 MB):        450 MB
echo - Frame buffers (6x 1 MB):    6 MB
echo - AMQP channels (6x 7 MB):   42 MB
echo - Go runtime overhead:       50 MB
echo - Local buffer pools:        10 MB
echo -------------------------------------------
echo TOTAL ESTIMATED:            ~558 MB
echo.

echo Frame Size Analysis (from logs):
echo --------------------------------
echo cam1 (RTMP):  ~320 KB/frame
echo cam2 (RTSP):   ~64 KB/frame
echo cam3 (RTSP):  ~180 KB/frame
echo cam4 (RTSP):  ~115 KB/frame
echo cam5 (RTSP):   ~97 KB/frame
echo cam6 (RTSP):  NOT WORKING (Circuit Breaker OPEN)
echo.
echo Average frame size: ~155 KB
echo Total bandwidth @ 15 FPS: 155 KB x 15 FPS x 5 cameras = ~11.6 MB/s
echo.

echo Memory Controller Settings:
echo ---------------------------
echo Max Memory: 2048 MB (2 GB)
echo WARNING at: 1229 MB (60%%)
echo CRITICAL at: 1536 MB (75%%)
echo EMERGENCY at: 1741 MB (85%%)
echo GC Trigger at: 1434 MB (70%%)
echo.

echo Recommended Adjustments:
echo ------------------------
echo Based on ~558 MB estimated usage:
echo.
echo CURRENT (config.yaml):
echo   max_memory_mb: 2048     # 2 GB - OK, but may be too high
echo.
echo RECOMMENDED:
echo   max_memory_mb: 1024     # 1 GB - More appropriate for 6 cameras
echo   warning_percent: 50.0   # WARNING at 512 MB
echo   critical_percent: 70.0  # CRITICAL at 716 MB
echo   emergency_percent: 85.0 # EMERGENCY at 870 MB
echo   gc_trigger_percent: 60.0 # GC at 614 MB
echo.

echo Why these values?
echo - Expected usage: ~558 MB
echo - Safety margin: 558 MB / 1024 MB = 54%% (comfortable)
echo - WARNING at 512 MB (92%% of expected) = early detection
echo - GC trigger at 614 MB (110%% of expected) = proactive cleanup
echo - EMERGENCY at 870 MB (156%% of expected) = severe problem
echo.

echo To monitor REAL usage:
echo ----------------------
echo 1. Run: .\bin\edge-video-v2.exe -config config.yaml
echo 2. Wait 30-60 seconds for stats
echo 3. Check logs for:
echo    - "Memory (Go Runtime): Alloc: XX MB"
echo    - "Sistema (Processo): RAM Usage: XX MB"
echo 4. Observe memory level changes
echo.

echo Press any key to exit...
pause >nul
