@echo off
REM Wrapper script to run the installer and keep window open on errors

REM Get the directory where this script is located
set WRAPPER_DIR=%~dp0
REM Remove trailing backslash
set WRAPPER_DIR=%WRAPPER_DIR:~0,-1%

echo ============================================
echo   Newslettar Installation
echo ============================================
echo.
echo Installation directory: %WRAPPER_DIR%
echo.

REM Check if install script exists
if not exist "%WRAPPER_DIR%\scripts\install-windows.ps1" (
    echo [ERROR] Install script not found!
    echo.
    echo Expected at: %WRAPPER_DIR%\scripts\install-windows.ps1
    echo.
    echo Checking what files are present...
    echo.
    echo Files in installation directory:
    dir /B "%WRAPPER_DIR%"
    echo.
    if exist "%WRAPPER_DIR%\scripts" (
        echo Files in scripts directory:
        dir /B "%WRAPPER_DIR%\scripts"
    ) else (
        echo [ERROR] Scripts directory does not exist!
    )
    echo.
    echo Please verify the MSI installation completed successfully.
    echo.
    pause
    exit /b 1
)

echo Running installer from: %WRAPPER_DIR%\scripts\install-windows.ps1
echo.

cd /d "%WRAPPER_DIR%"
powershell -ExecutionPolicy Bypass -NoProfile -File "%WRAPPER_DIR%\scripts\install-windows.ps1"

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
echo ========================================
echo   Installation Complete!
echo ========================================
echo.
echo This window will close in 3 seconds...
timeout /t 3 /nobreak >nul
