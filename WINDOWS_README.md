# Newslettar Windows MSI Installer

## Installation Steps

1. **Download** the MSI installer from the [releases page](https://github.com/MadSwell-dev/newslettar/releases)
2. **Run** the MSI installer (requires administrator privileges)
3. **Launch** Newslettar from the Start Menu

## First Launch

When you first launch Newslettar from the Start Menu:

1. A console window will appear asking for administrator privileges
2. Click "Yes" to grant admin rights
3. The Newslettar Windows service will be installed automatically
4. Your default web browser will open to http://localhost:8080
5. Configure your Sonarr/Radarr and email settings

## After First Launch

Every subsequent launch will simply:
- Ensure the service is running
- Open your browser to the Web UI

No more admin prompts needed!

## What Gets Installed

- **Newslettar service**: Runs in background, starts automatically with Windows
- **Start Menu shortcut**: Quick access to launch the Web UI
- **Installation directory**: `C:\Program Files\Newslettar`
- **Configuration**: Stored in `.env` file in installation directory

## Accessing the Web UI

Three ways:
1. Click **Start Menu → Newslettar**
2. Open browser to **http://localhost:8080**
3. Run **newslettar-ctl.bat web** from installation directory

## Service Management

The Newslettar service runs automatically. To manage it:

```cmd
# Check status
sc query Newslettar

# Start service
net start Newslettar

# Stop service
net stop Newslettar

# Restart service
net stop Newslettar && net start Newslettar
```

Or use the control script:
```cmd
cd "C:\Program Files\Newslettar"
newslettar-ctl.bat [start|stop|restart|status|logs]
```

## Configuration

All configuration is done through the Web UI at http://localhost:8080

For advanced users:
- Edit: `C:\Program Files\Newslettar\.env`
- After editing, restart: `net stop Newslettar && net start Newslettar`

## Logs

View service logs:
```cmd
cd "C:\Program Files\Newslettar"
type newslettar-service.out.log
```

Or view live logs:
```cmd
newslettar-ctl.bat logs
```

## Firewall

The installer automatically creates a Windows Firewall rule for port 8080.

## Uninstall

1. Stop the service: `net stop Newslettar`
2. Uninstall via **Windows Settings → Apps → Newslettar**
3. Manually remove service: `"C:\Program Files\Newslettar\newslettar-service.exe" uninstall`

## Troubleshooting

### Start Menu shortcut doesn't work
- Right-click the shortcut → Run as administrator
- Check that launcher.bat exists in `C:\Program Files\Newslettar\`

### Service won't start
- Check logs at `C:\Program Files\Newslettar\newslettar-service.out.log`
- Ensure port 8080 is not in use
- Try manual start: `net start Newslettar`

### Can't save configuration
- Ensure the service is installed (launch from Start Menu first)
- Check file permissions on `C:\Program Files\Newslettar\.env`
- Run launcher as administrator

### localhost:8080 doesn't work
- Verify service is running: `sc query Newslettar`
- Check for port conflicts: `netstat -ano | findstr :8080`
- Review service logs for errors

## Support

- **Documentation**: See WINDOWS_INSTALL.txt in installation directory
- **Issues**: https://github.com/MadSwell-dev/newslettar/issues
- **Releases**: https://github.com/MadSwell-dev/newslettar/releases
