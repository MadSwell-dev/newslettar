package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"sort"
	"strings"
	"sync"
	"time"
)

// Newsletter sending logic with parallel API calls
func runNewsletter() {
	cfg := getConfig()
	loc := getTimezone(cfg.Timezone)
	now := time.Now().In(loc)

	log.Println("🚀 Starting Newslettar - Weekly newsletter generation...")
	log.Printf("⏰ Current time: %s (%s)", now.Format("2006-01-02 15:04:05"), cfg.Timezone)

	weekStart := now.AddDate(0, 0, -7)
	weekEnd := now

	log.Printf("📅 Week range: %s to %s", weekStart.Format("2006-01-02"), weekEnd.Format("2006-01-02"))

	// Use a cancellable context for all fetches
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.APITimeout)*time.Second)
	defer cancel()

	// Parallel API calls (3-4x faster!)
	var wg sync.WaitGroup
	var downloadedEpisodes, upcomingEpisodes []Episode
	var downloadedMovies, upcomingMovies []Movie
	var traktAnticipatedSeries, traktWatchedSeries []TraktShow
	var traktAnticipatedMovies, traktWatchedMovies []TraktMovie
	var errSonarrHistory, errSonarrCalendar, errRadarrHistory, errRadarrCalendar error

	log.Println("📡 Fetching data in parallel...")
	startFetch := time.Now()

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

	// Only fetch from Sonarr if configured
	if hasSonarr {
		go func() {
			defer wg.Done()
			log.Println("📺 Fetching Sonarr history...")
			downloadedEpisodes, errSonarrHistory = fetchSonarrHistoryWithRetry(ctx, cfg, weekStart, cfg.MaxRetries)
			if errSonarrHistory != nil {
				log.Printf("⚠️  Sonarr history error: %v", errSonarrHistory)
			} else {
				log.Printf("✓ Found %d downloaded episodes", len(downloadedEpisodes))
			}
		}()

		go func() {
			defer wg.Done()
			log.Println("📺 Fetching Sonarr calendar...")
			upcomingEpisodes, errSonarrCalendar = fetchSonarrCalendarWithRetry(ctx, cfg, weekEnd, weekEnd.AddDate(0, 0, 7), cfg.MaxRetries)
			if errSonarrCalendar != nil {
				log.Printf("⚠️  Sonarr calendar error: %v", errSonarrCalendar)
			} else {
				log.Printf("✓ Found %d upcoming episodes", len(upcomingEpisodes))
			}
		}()
	} else {
		log.Println("📺 Sonarr not configured, skipping...")
	}

	// Only fetch from Radarr if configured
	if hasRadarr {
		go func() {
			defer wg.Done()
			log.Println("🎬 Fetching Radarr history...")
			downloadedMovies, errRadarrHistory = fetchRadarrHistoryWithRetry(ctx, cfg, weekStart, cfg.MaxRetries)
			if errRadarrHistory != nil {
				log.Printf("⚠️  Radarr history error: %v", errRadarrHistory)
			} else {
				log.Printf("✓ Found %d downloaded movies", len(downloadedMovies))
			}
		}()

		go func() {
			defer wg.Done()
			log.Println("🎬 Fetching Radarr calendar...")
			upcomingMovies, errRadarrCalendar = fetchRadarrCalendarWithRetry(ctx, cfg, weekEnd, weekEnd.AddDate(0, 0, 7), cfg.MaxRetries)
			if errRadarrCalendar != nil {
				log.Printf("⚠️  Radarr calendar error: %v", errRadarrCalendar)
			} else {
				log.Printf("✓ Found %d upcoming movies", len(upcomingMovies))
			}
		}()
	} else {
		log.Println("🎬 Radarr not configured, skipping...")
	}

	// Fetch Trakt data if enabled
	if cfg.ShowTraktAnticipatedSeries {
		go func() {
			defer wg.Done()
			log.Println("🔥 Fetching Trakt anticipated series...")
			series, err := fetchTraktAnticipatedSeries(ctx, cfg)
			if err != nil {
				log.Printf("⚠️  Trakt anticipated series error: %v", err)
			} else {
				traktAnticipatedSeries = series
				log.Printf("✓ Found %d anticipated series", len(series))
			}
		}()
	}

	if cfg.ShowTraktWatchedSeries {
		go func() {
			defer wg.Done()
			log.Println("👀 Fetching Trakt watched series...")
			series, err := fetchTraktWatchedSeries(ctx, cfg)
			if err != nil {
				log.Printf("⚠️  Trakt watched series error: %v", err)
			} else {
				traktWatchedSeries = series
				log.Printf("✓ Found %d watched series", len(series))
			}
		}()
	}

	if cfg.ShowTraktAnticipatedMovies {
		go func() {
			defer wg.Done()
			log.Println("🔥 Fetching Trakt anticipated movies...")
			movies, err := fetchTraktAnticipatedMovies(ctx, cfg)
			if err != nil {
				log.Printf("⚠️  Trakt anticipated movies error: %v", err)
			} else {
				traktAnticipatedMovies = movies
				log.Printf("✓ Found %d anticipated movies", len(movies))
			}
		}()
	}

	if cfg.ShowTraktWatchedMovies {
		go func() {
			defer wg.Done()
			log.Println("👀 Fetching Trakt watched movies...")
			movies, err := fetchTraktWatchedMovies(ctx, cfg)
			if err != nil {
				log.Printf("⚠️  Trakt watched movies error: %v", err)
			} else {
				traktWatchedMovies = movies
				log.Printf("✓ Found %d watched movies", len(movies))
			}
		}()
	}

	wg.Wait()
	fetchDuration := time.Since(startFetch)
	log.Printf("⚡ All data fetched in %v (parallel)", fetchDuration)

	// Check for partial failures and provide graceful degradation
	failedServices := []string{}
	workingServices := []string{}

	if errSonarrHistory != nil || errSonarrCalendar != nil {
		failedServices = append(failedServices, "Sonarr")
	} else {
		workingServices = append(workingServices, "Sonarr")
	}

	if errRadarrHistory != nil || errRadarrCalendar != nil {
		failedServices = append(failedServices, "Radarr")
	} else {
		workingServices = append(workingServices, "Radarr")
	}

	// Log graceful degradation status
	if len(failedServices) > 0 && len(workingServices) > 0 {
		log.Printf("⚠️  Graceful degradation: %s failed, continuing with %s only",
			strings.Join(failedServices, ", "),
			strings.Join(workingServices, ", "))
	} else if len(failedServices) > 0 && len(workingServices) == 0 {
		log.Printf("❌ All services failed - cannot generate newsletter")
		return
	}

	// Filter unmonitored items from next week releases only (last week already downloaded)
	if !cfg.ShowUnmonitored {
		log.Println("📋 Filtering out unmonitored items from upcoming releases...")
		upcomingEpisodes = filterMonitoredEpisodes(upcomingEpisodes)
		upcomingMovies = filterMonitoredMovies(upcomingMovies)
		log.Printf("✓ After filtering: %d upcoming episodes, %d upcoming movies",
			len(upcomingEpisodes), len(upcomingMovies))
	}

	// Check if we have any content to send
	hasContent := len(upcomingEpisodes) > 0 || len(upcomingMovies) > 0 ||
		(cfg.ShowDownloaded && (len(downloadedEpisodes) > 0 || len(downloadedMovies) > 0))

	if !hasContent {
		log.Println("ℹ️  No new content to report. Skipping email.")
		return
	}

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
		TraktAnticipatedSeries: traktAnticipatedSeries,
		TraktWatchedSeries:     traktWatchedSeries,
		TraktAnticipatedMovies: traktAnticipatedMovies,
		TraktWatchedMovies:     traktWatchedMovies,
		// Customizable strings
		EmailTitle:                cfg.EmailTitle,
		EmailIntro:                cfg.EmailIntro,
		WeekRangePrefix:           cfg.WeekRangePrefix,
		ComingThisWeekHeading:     cfg.ComingThisWeekHeading,
		TVShowsHeading:            cfg.TVShowsHeading,
		MoviesHeading:             cfg.MoviesHeading,
		NoShowsMessage:            cfg.NoShowsMessage,
		NoMoviesMessage:           cfg.NoMoviesMessage,
		DownloadedSectionHeading:  cfg.DownloadedSectionHeading,
		NoDownloadedShowsMessage:  cfg.NoDownloadedShowsMessage,
		NoDownloadedMoviesMessage: cfg.NoDownloadedMoviesMessage,
		TrendingSectionHeading:    cfg.TrendingSectionHeading,
		AnticipatedSeriesHeading:  cfg.AnticipatedSeriesHeading,
		WatchedSeriesHeading:      cfg.WatchedSeriesHeading,
		AnticipatedMoviesHeading:  cfg.AnticipatedMoviesHeading,
		WatchedMoviesHeading:      cfg.WatchedMoviesHeading,
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

	log.Println("📝 Generating newsletter HTML...")
	html, err := generateNewsletterHTML(data, cfg)
	if err != nil {
		log.Fatalf("❌ Failed to generate HTML: %v", err)
	}

	subject := fmt.Sprintf("📺 Your Weekly Newsletter - %s", weekEnd.Format("January 2, 2006"))

	log.Println("📧 Sending emails...")
	if err := sendEmail(cfg, subject, html); err != nil {
		log.Fatalf("❌ Failed to send email: %v", err)
	}

	// Update statistics after successful send
	stats.mu.Lock()
	stats.TotalEmailsSent += len(cfg.ToEmails)
	stats.LastSentDate = now
	stats.LastSentDateStr = now.Format("2006-01-02 15:04:05 MST")
	stats.mu.Unlock()

	// Persist statistics to disk
	if err := saveStats(); err != nil {
		log.Printf("⚠️  Failed to save statistics: %v", err)
	}

	log.Println("✅ Newsletter sent successfully!")

	// Clear data to free memory immediately
	downloadedEpisodes = nil
	upcomingEpisodes = nil
	downloadedMovies = nil
	upcomingMovies = nil
	data = NewsletterData{}
}

// Generate newsletter HTML using precompiled template
func generateNewsletterHTML(data NewsletterData, cfg *Config) (string, error) {
	var buf bytes.Buffer
	if err := emailTemplate.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// Send email (with batch support for large recipient lists)
func sendEmail(cfg *Config, subject, htmlBody string) error {
	if cfg.FromEmail == "" || len(cfg.ToEmails) == 0 {
		return fmt.Errorf("email configuration incomplete")
	}

	// If recipients fit in one batch, send normally
	if len(cfg.ToEmails) <= cfg.EmailBatchSize {
		return sendEmailBatch(cfg, subject, htmlBody, cfg.ToEmails)
	}

	// Send in batches to avoid SMTP rate limits
	log.Printf("📨 Sending to %d recipients in batches of %d...", len(cfg.ToEmails), cfg.EmailBatchSize)

	for i := 0; i < len(cfg.ToEmails); i += cfg.EmailBatchSize {
		end := i + cfg.EmailBatchSize
		if end > len(cfg.ToEmails) {
			end = len(cfg.ToEmails)
		}
		batch := cfg.ToEmails[i:end]

		log.Printf("📧 Sending batch %d/%d (%d recipients)...",
			(i/cfg.EmailBatchSize)+1,
			(len(cfg.ToEmails)+cfg.EmailBatchSize-1)/cfg.EmailBatchSize,
			len(batch))

		if err := sendEmailBatch(cfg, subject, htmlBody, batch); err != nil {
			return fmt.Errorf("batch %d failed: %w", (i/cfg.EmailBatchSize)+1, err)
		}

		// Add delay between batches (except for the last batch)
		if end < len(cfg.ToEmails) {
			time.Sleep(time.Duration(cfg.EmailBatchDelay) * time.Second)
		}
	}

	log.Printf("✅ Successfully sent to all %d recipients", len(cfg.ToEmails))
	return nil
}

// sanitizeHeader removes CRLF characters to prevent email header injection
func sanitizeHeader(value string) string {
	return strings.NewReplacer("\r", "", "\n", "").Replace(value)
}

// Send email to a single batch of recipients with TLS enforcement
func sendEmailBatch(cfg *Config, subject, htmlBody string, recipients []string) error {
	from := cfg.FromEmail
	if cfg.FromName != "" {
		from = fmt.Sprintf("%s <%s>", cfg.FromName, cfg.FromEmail)
	}

	headers := make(map[string]string)
	headers["From"] = sanitizeHeader(from)
	headers["To"] = sanitizeHeader(strings.Join(recipients, ", "))
	headers["Subject"] = sanitizeHeader(subject)
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + htmlBody

	addr := fmt.Sprintf("%s:%s", cfg.SMTPHost, cfg.SMTPPort)

	// Connect to SMTP server
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer client.Close()

	// Send EHLO
	if err = client.Hello("localhost"); err != nil {
		return fmt.Errorf("EHLO failed: %w", err)
	}

	// Try STARTTLS if available (with secure TLS configuration)
	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{
			ServerName: cfg.SMTPHost,
			MinVersion: tls.VersionTLS12,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			},
		}
		if err = client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("STARTTLS failed: %w", err)
		}
	}

	// Authenticate
	auth := smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPHost)
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Set sender
	if err = client.Mail(cfg.FromEmail); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipients
	for _, recipient := range recipients {
		if err = client.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", recipient, err)
		}
	}

	// Send message body
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to open data writer: %w", err)
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	if err = w.Close(); err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return client.Quit()
}

// Precompile template with custom functions
func initEmailTemplate() (*template.Template, error) {
	return template.New("email.html").Funcs(template.FuncMap{
		"formatDateWithDay": formatDateWithDay,
		"truncate":          truncateString,
	}).ParseFS(templateFS, "templates/email.html")
}

// Wrapper to get next scheduled run
func getNextScheduledRun(day, timeStr string, loc *time.Location) string {
	now := time.Now().In(loc)

	// Parse schedule time
	parts := strings.Split(timeStr, ":")
	hour, minute := 9, 0
	if len(parts) == 2 {
		fmt.Sscanf(parts[0], "%d", &hour)
		fmt.Sscanf(parts[1], "%d", &minute)
	}

	// Map day to weekday
	dayMap := map[string]time.Weekday{
		"Mon": time.Monday,
		"Tue": time.Tuesday,
		"Wed": time.Wednesday,
		"Thu": time.Thursday,
		"Fri": time.Friday,
		"Sat": time.Saturday,
		"Sun": time.Sunday,
	}

	targetWeekday := dayMap[day]
	daysUntil := int(targetWeekday - now.Weekday())
	if daysUntil <= 0 {
		daysUntil += 7
	}

	nextRun := now.AddDate(0, 0, daysUntil)
	nextRun = time.Date(nextRun.Year(), nextRun.Month(), nextRun.Day(), hour, minute, 0, 0, loc)

	// If today is the day and time hasn't passed
	if now.Weekday() == targetWeekday {
		today := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, loc)
		if now.Before(today) {
			nextRun = today
		}
	}

	return nextRun.Format("2006-01-02 15:04:05 MST")
}

// Monitorable is a constraint for types that have a Monitored field
type Monitorable interface {
	Episode | Movie
}

// Generic filter function to exclude unmonitored items - eliminates code duplication
func filterMonitored[T Monitorable](items []T) []T {
	filtered := make([]T, 0, len(items))
	for _, item := range items {
		// Type switch to access Monitored field (Go generics limitation workaround)
		var monitored bool
		switch any(item).(type) {
		case Episode:
			monitored = any(item).(Episode).Monitored
		case Movie:
			monitored = any(item).(Movie).Monitored
		}
		if monitored {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// Convenience wrappers for backward compatibility and type safety
func filterMonitoredEpisodes(episodes []Episode) []Episode {
	return filterMonitored[Episode](episodes)
}

func filterMonitoredMovies(movies []Movie) []Movie {
	return filterMonitored[Movie](movies)
}
