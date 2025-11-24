@echo off
REM ============================================================================
REM Newslettar Launcher - First-run setup and browser opener
REM ============================================================================

cd /d "%~dp0.."

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

    REM Run install script with admin privileges
    powershell -Command "Start-Process powershell -Verb RunAs -ArgumentList '-ExecutionPolicy Bypass -File \"%CD%\scripts\install-windows.ps1\"' -Wait"

    if %errorlevel% neq 0 (
        echo.
        echo Installation failed or was cancelled.
        echo Please run scripts\install-windows.ps1 as administrator.
        pause
        exit /b 1
    )
)

REM Check if service is running
sc query Newslettar | find "RUNNING" >nul 2>&1
if %errorlevel% neq 0 (
    echo Starting Newslettar service...
    net start Newslettar >nul 2>&1
    timeout /t 3 /nobreak >nul
)

REM Open Web UI in browser
start http://localhost:8080
