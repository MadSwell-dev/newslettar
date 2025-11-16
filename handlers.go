package main

import (
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"
)

// Register all HTTP handlers
func registerHandlers() {
	http.HandleFunc("/", withGzip(uiHandler))
	http.HandleFunc("/api/config", configHandler)
	http.HandleFunc("/api/test-sonarr", testSonarrHandler)
	http.HandleFunc("/api/test-radarr", testRadarrHandler)
	http.HandleFunc("/api/test-email", testEmailHandler)
	http.HandleFunc("/api/send", sendHandler)
	http.HandleFunc("/api/logs", logsHandler)
	http.HandleFunc("/api/version", versionHandler)
	http.HandleFunc("/api/update", updateHandler)
	http.HandleFunc("/api/preview", previewHandler)
	http.HandleFunc("/api/timezone-info", timezoneInfoHandler)
}

// Gzip compression middleware
func withGzip(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			handler(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()

		gzw := &gzipResponseWriter{Writer: gz, ResponseWriter: w}
		handler(gzw, r)
	}
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// Handlers

func uiHandler(w http.ResponseWriter, r *http.Request) {
	cfg := getConfig()
	loc := getTimezone(cfg.Timezone)
	nextRun := getNextScheduledRun(cfg.ScheduleDay, cfg.ScheduleTime, loc)

	html := getUIHTML(version, nextRun, cfg.Timezone)

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

func timezoneInfoHandler(w http.ResponseWriter, r *http.Request) {
	tz := r.URL.Query().Get("tz")
	if tz == "" {
		tz = "UTC"
	}

	loc := getTimezone(tz)
	now := time.Now().In(loc)

	_, offset := now.Zone()
	hours := offset / 3600
	minutes := (offset % 3600) / 60

	offsetStr := fmt.Sprintf("GMT%+d", hours)
	if minutes != 0 {
		offsetStr = fmt.Sprintf("GMT%+d:%02d", hours, minutes)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"current_time": now.Format("Monday, January 2, 2006 3:04 PM"),
		"offset":       offsetStr,
	})
}

// Preview handler for UI
func previewHandler(w http.ResponseWriter, r *http.Request) {
	cfg := getConfig()
	loc := getTimezone(cfg.Timezone)
	now := time.Now().In(loc)

	weekStart := now.AddDate(0, 0, -7)
	weekEnd := now

	// Parallel API calls with context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	var downloadedEpisodes, upcomingEpisodes []Episode
	var downloadedMovies, upcomingMovies []Movie

	wg.Add(4)

	go func() {
		defer wg.Done()
		downloadedEpisodes, _ = fetchSonarrHistoryWithRetry(ctx, cfg, weekStart, 2)
	}()

	go func() {
		defer wg.Done()
		upcomingEpisodes, _ = fetchSonarrCalendarWithRetry(ctx, cfg, weekEnd, weekEnd.AddDate(0, 0, 7), 2)
	}()

	go func() {
		defer wg.Done()
		downloadedMovies, _ = fetchRadarrHistoryWithRetry(ctx, cfg, weekStart, 2)
	}()

	go func() {
		defer wg.Done()
		upcomingMovies, _ = fetchRadarrCalendarWithRetry(ctx, cfg, weekEnd, weekEnd.AddDate(0, 0, 7), 2)
	}()

	wg.Wait()

	// Sort movies chronologically
	sort.Slice(upcomingMovies, func(i, j int) bool {
		return upcomingMovies[i].ReleaseDate < upcomingMovies[j].ReleaseDate
	})
	sort.Slice(downloadedMovies, func(i, j int) bool {
		return downloadedMovies[i].ReleaseDate < downloadedMovies[j].ReleaseDate
	})

	data := NewsletterData{
		WeekStart:              weekStart.Format("January 2, 2006"),
		WeekEnd:                weekEnd.Format("January 2, 2006"),
		UpcomingSeriesGroups:   groupEpisodesBySeries(upcomingEpisodes),
		UpcomingMovies:         upcomingMovies,
		DownloadedSeriesGroups: groupEpisodesBySeries(downloadedEpisodes),
		DownloadedMovies:       downloadedMovies,
	}

	html, err := generateNewsletterHTML(data, cfg.ShowPosters, cfg.ShowDownloaded)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Failed to generate preview: %v", err),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"html":    html,
	})
}

func configHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var webCfg WebConfig
		if err := json.NewDecoder(r.Body).Decode(&webCfg); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		envMap := readEnvFile()

		// Only update fields that were provided
		if webCfg.SonarrURL != "" {
			envMap["SONARR_URL"] = webCfg.SonarrURL
		}
		if webCfg.SonarrAPIKey != "" {
			envMap["SONARR_API_KEY"] = webCfg.SonarrAPIKey
		}
		if webCfg.RadarrURL != "" {
			envMap["RADARR_URL"] = webCfg.RadarrURL
		}
		if webCfg.RadarrAPIKey != "" {
			envMap["RADARR_API_KEY"] = webCfg.RadarrAPIKey
		}
		if webCfg.MailgunSMTP != "" {
			envMap["MAILGUN_SMTP"] = webCfg.MailgunSMTP
		}
		if webCfg.MailgunPort != "" {
			envMap["MAILGUN_PORT"] = webCfg.MailgunPort
		}
		if webCfg.MailgunUser != "" {
			envMap["MAILGUN_USER"] = webCfg.MailgunUser
		}
		if webCfg.MailgunPass != "" {
			envMap["MAILGUN_PASS"] = webCfg.MailgunPass
		}
		if webCfg.FromEmail != "" {
			envMap["FROM_EMAIL"] = webCfg.FromEmail
		}
		if webCfg.FromName != "" {
			envMap["FROM_NAME"] = webCfg.FromName
		}
		if webCfg.ToEmails != "" {
			envMap["TO_EMAILS"] = webCfg.ToEmails
		}
		if webCfg.Timezone != "" {
			envMap["TIMEZONE"] = webCfg.Timezone
		}
		if webCfg.ScheduleDay != "" {
			envMap["SCHEDULE_DAY"] = webCfg.ScheduleDay
		}
		if webCfg.ScheduleTime != "" {
			envMap["SCHEDULE_TIME"] = webCfg.ScheduleTime
		}
		if webCfg.ShowPosters != "" {
			envMap["SHOW_POSTERS"] = webCfg.ShowPosters
		}
		if webCfg.ShowDownloaded != "" {
			envMap["SHOW_DOWNLOADED"] = webCfg.ShowDownloaded
		}

		var envContent strings.Builder
		for key, value := range envMap {
			envContent.WriteString(fmt.Sprintf("%s=%s\n", key, value))
		}

		if err := os.WriteFile(".env", []byte(envContent.String()), 0644); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Reload config and restart scheduler
		reloadConfig()
		restartScheduler()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
		return
	}

	// GET request - return current config
	envMap := readEnvFile()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"sonarr_url":      getEnvFromFile(envMap, "SONARR_URL", ""),
		"sonarr_api_key":  getEnvFromFile(envMap, "SONARR_API_KEY", ""),
		"radarr_url":      getEnvFromFile(envMap, "RADARR_URL", ""),
		"radarr_api_key":  getEnvFromFile(envMap, "RADARR_API_KEY", ""),
		"mailgun_smtp":    getEnvFromFile(envMap, "MAILGUN_SMTP", "smtp.mailgun.org"),
		"mailgun_port":    getEnvFromFile(envMap, "MAILGUN_PORT", "587"),
		"mailgun_user":    getEnvFromFile(envMap, "MAILGUN_USER", ""),
		"mailgun_pass":    getEnvFromFile(envMap, "MAILGUN_PASS", ""),
		"from_email":      getEnvFromFile(envMap, "FROM_EMAIL", ""),
		"from_name":       getEnvFromFile(envMap, "FROM_NAME", "Newslettar"),
		"to_emails":       getEnvFromFile(envMap, "TO_EMAILS", ""),
		"timezone":        getEnvFromFile(envMap, "TIMEZONE", "UTC"),
		"schedule_day":    getEnvFromFile(envMap, "SCHEDULE_DAY", "Sun"),
		"schedule_time":   getEnvFromFile(envMap, "SCHEDULE_TIME", "09:00"),
		"show_posters":    getEnvFromFile(envMap, "SHOW_POSTERS", "true"),
		"show_downloaded": getEnvFromFile(envMap, "SHOW_DOWNLOADED", "true"),
	})
}

func testSonarrHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL    string `json:"url"`
		APIKey string `json:"api_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	success := false
	message := "Missing URL or API key"

	if req.URL != "" && req.APIKey != "" {
		httpReq, err := http.NewRequest("GET", req.URL+"/api/v3/system/status", nil)
		if err == nil {
			httpReq.Header.Set("X-Api-Key", req.APIKey)
			resp, err := httpClient.Do(httpReq)
			if err != nil {
				message = fmt.Sprintf("Connection failed: %v", err)
			} else if resp.StatusCode == 200 {
				success = true
				message = "Sonarr connection successful!"
				resp.Body.Close()
			} else {
				message = fmt.Sprintf("Connection failed: HTTP %d", resp.StatusCode)
				resp.Body.Close()
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": success,
		"message": message,
	})
}

func testRadarrHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL    string `json:"url"`
		APIKey string `json:"api_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	success := false
	message := "Missing URL or API key"

	if req.URL != "" && req.APIKey != "" {
		httpReq, err := http.NewRequest("GET", req.URL+"/api/v3/system/status", nil)
		if err == nil {
			httpReq.Header.Set("X-Api-Key", req.APIKey)
			resp, err := httpClient.Do(httpReq)
			if err != nil {
				message = fmt.Sprintf("Connection failed: %v", err)
			} else if resp.StatusCode == 200 {
				success = true
				message = "Radarr connection successful!"
				resp.Body.Close()
			} else {
				message = fmt.Sprintf("Connection failed: HTTP %d", resp.StatusCode)
				resp.Body.Close()
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": success,
		"message": message,
	})
}

func testEmailHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SMTP string `json:"smtp"`
		Port string `json:"port"`
		User string `json:"user"`
		Pass string `json:"pass"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	success := false
	message := "SMTP credentials missing"

	if req.User != "" && req.Pass != "" {
		addr := fmt.Sprintf("%s:%s", req.SMTP, req.Port)

		client, err := smtp.Dial(addr)
		if err != nil {
			message = fmt.Sprintf("Connection failed: %v", err)
		} else {
			defer client.Close()

			if err = client.Hello("localhost"); err != nil {
				message = fmt.Sprintf("EHLO failed: %v", err)
			} else if ok, _ := client.Extension("STARTTLS"); ok {
				config := &tls.Config{ServerName: req.SMTP}
				if err = client.StartTLS(config); err != nil {
					message = fmt.Sprintf("STARTTLS failed: %v", err)
				} else {
					auth := smtp.PlainAuth("", req.User, req.Pass, req.SMTP)
					if err = client.Auth(auth); err != nil {
						message = fmt.Sprintf("Authentication failed: %v", err)
					} else {
						success = true
						message = "SMTP authentication successful (with STARTTLS)"
					}
				}
			} else {
				auth := smtp.PlainAuth("", req.User, req.Pass, req.SMTP)
				if err = client.Auth(auth); err != nil {
					message = fmt.Sprintf("Authentication failed: %v", err)
				} else {
					success = true
					message = "SMTP authentication successful"
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": success,
		"message": message,
	})
}

func sendHandler(w http.ResponseWriter, r *http.Request) {
	// Send immediately
	go runNewsletter()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Newsletter generation started",
	})
}

func logsHandler(w http.ResponseWriter, r *http.Request) {
	logBufferMu.Lock()
	defer logBufferMu.Unlock()

	w.Header().Set("Content-Type", "text/plain")
	for _, line := range logBuffer {
		fmt.Fprint(w, line)
	}
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := httpClient.Get("https://raw.githubusercontent.com/agencefanfare/newslettar/main/version.json")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"current_version":  version,
			"latest_version":   version,
			"update_available": false,
			"error":            "Could not check for updates",
		})
		return
	}
	defer resp.Body.Close()

	var remoteVersion struct {
		Version   string   `json:"version"`
		Released  string   `json:"released"`
		Changelog []string `json:"changelog"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&remoteVersion); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"current_version":  version,
			"latest_version":   version,
			"update_available": false,
			"error":            "Could not parse version info",
		})
		return
	}

	updateAvailable := isNewerVersion(remoteVersion.Version, version)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"current_version":  version,
		"latest_version":   remoteVersion.Version,
		"update_available": updateAvailable,
		"released":         remoteVersion.Released,
		"changelog":        remoteVersion.Changelog,
	})
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Update started! Building in background...",
	})

	go func() {
		time.Sleep(1 * time.Second)

		log.Println("🔄 Starting update process...")

		cmd := exec.Command("bash", "-c", `
			set -e
			cd /opt/newslettar
			echo "Backing up .env..."
			cp .env .env.backup
			echo "Updating from GitHub..."
			git fetch origin main -q
			git reset --hard origin/main -q
			echo "Building with optimization flags..."
			/usr/local/go/bin/go build -ldflags="-s -w" -trimpath -o newslettar main.go
			echo "Restoring .env..."
			mv .env.backup .env
			echo "Restarting service..."
			systemctl restart newslettar.service
			echo "Update complete!"
		`)

		output, err := cmd.CombinedOutput()
		log.Printf("Update output: %s", string(output))
		if err != nil {
			log.Printf("❌ Update failed: %v", err)
		} else {
			log.Printf("✅ Update completed successfully")
		}
	}()
}
