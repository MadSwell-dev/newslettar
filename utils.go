package main

import (
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

// Log level priorities
var logLevels = map[string]int{
	"debug": 0,
	"info":  1,
	"warn":  2,
	"error": 3,
}

// Check if a message should be logged based on current log level
func shouldLog(messageLevel string) bool {
	cfg := getConfig()
	if cfg == nil {
		return true // Log everything if config not available
	}

	currentLevel := logLevels[strings.ToLower(cfg.LogLevel)]
	msgLevel := logLevels[strings.ToLower(messageLevel)]

	return msgLevel >= currentLevel
}

// Helper functions for different log levels
func logDebug(format string, v ...interface{}) {
	if shouldLog("debug") {
		log.Printf("[DEBUG] "+format, v...)
	}
}

func logInfo(format string, v ...interface{}) {
	if shouldLog("info") {
		log.Printf(format, v...)
	}
}

func logWarn(format string, v ...interface{}) {
	if shouldLog("warn") {
		log.Printf("⚠️  "+format, v...)
	}
}

func logError(format string, v ...interface{}) {
	if shouldLog("error") {
		log.Printf("❌ "+format, v...)
	}
}

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
				SeriesTitle: ep.SeriesTitle,
				PosterURL:   ep.PosterURL,
				Episodes:    []Episode{},
				IMDBID:      ep.IMDBID,
				TvdbID:      ep.TvdbID,
				Overview:    ep.SeriesOverview,
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
			group.Episodes = append(group.Episodes, ep)
		}
	}

	groups := make([]SeriesGroup, 0, len(seriesMap))
	for _, group := range seriesMap {
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

	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return dateStr
	}

	return t.Format("Monday, January 2, 2006")
}

// Convert day/time to cron expression
func convertToCronExpression(day, timeStr string) string {
	// Parse time (HH:MM)
	parts := strings.Split(timeStr, ":")
	hour := "9"
	minute := "0"
	if len(parts) == 2 {
		hour = parts[0]
		minute = parts[1]
	}

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
