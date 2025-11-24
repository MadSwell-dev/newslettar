@echo off
REM ============================================================================
REM Newslettar Launcher - First-run setup and browser opener
REM ============================================================================

REM Get the directory where this script is located (installation directory)
set INSTALL_DIR=%~dp0
REM Remove trailing backslash
set INSTALL_DIR=%INSTALL_DIR:~0,-1%

echo Newslettar Launcher
echo Install Directory: %INSTALL_DIR%
echo.

REM Check if service is installed
sc query Newslettar >nul 2>&1
if %errorlevel% neq 0 (
    echo ============================================
    echo   Newslettar First-Time Setup
    echo ============================================
    echo.
    echo The Newslettar service needs to be installed.
    echo This requires administrator privileges.
    echo.
    echo Press any key to install the service, or close this window to cancel.
    pause >nul

    REM Run install script with admin privileges using absolute path
    echo Running installer...
    powershell -Command "Start-Process powershell -Verb RunAs -ArgumentList '-ExecutionPolicy Bypass -File \"%INSTALL_DIR%\scripts\install-windows.ps1\"' -Wait"

    REM Check if service was installed successfully
    timeout /t 2 /nobreak >nul
    sc query Newslettar >nul 2>&1
    if %errorlevel% neq 0 (
        echo.
        echo [ERROR] Installation failed or was cancelled.
        echo.
        echo Troubleshooting:
        echo 1. Make sure you granted administrator privileges
        echo 2. Check the installation log for errors
        echo 3. Try manually running: %INSTALL_DIR%\scripts\install-windows.ps1
        echo.
        pause
        exit /b 1
    )

    echo.
    echo [OK] Service installed successfully!
    echo.
)

REM Check if service is running and start it if needed
sc query Newslettar | find "RUNNING" >nul 2>&1
if %errorlevel% neq 0 (
    echo Starting Newslettar service...
    net start Newslettar
    if %errorlevel% neq 0 (
        echo.
        echo [WARNING] Failed to start service automatically.
        echo Please try: net start Newslettar
        echo Or check logs at: %INSTALL_DIR%\newslettar-service.out.log
        echo.
        pause
    ) else (
        echo Service started successfully!
        timeout /t 3 /nobreak >nul
    )
) else (
    echo Service is already running.
)

echo.
echo Opening Web UI at http://localhost:8080
echo.

REM Open Web UI in browser
start http://localhost:8080
