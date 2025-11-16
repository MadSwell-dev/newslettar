package main

import (
	"bytes"
	"context"
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Parallel API calls (3-4x faster!)
	var wg sync.WaitGroup
	var downloadedEpisodes, upcomingEpisodes []Episode
	var downloadedMovies, upcomingMovies []Movie
	var errSonarrHistory, errSonarrCalendar, errRadarrHistory, errRadarrCalendar error

	log.Println("📡 Fetching data in parallel...")
	startFetch := time.Now()

	wg.Add(4)

	go func() {
		defer wg.Done()
		log.Println("📺 Fetching Sonarr history...")
		downloadedEpisodes, errSonarrHistory = fetchSonarrHistoryWithRetry(ctx, cfg, weekStart, 3)
		if errSonarrHistory != nil {
			log.Printf("⚠️  Sonarr history error: %v", errSonarrHistory)
		} else {
			log.Printf("✓ Found %d downloaded episodes", len(downloadedEpisodes))
		}
	}()

	go func() {
		defer wg.Done()
		log.Println("📺 Fetching Sonarr calendar...")
		upcomingEpisodes, errSonarrCalendar = fetchSonarrCalendarWithRetry(ctx, cfg, weekEnd, weekEnd.AddDate(0, 0, 7), 3)
		if errSonarrCalendar != nil {
			log.Printf("⚠️  Sonarr calendar error: %v", errSonarrCalendar)
		} else {
			log.Printf("✓ Found %d upcoming episodes", len(upcomingEpisodes))
		}
	}()

	go func() {
		defer wg.Done()
		log.Println("🎬 Fetching Radarr history...")
		downloadedMovies, errRadarrHistory = fetchRadarrHistoryWithRetry(ctx, cfg, weekStart, 3)
		if errRadarrHistory != nil {
			log.Printf("⚠️  Radarr history error: %v", errRadarrHistory)
		} else {
			log.Printf("✓ Found %d downloaded movies", len(downloadedMovies))
		}
	}()

	go func() {
		defer wg.Done()
		log.Println("🎬 Fetching Radarr calendar...")
		upcomingMovies, errRadarrCalendar = fetchRadarrCalendarWithRetry(ctx, cfg, weekEnd, weekEnd.AddDate(0, 0, 7), 3)
		if errRadarrCalendar != nil {
			log.Printf("⚠️  Radarr calendar error: %v", errRadarrCalendar)
		} else {
			log.Printf("✓ Found %d upcoming movies", len(upcomingMovies))
		}
	}()

	wg.Wait()
	fetchDuration := time.Since(startFetch)
	log.Printf("⚡ All data fetched in %v (parallel)", fetchDuration)

	// Filter unmonitored items if the setting is disabled
	if !cfg.ShowUnmonitored {
		log.Println("📋 Filtering out unmonitored items...")
		upcomingEpisodes = filterMonitoredEpisodes(upcomingEpisodes)
		downloadedEpisodes = filterMonitoredEpisodes(downloadedEpisodes)
		upcomingMovies = filterMonitoredMovies(upcomingMovies)
		downloadedMovies = filterMonitoredMovies(downloadedMovies)
		log.Printf("✓ After filtering: %d upcoming episodes, %d downloaded episodes, %d upcoming movies, %d downloaded movies",
			len(upcomingEpisodes), len(downloadedEpisodes), len(upcomingMovies), len(downloadedMovies))
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
	}

	log.Println("📝 Generating newsletter HTML...")
	html, err := generateNewsletterHTML(data, cfg.ShowPosters, cfg.ShowDownloaded, cfg.ShowSeriesOverview, cfg.ShowEpisodeOverview)
	if err != nil {
		log.Fatalf("❌ Failed to generate HTML: %v", err)
	}

	subject := fmt.Sprintf("📺 Your Weekly Newsletter - %s", weekEnd.Format("January 2, 2006"))

	log.Println("📧 Sending emails...")
	if err := sendEmail(cfg, subject, html); err != nil {
		log.Fatalf("❌ Failed to send email: %v", err)
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
func generateNewsletterHTML(data NewsletterData, showPosters, showDownloaded, showSeriesOverview, showEpisodeOverview bool) (string, error) {
	templateData := struct {
		NewsletterData
		ShowPosters         bool
		ShowDownloaded      bool
		ShowSeriesOverview  bool
		ShowEpisodeOverview bool
	}{
		NewsletterData:      data,
		ShowPosters:         showPosters,
		ShowDownloaded:      showDownloaded,
		ShowSeriesOverview:  showSeriesOverview,
		ShowEpisodeOverview: showEpisodeOverview,
	}

	var buf bytes.Buffer
	if err := emailTemplate.Execute(&buf, templateData); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// Send email
func sendEmail(cfg *Config, subject, htmlBody string) error {
	if cfg.FromEmail == "" || len(cfg.ToEmails) == 0 {
		return fmt.Errorf("email configuration incomplete")
	}

	from := cfg.FromEmail
	if cfg.FromName != "" {
		from = fmt.Sprintf("%s <%s>", cfg.FromName, cfg.FromEmail)
	}

	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = strings.Join(cfg.ToEmails, ", ")
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + htmlBody

	auth := smtp.PlainAuth("", cfg.MailgunUser, cfg.MailgunPass, cfg.MailgunSMTP)
	addr := fmt.Sprintf("%s:%s", cfg.MailgunSMTP, cfg.MailgunPort)

	return smtp.SendMail(addr, auth, cfg.FromEmail, cfg.ToEmails, []byte(message))
}

// Precompile template with custom functions
func initEmailTemplate() (*template.Template, error) {
	return template.New("email.html").Funcs(template.FuncMap{
		"formatDateWithDay": formatDateWithDay,
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

	return nextRun.Format("Monday, January 2, 2006 at 3:04 PM MST")
}

// Filter functions to exclude unmonitored items
func filterMonitoredEpisodes(episodes []Episode) []Episode {
	filtered := []Episode{}
	for _, ep := range episodes {
		if ep.Monitored {
			filtered = append(filtered, ep)
		}
	}
	return filtered
}

func filterMonitoredMovies(movies []Movie) []Movie {
	filtered := []Movie{}
	for _, mv := range movies {
		if mv.Monitored {
			filtered = append(filtered, mv)
		}
	}
	return filtered
}
