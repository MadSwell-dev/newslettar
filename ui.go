package main

// getUIHTML returns the full HTML for the web UI
func getUIHTML(version string, nextRun string, timezone string) string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Newslettar</title>
    <link rel="icon" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 100'><rect fill='%23667eea' width='100' height='100' rx='20'/><text x='50' y='70' font-size='60' text-anchor='middle' fill='white'>📺</text></svg>">
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
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            padding: 30px;
            border-radius: 12px;
            margin-bottom: 30px;
            text-align: center;
        }
        .header h1 {
            font-size: 2.5em;
            margin-bottom: 10px;
        }
        .version {
            opacity: 0.9;
            font-size: 0.9em;
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
        .logs-container {
            background: #0f1419;
            padding: 20px;
            border-radius: 8px;
            font-family: 'Courier New', monospace;
            font-size: 13px;
            max-height: 500px;
            overflow-y: auto;
            white-space: pre-wrap;
            border: 2px solid #2a3444;
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
            <h1>📺 Newslettar</h1>
            <p class="version">Version ` + version + `</p>
        </div>

        <div class="tabs" role="tablist">
            <button class="tab active" role="tab" aria-selected="true" aria-controls="config-tab" onclick="showTab('config')">⚙️ Configuration</button>
            <button class="tab" role="tab" aria-selected="false" aria-controls="template-tab" onclick="showTab('template')">📝 Email Template</button>
            <button class="tab" role="tab" aria-selected="false" aria-controls="logs-tab" onclick="showTab('logs')">📋 Logs</button>
            <button class="tab" id="update-tab-button" role="tab" aria-selected="false" aria-controls="update-tab" onclick="showTab('update')" style="display: none;">🔄 Update</button>
        </div>

        <div id="config-tab" class="tab-content active" role="tabpanel">
            <div class="info-banner">
                <p><strong>⏰ Next Scheduled Send:</strong> ` + nextRun + `</p>
                <p><strong>🌍 Timezone:</strong> <span id="current-timezone">` + timezone + `</span></p>
                <p style="margin-top: 10px; font-size: 0.9em; opacity: 0.8;">
                    ℹ️ Scheduler runs internally (no systemd timer needed). Changes apply immediately.
                </p>
            </div>

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

                <h3 style="margin-bottom: 15px; color: #667eea;">Email Settings</h3>
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
                <div class="form-group">
                    <label for="to_emails">To Emails (comma-separated)</label>
                    <input type="text" name="to_emails" id="to_emails" placeholder="user@example.com, user2@example.com" aria-label="To Emails">
                    <div class="error-message" id="to-emails-error">Please enter valid email addresses</div>
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

        <div id="logs-tab" class="tab-content" role="tabpanel">
            <h3 style="margin-bottom: 15px;">📋 Newsletter Logs</h3>
            <button class="btn btn-secondary" onclick="loadLogs()" style="margin-bottom: 15px;" aria-label="Refresh logs">
                <span>🔄 Refresh Logs</span>
            </button>
            <div class="logs-container" id="logs" role="log" aria-live="polite"></div>
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

    <!-- Loading Overlay -->
    <div id="loading-overlay" class="loading-overlay" role="status" aria-live="polite">
        <div class="loading-spinner"></div>
    </div>

    <script>
        let logsInterval;

        // Keyboard navigation
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                closePreview();
            }
        });

        function showTab(tabName) {
            document.querySelectorAll('.tab').forEach(t => {
                t.classList.remove('active');
                t.setAttribute('aria-selected', 'false');
            });
            document.querySelectorAll('.tab-content').forEach(t => t.classList.remove('active'));
            
            event.target.classList.add('active');
            event.target.setAttribute('aria-selected', 'true');
            document.getElementById(tabName + '-tab').classList.add('active');

            if (tabName === 'logs') {
                loadLogs();
                logsInterval = setInterval(loadLogs, 5000);
            } else {
                if (logsInterval) {
                    clearInterval(logsInterval);
                }
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

            toEmails.addEventListener('blur', function() {
                if (this.value && !validateEmails(this)) {
                    this.classList.add('error');
                    this.classList.remove('success');
                    document.getElementById('to-emails-error').classList.add('show');
                } else if (this.value) {
                    this.classList.remove('error');
                    this.classList.add('success');
                    document.getElementById('to-emails-error').classList.remove('show');
                }
            });

            // Update timezone info on change
            document.getElementById('timezone').addEventListener('change', updateTimezoneInfo);
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

        async function loadConfig() {
            showLoading();
            try {
                const resp = await fetch('/api/config');
                const data = await resp.json();
                
                document.querySelector('[name="sonarr_url"]').value = data.sonarr_url || '';
                document.querySelector('[name="sonarr_api_key"]').value = data.sonarr_api_key || '';
                document.querySelector('[name="radarr_url"]').value = data.radarr_url || '';
                document.querySelector('[name="radarr_api_key"]').value = data.radarr_api_key || '';
                document.querySelector('[name="smtp_host"]').value = data.smtp_host || 'smtp.mailgun.org';
                document.querySelector('[name="smtp_port"]').value = data.smtp_port || '587';
                document.querySelector('[name="smtp_user"]').value = data.smtp_user || '';
                document.querySelector('[name="smtp_pass"]').value = data.smtp_pass || '';
                document.querySelector('[name="from_email"]').value = data.from_email || '';
                document.querySelector('[name="from_name"]').value = data.from_name || 'Newslettar';
                document.querySelector('[name="to_emails"]').value = data.to_emails || '';
                document.querySelector('[name="timezone"]').value = data.timezone || 'UTC';
                document.querySelector('[name="schedule_day"]').value = data.schedule_day || 'Sun';
                document.querySelector('[name="schedule_time"]').value = data.schedule_time || '09:00';
                
                document.getElementById('show-posters').checked = data.show_posters !== 'false';
                document.getElementById('show-downloaded').checked = data.show_downloaded !== 'false';
                document.getElementById('show-series-overview').checked = data.show_series_overview !== 'false';
                document.getElementById('show-episode-overview').checked = data.show_episode_overview !== 'false';
                document.getElementById('show-unmonitored').checked = data.show_unmonitored !== 'false';

                document.getElementById('current-timezone').textContent = data.timezone || 'UTC';
                
                await updateTimezoneInfo();
            } catch (error) {
                showNotification('Failed to load configuration: ' + error.message, 'error');
            } finally {
                hideLoading();
            }
        }

        document.getElementById('config-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const formData = new FormData(e.target);
            const data = Object.fromEntries(formData);
            
            const submitBtn = e.target.querySelector('button[type="submit"]');
            submitBtn.classList.add('loading');
            submitBtn.disabled = true;
            
            try {
                const resp = await fetch('/api/config', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify(data)
                });

                if (resp.ok) {
                    showNotification('Configuration saved successfully!', 'success');
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

        async function sendNow() {
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

        async function loadLogs() {
            try {
                const resp = await fetch('/api/logs');
                const logs = await resp.text();
                document.getElementById('logs').textContent = logs;
                document.getElementById('logs').scrollTop = document.getElementById('logs').scrollHeight;
            } catch (error) {
                console.error('Failed to load logs:', error);
            }
        }

        async function saveTemplateSettings() {
            const showPosters = document.getElementById('show-posters').checked;
            const showDownloaded = document.getElementById('show-downloaded').checked;
            const showSeriesOverview = document.getElementById('show-series-overview').checked;
            const showEpisodeOverview = document.getElementById('show-episode-overview').checked;
            const showUnmonitored = document.getElementById('show-unmonitored').checked;

            try {
                await fetch('/api/config', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({
                        show_posters: showPosters ? 'true' : 'false',
                        show_downloaded: showDownloaded ? 'true' : 'false',
                        show_series_overview: showSeriesOverview ? 'true' : 'false',
                        show_episode_overview: showEpisodeOverview ? 'true' : 'false',
                        show_unmonitored: showUnmonitored ? 'true' : 'false'
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

        // Load config on page load
        loadConfig();
        checkUpdates();
    </script>
</body>
</html>`
}
