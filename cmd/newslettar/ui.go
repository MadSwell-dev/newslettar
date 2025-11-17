package main

// getUIHTML returns the full HTML for the web UI
func getUIHTML(version string, nextRun string, timezone string) string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Newslettar</title>
    <link rel="icon" href="/assets/newslettar_logo.svg" type="image/svg+xml">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: #0f1419;
            color: #e8e8e8;
            line-height: 1.6;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
        }
        
        /* Responsive design */
        @media (max-width: 768px) {
            .container { padding: 10px; }
            .header h1 { font-size: 1.8em; }
            .tabs { flex-wrap: wrap; }
            .tab { flex: 1 1 45%; font-size: 12px; padding: 10px; }
            .form-group { margin-bottom: 15px; }
            .action-buttons { flex-direction: column; }
            .action-buttons .btn { margin-bottom: 10px; }
        }
        
        .header {
            background: #1a2332;
            padding: 30px;
            border-radius: 12px;
            margin-bottom: 30px;
            text-align: center;
            border-bottom: 3px solid #667eea;
            position: relative;
        }
        .header-logo {
            max-width: 250px;
            height: auto;
            margin: 0 auto 15px;
            display: block;
        }
        .header h1 {
            font-size: 2.5em;
            margin-bottom: 10px;
            color: #e8e8e8;
        }
        .version {
            opacity: 0.7;
            font-size: 0.85em;
            color: #8899aa;
        }
        .tabs {
            display: flex;
            gap: 10px;
            margin-bottom: 20px;
            background: #1a2332;
            padding: 10px;
            border-radius: 10px;
        }
        .tab {
            flex: 1;
            padding: 12px 20px;
            background: transparent;
            border: none;
            color: #8899aa;
            cursor: pointer;
            border-radius: 8px;
            font-size: 14px;
            font-weight: 500;
            transition: all 0.3s;
        }
        .tab:hover { background: #252f3f; color: #fff; }
        .tab:focus { outline: 2px solid #667eea; outline-offset: 2px; }
        .tab.active {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: #fff;
        }
        .tab-content {
            display: none;
            background: #1a2332;
            padding: 30px;
            border-radius: 12px;
            min-height: 400px;
        }
        .tab-content.active { display: block; }
        .form-group {
            margin-bottom: 20px;
        }
        .form-group label {
            display: block;
            margin-bottom: 8px;
            color: #a0b0c0;
            font-weight: 500;
        }
        .form-group input, .form-group select {
            width: 100%;
            padding: 12px;
            background: #0f1419;
            border: 2px solid #2a3444;
            border-radius: 8px;
            color: #e8e8e8;
            font-size: 14px;
            transition: border-color 0.3s;
        }
        .form-group input:focus, .form-group select:focus {
            outline: none;
            border-color: #667eea;
        }
        .form-group input.error, .form-group select.error {
            border-color: #eb3349;
        }
        .form-group input.success, .form-group select.success {
            border-color: #38ef7d;
        }
        .error-message {
            color: #eb3349;
            font-size: 0.85em;
            margin-top: 5px;
            display: none;
        }
        .error-message.show { display: block; }
        .btn {
            padding: 12px 24px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            border-radius: 8px;
            cursor: pointer;
            font-size: 14px;
            font-weight: 600;
            transition: transform 0.2s, opacity 0.3s;
            position: relative;
        }
        .btn:hover { transform: translateY(-2px); opacity: 0.9; }
        .btn:active { transform: translateY(0); }
        .btn:disabled {
            opacity: 0.5;
            cursor: not-allowed;
            transform: none;
        }
        .btn:focus { outline: 2px solid #667eea; outline-offset: 2px; }
        .btn-secondary {
            background: #2a3444;
        }
        .btn-success {
            background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%);
        }
        .btn-danger {
            background: linear-gradient(135deg, #eb3349 0%, #f45c43 100%);
        }
        .btn.loading::after {
            content: "";
            position: absolute;
            width: 16px;
            height: 16px;
            top: 50%;
            left: 50%;
            margin-left: -8px;
            margin-top: -8px;
            border: 2px solid #ffffff40;
            border-top-color: #fff;
            border-radius: 50%;
            animation: spin 0.6s linear infinite;
        }
        .btn.loading span { opacity: 0; }
        @keyframes spin {
            to { transform: rotate(360deg); }
        }
        .notification {
            position: fixed;
            top: 20px;
            right: 20px;
            padding: 16px 24px;
            border-radius: 10px;
            color: white;
            font-weight: 500;
            animation: slideIn 0.3s;
            z-index: 1000;
            max-width: 400px;
        }
        .notification.success {
            background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%);
        }
        .notification.error {
            background: linear-gradient(135deg, #eb3349 0%, #f45c43 100%);
        }
        @keyframes slideIn {
            from { transform: translateX(400px); opacity: 0; }
            to { transform: translateX(0); opacity: 1; }
        }
        /* Dashboard Styles */
        .dashboard-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 20px;
            margin-bottom: 20px;
        }
        .dashboard-card {
            background: #252f3f;
            border-radius: 10px;
            overflow: hidden;
            border: 2px solid #2a3444;
        }
        .dashboard-logs {
            grid-column: 1 / -1;
        }
        .dashboard-card-header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            padding: 15px 20px;
            display: flex;
            align-items: center;
            gap: 10px;
        }
        .dashboard-card-header h4 {
            margin: 0;
            color: #fff;
            font-size: 1.1em;
        }
        .dashboard-icon {
            font-size: 1.5em;
        }
        .dashboard-card-content {
            padding: 20px;
        }
        .stat-row {
            display: flex;
            justify-content: space-between;
            padding: 10px 0;
            border-bottom: 1px solid #2a3444;
        }
        .stat-row:last-child {
            border-bottom: none;
        }
        .stat-label {
            color: #8899aa;
            font-size: 0.95em;
        }
        .stat-value {
            color: #e8e8e8;
            font-weight: 500;
        }
        .stat-highlight {
            color: #667eea;
            font-size: 1.2em;
            font-weight: 600;
        }
        .status-indicator {
            font-size: 0.8em;
            margin-right: 5px;
        }
        .dashboard-logs-container {
            background: #0f1419;
            padding: 15px;
            border-radius: 8px;
            font-family: 'Courier New', monospace;
            font-size: 12px;
            max-height: 300px;
            overflow-y: auto;
            white-space: pre-wrap;
            border: 2px solid #2a3444;
            line-height: 1.4;
        }
        @media (max-width: 768px) {
            .dashboard-grid {
                grid-template-columns: 1fr;
            }
        }
        .schedule-info {
            background: #252f3f;
            padding: 20px;
            border-radius: 10px;
            margin-bottom: 20px;
            border-left: 4px solid #667eea;
        }
        .schedule-info h3 {
            margin-bottom: 10px;
            color: #667eea;
        }
        .toggle-switch {
            position: relative;
            display: inline-block;
            width: 50px;
            height: 26px;
            margin-left: 10px;
        }
        .toggle-switch input {
            opacity: 0;
            width: 0;
            height: 0;
        }
        .toggle-slider {
            position: absolute;
            cursor: pointer;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background-color: #2a3444;
            transition: 0.3s;
            border-radius: 26px;
        }
        .toggle-slider:before {
            position: absolute;
            content: "";
            height: 20px;
            width: 20px;
            left: 3px;
            bottom: 3px;
            background-color: white;
            transition: 0.3s;
            border-radius: 50%;
        }
        input:checked + .toggle-slider {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
        }
        input:checked + .toggle-slider:before {
            transform: translateX(24px);
        }
        .template-option {
            display: flex;
            align-items: center;
            justify-content: space-between;
            padding: 15px;
            background: #252f3f;
            border-radius: 8px;
            margin-bottom: 12px;
        }
        .timezone-info {
            background: #252f3f;
            padding: 15px;
            border-radius: 8px;
            margin-top: 10px;
            font-size: 0.9em;
        }
        .timezone-info strong {
            color: #667eea;
        }
        .info-banner {
            background: #252f3f;
            padding: 15px 20px;
            border-radius: 8px;
            margin-bottom: 20px;
            border-left: 4px solid #11998e;
        }
        .info-banner p {
            margin: 5px 0;
            color: #a0b0c0;
        }
        .info-banner strong {
            color: #e8e8e8;
        }

        /* Email tags */
        .email-tags-container {
            width: 100%;
            min-height: 46px;
            padding: 8px;
            background: #0f1419;
            border: 2px solid #2a3444;
            border-radius: 8px;
            display: flex;
            flex-wrap: wrap;
            gap: 8px;
            align-items: center;
            cursor: text;
            transition: border-color 0.3s;
        }
        .email-tags-container:focus-within {
            border-color: #667eea;
        }
        .email-tags-container.error {
            border-color: #eb3349;
        }
        .email-tag {
            display: inline-flex;
            align-items: center;
            gap: 6px;
            padding: 6px 10px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border-radius: 6px;
            font-size: 14px;
            font-weight: 500;
            animation: tagSlideIn 0.2s ease-out;
        }
        .email-tag-remove {
            background: transparent;
            border: none;
            color: white;
            font-size: 16px;
            cursor: pointer;
            padding: 0;
            width: 18px;
            height: 18px;
            line-height: 1;
            border-radius: 3px;
            transition: background 0.2s;
        }
        .email-tag-remove:hover {
            background: rgba(255, 255, 255, 0.2);
        }
        .email-tag-input {
            flex: 1;
            min-width: 200px;
            border: none;
            background: transparent;
            color: #e8e8e8;
            font-size: 14px;
            outline: none;
            padding: 6px 0;
        }
        .email-tag-input::placeholder {
            color: #8899aa;
        }
        @keyframes tagSlideIn {
            from { transform: scale(0.8); opacity: 0; }
            to { transform: scale(1); opacity: 1; }
        }
        .email-section {
            background: #252f3f;
            padding: 20px;
            border-radius: 10px;
            margin-bottom: 20px;
            border-left: 4px solid #667eea;
        }
        .email-section h3 {
            margin-bottom: 10px;
            color: #667eea;
        }
        
        /* Preview Modal */
        .modal {
            display: none;
            position: fixed;
            z-index: 2000;
            left: 0;
            top: 0;
            width: 100%;
            height: 100%;
            background-color: rgba(0,0,0,0.8);
            animation: fadeIn 0.3s;
        }
        .modal.show { display: flex; align-items: center; justify-content: center; }
        .modal-content {
            background: #1a2332;
            width: 90%;
            max-width: 900px;
            max-height: 90vh;
            border-radius: 12px;
            overflow: hidden;
            animation: slideUp 0.3s;
        }
        .modal-header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            padding: 20px;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .modal-header h2 {
            margin: 0;
            color: white;
        }
        .modal-close {
            background: transparent;
            border: none;
            color: white;
            font-size: 28px;
            cursor: pointer;
            padding: 0;
            width: 30px;
            height: 30px;
            line-height: 1;
        }
        .modal-close:hover { opacity: 0.7; }
        .modal-close:focus { outline: 2px solid white; outline-offset: 2px; }
        .modal-body {
            padding: 20px;
            max-height: calc(90vh - 140px);
            overflow-y: auto;
        }
        .modal-body iframe {
            width: 100%;
            height: 600px;
            border: 2px solid #2a3444;
            border-radius: 8px;
            background: white;
        }
        @keyframes fadeIn {
            from { opacity: 0; }
            to { opacity: 1; }
        }
        @keyframes slideUp {
            from { transform: translateY(50px); opacity: 0; }
            to { transform: translateY(0); opacity: 1; }
        }
        
        .action-buttons {
            display: flex;
            gap: 10px;
            margin-top: 20px;
        }
        .action-buttons .btn {
            flex: 1;
        }
        
        /* Loading overlay */
        .loading-overlay {
            display: none;
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(0,0,0,0.7);
            z-index: 1500;
            align-items: center;
            justify-content: center;
        }
        .loading-overlay.show { display: flex; }
        .loading-spinner {
            width: 60px;
            height: 60px;
            border: 5px solid rgba(255,255,255,0.3);
            border-top-color: #667eea;
            border-radius: 50%;
            animation: spin 0.8s linear infinite;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <img src="/assets/newslettar_white.svg" alt="Newslettar" class="header-logo">
            <p class="version">v` + version + `</p>
        </div>

        <div class="tabs" role="tablist">
            <button class="tab active" role="tab" aria-selected="true" aria-controls="dashboard-tab" onclick="showTab('dashboard')">📊 Dashboard</button>
            <button class="tab" role="tab" aria-selected="false" aria-controls="config-tab" onclick="showTab('config')">⚙️ Configuration</button>
            <button class="tab" role="tab" aria-selected="false" aria-controls="template-tab" onclick="showTab('template')">📝 Email Template</button>
            <button class="tab" id="update-tab-button" role="tab" aria-selected="false" aria-controls="update-tab" onclick="showTab('update')" style="display: none;">🔄 Update</button>
        </div>

        <div id="dashboard-tab" class="tab-content active" role="tabpanel">
            <h3 style="margin-bottom: 20px; color: #667eea;">System Overview</h3>

            <div class="dashboard-grid">
                <div class="dashboard-card">
                    <div class="dashboard-card-header">
                        <span class="dashboard-icon">🖥️</span>
                        <h4>System Stats</h4>
                    </div>
                    <div class="dashboard-card-content">
                        <div class="stat-row">
                            <span class="stat-label">Version:</span>
                            <span class="stat-value" id="dash-version">` + version + `</span>
                        </div>
                        <div class="stat-row">
                            <span class="stat-label">Running on Port:</span>
                            <span class="stat-value" id="dash-port">Loading...</span>
                        </div>
                        <div class="stat-row">
                            <span class="stat-label">Uptime:</span>
                            <span class="stat-value" id="dash-uptime">Loading...</span>
                        </div>
                        <div class="stat-row">
                            <span class="stat-label">Memory Usage:</span>
                            <span class="stat-value" id="dash-memory">~12 MB</span>
                        </div>
                    </div>
                </div>

                <div class="dashboard-card">
                    <div class="dashboard-card-header">
                        <span class="dashboard-icon">📧</span>
                        <h4>Newsletter Stats</h4>
                    </div>
                    <div class="dashboard-card-content">
                        <div class="stat-row">
                            <span class="stat-label">Total Emails Sent:</span>
                            <span class="stat-value stat-highlight" id="dash-emails-sent">0</span>
                        </div>
                        <div class="stat-row">
                            <span class="stat-label">Last Sent:</span>
                            <span class="stat-value" id="dash-last-sent">Never</span>
                        </div>
                        <div class="stat-row">
                            <span class="stat-label">Next Scheduled:</span>
                            <span class="stat-value" style="text-align: right; line-height: 1.4;" id="dash-next-run">` + nextRun + `</span>
                        </div>
                        <div class="stat-row">
                            <span class="stat-label">Timezone:</span>
                            <span class="stat-value" id="dash-timezone">` + timezone + `</span>
                        </div>
                    </div>
                </div>

                <div class="dashboard-card">
                    <div class="dashboard-card-header">
                        <span class="dashboard-icon">🔌</span>
                        <h4>Service Status</h4>
                    </div>
                    <div class="dashboard-card-content">
                        <div class="stat-row">
                            <span class="stat-label">Sonarr:</span>
                            <span class="stat-value"><span id="status-sonarr" class="status-indicator">⚫</span> <span id="status-sonarr-text">Checking...</span></span>
                        </div>
                        <div class="stat-row">
                            <span class="stat-label">Radarr:</span>
                            <span class="stat-value"><span id="status-radarr" class="status-indicator">⚫</span> <span id="status-radarr-text">Checking...</span></span>
                        </div>
                        <div class="stat-row">
                            <span class="stat-label">Email:</span>
                            <span class="stat-value"><span id="status-email" class="status-indicator">⚫</span> <span id="status-email-text">Checking...</span></span>
                        </div>
                        <div class="stat-row">
                            <span class="stat-label">Trakt:</span>
                            <span class="stat-value"><span id="status-trakt" class="status-indicator">⚫</span> <span id="status-trakt-text">Checking...</span></span>
                        </div>
                    </div>
                </div>

                <div class="dashboard-card dashboard-logs">
                    <div class="dashboard-card-header">
                        <span class="dashboard-icon">📋</span>
                        <h4>Recent Logs</h4>
                    </div>
                    <div class="dashboard-card-content">
                        <div id="dashboard-logs" class="dashboard-logs-container">
                            Loading logs...
                        </div>
                    </div>
                </div>
            </div>

            <div class="action-buttons" style="margin-top: 30px;">
                <button class="btn" onclick="previewNewsletter()" aria-label="Generate newsletter preview">
                    <span>👁️ Preview Newsletter</span>
                </button>
                <button class="btn" onclick="sendNow()" aria-label="Send newsletter immediately">
                    <span>📤 Send Now</span>
                </button>
                <button class="btn btn-secondary" onclick="showTab('config')" aria-label="Go to configuration">
                    <span>⚙️ Configuration</span>
                </button>
            </div>
        </div>

        <div id="config-tab" class="tab-content" role="tabpanel">
            <form id="config-form">
                <h3 style="margin-bottom: 15px; color: #667eea;">Schedule Settings</h3>
                
                <div class="form-group">
                    <label for="timezone">Timezone</label>
                    <select name="timezone" id="timezone" aria-label="Select timezone">
                        <option value="UTC">UTC (GMT+0)</option>
                        <option value="America/New_York">Eastern Time (GMT-5/-4)</option>
                        <option value="America/Chicago">Central Time (GMT-6/-5)</option>
                        <option value="America/Denver">Mountain Time (GMT-7/-6)</option>
                        <option value="America/Los_Angeles">Pacific Time (GMT-8/-7)</option>
                        <option value="America/Toronto">Toronto (GMT-5/-4)</option>
                        <option value="America/Vancouver">Vancouver (GMT-8/-7)</option>
                        <option value="America/Montreal">Montreal (GMT-5/-4)</option>
                        <option value="Europe/London">London (GMT+0/+1)</option>
                        <option value="Europe/Paris">Paris (GMT+1/+2)</option>
                        <option value="Europe/Berlin">Berlin (GMT+1/+2)</option>
                        <option value="Asia/Tokyo">Tokyo (GMT+9)</option>
                        <option value="Asia/Shanghai">Shanghai (GMT+8)</option>
                        <option value="Australia/Sydney">Sydney (GMT+10/+11)</option>
                    </select>
                    <div class="timezone-info" id="timezone-info"></div>
                </div>

                <div class="form-group">
                    <label for="schedule_day">Day of Week</label>
                    <select name="schedule_day" id="schedule_day" aria-label="Select day of week">
                        <option value="Sun">Sunday</option>
                        <option value="Mon">Monday</option>
                        <option value="Tue">Tuesday</option>
                        <option value="Wed">Wednesday</option>
                        <option value="Thu">Thursday</option>
                        <option value="Fri">Friday</option>
                        <option value="Sat">Saturday</option>
                    </select>
                </div>

                <div class="form-group">
                    <label for="schedule_time">Time (24-hour format, HH:MM)</label>
                    <input type="time" name="schedule_time" id="schedule_time" required aria-label="Select time">
                    <div class="error-message" id="time-error">Please enter a valid time (HH:MM)</div>
                </div>

                <hr style="margin: 30px 0; border: none; border-top: 2px solid #2a3444;">

                <h3 style="margin-bottom: 15px; color: #667eea;">Sonarr Settings</h3>
                <div class="form-group">
                    <label for="sonarr_url">Sonarr URL</label>
                    <input type="url" name="sonarr_url" id="sonarr_url" placeholder="http://localhost:8989" aria-label="Sonarr URL">
                    <div class="error-message" id="sonarr-url-error">Please enter a valid URL</div>
                </div>
                <div class="form-group">
                    <label for="sonarr_api_key">Sonarr API Key</label>
                    <input type="text" name="sonarr_api_key" id="sonarr_api_key" placeholder="Your Sonarr API key" aria-label="Sonarr API Key">
                </div>
                <button type="button" class="btn btn-secondary" onclick="testConnection('sonarr')" aria-label="Test Sonarr connection">
                    <span>Test Sonarr</span>
                </button>

                <hr style="margin: 30px 0; border: none; border-top: 2px solid #2a3444;">

                <h3 style="margin-bottom: 15px; color: #667eea;">Radarr Settings</h3>
                <div class="form-group">
                    <label for="radarr_url">Radarr URL</label>
                    <input type="url" name="radarr_url" id="radarr_url" placeholder="http://localhost:7878" aria-label="Radarr URL">
                    <div class="error-message" id="radarr-url-error">Please enter a valid URL</div>
                </div>
                <div class="form-group">
                    <label for="radarr_api_key">Radarr API Key</label>
                    <input type="text" name="radarr_api_key" id="radarr_api_key" placeholder="Your Radarr API key" aria-label="Radarr API Key">
                </div>
                <button type="button" class="btn btn-secondary" onclick="testConnection('radarr')" aria-label="Test Radarr connection">
                    <span>Test Radarr</span>
                </button>

                <hr style="margin: 30px 0; border: none; border-top: 2px solid #2a3444;">

                <h3 style="margin-bottom: 15px; color: #667eea;">Trakt Settings (Optional)</h3>
                <div class="info-banner" style="margin-bottom: 20px;">
                    <p style="font-size: 0.9em;">
                        ℹ️ Trakt integration enables trending content sections in your newsletter.
                        <a href="https://trakt.tv/oauth/applications" target="_blank" style="color: #667eea; text-decoration: underline;">Create an app</a>
                        and copy your <strong>Client ID</strong> (Client Secret is not needed).
                    </p>
                </div>
                <div class="form-group">
                    <label for="trakt_client_id">Trakt Client ID</label>
                    <input type="text" name="trakt_client_id" id="trakt_client_id" placeholder="Your Trakt Client ID" aria-label="Trakt Client ID">
                </div>
                <button type="button" class="btn btn-secondary" onclick="testConnection('trakt')" aria-label="Test Trakt connection">
                    <span>Test Trakt</span>
                </button>

                <hr style="margin: 30px 0; border: none; border-top: 2px solid #2a3444;">

                <h3 style="margin-bottom: 15px; color: #667eea;">Email Settings</h3>

                <div class="email-section">
                    <h3>📧 Email Recipients</h3>
                    <p style="color: #8899aa; font-size: 0.9em; margin-bottom: 15px;">Add email addresses to receive the newsletter. Type an email and press comma or Enter to add it.</p>
                    <div class="form-group">
                        <label for="email-tag-input">Recipient Email Addresses</label>
                        <div id="email-tags-container" class="email-tags-container" onclick="document.getElementById('email-tag-input').focus()">
                            <input type="text" id="email-tag-input" class="email-tag-input" placeholder="Add email address..." aria-label="Add recipient email">
                        </div>
                        <input type="hidden" name="to_emails" id="to_emails">
                        <div class="error-message" id="to-emails-error">Please enter a valid email address</div>
                    </div>
                </div>

                <div class="form-group">
                    <label for="smtp_host">SMTP Server</label>
                    <input type="text" name="smtp_host" id="smtp_host" placeholder="smtp.mailgun.org" aria-label="SMTP Server">
                </div>
                <div class="form-group">
                    <label for="smtp_port">SMTP Port</label>
                    <input type="number" name="smtp_port" id="smtp_port" placeholder="587" aria-label="SMTP Port">
                </div>
                <div class="form-group">
                    <label for="smtp_user">SMTP Username</label>
                    <input type="text" name="smtp_user" id="smtp_user" placeholder="postmaster@yourdomain.com" aria-label="SMTP Username">
                </div>
                <div class="form-group">
                    <label for="smtp_pass">SMTP Password</label>
                    <input type="password" name="smtp_pass" id="smtp_pass" placeholder="Your SMTP password" aria-label="SMTP Password">
                </div>
                <div class="form-group">
                    <label for="from_name">From Name</label>
                    <input type="text" name="from_name" id="from_name" placeholder="Newslettar" aria-label="From Name">
                </div>
                <div class="form-group">
                    <label for="from_email">From Email</label>
                    <input type="email" name="from_email" id="from_email" placeholder="newsletter@yourdomain.com" aria-label="From Email">
                    <div class="error-message" id="from-email-error">Please enter a valid email address</div>
                </div>
                <button type="button" class="btn btn-secondary" onclick="testConnection('email')" aria-label="Test email authentication">
                    <span>Test Email Auth</span>
                </button>

                <hr style="margin: 30px 0; border: none; border-top: 2px solid #2a3444;">

                <button type="submit" class="btn" aria-label="Save configuration">
                    <span>💾 Save Configuration</span>
                </button>
            </form>
        </div>

        <div id="template-tab" class="tab-content" role="tabpanel">
            <h3 style="margin-bottom: 20px;">Email Template Options</h3>

            <h3 style="margin-bottom: 20px;">Customize Email Text</h3>

            <div class="info-banner" style="margin-bottom: 20px;">
                <p style="font-size: 0.9em;">
                    ℹ️ Customize all static text in your newsletter, including headings, messages, and the footer. Perfect for translating your newsletter to other languages or personalizing the content.
                </p>
            </div>

            <button class="btn" onclick="openEditStringsModal()" aria-label="Edit email strings">
                <span>✏️ Edit Email Strings</span>
            </button>

            <hr style="margin: 30px 0; border: none; border-top: 2px solid #2a3444;">

            <h3 style="margin-bottom: 15px;">Display Options</h3>

            <div class="template-option">
                <div>
                    <strong>Show Movie/Series Posters</strong>
                    <p style="font-size: 0.9em; color: #8899aa; margin-top: 5px;">
                        Display poster images in the newsletter
                    </p>
                </div>
                <label class="toggle-switch">
                    <input type="checkbox" id="show-posters" onchange="saveTemplateSettings()" aria-label="Toggle poster display">
                    <span class="toggle-slider"></span>
                </label>
            </div>

            <div class="template-option">
                <div>
                    <strong>Show Downloaded Section</strong>
                    <p style="font-size: 0.9em; color: #8899aa; margin-top: 5px;">
                        Include "Downloaded Last Week" section
                    </p>
                </div>
                <label class="toggle-switch">
                    <input type="checkbox" id="show-downloaded" onchange="saveTemplateSettings()" aria-label="Toggle downloaded section">
                    <span class="toggle-slider"></span>
                </label>
            </div>

            <div class="template-option">
                <div>
                    <strong>Show Series Descriptions</strong>
                    <p style="font-size: 0.9em; color: #8899aa; margin-top: 5px;">
                        Display short description for each TV series
                    </p>
                </div>
                <label class="toggle-switch">
                    <input type="checkbox" id="show-series-overview" onchange="saveTemplateSettings()" aria-label="Toggle series descriptions">
                    <span class="toggle-slider"></span>
                </label>
            </div>

            <div class="template-option">
                <div>
                    <strong>Show Episode Descriptions</strong>
                    <p style="font-size: 0.9em; color: #8899aa; margin-top: 5px;">
                        Display description for each episode
                    </p>
                </div>
                <label class="toggle-switch">
                    <input type="checkbox" id="show-episode-overview" onchange="saveTemplateSettings()" aria-label="Toggle episode descriptions">
                    <span class="toggle-slider"></span>
                </label>
            </div>

            <div class="template-option">
                <div>
                    <strong>Include Unmonitored Items</strong>
                    <p style="font-size: 0.9em; color: #8899aa; margin-top: 5px;">
                        Include unmonitored series and movies from Sonarr/Radarr. When disabled, only shows monitored items that will be downloaded automatically.
                    </p>
                </div>
                <label class="toggle-switch">
                    <input type="checkbox" id="show-unmonitored" onchange="saveTemplateSettings()" aria-label="Toggle unmonitored items">
                    <span class="toggle-slider"></span>
                </label>
            </div>

            <div class="template-option">
                <div>
                    <strong>Show Series Ratings</strong>
                    <p style="font-size: 0.9em; color: #8899aa; margin-top: 5px;">
                        Display series ratings in series headers from Sonarr/Radarr
                    </p>
                </div>
                <label class="toggle-switch">
                    <input type="checkbox" id="show-series-ratings" onchange="saveTemplateSettings()" aria-label="Toggle series ratings display">
                    <span class="toggle-slider"></span>
                </label>
            </div>

            <div class="template-option">
                <div>
                    <strong>Dark Mode</strong>
                    <p style="font-size: 0.9em; color: #8899aa; margin-top: 5px;">
                        Use dark theme for email newsletters (recommended). When disabled, uses traditional light theme with white background.
                    </p>
                </div>
                <label class="toggle-switch">
                    <input type="checkbox" id="dark-mode" onchange="saveTemplateSettings()" aria-label="Toggle dark mode">
                    <span class="toggle-slider"></span>
                </label>
            </div>

            <hr style="margin: 30px 0; border: none; border-top: 2px solid #2a3444;">

            <h3 style="margin-bottom: 15px;">Trakt Trending Sections</h3>
            <div class="info-banner" style="margin-bottom: 20px;">
                <p style="font-size: 0.9em;">
                    ℹ️ Requires Trakt Client ID in Configuration tab. Toggles work only when Client ID is configured.
                </p>
            </div>

            <div class="template-option">
                <div style="flex: 1;">
                    <div style="display: flex; justify-content: space-between; align-items: center;">
                        <div>
                            <strong>Show Most Anticipated Series</strong>
                            <p style="font-size: 0.9em; color: #8899aa; margin-top: 5px;">
                                Display trending series people are most excited about
                            </p>
                        </div>
                        <label class="toggle-switch">
                            <input type="checkbox" id="show-trakt-anticipated-series" onchange="toggleTraktLimit('anticipated-series')" aria-label="Toggle Trakt anticipated series">
                            <span class="toggle-slider"></span>
                        </label>
                    </div>
                    <div id="trakt-anticipated-series-limit-container" style="display: none; margin-top: 10px; padding: 10px; background: #1a2332; border-radius: 6px; border-left: 3px solid #667eea;">
                        <label for="trakt-anticipated-series-limit" style="font-size: 0.85em; color: #a0b0c0; display: block; margin-bottom: 6px;">Number of results (1-20, default: 5)</label>
                        <input type="number" id="trakt-anticipated-series-limit" min="1" max="20" placeholder="5" style="width: 80px; padding: 6px 10px; background: #0f1419; border: 2px solid #2a3444; border-radius: 6px; color: #e8e8e8; font-size: 14px;" onchange="saveTemplateSettings()">
                    </div>
                </div>
            </div>

            <div class="template-option">
                <div style="flex: 1;">
                    <div style="display: flex; justify-content: space-between; align-items: center;">
                        <div>
                            <strong>Show Most Watched Series</strong>
                            <p style="font-size: 0.9em; color: #8899aa; margin-top: 5px;">
                                Display most watched series from the last week
                            </p>
                        </div>
                        <label class="toggle-switch">
                            <input type="checkbox" id="show-trakt-watched-series" onchange="toggleTraktLimit('watched-series')" aria-label="Toggle Trakt watched series">
                            <span class="toggle-slider"></span>
                        </label>
                    </div>
                    <div id="trakt-watched-series-limit-container" style="display: none; margin-top: 10px; padding: 10px; background: #1a2332; border-radius: 6px; border-left: 3px solid #667eea;">
                        <label for="trakt-watched-series-limit" style="font-size: 0.85em; color: #a0b0c0; display: block; margin-bottom: 6px;">Number of results (1-20, default: 5)</label>
                        <input type="number" id="trakt-watched-series-limit" min="1" max="20" placeholder="5" style="width: 80px; padding: 6px 10px; background: #0f1419; border: 2px solid #2a3444; border-radius: 6px; color: #e8e8e8; font-size: 14px;" onchange="saveTemplateSettings()">
                    </div>
                </div>
            </div>

            <div class="template-option">
                <div style="flex: 1;">
                    <div style="display: flex; justify-content: space-between; align-items: center;">
                        <div>
                            <strong>Show Most Anticipated Movies</strong>
                            <p style="font-size: 0.9em; color: #8899aa; margin-top: 5px;">
                                Display upcoming movies generating the most buzz
                            </p>
                        </div>
                        <label class="toggle-switch">
                            <input type="checkbox" id="show-trakt-anticipated-movies" onchange="toggleTraktLimit('anticipated-movies')" aria-label="Toggle Trakt anticipated movies">
                            <span class="toggle-slider"></span>
                        </label>
                    </div>
                    <div id="trakt-anticipated-movies-limit-container" style="display: none; margin-top: 10px; padding: 10px; background: #1a2332; border-radius: 6px; border-left: 3px solid #667eea;">
                        <label for="trakt-anticipated-movies-limit" style="font-size: 0.85em; color: #a0b0c0; display: block; margin-bottom: 6px;">Number of results (1-20, default: 5)</label>
                        <input type="number" id="trakt-anticipated-movies-limit" min="1" max="20" placeholder="5" style="width: 80px; padding: 6px 10px; background: #0f1419; border: 2px solid #2a3444; border-radius: 6px; color: #e8e8e8; font-size: 14px;" onchange="saveTemplateSettings()">
                    </div>
                </div>
            </div>

            <div class="template-option">
                <div style="flex: 1;">
                    <div style="display: flex; justify-content: space-between; align-items: center;">
                        <div>
                            <strong>Show Most Watched Movies</strong>
                            <p style="font-size: 0.9em; color: #8899aa; margin-top: 5px;">
                                Display most watched movies from the last week
                            </p>
                        </div>
                        <label class="toggle-switch">
                            <input type="checkbox" id="show-trakt-watched-movies" onchange="toggleTraktLimit('watched-movies')" aria-label="Toggle Trakt watched movies">
                            <span class="toggle-slider"></span>
                        </label>
                    </div>
                    <div id="trakt-watched-movies-limit-container" style="display: none; margin-top: 10px; padding: 10px; background: #1a2332; border-radius: 6px; border-left: 3px solid #667eea;">
                        <label for="trakt-watched-movies-limit" style="font-size: 0.85em; color: #a0b0c0; display: block; margin-bottom: 6px;">Number of results (1-20, default: 5)</label>
                        <input type="number" id="trakt-watched-movies-limit" min="1" max="20" placeholder="5" style="width: 80px; padding: 6px 10px; background: #0f1419; border: 2px solid #2a3444; border-radius: 6px; color: #e8e8e8; font-size: 14px;" onchange="saveTemplateSettings()">
                    </div>
                </div>
            </div>

            <p style="margin-top: 20px; color: #8899aa; font-size: 0.9em;">
                ℹ️ Changes are saved automatically when you toggle switches.
            </p>

            <hr style="margin: 30px 0; border: none; border-top: 2px solid #2a3444;">

            <h3 style="margin-bottom: 20px;">Actions</h3>

            <div class="action-buttons">
                <button class="btn btn-secondary" onclick="previewNewsletter()" aria-label="Preview newsletter">
                    <span>👁️ Preview Newsletter</span>
                </button>
                <button class="btn btn-success" onclick="sendNow()" aria-label="Send newsletter now">
                    <span>📧 Send Newsletter Now</span>
                </button>
            </div>

            <p style="margin-top: 15px; color: #8899aa; font-size: 0.9em;">
                Preview generates the email based on current settings without sending. Send Now will generate and send immediately.
            </p>
        </div>

        <div id="update-tab" class="tab-content" role="tabpanel">
            <h3 style="margin-bottom: 20px;">🔄 Update Newslettar</h3>
            
            <div id="version-info" aria-live="polite">
                <p>Checking for updates...</p>
            </div>

            <button class="btn" onclick="checkUpdates()" style="margin-right: 10px;" aria-label="Check for updates">
                <span>🔍 Check for Updates</span>
            </button>
            <button class="btn btn-success" id="update-btn" onclick="performUpdate()" style="display: none;" aria-label="Update now">
                <span>⬇️ Update Now</span>
            </button>
        </div>
    </div>

    <!-- Preview Modal -->
    <div id="preview-modal" class="modal" role="dialog" aria-labelledby="preview-title" aria-modal="true">
        <div class="modal-content">
            <div class="modal-header">
                <h2 id="preview-title">Email Preview</h2>
                <button class="modal-close" onclick="closePreview()" aria-label="Close preview">&times;</button>
            </div>
            <div class="modal-body">
                <iframe id="preview-frame" title="Email preview"></iframe>
            </div>
        </div>
    </div>

    <!-- Edit Strings Modal -->
    <div id="edit-strings-modal" class="modal" role="dialog" aria-labelledby="edit-strings-title" aria-modal="true">
        <div class="modal-content">
            <div class="modal-header">
                <h2 id="edit-strings-title">✏️ Edit Email Strings</h2>
                <button class="modal-close" onclick="closeEditStringsModal()" aria-label="Close edit strings">&times;</button>
            </div>
            <div class="modal-body" style="max-height: calc(90vh - 140px); overflow-y: auto;">
                <div style="margin-bottom: 20px; background: #252f3f; padding: 15px; border-radius: 8px; border-left: 3px solid #11998e;">
                    <p style="color: #a0b0c0; font-size: 0.9em; margin: 0;">
                        💡 <strong style="color: #e8e8e8;">Tip:</strong> Leave fields empty to hide that section from your emails. Perfect for customization and translation!
                    </p>
                </div>

                <form id="edit-strings-form">
                    <div class="form-group">
                        <label for="email-title">Email Title</label>
                        <input type="text" id="email-title" name="email_title" placeholder="e.g., 📺 Your Weekly Newslettar">
                    </div>

                    <div class="form-group">
                        <label for="email-intro">Email Introduction (optional)</label>
                        <textarea id="email-intro" name="email_intro" rows="3" style="width: 100%; padding: 12px; background: #0f1419; border: 2px solid #2a3444; border-radius: 8px; color: #e8e8e8; font-size: 14px; font-family: inherit; resize: vertical;" placeholder="A brief introduction paragraph shown under the title (leave empty to hide)"></textarea>
                        <small style="color: #8899aa; font-size: 0.85em; display: block; margin-top: 5px;">This paragraph appears below the email title. Leave empty if you don't want an introduction.</small>
                    </div>

                    <hr style="margin: 25px 0; border: none; border-top: 1px solid #2a3444;">
                    <h4 style="color: #667eea; margin-bottom: 15px;">Section Headings</h4>

                    <div class="form-group">
                        <label for="week-range-prefix">Week Range Prefix</label>
                        <input type="text" id="week-range-prefix" name="week_range_prefix" placeholder="e.g., Week of">
                    </div>

                    <div class="form-group">
                        <label for="coming-this-week-heading">Coming This Week Heading</label>
                        <input type="text" id="coming-this-week-heading" name="coming_this_week_heading" placeholder="e.g., 📅 Coming This Week">
                    </div>

                    <div class="form-group">
                        <label for="tv-shows-heading">TV Shows Heading</label>
                        <input type="text" id="tv-shows-heading" name="tv_shows_heading" placeholder="e.g., TV Shows">
                    </div>

                    <div class="form-group">
                        <label for="movies-heading">Movies Heading</label>
                        <input type="text" id="movies-heading" name="movies_heading" placeholder="e.g., Movies">
                    </div>

                    <div class="form-group">
                        <label for="downloaded-section-heading">Downloaded Section Heading</label>
                        <input type="text" id="downloaded-section-heading" name="downloaded_section_heading" placeholder="e.g., 📥 Downloaded Last Week">
                    </div>

                    <div class="form-group">
                        <label for="trending-section-heading">Trending Section Heading</label>
                        <input type="text" id="trending-section-heading" name="trending_section_heading" placeholder="e.g., 🔥 Trending">
                    </div>

                    <hr style="margin: 25px 0; border: none; border-top: 1px solid #2a3444;">
                    <h4 style="color: #667eea; margin-bottom: 15px;">Empty State Messages</h4>

                    <div class="form-group">
                        <label for="no-shows-message">No Shows Message</label>
                        <input type="text" id="no-shows-message" name="no_shows_message" placeholder="e.g., No shows scheduled for this week">
                    </div>

                    <div class="form-group">
                        <label for="no-movies-message">No Movies Message</label>
                        <input type="text" id="no-movies-message" name="no_movies_message" placeholder="e.g., No movies scheduled for this week">
                    </div>

                    <div class="form-group">
                        <label for="no-downloaded-shows-message">No Downloaded Shows Message</label>
                        <input type="text" id="no-downloaded-shows-message" name="no_downloaded_shows_message" placeholder="e.g., No shows downloaded this week">
                    </div>

                    <div class="form-group">
                        <label for="no-downloaded-movies-message">No Downloaded Movies Message</label>
                        <input type="text" id="no-downloaded-movies-message" name="no_downloaded_movies_message" placeholder="e.g., No movies downloaded this week">
                    </div>

                    <hr style="margin: 25px 0; border: none; border-top: 1px solid #2a3444;">
                    <h4 style="color: #667eea; margin-bottom: 15px;">Trakt Trending Headings</h4>

                    <div class="form-group">
                        <label for="anticipated-series-heading">Anticipated Series Heading</label>
                        <input type="text" id="anticipated-series-heading" name="anticipated_series_heading" placeholder="e.g., Most Anticipated Series (Next Week)">
                    </div>

                    <div class="form-group">
                        <label for="watched-series-heading">Watched Series Heading</label>
                        <input type="text" id="watched-series-heading" name="watched_series_heading" placeholder="e.g., Most Watched Series (Last Week)">
                    </div>

                    <div class="form-group">
                        <label for="anticipated-movies-heading">Anticipated Movies Heading</label>
                        <input type="text" id="anticipated-movies-heading" name="anticipated_movies_heading" placeholder="e.g., Most Anticipated Movies (Next Week)">
                    </div>

                    <div class="form-group">
                        <label for="watched-movies-heading">Watched Movies Heading</label>
                        <input type="text" id="watched-movies-heading" name="watched_movies_heading" placeholder="e.g., Most Watched Movies (Last Week)">
                    </div>

                    <hr style="margin: 25px 0; border: none; border-top: 1px solid #2a3444;">
                    <h4 style="color: #667eea; margin-bottom: 15px;">Footer</h4>

                    <div class="form-group">
                        <label for="footer-text">Footer Text</label>
                        <input type="text" id="footer-text" name="footer_text" placeholder="e.g., Generated by Newslettar">
                    </div>

                    <div style="display: flex; gap: 10px; margin-top: 30px;">
                        <button type="button" class="btn btn-success" onclick="saveEmailStrings()" style="flex: 1;" aria-label="Save email strings">
                            <span>💾 Save Changes</span>
                        </button>
                        <button type="button" class="btn btn-secondary" onclick="resetEmailStrings()" style="flex: 1;" aria-label="Reset to defaults">
                            <span>↩️ Reset to Defaults</span>
                        </button>
                        <button type="button" class="btn btn-secondary" onclick="closeEditStringsModal()" aria-label="Cancel">
                            <span>✖️ Cancel</span>
                        </button>
                    </div>
                </form>
            </div>
        </div>
    </div>

    <!-- Loading Overlay -->
    <div id="loading-overlay" class="loading-overlay" role="status" aria-live="polite">
        <div class="loading-spinner"></div>
    </div>

    <script>
        // Keyboard navigation and click-outside-to-close
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                closePreview();
                closeEditStringsModal();
            }
        });

        // Click outside modal to close
        document.getElementById('preview-modal').addEventListener('click', (e) => {
            if (e.target.id === 'preview-modal') {
                closePreview();
            }
        });

        document.getElementById('edit-strings-modal').addEventListener('click', (e) => {
            if (e.target.id === 'edit-strings-modal') {
                closeEditStringsModal();
            }
        });

        let dashboardInterval;
        let timezoneInterval;

        function showTab(tabName) {
            document.querySelectorAll('.tab').forEach(t => {
                t.classList.remove('active');
                t.setAttribute('aria-selected', 'false');
            });
            document.querySelectorAll('.tab-content').forEach(t => t.classList.remove('active'));

            // Find and activate the correct tab button by matching the tab name
            const tabs = document.querySelectorAll('.tab');
            tabs.forEach(t => {
                const onclick = t.getAttribute('onclick');
                if (onclick && onclick.includes("'" + tabName + "'")) {
                    t.classList.add('active');
                    t.setAttribute('aria-selected', 'true');
                }
            });
            document.getElementById(tabName + '-tab').classList.add('active');

            // Clear intervals
            if (dashboardInterval) {
                clearInterval(dashboardInterval);
            }
            if (timezoneInterval) {
                clearInterval(timezoneInterval);
            }

            // Handle tab-specific actions
            if (tabName === 'dashboard') {
                loadDashboard();
                dashboardInterval = setInterval(loadDashboard, 10000); // Update every 10 seconds
            } else if (tabName === 'config') {
                // Refresh timezone info every 60 seconds when config tab is active
                updateTimezoneInfo();
                timezoneInterval = setInterval(updateTimezoneInfo, 60000); // Update every 60 seconds
            }
        }

        async function loadDashboard() {
            try {
                const resp = await fetch('/api/dashboard');
                const data = await resp.json();

                // Update system stats
                document.getElementById('dash-version').textContent = data.version;
                document.getElementById('dash-port').textContent = data.port;
                document.getElementById('dash-uptime').textContent = data.uptime;
                document.getElementById('dash-memory').textContent = '~' + data.memory_usage_mb.toFixed(1) + ' MB';

                // Update newsletter stats
                document.getElementById('dash-emails-sent').textContent = data.total_emails_sent;
                document.getElementById('dash-last-sent').textContent = data.last_sent_date;
                document.getElementById('dash-next-run').textContent = data.next_scheduled_run;
                document.getElementById('dash-timezone').textContent = data.timezone;

                // Update service status
                updateServiceStatus('sonarr', data.service_status.sonarr);
                updateServiceStatus('radarr', data.service_status.radarr);
                updateServiceStatus('email', data.service_status.email);
                updateServiceStatus('trakt', data.service_status.trakt);

                // Update dashboard logs (last 20 lines)
                const logsResp = await fetch('/api/logs');
                const logs = await logsResp.text();
                const logLines = logs.trim().split('\n');
                const recentLogs = logLines.slice(-20).join('\n');
                document.getElementById('dashboard-logs').textContent = recentLogs || 'No logs available';
            } catch (error) {
                console.error('Failed to load dashboard:', error);
            }
        }

        function updateServiceStatus(service, status) {
            const indicator = document.getElementById('status-' + service);
            const text = document.getElementById('status-' + service + '-text');

            if (status === 'ok') {
                indicator.textContent = '🟢';
                text.textContent = 'Connected';
                text.style.color = '#10b981';
            } else if (status === 'configured') {
                indicator.textContent = '🟡';
                text.textContent = 'Configured';
                text.style.color = '#f59e0b';
            } else if (status === 'not_configured') {
                indicator.textContent = '⚪';
                text.textContent = 'Not Configured';
                text.style.color = '#8899aa';
            } else if (status === 'error') {
                indicator.textContent = '🔴';
                text.textContent = 'Error';
                text.style.color = '#ef4444';
            }
        }

        // Email tag management
        let emailTags = [];

        function addEmailTag(email) {
            if (!email) return;

            // Validate email
            if (!validateEmail(email)) {
                document.getElementById('to-emails-error').classList.add('show');
                document.getElementById('email-tags-container').classList.add('error');
                setTimeout(() => {
                    document.getElementById('to-emails-error').classList.remove('show');
                    document.getElementById('email-tags-container').classList.remove('error');
                }, 3000);
                return;
            }

            // Check for duplicates
            if (emailTags.includes(email)) {
                return;
            }

            // Add to array
            emailTags.push(email);

            // Create tag element
            const tag = document.createElement('div');
            tag.className = 'email-tag';
            tag.dataset.email = email;
            tag.innerHTML = '<span>' + email + '</span>' +
                '<button type="button" class="email-tag-remove" onclick="removeEmailTag(\'' + email + '\')" aria-label="Remove ' + email + '">&times;</button>';

            // Insert before input
            const input = document.getElementById('email-tag-input');
            input.parentElement.insertBefore(tag, input);

            // Update hidden field
            updateEmailsField();
        }

        function removeEmailTag(email) {
            // Remove from array
            emailTags = emailTags.filter(e => e !== email);

            // Remove element
            const tags = document.querySelectorAll('.email-tag');
            tags.forEach(tag => {
                if (tag.dataset.email === email) {
                    tag.remove();
                }
            });

            // Update hidden field
            updateEmailsField();
        }

        function updateEmailsField() {
            document.getElementById('to_emails').value = emailTags.join(', ');
        }

        function loadEmailTags(emailsString) {
            // Clear existing tags
            emailTags = [];
            const existingTags = document.querySelectorAll('.email-tag');
            existingTags.forEach(tag => tag.remove());

            // Add tags from string
            if (emailsString) {
                const emails = emailsString.split(',').map(e => e.trim()).filter(e => e);
                emails.forEach(email => addEmailTag(email));
            }
        }

        // Real-time validation
        function validateURL(input) {
            const value = input.value.trim();
            if (!value) return true;

            try {
                new URL(value);
                return true;
            } catch {
                return false;
            }
        }

        function validateEmail(email) {
            return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
        }

        function validateEmails(input) {
            const value = input.value.trim();
            if (!value) return true;

            const emails = value.split(',').map(e => e.trim());
            return emails.every(email => validateEmail(email));
        }

        // Add validation listeners
        document.addEventListener('DOMContentLoaded', () => {
            const sonarrUrl = document.getElementById('sonarr_url');
            const radarrUrl = document.getElementById('radarr_url');
            const fromEmail = document.getElementById('from_email');
            const toEmails = document.getElementById('to_emails');

            sonarrUrl.addEventListener('blur', function() {
                if (this.value && !validateURL(this)) {
                    this.classList.add('error');
                    this.classList.remove('success');
                    document.getElementById('sonarr-url-error').classList.add('show');
                } else if (this.value) {
                    this.classList.remove('error');
                    this.classList.add('success');
                    document.getElementById('sonarr-url-error').classList.remove('show');
                }
            });

            radarrUrl.addEventListener('blur', function() {
                if (this.value && !validateURL(this)) {
                    this.classList.add('error');
                    this.classList.remove('success');
                    document.getElementById('radarr-url-error').classList.add('show');
                } else if (this.value) {
                    this.classList.remove('error');
                    this.classList.add('success');
                    document.getElementById('radarr-url-error').classList.remove('show');
                }
            });

            fromEmail.addEventListener('blur', function() {
                if (this.value && !validateEmail(this.value)) {
                    this.classList.add('error');
                    this.classList.remove('success');
                    document.getElementById('from-email-error').classList.add('show');
                } else if (this.value) {
                    this.classList.remove('error');
                    this.classList.add('success');
                    document.getElementById('from-email-error').classList.remove('show');
                }
            });

            // Email tags functionality
            const emailTagInput = document.getElementById('email-tag-input');
            const emailTagsContainer = document.getElementById('email-tags-container');

            emailTagInput.addEventListener('keydown', (e) => {
                if (e.key === ',' || e.key === 'Enter') {
                    e.preventDefault();
                    addEmailTag(emailTagInput.value.trim());
                    emailTagInput.value = '';
                } else if (e.key === 'Backspace' && emailTagInput.value === '') {
                    // Remove last tag if backspace is pressed on empty input
                    const tags = emailTagsContainer.querySelectorAll('.email-tag');
                    if (tags.length > 0) {
                        removeEmailTag(tags[tags.length - 1].dataset.email);
                    }
                }
            });

            emailTagInput.addEventListener('blur', () => {
                // Add email if there's text when input loses focus
                if (emailTagInput.value.trim()) {
                    addEmailTag(emailTagInput.value.trim());
                    emailTagInput.value = '';
                }
            });

            // Update timezone info on change
            document.getElementById('timezone').addEventListener('change', updateTimezoneInfo);

            // Enable/disable Trakt toggles based on Client ID
            const traktClientId = document.getElementById('trakt_client_id');
            const traktToggles = [
                document.getElementById('show-trakt-anticipated-series'),
                document.getElementById('show-trakt-watched-series'),
                document.getElementById('show-trakt-anticipated-movies'),
                document.getElementById('show-trakt-watched-movies')
            ];

            function updateTraktToggles() {
                const hasClientId = traktClientId.value.trim() !== '';
                traktToggles.forEach(toggle => {
                    toggle.disabled = !hasClientId;
                    toggle.closest('.template-option').style.opacity = hasClientId ? '1' : '0.5';
                    toggle.closest('.template-option').style.pointerEvents = hasClientId ? 'auto' : 'none';
                });
            }

            // Check on page load and on input change
            traktClientId.addEventListener('input', updateTraktToggles);
            // Call once after config is loaded
            setTimeout(updateTraktToggles, 100);
        });

        async function updateTimezoneInfo() {
            const tz = document.getElementById('timezone').value;
            try {
                const resp = await fetch('/api/timezone-info?tz=' + encodeURIComponent(tz));
                const data = await resp.json();
                
                document.getElementById('timezone-info').innerHTML = 
                    '<strong>Current time:</strong> ' + data.current_time + 
                    ' <strong>•</strong> Offset: ' + data.offset;
            } catch (error) {
                console.error('Failed to fetch timezone info:', error);
            }
        }

        let originalTraktClientId = '';

        async function loadConfig() {
            showLoading();
            try {
                const resp = await fetch('/api/config');
                const data = await resp.json();

                document.querySelector('[name="sonarr_url"]').value = data.sonarr_url || '';
                document.querySelector('[name="sonarr_api_key"]').value = data.sonarr_api_key || '';
                document.querySelector('[name="radarr_url"]').value = data.radarr_url || '';
                document.querySelector('[name="radarr_api_key"]').value = data.radarr_api_key || '';
                document.querySelector('[name="trakt_client_id"]').value = data.trakt_client_id || '';
                originalTraktClientId = data.trakt_client_id || '';
                document.querySelector('[name="smtp_host"]').value = data.smtp_host || 'smtp.mailgun.org';
                document.querySelector('[name="smtp_port"]').value = data.smtp_port || '587';
                document.querySelector('[name="smtp_user"]').value = data.smtp_user || '';
                document.querySelector('[name="smtp_pass"]').value = data.smtp_pass || '';
                document.querySelector('[name="from_email"]').value = data.from_email || '';
                document.querySelector('[name="from_name"]').value = data.from_name || 'Newslettar';
                document.querySelector('[name="to_emails"]').value = data.to_emails || '';
                loadEmailTags(data.to_emails || '');
                document.querySelector('[name="timezone"]').value = data.timezone || 'UTC';
                document.querySelector('[name="schedule_day"]').value = data.schedule_day || 'Sun';
                document.querySelector('[name="schedule_time"]').value = data.schedule_time || '09:00';
                
                document.getElementById('show-posters').checked = data.show_posters !== 'false';
                document.getElementById('show-downloaded').checked = data.show_downloaded !== 'false';
                document.getElementById('show-series-overview').checked = data.show_series_overview !== 'false';
                document.getElementById('show-episode-overview').checked = data.show_episode_overview !== 'false';
                document.getElementById('show-unmonitored').checked = data.show_unmonitored !== 'false';
                document.getElementById('show-series-ratings').checked = data.show_series_ratings !== 'false';
                document.getElementById('dark-mode').checked = data.dark_mode !== 'false';
                document.getElementById('show-trakt-anticipated-series').checked = data.show_trakt_anticipated_series !== 'false';
                document.getElementById('show-trakt-watched-series').checked = data.show_trakt_watched_series !== 'false';
                document.getElementById('show-trakt-anticipated-movies').checked = data.show_trakt_anticipated_movies !== 'false';
                document.getElementById('show-trakt-watched-movies').checked = data.show_trakt_watched_movies !== 'false';

                document.getElementById('trakt-anticipated-series-limit').value = data.trakt_anticipated_series_limit || '5';
                document.getElementById('trakt-watched-series-limit').value = data.trakt_watched_series_limit || '5';
                document.getElementById('trakt-anticipated-movies-limit').value = data.trakt_anticipated_movies_limit || '5';
                document.getElementById('trakt-watched-movies-limit').value = data.trakt_watched_movies_limit || '5';

                // Show/hide limit containers based on toggle state
                document.getElementById('trakt-anticipated-series-limit-container').style.display =
                    data.show_trakt_anticipated_series !== 'false' ? 'block' : 'none';
                document.getElementById('trakt-watched-series-limit-container').style.display =
                    data.show_trakt_watched_series !== 'false' ? 'block' : 'none';
                document.getElementById('trakt-anticipated-movies-limit-container').style.display =
                    data.show_trakt_anticipated_movies !== 'false' ? 'block' : 'none';
                document.getElementById('trakt-watched-movies-limit-container').style.display =
                    data.show_trakt_watched_movies !== 'false' ? 'block' : 'none';

                await updateTimezoneInfo();
            } catch (error) {
                showNotification('Failed to load configuration: ' + error.message, 'error');
            } finally {
                hideLoading();
            }
        }

        document.getElementById('config-form').addEventListener('submit', async (e) => {
            e.preventDefault();

            // Ensure the to_emails field is updated with current tags before submitting
            updateEmailsField();

            const formData = new FormData(e.target);
            const data = Object.fromEntries(formData);

            // Explicitly ensure to_emails is included even if empty
            data.to_emails = emailTags.join(', ');

            const submitBtn = e.target.querySelector('button[type="submit"]');
            submitBtn.classList.add('loading');
            submitBtn.disabled = true;

            try {
                // Validate Trakt Client ID if it changed and is not empty
                const currentTraktClientId = data.trakt_client_id || '';
                if (currentTraktClientId !== originalTraktClientId && currentTraktClientId !== '') {
                    showNotification('Validating Trakt Client ID...', 'info');

                    const testResp = await fetch('/api/test-trakt', {
                        method: 'POST',
                        headers: {'Content-Type': 'application/json'},
                        body: JSON.stringify({ client_id: currentTraktClientId })
                    });

                    const testResult = await testResp.json();
                    if (!testResult.success) {
                        showNotification('Invalid Trakt Client ID: ' + testResult.message, 'error');
                        submitBtn.classList.remove('loading');
                        submitBtn.disabled = false;
                        return;
                    }
                }

                const resp = await fetch('/api/config', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify(data)
                });

                if (resp.ok) {
                    showNotification('Configuration saved successfully!', 'success');
                    originalTraktClientId = currentTraktClientId;
                    setTimeout(() => location.reload(), 2000);
                } else {
                    showNotification('Failed to save configuration', 'error');
                }
            } catch (error) {
                showNotification('Network error: ' + error.message, 'error');
            } finally {
                submitBtn.classList.remove('loading');
                submitBtn.disabled = false;
            }
        });

        async function testConnection(type) {
            const form = document.getElementById('config-form');
            const formData = new FormData(form);
            const data = Object.fromEntries(formData);
            
            const button = event.target.closest('button');
            button.classList.add('loading');
            button.disabled = true;

            let endpoint, payload;

            if (type === 'sonarr') {
                endpoint = '/api/test-sonarr';
                payload = { url: data.sonarr_url, api_key: data.sonarr_api_key };
            } else if (type === 'radarr') {
                endpoint = '/api/test-radarr';
                payload = { url: data.radarr_url, api_key: data.radarr_api_key };
            } else if (type === 'trakt') {
                endpoint = '/api/test-trakt';
                payload = { client_id: data.trakt_client_id };
            } else {
                endpoint = '/api/test-email';
                payload = {
                    smtp: data.smtp_host,
                    port: data.smtp_port,
                    user: data.smtp_user,
                    pass: data.smtp_pass
                };
            }

            try {
                const resp = await fetch(endpoint, {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify(payload)
                });

                const result = await resp.json();
                showNotification(result.message, result.success ? 'success' : 'error');
            } catch (error) {
                showNotification('Connection test failed: ' + error.message, 'error');
            } finally {
                button.classList.remove('loading');
                button.disabled = false;
            }
        }

        async function previewNewsletter() {
            const button = event.target.closest('button');
            button.classList.add('loading');
            button.disabled = true;
            
            showLoading();

            try {
                const resp = await fetch('/api/preview', { method: 'POST' });
                const data = await resp.json();

                if (data.success) {
                    const iframe = document.getElementById('preview-frame');
                    iframe.srcdoc = data.html;
                    document.getElementById('preview-modal').classList.add('show');
                } else {
                    showNotification(data.error || 'Failed to generate preview', 'error');
                }
            } catch (error) {
                showNotification('Preview failed: ' + error.message, 'error');
            } finally {
                button.classList.remove('loading');
                button.disabled = false;
                hideLoading();
            }
        }

        function closePreview() {
            document.getElementById('preview-modal').classList.remove('show');
        }

        // Edit Strings Modal Functions
        async function openEditStringsModal() {
            try {
                // Fetch current config
                const resp = await fetch('/api/config');
                const config = await resp.json();

                // Populate form fields
                document.getElementById('email-title').value = config.email_title || '';
                document.getElementById('email-intro').value = config.email_intro || '';
                document.getElementById('week-range-prefix').value = config.week_range_prefix || '';
                document.getElementById('coming-this-week-heading').value = config.coming_this_week_heading || '';
                document.getElementById('tv-shows-heading').value = config.tv_shows_heading || '';
                document.getElementById('movies-heading').value = config.movies_heading || '';
                document.getElementById('no-shows-message').value = config.no_shows_message || '';
                document.getElementById('no-movies-message').value = config.no_movies_message || '';
                document.getElementById('downloaded-section-heading').value = config.downloaded_section_heading || '';
                document.getElementById('no-downloaded-shows-message').value = config.no_downloaded_shows_message || '';
                document.getElementById('no-downloaded-movies-message').value = config.no_downloaded_movies_message || '';
                document.getElementById('trending-section-heading').value = config.trending_section_heading || '';
                document.getElementById('anticipated-series-heading').value = config.anticipated_series_heading || '';
                document.getElementById('watched-series-heading').value = config.watched_series_heading || '';
                document.getElementById('anticipated-movies-heading').value = config.anticipated_movies_heading || '';
                document.getElementById('watched-movies-heading').value = config.watched_movies_heading || '';
                document.getElementById('footer-text').value = config.footer_text || '';

                // Show modal
                document.getElementById('edit-strings-modal').classList.add('show');
            } catch (error) {
                console.error('Failed to load config:', error);
                showNotification('Failed to load current configuration', 'error');
            }
        }

        function closeEditStringsModal() {
            document.getElementById('edit-strings-modal').classList.remove('show');
        }

        async function saveEmailStrings() {
            const button = event.target.closest('button');
            button.classList.add('loading');
            button.disabled = true;

            try {
                // Get current config first
                const respGet = await fetch('/api/config');
                const currentConfig = await respGet.json();

                // Update with email string values
                const updatedConfig = {
                    ...currentConfig,
                    email_title: document.getElementById('email-title').value,
                    email_intro: document.getElementById('email-intro').value,
                    week_range_prefix: document.getElementById('week-range-prefix').value,
                    coming_this_week_heading: document.getElementById('coming-this-week-heading').value,
                    tv_shows_heading: document.getElementById('tv-shows-heading').value,
                    movies_heading: document.getElementById('movies-heading').value,
                    no_shows_message: document.getElementById('no-shows-message').value,
                    no_movies_message: document.getElementById('no-movies-message').value,
                    downloaded_section_heading: document.getElementById('downloaded-section-heading').value,
                    no_downloaded_shows_message: document.getElementById('no-downloaded-shows-message').value,
                    no_downloaded_movies_message: document.getElementById('no-downloaded-movies-message').value,
                    trending_section_heading: document.getElementById('trending-section-heading').value,
                    anticipated_series_heading: document.getElementById('anticipated-series-heading').value,
                    watched_series_heading: document.getElementById('watched-series-heading').value,
                    anticipated_movies_heading: document.getElementById('anticipated-movies-heading').value,
                    watched_movies_heading: document.getElementById('watched-movies-heading').value,
                    footer_text: document.getElementById('footer-text').value
                };

                // Save to backend
                const respPost = await fetch('/api/config', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(updatedConfig)
                });

                if (respPost.ok) {
                    showNotification('Email strings saved successfully!', 'success');
                    closeEditStringsModal();
                } else {
                    throw new Error('Failed to save configuration');
                }
            } catch (error) {
                console.error('Error saving email strings:', error);
                showNotification('Failed to save email strings', 'error');
            } finally {
                button.classList.remove('loading');
                button.disabled = false;
            }
        }

        async function resetEmailStrings() {
            if (!confirm('Reset all email strings to default values? This will overwrite your current customizations.')) {
                return;
            }

            const button = event.target.closest('button');
            button.classList.add('loading');
            button.disabled = true;

            try {
                // Set default values (from constants.go)
                document.getElementById('email-title').value = '📺 Your Weekly Newslettar';
                document.getElementById('email-intro').value = '';
                document.getElementById('week-range-prefix').value = 'Week of';
                document.getElementById('coming-this-week-heading').value = '📅 Coming This Week';
                document.getElementById('tv-shows-heading').value = 'TV Shows';
                document.getElementById('movies-heading').value = 'Movies';
                document.getElementById('no-shows-message').value = 'No shows scheduled for this week';
                document.getElementById('no-movies-message').value = 'No movies scheduled for this week';
                document.getElementById('downloaded-section-heading').value = '📥 Downloaded Last Week';
                document.getElementById('no-downloaded-shows-message').value = 'No shows downloaded this week';
                document.getElementById('no-downloaded-movies-message').value = 'No movies downloaded this week';
                document.getElementById('trending-section-heading').value = '🔥 Trending';
                document.getElementById('anticipated-series-heading').value = 'Most Anticipated Series (Next Week)';
                document.getElementById('watched-series-heading').value = 'Most Watched Series (Last Week)';
                document.getElementById('anticipated-movies-heading').value = 'Most Anticipated Movies (Next Week)';
                document.getElementById('watched-movies-heading').value = 'Most Watched Movies (Last Week)';
                document.getElementById('footer-text').value = 'Generated by Newslettar';

                showNotification('Reset to default values. Click "Save Changes" to apply.', 'success');
            } catch (error) {
                console.error('Error resetting email strings:', error);
                showNotification('Failed to reset email strings', 'error');
            } finally {
                button.classList.remove('loading');
                button.disabled = false;
            }
        }

        async function sendNow() {
            // Check if there are any recipient emails configured
            const toEmailsValue = document.getElementById('to_emails').value.trim();
            if (emailTags.length === 0 && !toEmailsValue) {
                showNotification('Cannot send newsletter: No recipient email addresses configured. Please add at least one email address in the Configuration tab.', 'error');
                return;
            }

            if (!confirm('Send newsletter now?')) return;

            const button = event.target.closest('button');
            button.classList.add('loading');
            button.disabled = true;

            showNotification('Sending newsletter...', 'success');

            try {
                const resp = await fetch('/api/send', { method: 'POST' });
                const data = await resp.json();

                if (data.success) {
                    showNotification('Newsletter sent successfully!', 'success');
                } else {
                    showNotification('Failed to send newsletter', 'error');
                }
            } catch (error) {
                showNotification('Send failed: ' + error.message, 'error');
            } finally {
                button.classList.remove('loading');
                button.disabled = false;
            }
        }

        function toggleTraktLimit(type) {
            const checkbox = document.getElementById('show-trakt-' + type);
            const container = document.getElementById('trakt-' + type + '-limit-container');
            container.style.display = checkbox.checked ? 'block' : 'none';
            saveTemplateSettings();
        }

        async function saveTemplateSettings() {
            const showPosters = document.getElementById('show-posters').checked;
            const showDownloaded = document.getElementById('show-downloaded').checked;
            const showSeriesOverview = document.getElementById('show-series-overview').checked;
            const showEpisodeOverview = document.getElementById('show-episode-overview').checked;
            const showUnmonitored = document.getElementById('show-unmonitored').checked;
            const showSeriesRatings = document.getElementById('show-series-ratings').checked;
            const darkMode = document.getElementById('dark-mode').checked;
            const showTraktAnticipatedSeries = document.getElementById('show-trakt-anticipated-series').checked;
            const showTraktWatchedSeries = document.getElementById('show-trakt-watched-series').checked;
            const showTraktAnticipatedMovies = document.getElementById('show-trakt-anticipated-movies').checked;
            const showTraktWatchedMovies = document.getElementById('show-trakt-watched-movies').checked;
            const traktAnticipatedSeriesLimit = document.getElementById('trakt-anticipated-series-limit').value || '5';
            const traktWatchedSeriesLimit = document.getElementById('trakt-watched-series-limit').value || '5';
            const traktAnticipatedMoviesLimit = document.getElementById('trakt-anticipated-movies-limit').value || '5';
            const traktWatchedMoviesLimit = document.getElementById('trakt-watched-movies-limit').value || '5';

            try {
                await fetch('/api/config', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({
                        show_posters: showPosters ? 'true' : 'false',
                        show_downloaded: showDownloaded ? 'true' : 'false',
                        show_series_overview: showSeriesOverview ? 'true' : 'false',
                        show_episode_overview: showEpisodeOverview ? 'true' : 'false',
                        show_unmonitored: showUnmonitored ? 'true' : 'false',
                        show_series_ratings: showSeriesRatings ? 'true' : 'false',
                        dark_mode: darkMode ? 'true' : 'false',
                        show_trakt_anticipated_series: showTraktAnticipatedSeries ? 'true' : 'false',
                        show_trakt_watched_series: showTraktWatchedSeries ? 'true' : 'false',
                        show_trakt_anticipated_movies: showTraktAnticipatedMovies ? 'true' : 'false',
                        show_trakt_watched_movies: showTraktWatchedMovies ? 'true' : 'false',
                        trakt_anticipated_series_limit: traktAnticipatedSeriesLimit,
                        trakt_watched_series_limit: traktWatchedSeriesLimit,
                        trakt_anticipated_movies_limit: traktAnticipatedMoviesLimit,
                        trakt_watched_movies_limit: traktWatchedMoviesLimit
                    })
                });

                showNotification('Template settings saved', 'success');
            } catch (error) {
                showNotification('Failed to save settings: ' + error.message, 'error');
            }
        }

        async function checkUpdates(event) {
            const button = event ? event.target : null;
            if (button) {
                button.classList.add('loading');
                button.disabled = true;
            }

            try {
                const resp = await fetch('/api/version');
                const data = await resp.json();

                let html = '<div style="background: #252f3f; padding: 20px; border-radius: 10px; margin-bottom: 20px;">';
                html += '<p><strong>Current Version:</strong> ' + data.current_version + '</p>';
                html += '<p><strong>Latest Version:</strong> ' + data.latest_version + '</p>';

                if (data.update_available) {
                    html += '<p style="color: #38ef7d; margin-top: 15px;"><strong>Update Available!</strong></p>';
                    html += '<h4 style="margin-top: 15px;">What\'s New:</h4>';
                    html += '<ul style="margin-left: 20px; margin-top: 10px;">';
                    data.changelog.forEach(item => {
                        html += '<li style="margin: 5px 0;">' + item + '</li>';
                    });
                    html += '</ul>';
                    document.getElementById('update-btn').style.display = 'inline-block';
                    document.getElementById('update-tab-button').style.display = 'inline-block';
                } else {
                    html += '<p style="color: #8899aa; margin-top: 15px;">You are running the latest version!</p>';
                    document.getElementById('update-btn').style.display = 'none';
                    document.getElementById('update-tab-button').style.display = 'none';
                }

                html += '</div>';
                document.getElementById('version-info').innerHTML = html;
            } catch (error) {
                showNotification('Failed to check updates: ' + error.message, 'error');
            } finally {
                if (button) {
                    button.classList.remove('loading');
                    button.disabled = false;
                }
            }
        }

        async function performUpdate() {
            if (!confirm('Update Newslettar? The page will reload automatically when the update completes.')) return;

            const button = document.getElementById('update-btn');
            button.classList.add('loading');
            button.disabled = true;

            showNotification('Starting update... Please wait...', 'success');

            try {
                await fetch('/api/update', { method: 'POST' });

                // Poll server to check when it's back up
                let attempts = 0;
                const maxAttempts = 60; // 60 attempts * 2 seconds = 2 minutes max

                const pollServer = async () => {
                    attempts++;

                    if (attempts > maxAttempts) {
                        showNotification('Update may have failed. Please refresh manually.', 'error');
                        button.classList.remove('loading');
                        button.disabled = false;
                        return;
                    }

                    try {
                        // Try to fetch the version endpoint to see if server is back
                        const response = await fetch('/api/version');
                        if (response.ok) {
                            // Server is back, reload the page
                            showNotification('Update complete! Reloading...', 'success');
                            setTimeout(() => location.reload(), 500);
                        } else {
                            // Server returned an error, keep polling
                            setTimeout(pollServer, 2000);
                        }
                    } catch (error) {
                        // Server not ready yet, keep polling
                        setTimeout(pollServer, 2000);
                    }
                };

                // Start polling after 5 seconds (give the update process time to start)
                setTimeout(pollServer, 5000);

            } catch (error) {
                showNotification('Update failed: ' + error.message, 'error');
                button.classList.remove('loading');
                button.disabled = false;
            }
        }

        function showNotification(message, type) {
            const notification = document.createElement('div');
            notification.className = 'notification ' + type;
            notification.textContent = message;
            notification.setAttribute('role', 'alert');
            document.body.appendChild(notification);

            setTimeout(() => {
                notification.remove();
            }, 10000);
        }

        function showLoading() {
            document.getElementById('loading-overlay').classList.add('show');
        }

        function hideLoading() {
            document.getElementById('loading-overlay').classList.remove('show');
        }

        // Load data on page load
        loadConfig();
        loadDashboard(); // Load dashboard data immediately
        checkUpdates();
    </script>
</body>
</html>`
}
