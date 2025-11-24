@echo off
REM Wrapper script to run the installer and keep window open on errors

echo Running Newslettar installer...
echo.

cd /d "%~dp0"
powershell -ExecutionPolicy Bypass -NoProfile -File "scripts\install-windows.ps1"

if %errorlevel% neq 0 (
    echo.
    echo ========================================
    echo   Installation Failed!
    echo ========================================
    echo.
    echo Please review the errors above.
    echo.
    pause
    exit /b 1
)

echo.
echo Installation complete!
timeout /t 3 /nobreak >nul
