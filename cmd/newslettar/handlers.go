package main

import (
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/smtp"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

// Register all HTTP handlers
func registerHandlers() {
	// Serve embedded assets (images, etc.)
	assetsHandler := http.FileServer(http.FS(assetsFS))
	http.Handle("/assets/", assetsHandler)
	http.HandleFunc("/", withGzip(uiHandler))
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/api/config", configHandler)
	http.HandleFunc("/api/test-sonarr", testSonarrHandler)
	http.HandleFunc("/api/test-radarr", testRadarrHandler)
	http.HandleFunc("/api/test-trakt", testTraktHandler)
	http.HandleFunc("/api/test-email", testEmailHandler)
	http.HandleFunc("/api/send", sendHandler)
	http.HandleFunc("/api/logs", logsHandler)
	http.HandleFunc("/api/version", versionHandler)
	http.HandleFunc("/api/preview", previewHandler)
	http.HandleFunc("/api/timezone-info", timezoneInfoHandler)
	http.HandleFunc("/api/dashboard", dashboardHandler)
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
	nextRun := getNextScheduledRun(cfg.ScheduleDay, cfg.ScheduleTime, cfg.ScheduleType, cfg.ScheduleDayOfMonth, loc)

	// Detect installation type: docker, native-windows, native-linux, or unknown
	installType := detectInstallationType()

	html := getUIHTML(version, nextRun, cfg.Timezone, installType)

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

// detectInstallationType determines how Newslettar was installed
func detectInstallationType() string {
	// Check for Docker first (most specific)
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return "docker"
	}

	// Check for Windows
	if runtime.GOOS == "windows" {
		// Check if running as Windows service or from Program Files
		if _, err := os.Stat("C:\\Program Files\\Newslettar"); err == nil {
			return "native-windows"
		}
		// Still Windows, even if not in Program Files
		return "native-windows"
	}

	// Check for Linux native installation
	if runtime.GOOS == "linux" {
		// Check for systemd service file (native Linux installation)
		if _, err := os.Stat("/etc/systemd/system/newslettar.service"); err == nil {
			return "native-linux"
		}
		// Check if installed in /opt (from install script)
		if _, err := os.Stat("/opt/newslettar/newslettar"); err == nil {
			return "native-linux"
		}
	}

	// Default to unknown for other scenarios
	return "unknown"
}

// Health check endpoint for monitoring and load balancers
func healthHandler(w http.ResponseWriter, r *http.Request) {
	cfg := getConfig()

	// Check if basic configuration is present
	healthy := true
	checks := make(map[string]string)

	// Check Sonarr configuration
	if cfg.SonarrURL != "" && cfg.SonarrAPIKey != "" {
		checks["sonarr"] = "configured"
	} else if cfg.SonarrURL == "" && cfg.SonarrAPIKey == "" {
		checks["sonarr"] = "not_configured"
	} else {
		checks["sonarr"] = "misconfigured"
		healthy = false
	}

	// Check Radarr configuration
	if cfg.RadarrURL != "" && cfg.RadarrAPIKey != "" {
		checks["radarr"] = "configured"
	} else if cfg.RadarrURL == "" && cfg.RadarrAPIKey == "" {
		checks["radarr"] = "not_configured"
	} else {
		checks["radarr"] = "misconfigured"
		healthy = false
	}

	// Check if at least one service is configured
	if checks["sonarr"] == "not_configured" && checks["radarr"] == "not_configured" {
		checks["services"] = "none_configured"
		healthy = false
	} else {
		checks["services"] = "ok"
	}

	// Check email configuration
	// Email is configured if SMTP settings are present (recipients can be added later)
	if cfg.SMTPHost != "" && cfg.SMTPPort != "" && cfg.SMTPUser != "" && cfg.SMTPPass != "" && cfg.FromEmail != "" {
		checks["email"] = "configured"
	} else if cfg.SMTPHost == "" && cfg.SMTPPort == "" && cfg.FromEmail == "" {
		checks["email"] = "not_configured"
		// Don't mark as unhealthy if email is simply not configured
	} else {
		checks["email"] = "misconfigured"
		healthy = false
	}

	status := "healthy"
	statusCode := http.StatusOK
	if !healthy {
		status = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	}

	response := map[string]interface{}{
		"status":  status,
		"version": version,
		"checks":  checks,
		"uptime":  time.Since(startTime).String(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func timezoneInfoHandler(w http.ResponseWriter, r *http.Request) {
	tz := r.URL.Query().Get("tz")
	if tz == "" {
		tz = DefaultTimezone
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

	// Calculate timeframe based on schedule type
	var weekStart, weekEnd time.Time
	if cfg.ScheduleType == "monthly" {
		// Monthly: previous month and current month
		weekStart = time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, loc)
		weekEnd = now
	} else {
		// Weekly: last 7 days
		weekStart = now.AddDate(0, 0, -7)
		weekEnd = now
	}

	// Parallel API calls with context
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.APITimeout)*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	var downloadedEpisodes, upcomingEpisodes []Episode
	var downloadedMovies, upcomingMovies []Movie
	var traktAnticipatedSeries, traktWatchedSeries []TraktShow
	var traktAnticipatedMovies, traktWatchedMovies []TraktMovie

	// Check which services are configured
	hasSonarr := cfg.SonarrURL != "" && cfg.SonarrAPIKey != ""
	hasRadarr := cfg.RadarrURL != "" && cfg.RadarrAPIKey != ""

	// Count API calls (only for configured services + Trakt if enabled)
	apiCalls := 0
	if hasSonarr {
		apiCalls += 2 // history + calendar
	}
	if hasRadarr {
		apiCalls += 2 // history + calendar
	}
	if cfg.ShowTraktAnticipatedSeries {
		apiCalls++
	}
	if cfg.ShowTraktWatchedSeries {
		apiCalls++
	}
	if cfg.ShowTraktAnticipatedMovies {
		apiCalls++
	}
	if cfg.ShowTraktWatchedMovies {
		apiCalls++
	}

	wg.Add(apiCalls)

	// Calculate upcoming period based on schedule type
	var upcomingEnd time.Time
	if cfg.ScheduleType == "monthly" {
		upcomingEnd = weekEnd.AddDate(0, 1, 0) // Next month
	} else {
		upcomingEnd = weekEnd.AddDate(0, 0, 7) // Next 7 days
	}

	// Only fetch from Sonarr if configured
	if hasSonarr {
		go func() {
			defer wg.Done()
			downloadedEpisodes, _ = fetchSonarrHistoryWithRetry(ctx, cfg, weekStart, cfg.PreviewRetries)
		}()

		go func() {
			defer wg.Done()
			upcomingEpisodes, _ = fetchSonarrCalendarWithRetry(ctx, cfg, weekEnd, upcomingEnd, cfg.PreviewRetries)
		}()
	}

	// Only fetch from Radarr if configured
	if hasRadarr {
		go func() {
			defer wg.Done()
			downloadedMovies, _ = fetchRadarrHistoryWithRetry(ctx, cfg, weekStart, cfg.PreviewRetries)
		}()

		go func() {
			defer wg.Done()
			upcomingMovies, _ = fetchRadarrCalendarWithRetry(ctx, cfg, weekEnd, upcomingEnd, cfg.PreviewRetries)
		}()
	}

	// Fetch Trakt data if enabled
	if cfg.ShowTraktAnticipatedSeries {
		go func() {
			defer wg.Done()
			traktAnticipatedSeries, _ = fetchTraktAnticipatedSeries(ctx, cfg)
		}()
	}

	if cfg.ShowTraktWatchedSeries {
		go func() {
			defer wg.Done()
			traktWatchedSeries, _ = fetchTraktWatchedSeries(ctx, cfg)
		}()
	}

	if cfg.ShowTraktAnticipatedMovies {
		go func() {
			defer wg.Done()
			traktAnticipatedMovies, _ = fetchTraktAnticipatedMovies(ctx, cfg)
		}()
	}

	if cfg.ShowTraktWatchedMovies {
		go func() {
			defer wg.Done()
			traktWatchedMovies, _ = fetchTraktWatchedMovies(ctx, cfg)
		}()
	}

	wg.Wait()

	// Filter unmonitored items from next week releases only (last week already downloaded)
	if !cfg.ShowUnmonitored {
		upcomingEpisodes = filterMonitoredEpisodes(upcomingEpisodes)
		upcomingMovies = filterMonitoredMovies(upcomingMovies)
	}

	// Sort movies chronologically
	sort.Slice(upcomingMovies, func(i, j int) bool {
		return upcomingMovies[i].ReleaseDate < upcomingMovies[j].ReleaseDate
	})
	sort.Slice(downloadedMovies, func(i, j int) bool {
		return downloadedMovies[i].ReleaseDate < downloadedMovies[j].ReleaseDate
	})

	// Select appropriate strings based on schedule type
	var emailTitle, weekRangePrefix, comingThisWeekHeading string
	var noShowsMessage, noMoviesMessage string
	var downloadedSectionHeading, noDownloadedShowsMessage, noDownloadedMoviesMessage string
	var anticipatedSeriesHeading, watchedSeriesHeading string
	var anticipatedMoviesHeading, watchedMoviesHeading string

	if cfg.ScheduleType == "monthly" {
		emailTitle = cfg.MonthlyEmailTitle
		weekRangePrefix = cfg.MonthlyWeekRangePrefix
		comingThisWeekHeading = cfg.MonthlyComingThisWeekHeading
		noShowsMessage = cfg.MonthlyNoShowsMessage
		noMoviesMessage = cfg.MonthlyNoMoviesMessage
		downloadedSectionHeading = cfg.MonthlyDownloadedSectionHeading
		noDownloadedShowsMessage = cfg.MonthlyNoDownloadedShowsMessage
		noDownloadedMoviesMessage = cfg.MonthlyNoDownloadedMoviesMessage
		anticipatedSeriesHeading = cfg.MonthlyAnticipatedSeriesHeading
		watchedSeriesHeading = cfg.MonthlyWatchedSeriesHeading
		anticipatedMoviesHeading = cfg.MonthlyAnticipatedMoviesHeading
		watchedMoviesHeading = cfg.MonthlyWatchedMoviesHeading
	} else {
		emailTitle = cfg.EmailTitle
		weekRangePrefix = cfg.WeekRangePrefix
		comingThisWeekHeading = cfg.ComingThisWeekHeading
		noShowsMessage = cfg.NoShowsMessage
		noMoviesMessage = cfg.NoMoviesMessage
		downloadedSectionHeading = cfg.DownloadedSectionHeading
		noDownloadedShowsMessage = cfg.NoDownloadedShowsMessage
		noDownloadedMoviesMessage = cfg.NoDownloadedMoviesMessage
		anticipatedSeriesHeading = cfg.AnticipatedSeriesHeading
		watchedSeriesHeading = cfg.WatchedSeriesHeading
		anticipatedMoviesHeading = cfg.AnticipatedMoviesHeading
		watchedMoviesHeading = cfg.WatchedMoviesHeading
	}

	// Format dates based on schedule type
	var weekStartStr, weekEndStr, upcomingStartStr, upcomingEndStr string
	if cfg.ScheduleType == "monthly" {
		// For monthly, just show the current month name and year
		currentMonth := weekEnd.Format("January 2006")
		weekStartStr = currentMonth
		weekEndStr = currentMonth
		upcomingStartStr = currentMonth
		upcomingEndStr = currentMonth
	} else {
		// For weekly, show full dates
		weekStartStr = weekStart.Format("January 2, 2006")
		weekEndStr = weekEnd.Format("January 2, 2006")
		upcomingStartStr = weekEnd.Format("January 2, 2006")
		upcomingEndStr = upcomingEnd.Format("January 2, 2006")
	}

	// Deduplicate episodes and movies
	upcomingEpisodes = deduplicateEpisodes(upcomingEpisodes)
	downloadedEpisodes = deduplicateEpisodes(downloadedEpisodes)
	upcomingMovies = deduplicateMovies(upcomingMovies)
	downloadedMovies = deduplicateMovies(downloadedMovies)

	data := NewsletterData{
		WeekStart:              weekStartStr,
		WeekEnd:                weekEndStr,
		UpcomingStart:          upcomingStartStr,
		UpcomingEnd:            upcomingEndStr,
		UpcomingSeriesGroups:   groupEpisodesBySeries(upcomingEpisodes),
		UpcomingMovies:         upcomingMovies,
		DownloadedSeriesGroups: groupEpisodesBySeries(downloadedEpisodes),
		DownloadedMovies:       downloadedMovies,
		TraktAnticipatedSeries: traktAnticipatedSeries,
		TraktWatchedSeries:     traktWatchedSeries,
		TraktAnticipatedMovies: traktAnticipatedMovies,
		TraktWatchedMovies:     traktWatchedMovies,
		// Customizable strings (schedule-aware)
		EmailTitle:                emailTitle,
		EmailIntro:                cfg.EmailIntro,
		WeekRangePrefix:           weekRangePrefix,
		ComingThisWeekHeading:     comingThisWeekHeading,
		TVShowsHeading:            cfg.TVShowsHeading,
		MoviesHeading:             cfg.MoviesHeading,
		NoShowsMessage:            noShowsMessage,
		NoMoviesMessage:           noMoviesMessage,
		DownloadedSectionHeading:  downloadedSectionHeading,
		NoDownloadedShowsMessage:  noDownloadedShowsMessage,
		NoDownloadedMoviesMessage: noDownloadedMoviesMessage,
		TrendingSectionHeading:    cfg.TrendingSectionHeading,
		AnticipatedSeriesHeading:  anticipatedSeriesHeading,
		WatchedSeriesHeading:      watchedSeriesHeading,
		AnticipatedMoviesHeading:  anticipatedMoviesHeading,
		WatchedMoviesHeading:      watchedMoviesHeading,
		FooterText:                cfg.FooterText,
		// Display options
		ShowPosters:                cfg.ShowPosters,
		ShowDownloaded:             cfg.ShowDownloaded,
		ShowSeriesOverview:         cfg.ShowSeriesOverview,
		ShowEpisodeOverview:        cfg.ShowEpisodeOverview,
		ShowSeriesRatings:          cfg.ShowSeriesRatings,
		DarkMode:                   cfg.DarkMode,
		ShowTraktAnticipatedSeries: cfg.ShowTraktAnticipatedSeries,
		ShowTraktWatchedSeries:     cfg.ShowTraktWatchedSeries,
		ShowTraktAnticipatedMovies: cfg.ShowTraktAnticipatedMovies,
		ShowTraktWatchedMovies:     cfg.ShowTraktWatchedMovies,
	}

	html, err := generateNewsletterHTML(data, cfg)
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

		// Masked placeholder - don't update if this value is sent back
		const maskedPlaceholder = "••••••••"

		// Detect if this is a main config form submission (has API/connection fields)
		// vs template settings submission (only has toggle fields)
		// This prevents template settings from clearing API keys
		hasMainConfigFields := webCfg.SonarrURL != "" || webCfg.SonarrAPIKey != "" ||
			webCfg.RadarrURL != "" || webCfg.RadarrAPIKey != "" ||
			webCfg.TraktClientID != "" || webCfg.SMTPHost != "" ||
			webCfg.SMTPPort != "" || webCfg.SMTPUser != "" ||
			webCfg.SMTPPass != "" || webCfg.FromEmail != "" ||
			webCfg.FromName != "" || webCfg.ToEmails != "" ||
			webCfg.Timezone != "" || webCfg.ScheduleDay != "" ||
			webCfg.ScheduleTime != "" ||
			webCfg.SonarrAPIKey == maskedPlaceholder ||
			webCfg.RadarrAPIKey == maskedPlaceholder ||
			webCfg.TraktClientID == maskedPlaceholder ||
			webCfg.SMTPPass == maskedPlaceholder

		// Only update main config fields if they're being submitted
		if hasMainConfigFields {
			// Allow clearing URLs - update even if empty (same as API keys)
			envMap["SONARR_URL"] = webCfg.SonarrURL
			// Allow clearing API keys - update if not masked (even if empty)
			if webCfg.SonarrAPIKey != maskedPlaceholder {
				envMap["SONARR_API_KEY"] = webCfg.SonarrAPIKey
			}
			// Allow clearing URLs - update even if empty (same as API keys)
			envMap["RADARR_URL"] = webCfg.RadarrURL
			// Allow clearing API keys - update if not masked (even if empty)
			if webCfg.RadarrAPIKey != maskedPlaceholder {
				envMap["RADARR_API_KEY"] = webCfg.RadarrAPIKey
			}
			// Allow clearing Trakt Client ID - update if not masked (even if empty)
			if webCfg.TraktClientID != maskedPlaceholder {
				envMap["TRAKT_CLIENT_ID"] = webCfg.TraktClientID
			}
			if webCfg.SMTPHost != "" {
				envMap["SMTP_HOST"] = webCfg.SMTPHost
			}
			if webCfg.SMTPPort != "" {
				envMap["SMTP_PORT"] = webCfg.SMTPPort
			}
			if webCfg.SMTPUser != "" {
				envMap["SMTP_USER"] = webCfg.SMTPUser
			}
			// Allow clearing password - update if not masked (even if empty)
			if webCfg.SMTPPass != maskedPlaceholder {
				envMap["SMTP_PASS"] = webCfg.SMTPPass
			}
			if webCfg.FromEmail != "" {
				envMap["FROM_EMAIL"] = webCfg.FromEmail
			}
			if webCfg.FromName != "" {
				envMap["FROM_NAME"] = webCfg.FromName
			}
			// Always update TO_EMAILS, even if empty (allows clearing all recipients)
			envMap["TO_EMAILS"] = webCfg.ToEmails
			if webCfg.Timezone != "" {
				envMap["TIMEZONE"] = webCfg.Timezone
			}
			if webCfg.ScheduleDay != "" {
				envMap["SCHEDULE_DAY"] = webCfg.ScheduleDay
			}
			if webCfg.ScheduleTime != "" {
				envMap["SCHEDULE_TIME"] = webCfg.ScheduleTime
			}
			if webCfg.ScheduleType != "" {
				envMap["SCHEDULE_TYPE"] = webCfg.ScheduleType
			}
			if webCfg.ScheduleDayOfMonth != "" {
				envMap["SCHEDULE_DAY_OF_MONTH"] = webCfg.ScheduleDayOfMonth
			}
		}
		if webCfg.ShowPosters != "" {
			envMap["SHOW_POSTERS"] = webCfg.ShowPosters
		}
		if webCfg.ShowDownloaded != "" {
			envMap["SHOW_DOWNLOADED"] = webCfg.ShowDownloaded
		}
		if webCfg.ShowSeriesOverview != "" {
			envMap["SHOW_SERIES_OVERVIEW"] = webCfg.ShowSeriesOverview
		}
		if webCfg.ShowEpisodeOverview != "" {
			envMap["SHOW_EPISODE_OVERVIEW"] = webCfg.ShowEpisodeOverview
		}
		if webCfg.ShowUnmonitored != "" {
			envMap["SHOW_UNMONITORED"] = webCfg.ShowUnmonitored
		}
		if webCfg.ShowSeriesRatings != "" {
			envMap["SHOW_SERIES_RATINGS"] = webCfg.ShowSeriesRatings
		}
		if webCfg.DarkMode != "" {
			envMap["DARK_MODE"] = webCfg.DarkMode
		}
		if webCfg.ShowTraktAnticipatedSeries != "" {
			envMap["SHOW_TRAKT_ANTICIPATED_SERIES"] = webCfg.ShowTraktAnticipatedSeries
		}
		if webCfg.ShowTraktWatchedSeries != "" {
			envMap["SHOW_TRAKT_WATCHED_SERIES"] = webCfg.ShowTraktWatchedSeries
		}
		if webCfg.ShowTraktAnticipatedMovies != "" {
			envMap["SHOW_TRAKT_ANTICIPATED_MOVIES"] = webCfg.ShowTraktAnticipatedMovies
		}
		if webCfg.ShowTraktWatchedMovies != "" {
			envMap["SHOW_TRAKT_WATCHED_MOVIES"] = webCfg.ShowTraktWatchedMovies
		}
		if webCfg.TraktAnticipatedSeriesLimit != "" {
			envMap["TRAKT_ANTICIPATED_SERIES_LIMIT"] = webCfg.TraktAnticipatedSeriesLimit
		}
		if webCfg.TraktWatchedSeriesLimit != "" {
			envMap["TRAKT_WATCHED_SERIES_LIMIT"] = webCfg.TraktWatchedSeriesLimit
		}
		if webCfg.TraktAnticipatedMoviesLimit != "" {
			envMap["TRAKT_ANTICIPATED_MOVIES_LIMIT"] = webCfg.TraktAnticipatedMoviesLimit
		}
		if webCfg.TraktWatchedMoviesLimit != "" {
			envMap["TRAKT_WATCHED_MOVIES_LIMIT"] = webCfg.TraktWatchedMoviesLimit
		}
		// Email string customization
		// Only update if at least one custom string field is provided
		// This prevents wiping them when saving from other tabs
		// Check if any custom string field has a value (indicates template tab submission)
		hasCustomStrings := webCfg.EmailTitle != "" || webCfg.EmailIntro != "" ||
			webCfg.WeekRangePrefix != "" || webCfg.ComingThisWeekHeading != "" ||
			webCfg.TVShowsHeading != "" || webCfg.MoviesHeading != "" ||
			webCfg.NoShowsMessage != "" || webCfg.NoMoviesMessage != "" ||
			webCfg.DownloadedSectionHeading != "" || webCfg.NoDownloadedShowsMessage != "" ||
			webCfg.NoDownloadedMoviesMessage != "" || webCfg.TrendingSectionHeading != "" ||
			webCfg.AnticipatedSeriesHeading != "" || webCfg.WatchedSeriesHeading != "" ||
			webCfg.AnticipatedMoviesHeading != "" || webCfg.WatchedMoviesHeading != "" ||
			webCfg.FooterText != ""

		// Only update custom strings if they're being submitted (from template tab)
		// Allow empty strings for intentional clearing when submitting from template tab
		if hasCustomStrings {
			envMap["EMAIL_TITLE"] = webCfg.EmailTitle
			envMap["EMAIL_INTRO"] = webCfg.EmailIntro
			envMap["WEEK_RANGE_PREFIX"] = webCfg.WeekRangePrefix
			envMap["COMING_THIS_WEEK_HEADING"] = webCfg.ComingThisWeekHeading
			envMap["TV_SHOWS_HEADING"] = webCfg.TVShowsHeading
			envMap["MOVIES_HEADING"] = webCfg.MoviesHeading
			envMap["NO_SHOWS_MESSAGE"] = webCfg.NoShowsMessage
			envMap["NO_MOVIES_MESSAGE"] = webCfg.NoMoviesMessage
			envMap["DOWNLOADED_SECTION_HEADING"] = webCfg.DownloadedSectionHeading
			envMap["NO_DOWNLOADED_SHOWS_MESSAGE"] = webCfg.NoDownloadedShowsMessage
			envMap["NO_DOWNLOADED_MOVIES_MESSAGE"] = webCfg.NoDownloadedMoviesMessage
			envMap["TRENDING_SECTION_HEADING"] = webCfg.TrendingSectionHeading
			envMap["ANTICIPATED_SERIES_HEADING"] = webCfg.AnticipatedSeriesHeading
			envMap["WATCHED_SERIES_HEADING"] = webCfg.WatchedSeriesHeading
			envMap["ANTICIPATED_MOVIES_HEADING"] = webCfg.AnticipatedMoviesHeading
			envMap["WATCHED_MOVIES_HEADING"] = webCfg.WatchedMoviesHeading
			envMap["FOOTER_TEXT"] = webCfg.FooterText
			// Monthly versions
			envMap["MONTHLY_EMAIL_TITLE"] = webCfg.MonthlyEmailTitle
			envMap["MONTHLY_WEEK_RANGE_PREFIX"] = webCfg.MonthlyWeekRangePrefix
			envMap["MONTHLY_COMING_THIS_WEEK_HEADING"] = webCfg.MonthlyComingThisWeekHeading
			envMap["MONTHLY_NO_SHOWS_MESSAGE"] = webCfg.MonthlyNoShowsMessage
			envMap["MONTHLY_NO_MOVIES_MESSAGE"] = webCfg.MonthlyNoMoviesMessage
			envMap["MONTHLY_DOWNLOADED_SECTION_HEADING"] = webCfg.MonthlyDownloadedSectionHeading
			envMap["MONTHLY_NO_DOWNLOADED_SHOWS_MESSAGE"] = webCfg.MonthlyNoDownloadedShowsMessage
			envMap["MONTHLY_NO_DOWNLOADED_MOVIES_MESSAGE"] = webCfg.MonthlyNoDownloadedMoviesMessage
			envMap["MONTHLY_ANTICIPATED_SERIES_HEADING"] = webCfg.MonthlyAnticipatedSeriesHeading
			envMap["MONTHLY_WATCHED_SERIES_HEADING"] = webCfg.MonthlyWatchedSeriesHeading
			envMap["MONTHLY_ANTICIPATED_MOVIES_HEADING"] = webCfg.MonthlyAnticipatedMoviesHeading
			envMap["MONTHLY_WATCHED_MOVIES_HEADING"] = webCfg.MonthlyWatchedMoviesHeading
		}

		var envContent strings.Builder
		for key, value := range envMap {
			envContent.WriteString(fmt.Sprintf("%s=%s\n", key, value))
		}

		// Write .env file with restricted permissions (owner read/write only)
		// 0600 prevents other users from reading API keys and passwords
		if err := os.WriteFile(".env", []byte(envContent.String()), 0600); err != nil {
			log.Printf("❌ Failed to write .env file: %v", err)
			http.Error(w, fmt.Sprintf("Failed to save configuration: %v", err), http.StatusInternalServerError)
			return
		}

		log.Println("✅ Configuration saved successfully")

		// Reload config and restart scheduler
		reloadConfig()
		restartScheduler()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
		return
	}

	// GET request - return current config
	envMap := readEnvFile()
	cfg := getConfig()
	w.Header().Set("Content-Type", "application/json")

	// Mask sensitive fields (API keys, passwords) for security
	// Only return masked values, never expose actual credentials in GET responses
	maskedSonarrKey := ""
	if key := getEnvFromFileOnly(envMap, "SONARR_API_KEY", ""); key != "" {
		maskedSonarrKey = "••••••••"
	}
	maskedRadarrKey := ""
	if key := getEnvFromFileOnly(envMap, "RADARR_API_KEY", ""); key != "" {
		maskedRadarrKey = "••••••••"
	}
	maskedTraktKey := ""
	if key := getEnvFromFileOnly(envMap, "TRAKT_CLIENT_ID", ""); key != "" {
		maskedTraktKey = "••••••••"
	}
	maskedSMTPPass := ""
	if cfg.SMTPPass != "" {
		maskedSMTPPass = "••••••••"
	}

	json.NewEncoder(w).Encode(map[string]string{
		"sonarr_url":                     getEnvFromFileOnly(envMap, "SONARR_URL", ""),
		"sonarr_api_key":                 maskedSonarrKey,
		"radarr_url":                     getEnvFromFileOnly(envMap, "RADARR_URL", ""),
		"radarr_api_key":                 maskedRadarrKey,
		"trakt_client_id":                maskedTraktKey,
		"smtp_host":                      cfg.SMTPHost,
		"smtp_port":                      cfg.SMTPPort,
		"smtp_user":                      cfg.SMTPUser,
		"smtp_pass":                      maskedSMTPPass,
		"from_email":                     getEnvFromFile(envMap, "FROM_EMAIL", ""),
		"from_name":                      getEnvFromFile(envMap, "FROM_NAME", DefaultFromName),
		"to_emails":                      getEnvFromFile(envMap, "TO_EMAILS", ""),
		"timezone":                       getEnvFromFile(envMap, "TIMEZONE", DefaultTimezone),
		"schedule_day":                   getEnvFromFile(envMap, "SCHEDULE_DAY", DefaultScheduleDay),
		"schedule_time":                  getEnvFromFile(envMap, "SCHEDULE_TIME", DefaultScheduleTime),
		"schedule_type":                  getEnvFromFile(envMap, "SCHEDULE_TYPE", DefaultScheduleType),
		"schedule_day_of_month":          fmt.Sprintf("%d", cfg.ScheduleDayOfMonth),
		"show_posters":                   getEnvFromFile(envMap, "SHOW_POSTERS", DefaultShowPosters),
		"show_downloaded":                getEnvFromFile(envMap, "SHOW_DOWNLOADED", DefaultShowDownloaded),
		"show_series_overview":           getEnvFromFile(envMap, "SHOW_SERIES_OVERVIEW", DefaultShowSeriesOverview),
		"show_episode_overview":          getEnvFromFile(envMap, "SHOW_EPISODE_OVERVIEW", DefaultShowEpisodeOverview),
		"show_unmonitored":               getEnvFromFile(envMap, "SHOW_UNMONITORED", DefaultShowUnmonitored),
		"show_series_ratings":            getEnvFromFile(envMap, "SHOW_SERIES_RATINGS", DefaultShowSeriesRatings),
		"dark_mode":                      getEnvFromFile(envMap, "DARK_MODE", DefaultDarkMode),
		"show_trakt_anticipated_series":  getEnvFromFile(envMap, "SHOW_TRAKT_ANTICIPATED_SERIES", DefaultShowTraktAnticipatedSeries),
		"show_trakt_watched_series":      getEnvFromFile(envMap, "SHOW_TRAKT_WATCHED_SERIES", DefaultShowTraktWatchedSeries),
		"show_trakt_anticipated_movies":  getEnvFromFile(envMap, "SHOW_TRAKT_ANTICIPATED_MOVIES", DefaultShowTraktAnticipatedMovies),
		"show_trakt_watched_movies":      getEnvFromFile(envMap, "SHOW_TRAKT_WATCHED_MOVIES", DefaultShowTraktWatchedMovies),
		"trakt_anticipated_series_limit": getEnvFromFile(envMap, "TRAKT_ANTICIPATED_SERIES_LIMIT", "5"),
		"trakt_watched_series_limit":     getEnvFromFile(envMap, "TRAKT_WATCHED_SERIES_LIMIT", "5"),
		"trakt_anticipated_movies_limit": getEnvFromFile(envMap, "TRAKT_ANTICIPATED_MOVIES_LIMIT", "5"),
		"trakt_watched_movies_limit":     getEnvFromFile(envMap, "TRAKT_WATCHED_MOVIES_LIMIT", "5"),
		// Email string customization
		"email_title":                  getEnvFromFile(envMap, "EMAIL_TITLE", DefaultEmailTitle),
		"email_intro":                  getEnvFromFile(envMap, "EMAIL_INTRO", DefaultEmailIntro),
		"week_range_prefix":            getEnvFromFile(envMap, "WEEK_RANGE_PREFIX", DefaultWeekRangePrefix),
		"coming_this_week_heading":     getEnvFromFile(envMap, "COMING_THIS_WEEK_HEADING", DefaultComingThisWeekHeading),
		"tv_shows_heading":             getEnvFromFile(envMap, "TV_SHOWS_HEADING", DefaultTVShowsHeading),
		"movies_heading":               getEnvFromFile(envMap, "MOVIES_HEADING", DefaultMoviesHeading),
		"no_shows_message":             getEnvFromFile(envMap, "NO_SHOWS_MESSAGE", DefaultNoShowsMessage),
		"no_movies_message":            getEnvFromFile(envMap, "NO_MOVIES_MESSAGE", DefaultNoMoviesMessage),
		"downloaded_section_heading":   getEnvFromFile(envMap, "DOWNLOADED_SECTION_HEADING", DefaultDownloadedSectionHeading),
		"no_downloaded_shows_message":  getEnvFromFile(envMap, "NO_DOWNLOADED_SHOWS_MESSAGE", DefaultNoDownloadedShowsMessage),
		"no_downloaded_movies_message": getEnvFromFile(envMap, "NO_DOWNLOADED_MOVIES_MESSAGE", DefaultNoDownloadedMoviesMessage),
		"trending_section_heading":     getEnvFromFile(envMap, "TRENDING_SECTION_HEADING", DefaultTrendingSectionHeading),
		"anticipated_series_heading":   getEnvFromFile(envMap, "ANTICIPATED_SERIES_HEADING", DefaultAnticipatedSeriesHeading),
		"watched_series_heading":       getEnvFromFile(envMap, "WATCHED_SERIES_HEADING", DefaultWatchedSeriesHeading),
		"anticipated_movies_heading":   getEnvFromFile(envMap, "ANTICIPATED_MOVIES_HEADING", DefaultAnticipatedMoviesHeading),
		"watched_movies_heading":       getEnvFromFile(envMap, "WATCHED_MOVIES_HEADING", DefaultWatchedMoviesHeading),
		"footer_text":                  getEnvFromFile(envMap, "FOOTER_TEXT", DefaultFooterText),
		// Monthly email string customization
		"monthly_email_title":                  getEnvFromFile(envMap, "MONTHLY_EMAIL_TITLE", DefaultMonthlyEmailTitle),
		"monthly_week_range_prefix":            getEnvFromFile(envMap, "MONTHLY_WEEK_RANGE_PREFIX", DefaultMonthlyWeekRangePrefix),
		"monthly_coming_this_week_heading":     getEnvFromFile(envMap, "MONTHLY_COMING_THIS_WEEK_HEADING", DefaultMonthlyComingThisWeekHeading),
		"monthly_no_shows_message":             getEnvFromFile(envMap, "MONTHLY_NO_SHOWS_MESSAGE", DefaultMonthlyNoShowsMessage),
		"monthly_no_movies_message":            getEnvFromFile(envMap, "MONTHLY_NO_MOVIES_MESSAGE", DefaultMonthlyNoMoviesMessage),
		"monthly_downloaded_section_heading":   getEnvFromFile(envMap, "MONTHLY_DOWNLOADED_SECTION_HEADING", DefaultMonthlyDownloadedSectionHeading),
		"monthly_no_downloaded_shows_message":  getEnvFromFile(envMap, "MONTHLY_NO_DOWNLOADED_SHOWS_MESSAGE", DefaultMonthlyNoDownloadedShowsMessage),
		"monthly_no_downloaded_movies_message": getEnvFromFile(envMap, "MONTHLY_NO_DOWNLOADED_MOVIES_MESSAGE", DefaultMonthlyNoDownloadedMoviesMessage),
		"monthly_anticipated_series_heading":   getEnvFromFile(envMap, "MONTHLY_ANTICIPATED_SERIES_HEADING", DefaultMonthlyAnticipatedSeriesHeading),
		"monthly_watched_series_heading":       getEnvFromFile(envMap, "MONTHLY_WATCHED_SERIES_HEADING", DefaultMonthlyWatchedSeriesHeading),
		"monthly_anticipated_movies_heading":   getEnvFromFile(envMap, "MONTHLY_ANTICIPATED_MOVIES_HEADING", DefaultMonthlyAnticipatedMoviesHeading),
		"monthly_watched_movies_heading":       getEnvFromFile(envMap, "MONTHLY_WATCHED_MOVIES_HEADING", DefaultMonthlyWatchedMoviesHeading),
	})
}

// Generic API test handler - eliminates 74 lines of duplication
func testAPIHandler(w http.ResponseWriter, r *http.Request, serviceName string) {
	const maskedPlaceholder = "••••••••"

	var req struct {
		URL    string `json:"url"`
		APIKey string `json:"api_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// If API key is masked, load the real one from .env
	if req.APIKey == maskedPlaceholder {
		envMap := readEnvFile()
		if serviceName == "Sonarr" {
			req.APIKey = getEnvFromFile(envMap, "SONARR_API_KEY", "")
		} else if serviceName == "Radarr" {
			req.APIKey = getEnvFromFile(envMap, "RADARR_API_KEY", "")
		}
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
				message = fmt.Sprintf("%s connection successful!", serviceName)
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

func testSonarrHandler(w http.ResponseWriter, r *http.Request) {
	testAPIHandler(w, r, "Sonarr")
}

func testRadarrHandler(w http.ResponseWriter, r *http.Request) {
	testAPIHandler(w, r, "Radarr")
}

func testTraktHandler(w http.ResponseWriter, r *http.Request) {
	const maskedPlaceholder = "••••••••"

	var req struct {
		ClientID string `json:"client_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// If Client ID is masked, load the real one from .env
	if req.ClientID == maskedPlaceholder {
		envMap := readEnvFile()
		req.ClientID = getEnvFromFile(envMap, "TRAKT_CLIENT_ID", "")
	}

	success := false
	message := "Missing Client ID"

	if req.ClientID != "" {
		// Test Trakt API by fetching trending shows (limit 1 for speed)
		httpReq, err := http.NewRequest("GET", "https://api.trakt.tv/shows/trending?limit=1", nil)
		if err == nil {
			httpReq.Header.Set("Content-Type", "application/json")
			httpReq.Header.Set("trakt-api-version", "2")
			httpReq.Header.Set("trakt-api-key", req.ClientID)

			resp, err := httpClient.Do(httpReq)
			if err != nil {
				message = fmt.Sprintf("Connection failed: %v", err)
			} else if resp.StatusCode == 200 {
				success = true
				message = "Trakt connection successful!"
				resp.Body.Close()
			} else if resp.StatusCode == 401 {
				message = "Invalid Client ID"
				resp.Body.Close()
			} else {
				message = fmt.Sprintf("Connection failed: HTTP %d", resp.StatusCode)
				resp.Body.Close()
			}
		} else {
			message = fmt.Sprintf("Failed to create request: %v", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": success,
		"message": message,
	})
}

func testEmailHandler(w http.ResponseWriter, r *http.Request) {
	const maskedPlaceholder = "••••••••"

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

	// If password is masked, load the real one from .env
	if req.Pass == maskedPlaceholder {
		envMap := readEnvFile()
		req.Pass = getEnvFromFile(envMap, "SMTP_PASS", "")
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
				// Secure TLS configuration - require TLS 1.2+ and strong ciphers
				config := &tls.Config{
					ServerName: req.SMTP,
					MinVersion: tls.VersionTLS12,
					CipherSuites: []uint16{
						tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
						tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
						tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
						tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					},
				}
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
	resp, err := httpClient.Get("https://raw.githubusercontent.com/MadSwell-dev/newslettar/main/version.json")
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

// Dashboard handler - returns system stats, newsletter stats, and service status
func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cfg := getConfig()
	loc := getTimezone(cfg.Timezone)
	nextRun := getNextScheduledRun(cfg.ScheduleDay, cfg.ScheduleTime, cfg.ScheduleType, cfg.ScheduleDayOfMonth, loc)

	// Calculate uptime
	uptime := time.Since(startTime)
	uptimeStr := fmt.Sprintf("%dd %dh %dm",
		int(uptime.Hours()/24),
		int(uptime.Hours())%24,
		int(uptime.Minutes())%60)

	// Get actual memory usage
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memoryMB := float64(m.Alloc) / 1024 / 1024

	// Get statistics
	stats.mu.RLock()
	totalEmails := stats.TotalEmailsSent
	lastSent := stats.LastSentDateStr
	stats.mu.RUnlock()

	if lastSent == "" {
		lastSent = "Never"
	}

	// Check service status (config only - no API calls for performance)
	serviceStatus := make(map[string]string)

	// Check Sonarr configuration
	if cfg.SonarrURL != "" && cfg.SonarrAPIKey != "" {
		serviceStatus["sonarr"] = "configured"
	} else {
		serviceStatus["sonarr"] = "not_configured"
	}

	// Check Radarr configuration
	if cfg.RadarrURL != "" && cfg.RadarrAPIKey != "" {
		serviceStatus["radarr"] = "configured"
	} else {
		serviceStatus["radarr"] = "not_configured"
	}

	// Check Email configuration
	// Show as configured if SMTP settings are present (recipients can be added later)
	if cfg.SMTPHost != "" && cfg.SMTPPort != "" && cfg.SMTPUser != "" && cfg.SMTPPass != "" && cfg.FromEmail != "" {
		serviceStatus["email"] = "configured"
	} else if cfg.SMTPHost != "" || cfg.SMTPPort != "" || cfg.FromEmail != "" {
		serviceStatus["email"] = "misconfigured" // Partially configured
	} else {
		serviceStatus["email"] = "not_configured"
	}

	// Check Trakt configuration
	if cfg.TraktClientID != "" {
		serviceStatus["trakt"] = "configured"
	} else {
		serviceStatus["trakt"] = "not_configured"
	}

	dashboard := DashboardData{
		Version:          version,
		Uptime:           uptimeStr,
		UptimeSeconds:    int64(uptime.Seconds()),
		MemoryUsageMB:    memoryMB,
		Port:             cfg.WebUIPort,
		TotalEmailsSent:  totalEmails,
		LastSentDate:     lastSent,
		NextScheduledRun: nextRun,
		Timezone:         cfg.Timezone,
		ServiceStatus:    serviceStatus,
	}

	json.NewEncoder(w).Encode(dashboard)
}
