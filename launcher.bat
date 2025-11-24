@echo off
REM ============================================================================
REM Newslettar Launcher - First-run setup and browser opener
REM ============================================================================

REM Get the directory where this script is located (installation directory)
set INSTALL_DIR=%~dp0
REM Remove trailing backslash
set INSTALL_DIR=%INSTALL_DIR:~0,-1%

echo.
echo ============================================
echo   Newslettar Launcher
echo ============================================
echo.
echo Install Directory: %INSTALL_DIR%
echo.

REM Check if service is installed
sc query Newslettar >nul 2>&1
if %errorlevel% neq 0 (
    echo [INFO] Newslettar service not found. First-time setup required.
    echo.
    echo ============================================
    echo   Newslettar First-Time Setup
    echo ============================================
    echo.
    echo This will install the Newslettar Windows service.
    echo Administrator privileges are required.
    echo.
    echo Press any key to continue, or close this window to cancel.
    pause >nul

    echo.
    echo [STEP 1/3] Requesting administrator privileges...
    echo.

    REM Run the installer with admin privileges
    powershell -Command "Start-Process powershell -Verb RunAs -ArgumentList '-ExecutionPolicy Bypass -NoProfile -Command \"& \"\"\"%INSTALL_DIR%\scripts\install-windows.ps1\"\"\"\"' -Wait"

    echo.
    echo [STEP 2/3] Verifying service installation...
    timeout /t 3 /nobreak >nul

    REM Check if service was installed successfully
    sc query Newslettar >nul 2>&1
    if %errorlevel% neq 0 (
        echo.
        echo ============================================
        echo   [ERROR] Installation Failed
        echo ============================================
        echo.
        echo The service was not installed successfully.
        echo.
        echo Troubleshooting:
        echo   1. Make sure you clicked "Yes" when prompted for admin rights
        echo   2. Check if antivirus is blocking the installation
        echo   3. Try running manually as administrator:
        echo      Right-click: %INSTALL_DIR%\scripts\install-windows.ps1
        echo      Select: "Run with PowerShell"
        echo.
        echo For logs, check:
        echo   %INSTALL_DIR%\newslettar-service.out.log
        echo.
        pause
        exit /b 1
    )

    echo [OK] Service installed successfully!
    echo.
)

echo [STEP 3/3] Starting service and opening Web UI...
echo.

REM Check if service is running and start it if needed
sc query Newslettar | find "RUNNING" >nul 2>&1
if %errorlevel% neq 0 (
    echo Starting Newslettar service...
    net start Newslettar
    if %errorlevel% neq 0 (
        echo.
        echo [WARNING] Failed to start service.
        echo.
        echo Please check the logs at:
        echo   %INSTALL_DIR%\newslettar-service.out.log
        echo.
        echo Or try manually:
        echo   net start Newslettar
        echo.
        pause
        exit /b 1
    )
    echo [OK] Service started!
    timeout /t 2 /nobreak >nul
) else (
    echo [OK] Service is already running.
)

echo.
echo ============================================
echo   Opening Web UI
echo ============================================
echo.
echo URL: http://localhost:8080
echo.

REM Open Web UI in browser
start http://localhost:8080

echo.
echo You can close this window now.
timeout /t 5 /nobreak >nul
