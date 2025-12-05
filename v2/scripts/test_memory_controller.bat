@echo off
REM Test Memory Controller - Edge Video V2
REM This script tests the memory controller with different configurations

echo ================================================
echo   Memory Controller Test - Edge Video V2
echo ================================================
echo.

REM Ensure we're in the v2 directory
cd /d "%~dp0.."

REM Check if binary exists
if not exist "bin\edge-video-v2.exe" (
    echo ERROR: Binary not found. Building...
    go build -o bin\edge-video-v2.exe .\src
    if errorlevel 1 (
        echo ERROR: Build failed!
        pause
        exit /b 1
    )
)

echo.
echo Test 1: Memory Controller ENABLED (max 2048 MB)
echo ------------------------------------------------
echo Starting system with memory controller enabled...
echo Watch for memory level changes: NORMAL -> WARNING -> CRITICAL -> EMERGENCY
echo.
echo Press Ctrl+C to stop after 2 minutes or when you see memory warnings
echo.
pause

.\bin\edge-video-v2.exe -config config.yaml

echo.
echo ================================================
echo Test completed!
echo.
echo Expected behavior:
echo - Memory Controller should monitor RAM every 5 seconds
echo - When memory usage exceeds 60%%, you'll see WARNING
echo - When memory usage exceeds 75%%, you'll see CRITICAL
echo - When memory usage exceeds 85%%, you'll see EMERGENCY
echo - GC should be triggered automatically at 70%%
echo - Check the final report for memory controller stats
echo ================================================
echo.
pause
