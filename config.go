package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Global config cache (loaded once at startup, reloaded on save)
var (
	configMu     sync.RWMutex
	cachedConfig *Config
)

// Get config (cached, thread-safe)
func getConfig() *Config {
	configMu.RLock()
	defer configMu.RUnlock()
	return cachedConfig
}

// Reload config (called when user saves configuration)
func reloadConfig() {
	configMu.Lock()
	defer configMu.Unlock()
	cachedConfig = loadConfig()
	log.Println("🔄 Configuration reloaded from .env")
}

// Load configuration from .env file (only called at startup and on reload)
func loadConfig() *Config {
	envMap := readEnvFile()

	toEmailsStr := getEnvFromFile(envMap, "TO_EMAILS", "")
	toEmails := []string{}
	if toEmailsStr != "" {
		toEmails = strings.Split(toEmailsStr, ",")
		for i := range toEmails {
			toEmails[i] = strings.TrimSpace(toEmails[i])
		}
	}

	// Support backward compatibility with old MAILGUN_* env vars
	smtpHost := getEnvFromFile(envMap, "SMTP_HOST", "")
	if smtpHost == "" {
		smtpHost = getEnvFromFile(envMap, "MAILGUN_SMTP", DefaultSMTPHost)
	}
	smtpPort := getEnvFromFile(envMap, "SMTP_PORT", "")
	if smtpPort == "" {
		smtpPort = getEnvFromFile(envMap, "MAILGUN_PORT", DefaultSMTPPort)
	}
	smtpUser := getEnvFromFile(envMap, "SMTP_USER", "")
	if smtpUser == "" {
		smtpUser = getEnvFromFile(envMap, "MAILGUN_USER", "")
	}
	smtpPass := getEnvFromFile(envMap, "SMTP_PASS", "")
	if smtpPass == "" {
		smtpPass = getEnvFromFile(envMap, "MAILGUN_PASS", "")
	}

	return &Config{
		SonarrURL:                  getEnvFromFileOnly(envMap, "SONARR_URL", ""),
		SonarrAPIKey:               getEnvFromFileOnly(envMap, "SONARR_API_KEY", ""),
		RadarrURL:                  getEnvFromFileOnly(envMap, "RADARR_URL", ""),
		RadarrAPIKey:               getEnvFromFileOnly(envMap, "RADARR_API_KEY", ""),
		TraktClientID:              getEnvFromFileOnly(envMap, "TRAKT_CLIENT_ID", ""),
		SMTPHost:                   smtpHost,
		SMTPPort:                   smtpPort,
		SMTPUser:                   smtpUser,
		SMTPPass:                   smtpPass,
		FromEmail:                  getEnvFromFile(envMap, "FROM_EMAIL", ""),
		FromName:                   getEnvFromFile(envMap, "FROM_NAME", DefaultFromName),
		ToEmails:                   toEmails,
		Timezone:                   getEnvFromFile(envMap, "TIMEZONE", DefaultTimezone),
		ScheduleDay:                getEnvFromFile(envMap, "SCHEDULE_DAY", DefaultScheduleDay),
		ScheduleTime:               getEnvFromFile(envMap, "SCHEDULE_TIME", DefaultScheduleTime),
		ShowPosters:                getEnvFromFile(envMap, "SHOW_POSTERS", DefaultShowPosters) != "false",
		ShowDownloaded:             getEnvFromFile(envMap, "SHOW_DOWNLOADED", DefaultShowDownloaded) != "false",
		ShowSeriesOverview:         getEnvFromFile(envMap, "SHOW_SERIES_OVERVIEW", DefaultShowSeriesOverview) != "false",
		ShowEpisodeOverview:        getEnvFromFile(envMap, "SHOW_EPISODE_OVERVIEW", DefaultShowEpisodeOverview) != "false",
		ShowUnmonitored:            getEnvFromFile(envMap, "SHOW_UNMONITORED", DefaultShowUnmonitored) != "false",
		DarkMode:                   getEnvFromFile(envMap, "DARK_MODE", DefaultDarkMode) != "false",
		ShowTraktAnticipatedSeries: getEnvFromFile(envMap, "SHOW_TRAKT_ANTICIPATED_SERIES", DefaultShowTraktAnticipatedSeries) != "false",
		ShowTraktWatchedSeries:     getEnvFromFile(envMap, "SHOW_TRAKT_WATCHED_SERIES", DefaultShowTraktWatchedSeries) != "false",
		ShowTraktAnticipatedMovies: getEnvFromFile(envMap, "SHOW_TRAKT_ANTICIPATED_MOVIES", DefaultShowTraktAnticipatedMovies) != "false",
		ShowTraktWatchedMovies:     getEnvFromFile(envMap, "SHOW_TRAKT_WATCHED_MOVIES", DefaultShowTraktWatchedMovies) != "false",
		// Performance tuning - parse as integers with defaults
		APIPageSize:     getEnvIntFromFile(envMap, "API_PAGE_SIZE", DefaultAPIPageSize),
		MaxRetries:      getEnvIntFromFile(envMap, "MAX_RETRIES", DefaultMaxRetries),
		PreviewRetries:  getEnvIntFromFile(envMap, "PREVIEW_RETRIES", DefaultPreviewRetries),
		APITimeout:      getEnvIntFromFile(envMap, "API_TIMEOUT", int(DefaultAPITimeout/time.Second)),
		WebUIPort:       getEnvFromFile(envMap, "WEBUI_PORT", DefaultWebUIPort),
		EmailBatchSize:  getEnvIntFromFile(envMap, "EMAIL_BATCH_SIZE", DefaultEmailBatchSize),
		EmailBatchDelay: getEnvIntFromFile(envMap, "EMAIL_BATCH_DELAY", int(DefaultEmailBatchDelay/time.Second)),
		LogLevel:        getEnvFromFile(envMap, "LOG_LEVEL", DefaultLogLevel),
	}
}

func readEnvFile() map[string]string {
	envMap := make(map[string]string)

	data, err := os.ReadFile(".env")
	if err != nil {
		return envMap
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			envMap[key] = value
		}
	}

	return envMap
}

func getEnvFromFile(envMap map[string]string, key, defaultValue string) string {
	if val, exists := envMap[key]; exists {
		return val
	}
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

// getEnvFromFileOnly reads only from the env file, not from system environment variables
// Use this for user-configurable fields that can be deleted via UI
func getEnvFromFileOnly(envMap map[string]string, key, defaultValue string) string {
	if val, exists := envMap[key]; exists {
		return val
	}
	return defaultValue
}

func getEnvIntFromFile(envMap map[string]string, key string, defaultValue int) int {
	strVal := getEnvFromFile(envMap, key, "")
	if strVal == "" {
		return defaultValue
	}
	intVal, err := strconv.Atoi(strVal)
	if err != nil {
		log.Printf("⚠️  Invalid integer value for %s: %s, using default %d", key, strVal, defaultValue)
		return defaultValue
	}
	return intVal
}

// Validate configuration and return warnings/errors
func validateConfig(cfg *Config) []string {
	var warnings []string

	// Check for email configuration (only warn if partially configured)
	hasEmailConfig := cfg.SMTPUser != "" || cfg.SMTPPass != "" || cfg.FromEmail != "" || len(cfg.ToEmails) > 0
	if hasEmailConfig {
		if cfg.FromEmail == "" {
			warnings = append(warnings, "FROM_EMAIL is not set - email sending will fail")
		}
		if len(cfg.ToEmails) == 0 {
			warnings = append(warnings, "TO_EMAILS is not set - email sending will fail")
		}
		if cfg.SMTPUser == "" {
			warnings = append(warnings, "SMTP_USER is not set - email sending may fail")
		}
		if cfg.SMTPPass == "" {
			warnings = append(warnings, "SMTP_PASS is not set - email sending may fail")
		}
	}

	// Check for API configuration (warn if none configured)
	hasSonarr := cfg.SonarrURL != "" && cfg.SonarrAPIKey != ""
	hasRadarr := cfg.RadarrURL != "" && cfg.RadarrAPIKey != ""

	if !hasSonarr && !hasRadarr {
		warnings = append(warnings, "Neither Sonarr nor Radarr is configured - newsletter will have no content")
	}

	// Warn about partial configuration
	if cfg.SonarrURL != "" && cfg.SonarrAPIKey == "" {
		warnings = append(warnings, "SONARR_URL is set but SONARR_API_KEY is missing")
	}
	if cfg.SonarrAPIKey != "" && cfg.SonarrURL == "" {
		warnings = append(warnings, "SONARR_API_KEY is set but SONARR_URL is missing")
	}
	if cfg.RadarrURL != "" && cfg.RadarrAPIKey == "" {
		warnings = append(warnings, "RADARR_URL is set but RADARR_API_KEY is missing")
	}
	if cfg.RadarrAPIKey != "" && cfg.RadarrURL == "" {
		warnings = append(warnings, "RADARR_API_KEY is set but RADARR_URL is missing")
	}

	// Validate timezone
	if _, err := time.LoadLocation(cfg.Timezone); err != nil {
		warnings = append(warnings, "Invalid TIMEZONE '"+cfg.Timezone+"' - using UTC")
	}

	return warnings
}
