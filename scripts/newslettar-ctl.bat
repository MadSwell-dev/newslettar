@echo off
REM ============================================================================
REM Newslettar Control Script for Windows
REM ============================================================================
REM Usage: newslettar-ctl.bat {start|stop|restart|status|logs|web}

setlocal

set INSTALL_DIR=%ProgramFiles%\Newslettar

if "%1"=="" goto usage
if "%1"=="start" goto start
if "%1"=="stop" goto stop
if "%1"=="restart" goto restart
if "%1"=="status" goto status
if "%1"=="logs" goto logs
if "%1"=="web" goto web
if "%1"=="edit" goto edit
goto usage

:start
echo Starting Newslettar service...
net start Newslettar
if %errorlevel% equ 0 (
    echo [32m✓ Newslettar started[0m
) else (
    echo [31m✗ Failed to start service[0m
)
goto end

:stop
echo Stopping Newslettar service...
net stop Newslettar
if %errorlevel% equ 0 (
    echo [33mNewslettar stopped[0m
) else (
    echo [31m✗ Failed to stop service[0m
)
goto end

:restart
echo Restarting Newslettar service...
net stop Newslettar
timeout /t 2 /nobreak >nul
net start Newslettar
if %errorlevel% equ 0 (
    echo [32m✓ Newslettar restarted[0m
) else (
    echo [31m✗ Failed to restart service[0m
)
goto end

:status
sc query Newslettar
goto end

:logs
echo Opening log file...
if exist "%INSTALL_DIR%\newslettar-service.out.log" (
    powershell -Command "Get-Content '%INSTALL_DIR%\newslettar-service.out.log' -Tail 50 -Wait"
) else (
    echo [33mLog file not found[0m
)
goto end

:web
for /f "tokens=2 delims=:" %%a in ('ipconfig ^| findstr /c:"IPv4"') do set IP=%%a
set IP=%IP:~1%
echo.
echo [32mWeb UI:[0m http://localhost:8080
echo [32mOr:[0m http://%IP%:8080
echo.
goto end

:edit
if exist "%INSTALL_DIR%\.env" (
    notepad "%INSTALL_DIR%\.env"
    echo.
    echo [33mRemember to restart: newslettar-ctl.bat restart[0m
) else (
    echo [31mConfiguration file not found[0m
)
goto end

:usage
echo.
echo ╔════════════════════════════════════════╗
echo ║     Newslettar Control                 ║
echo ╚════════════════════════════════════════╝
echo.
echo Usage: newslettar-ctl.bat {command}
echo.
echo Commands:
echo   start    - Start Newslettar service
echo   stop     - Stop Newslettar service
echo   restart  - Restart Newslettar service
echo   status   - Show service status
echo   logs     - View logs (live)
echo   web      - Show Web UI URL
echo   edit     - Edit configuration file
echo.
goto end

:end
endlocal
