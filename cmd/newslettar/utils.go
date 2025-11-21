package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// Ring buffer for logs (no disk writes, configurable lines in memory)
var (
	logBuffer   []string
	logBufferMu sync.Mutex
	maxLogLines = DefaultMaxLogLines
)

// Custom log writer that maintains ring buffer
type logWriter struct{}

func (w *logWriter) Write(p []byte) (n int, err error) {
	logBufferMu.Lock()
	defer logBufferMu.Unlock()

	line := string(p)
	logBuffer = append(logBuffer, line)

	// Keep only last maxLogLines
	if len(logBuffer) > maxLogLines {
		logBuffer = logBuffer[len(logBuffer)-maxLogLines:]
	}

	// Also write to stdout for external logging if needed
	return os.Stdout.Write(p)
}

// Get timezone location
func getTimezone(tz string) *time.Location {
	if tz == "" {
		tz = DefaultTimezone
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		log.Printf("⚠️  Invalid timezone '%s', using %s", tz, DefaultTimezone)
		return time.UTC
	}
	return loc
}

// Group episodes by series
func groupEpisodesBySeries(episodes []Episode) []SeriesGroup {
	seriesMap := make(map[string]*SeriesGroup)

	// Sort episodes by air date first
	sort.Slice(episodes, func(i, j int) bool {
		return episodes[i].AirDate < episodes[j].AirDate
	})

	// Track seen episodes per series using map for O(1) lookups instead of O(N)
	type episodeKey struct {
		seriesTitle string
		seasonNum   int
		episodeNum  int
	}
	seenEpisodes := make(map[episodeKey]bool)

	for _, ep := range episodes {
		group, exists := seriesMap[ep.SeriesTitle]
		if !exists {
			group = &SeriesGroup{
				SeriesTitle:  ep.SeriesTitle,
				PosterURL:    ep.PosterURL,
				Episodes:     []Episode{},
				IMDBID:       ep.IMDBID,
				TvdbID:       ep.TvdbID,
				Overview:     ep.SeriesOverview,
				SeriesRating: ep.Rating, // Get series rating from first episode
			}
			seriesMap[ep.SeriesTitle] = group
		}

		// Check for duplicate episodes using O(1) map lookup instead of O(N) loop
		key := episodeKey{
			seriesTitle: ep.SeriesTitle,
			seasonNum:   ep.SeasonNum,
			episodeNum:  ep.EpisodeNum,
		}

		if !seenEpisodes[key] {
			seenEpisodes[key] = true
			// Clear episode rating since it's actually the series rating (episodes don't have individual ratings in Sonarr)
			ep.Rating = 0.0
			group.Episodes = append(group.Episodes, ep)
		}
	}

	groups := make([]SeriesGroup, 0, len(seriesMap))
	for _, group := range seriesMap {
		// Sort episodes within each series by season and episode number
		sort.Slice(group.Episodes, func(i, j int) bool {
			if group.Episodes[i].SeasonNum != group.Episodes[j].SeasonNum {
				return group.Episodes[i].SeasonNum < group.Episodes[j].SeasonNum
			}
			return group.Episodes[i].EpisodeNum < group.Episodes[j].EpisodeNum
		})
		groups = append(groups, *group)
	}

	sort.Slice(groups, func(i, j int) bool {
		return groups[i].SeriesTitle < groups[j].SeriesTitle
	})

	return groups
}

func formatDateWithDay(dateStr string) string {
	if dateStr == "" {
		return "Date TBA"
	}

	// Try RFC3339 first (ISO8601 with time: "2025-11-19T08:00:00.000Z")
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		// Try simple date format (YYYY-MM-DD)
		t, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return dateStr
		}
	}

	return t.Format("Monday, January 2, 2006")
}

// Truncate string to maxLength characters, adding "..." if truncated
func truncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength] + "..."
}

// Convert day/time to cron expression
func convertToCronExpression(day, timeStr, scheduleType string, dayOfMonth int) string {
	// Parse time (HH:MM)
	parts := strings.Split(timeStr, ":")
	hour := "9"
	minute := "0"
	if len(parts) == 2 {
		hour = parts[0]
		minute = parts[1]
	}

	// Handle monthly schedules
	if scheduleType == "monthly" {
		// Validate day of month (1-31)
		if dayOfMonth < 1 || dayOfMonth > 31 {
			dayOfMonth = 1
		}
		// Cron format for monthly: minute hour dayOfMonth * *
		return fmt.Sprintf("%s %s %d * *", minute, hour, dayOfMonth)
	}

	// Weekly schedule (default)
	// Convert day to cron weekday (0 = Sunday, 6 = Saturday)
	dayMap := map[string]string{
		"Sun": "0",
		"Mon": "1",
		"Tue": "2",
		"Wed": "3",
		"Thu": "4",
		"Fri": "5",
		"Sat": "6",
	}

	cronDay := dayMap[day]
	if cronDay == "" {
		cronDay = "0" // Default to Sunday
	}

	// Cron format: minute hour day month weekday
	return fmt.Sprintf("%s %s * * %s", minute, hour, cronDay)
}

func isNewerVersion(remote, current string) bool {
	remote = strings.TrimPrefix(remote, "v")
	current = strings.TrimPrefix(current, "v")

	remoteParts := strings.Split(remote, ".")
	currentParts := strings.Split(current, ".")

	maxLen := len(remoteParts)
	if len(currentParts) > maxLen {
		maxLen = len(currentParts)
	}

	for len(remoteParts) < maxLen {
		remoteParts = append(remoteParts, "0")
	}
	for len(currentParts) < maxLen {
		currentParts = append(currentParts, "0")
	}

	for i := 0; i < maxLen; i++ {
		var remoteNum, currentNum int
		fmt.Sscanf(remoteParts[i], "%d", &remoteNum)
		fmt.Sscanf(currentParts[i], "%d", &currentNum)

		if remoteNum > currentNum {
			return true
		} else if remoteNum < currentNum {
			return false
		}
	}

	return false
}

// Load statistics from disk (persistent across restarts)
func loadStats() error {
	const statsFile = ".stats.json"

	data, err := os.ReadFile(statsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No stats file yet, that's fine
		}
		return err
	}

	stats.mu.Lock()
	defer stats.mu.Unlock()

	var savedStats struct {
		TotalEmailsSent int    `json:"total_emails_sent"`
		LastSentDate    string `json:"last_sent_date"`
	}

	if err := json.Unmarshal(data, &savedStats); err != nil {
		return err
	}

	stats.TotalEmailsSent = savedStats.TotalEmailsSent
	stats.LastSentDateStr = savedStats.LastSentDate

	if savedStats.LastSentDate != "" {
		if t, err := time.Parse(time.RFC3339, savedStats.LastSentDate); err == nil {
			stats.LastSentDate = t
		}
	}

	log.Printf("✓ Loaded statistics: %d emails sent, last sent: %s", stats.TotalEmailsSent, stats.LastSentDateStr)
	return nil
}

// Save statistics to disk
func saveStats() error {
	const statsFile = ".stats.json"

	stats.mu.RLock()
	savedStats := struct {
		TotalEmailsSent int    `json:"total_emails_sent"`
		LastSentDate    string `json:"last_sent_date"`
	}{
		TotalEmailsSent: stats.TotalEmailsSent,
		LastSentDate:    stats.LastSentDateStr,
	}
	stats.mu.RUnlock()

	data, err := json.MarshalIndent(savedStats, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(statsFile, data, 0600)
}
