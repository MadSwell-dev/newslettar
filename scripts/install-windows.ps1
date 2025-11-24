# =============================================================================
# Newslettar Windows Installer Script
# =============================================================================
# This script installs Newslettar as a Windows service
# Requires Administrator privileges
#
# USAGE:
#   powershell -ExecutionPolicy Bypass -File install-windows.ps1
#

# Requires Administrator
#Requires -RunAsAdministrator

$ErrorActionPreference = "Stop"

# Auto-detect installation directory from script location
# Note: go-msi flattens all files to the root installation directory
# This works regardless of Windows language (Program Files, Programmes, etc.)
$ScriptPath = Split-Path -Parent $MyInvocation.MyCommand.Path

# Check if we're already in the installation directory (go-msi flattens structure)
if (Test-Path "$ScriptPath\newslettar.exe") {
    $InstallDir = $ScriptPath
} else {
    # Maybe script is in a scripts/ subdirectory (manual installation)
    $InstallDir = Split-Path -Parent $ScriptPath
}

# Fallback: if still can't find it, try to detect
if (-not (Test-Path "$InstallDir\newslettar.exe")) {
    Write-Host "Auto-detection failed. Searching for Newslettar installation..." -ForegroundColor Yellow

    # Search in common locations
    $possiblePaths = @(
        "$env:ProgramFiles\Newslettar",
        "${env:ProgramFiles(x86)}\Newslettar",
        "C:\Newslettar"
    )

    foreach ($path in $possiblePaths) {
        if (Test-Path "$path\newslettar.exe") {
            $InstallDir = $path
            Write-Host "Found installation at: $InstallDir" -ForegroundColor Green
            break
        }
    }

    if (-not (Test-Path "$InstallDir\newslettar.exe")) {
        Write-Host "ERROR: Cannot find Newslettar installation!" -ForegroundColor Red
        Write-Host "Please run this script from the Newslettar installation directory." -ForegroundColor Yellow
        pause
        exit 1
    }
}

$ServiceName = "Newslettar"
$WinswVersion = "3.0.0-alpha.11"
$WinswUrl = "https://github.com/winsw/winsw/releases/download/v$WinswVersion/WinSW-x64.exe"

Write-Host "========================================" -ForegroundColor Green
Write-Host "    Newslettar Windows Installer        " -ForegroundColor Green
Write-Host "    Version 1.0.2                       " -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Installation Directory: $InstallDir" -ForegroundColor Cyan
Write-Host ""

# Function to check if service exists
function Test-ServiceExists {
    param($ServiceName)
    if (Get-Service -Name $ServiceName -ErrorAction SilentlyContinue) {
        return $true
    }
    return $false
}

# Stop and remove existing service if it exists
if (Test-ServiceExists -ServiceName $ServiceName) {
    Write-Host "[1/6] Stopping existing service..." -ForegroundColor Yellow
    Stop-Service -Name $ServiceName -Force -ErrorAction SilentlyContinue

    Write-Host "[1/6] Removing existing service..." -ForegroundColor Yellow
    & "$InstallDir\newslettar-service.exe" uninstall
    Start-Sleep -Seconds 2
    Write-Host "OK: Existing service removed" -ForegroundColor Green
}

# Create installation directory
Write-Host "[2/6] Creating installation directory..." -ForegroundColor Yellow
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

# Grant write permissions to LocalSystem account
try {
    $Acl = Get-Acl $InstallDir
    $Permission = "NT AUTHORITY\SYSTEM", "FullControl", "ContainerInherit,ObjectInherit", "None", "Allow"
    $AccessRule = New-Object System.Security.AccessControl.FileSystemAccessRule $Permission
    $Acl.SetAccessRule($AccessRule)
    Set-Acl $InstallDir $Acl
    Write-Host "OK: Directory created with proper permissions: $InstallDir" -ForegroundColor Green
} catch {
    Write-Host "OK: Directory created: $InstallDir" -ForegroundColor Green
    Write-Host "  Warning: Could not set permissions, using defaults" -ForegroundColor Yellow
}

# Copy newslettar.exe to installation directory
Write-Host "[3/6] Installing Newslettar..." -ForegroundColor Yellow

# Get current directory to check if we're already in the installation directory
$CurrentDir = (Get-Location).Path

if ($CurrentDir -ne $InstallDir) {
    # We're in a different directory (manual install), need to copy
    if (Test-Path ".\newslettar.exe") {
        Copy-Item ".\newslettar.exe" -Destination $InstallDir -Force
        Write-Host "OK: Newslettar binary installed" -ForegroundColor Green
    } else {
        Write-Host "ERROR: newslettar.exe not found in current directory" -ForegroundColor Red
        Write-Host "  Please ensure newslettar.exe is in the same directory as this script" -ForegroundColor Yellow
        exit 1
    }
} else {
    # Already in installation directory (MSI case), just verify file exists
    if (Test-Path "$InstallDir\newslettar.exe") {
        Write-Host "OK: Newslettar binary already in place" -ForegroundColor Green
    } else {
        Write-Host "ERROR: newslettar.exe not found in installation directory" -ForegroundColor Red
        exit 1
    }
}

# Download WinSW (Windows Service Wrapper)
Write-Host "[4/6] Downloading Windows Service Wrapper..." -ForegroundColor Yellow
$WinswPath = "$InstallDir\newslettar-service.exe"
try {
    Invoke-WebRequest -Uri $WinswUrl -OutFile $WinswPath -UseBasicParsing
    Write-Host "OK: Service wrapper downloaded" -ForegroundColor Green
} catch {
    Write-Host "ERROR: Failed to download WinSW: $_" -ForegroundColor Red
    exit 1
}

# Copy service configuration
Write-Host "[5/6] Configuring Windows service..." -ForegroundColor Yellow

if ($CurrentDir -ne $InstallDir) {
    # We're in a different directory (manual install), need to copy
    if (Test-Path ".\newslettar-service.xml") {
        Copy-Item ".\newslettar-service.xml" -Destination $InstallDir -Force
        Write-Host "OK: Service configuration installed" -ForegroundColor Green
    } else {
        # Create minimal service configuration if not found
        $ServiceConfig = @"
<?xml version="1.0" encoding="UTF-8"?>
<service>
  <id>Newslettar</id>
  <name>Newslettar</name>
  <description>Automated newsletter generator for Sonarr and Radarr</description>
  <executable>${InstallDir}\newslettar.exe</executable>
  <arguments>-web</arguments>
  <startmode>Automatic</startmode>
  <log mode="roll-by-size">
    <sizeThreshold>10240</sizeThreshold>
    <keepFiles>8</keepFiles>
  </log>
</service>
"@
        $ServiceConfig | Out-File -FilePath (Join-Path $InstallDir "newslettar-service.xml") -Encoding UTF8
        Write-Host "OK: Service configuration created" -ForegroundColor Green
    }
} else {
    # Already in installation directory (MSI case), verify file exists or create it
    if (Test-Path "$InstallDir\newslettar-service.xml") {
        Write-Host "OK: Service configuration already in place" -ForegroundColor Green
    } else {
        # Create minimal service configuration if not found
        $ServiceConfig = @"
<?xml version="1.0" encoding="UTF-8"?>
<service>
  <id>Newslettar</id>
  <name>Newslettar</name>
  <description>Automated newsletter generator for Sonarr and Radarr</description>
  <executable>${InstallDir}\newslettar.exe</executable>
  <arguments>-web</arguments>
  <startmode>Automatic</startmode>
  <log mode="roll-by-size">
    <sizeThreshold>10240</sizeThreshold>
    <keepFiles>8</keepFiles>
  </log>
</service>
"@
        $ServiceConfig | Out-File -FilePath (Join-Path $InstallDir "newslettar-service.xml") -Encoding UTF8
        Write-Host "OK: Service configuration created" -ForegroundColor Green
    }
}

# Create default .env file if it doesn't exist
if (-not (Test-Path "$InstallDir\.env")) {
    $EnvTemplate = @"
# Sonarr Configuration
SONARR_URL=http://localhost:8989
SONARR_API_KEY=

# Radarr Configuration
RADARR_URL=http://localhost:7878
RADARR_API_KEY=

# Email Configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=
SMTP_PASS=
FROM_NAME=Newslettar
FROM_EMAIL=newsletter@yourdomain.com
TO_EMAILS=user@example.com

# Schedule Settings
TIMEZONE=UTC
SCHEDULE_DAY=Sun
SCHEDULE_TIME=09:00

# Template Settings
SHOW_POSTERS=true
SHOW_DOWNLOADED=true

# Web UI Port
WEBUI_PORT=8080
"@
    $EnvTemplate | Out-File -FilePath (Join-Path $InstallDir ".env") -Encoding UTF8
    Write-Host "OK: Configuration file created" -ForegroundColor Green
} else {
    Write-Host "OK: Existing configuration preserved" -ForegroundColor Green
}

# Ensure .env file is writable
if (Test-Path "$InstallDir\.env") {
    try {
        $EnvFile = Get-Item "$InstallDir\.env" -Force
        $EnvFile.IsReadOnly = $false
    } catch {
        Write-Host "  Warning: Could not verify .env file permissions" -ForegroundColor Yellow
    }
}

# Install and start service
Write-Host "[6/6] Installing Windows service..." -ForegroundColor Yellow
Set-Location $InstallDir
& ".\newslettar-service.exe" install
Start-Sleep -Seconds 2

Write-Host "[6/6] Starting service..." -ForegroundColor Yellow
& ".\newslettar-service.exe" start
Start-Sleep -Seconds 3

Write-Host "OK: Service installed and started" -ForegroundColor Green
Write-Host ""

# Add firewall rule for web UI
Write-Host "Configuring Windows Firewall..." -ForegroundColor Yellow
try {
    New-NetFirewallRule -DisplayName "Newslettar Web UI" -Direction Inbound -LocalPort 8080 -Protocol TCP -Action Allow -ErrorAction SilentlyContinue | Out-Null
    Write-Host "OK: Firewall rule added for port 8080" -ForegroundColor Green
} catch {
    Write-Host "WARNING: Could not add firewall rule (may already exist)" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "     Installation Complete!             " -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "                                        " -ForegroundColor Cyan
Write-Host "      Web UI:                           " -ForegroundColor Cyan
Write-Host "      http://localhost:8080             " -ForegroundColor Cyan
Write-Host "                                        " -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Quick Start:" -ForegroundColor Yellow
Write-Host "  1. Open http://localhost:8080 in your browser"
Write-Host "  2. Configure Sonarr/Radarr in Configuration tab"
Write-Host "  3. Configure email settings"
Write-Host "  4. Test connections and send test newsletter"
Write-Host ""
Write-Host "Service Management:" -ForegroundColor Yellow
Write-Host "  Start:   Start-Service Newslettar"
Write-Host "  Stop:    Stop-Service Newslettar"
Write-Host "  Status:  Get-Service Newslettar"
Write-Host "  Logs:    Get-Content (Join-Path `"$InstallDir`" `"newslettar-service.out.log`") -Tail 50 -Wait"
Write-Host ""
Write-Host "Configuration:" -ForegroundColor Yellow
Write-Host "  Edit: ${InstallDir}\.env"
Write-Host "  After editing, restart the service: Restart-Service Newslettar"
Write-Host ""
Write-Host "Installation Directory: $InstallDir" -ForegroundColor Green
Write-Host ""
Write-Host "Opening Web UI in your browser..." -ForegroundColor Cyan
Start-Sleep -Seconds 2
Start-Process "http://localhost:8080"
Write-Host ""
