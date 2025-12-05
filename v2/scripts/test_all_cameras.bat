@echo off
REM ========================================
REM  Edge Video V2 - Multi-Camera Viewer
REM  Abre 6 viewers simultaneamente
REM ========================================

echo ========================================
echo  INICIANDO VIEWERS DE TODAS AS CAMERAS
echo ========================================
echo.

REM Define o caminho do Python (ajuste se necessário)
set PYTHON=python

REM Define o caminho do script viewer
set VIEWER_SCRIPT=..\examples\viewer_cam1_sync.py

REM Abre cada viewer em um terminal separado
echo [1/6] Abrindo viewer CAM1...
start "CAM1 Viewer" cmd /k "%PYTHON% %VIEWER_SCRIPT% cam1"
timeout /t 1 /nobreak >nul

echo [2/6] Abrindo viewer CAM2...
start "CAM2 Viewer" cmd /k "%PYTHON% %VIEWER_SCRIPT% cam2"
timeout /t 1 /nobreak >nul

echo [3/6] Abrindo viewer CAM3...
start "CAM3 Viewer" cmd /k "%PYTHON% %VIEWER_SCRIPT% cam3"
timeout /t 1 /nobreak >nul

echo [4/6] Abrindo viewer CAM4...
start "CAM4 Viewer" cmd /k "%PYTHON% %VIEWER_SCRIPT% cam4"
timeout /t 1 /nobreak >nul

echo [5/6] Abrindo viewer CAM5...
start "CAM5 Viewer" cmd /k "%PYTHON% %VIEWER_SCRIPT% cam5"
timeout /t 1 /nobreak >nul

echo [6/6] Abrindo viewer CAM6...
start "CAM6 Viewer" cmd /k "%PYTHON% %VIEWER_SCRIPT% cam6"
timeout /t 1 /nobreak >nul

echo.
echo ========================================
echo  TODOS OS VIEWERS FORAM INICIADOS!
echo ========================================
echo.
echo 6 janelas foram abertas, uma para cada camera.
echo.
echo Para fechar todos os viewers:
echo   - Feche cada janela individualmente OU
echo   - Use Ctrl+C em cada terminal
echo.
echo Pressione qualquer tecla para fechar este terminal...
pause >nul
