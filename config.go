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

	return &Config{
		SonarrURL:           getEnvFromFile(envMap, "SONARR_URL", ""),
		SonarrAPIKey:        getEnvFromFile(envMap, "SONARR_API_KEY", ""),
		RadarrURL:           getEnvFromFile(envMap, "RADARR_URL", ""),
		RadarrAPIKey:        getEnvFromFile(envMap, "RADARR_API_KEY", ""),
		MailgunSMTP:         getEnvFromFile(envMap, "MAILGUN_SMTP", DefaultSMTPHost),
		MailgunPort:         getEnvFromFile(envMap, "MAILGUN_PORT", DefaultSMTPPort),
		MailgunUser:         getEnvFromFile(envMap, "MAILGUN_USER", ""),
		MailgunPass:         getEnvFromFile(envMap, "MAILGUN_PASS", ""),
		FromEmail:           getEnvFromFile(envMap, "FROM_EMAIL", ""),
		FromName:            getEnvFromFile(envMap, "FROM_NAME", DefaultFromName),
		ToEmails:            toEmails,
		Timezone:            getEnvFromFile(envMap, "TIMEZONE", DefaultTimezone),
		ScheduleDay:         getEnvFromFile(envMap, "SCHEDULE_DAY", DefaultScheduleDay),
		ScheduleTime:        getEnvFromFile(envMap, "SCHEDULE_TIME", DefaultScheduleTime),
		ShowPosters:         getEnvFromFile(envMap, "SHOW_POSTERS", DefaultShowPosters) != "false",
		ShowDownloaded:      getEnvFromFile(envMap, "SHOW_DOWNLOADED", DefaultShowDownloaded) != "false",
		ShowSeriesOverview:  getEnvFromFile(envMap, "SHOW_SERIES_OVERVIEW", DefaultShowSeriesOverview) != "false",
		ShowEpisodeOverview: getEnvFromFile(envMap, "SHOW_EPISODE_OVERVIEW", DefaultShowEpisodeOverview) != "false",
		ShowUnmonitored:     getEnvFromFile(envMap, "SHOW_UNMONITORED", DefaultShowUnmonitored) != "false",
		// Performance tuning - parse as integers with defaults
		APIPageSize:    getEnvIntFromFile(envMap, "API_PAGE_SIZE", DefaultAPIPageSize),
		MaxRetries:     getEnvIntFromFile(envMap, "MAX_RETRIES", DefaultMaxRetries),
		PreviewRetries: getEnvIntFromFile(envMap, "PREVIEW_RETRIES", DefaultPreviewRetries),
		APITimeout:     getEnvIntFromFile(envMap, "API_TIMEOUT", int(DefaultAPITimeout/time.Second)),
		WebUIPort:      getEnvFromFile(envMap, "WEBUI_PORT", DefaultWebUIPort),
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
	hasEmailConfig := cfg.MailgunUser != "" || cfg.MailgunPass != "" || cfg.FromEmail != "" || len(cfg.ToEmails) > 0
	if hasEmailConfig {
		if cfg.FromEmail == "" {
			warnings = append(warnings, "FROM_EMAIL is not set - email sending will fail")
		}
		if len(cfg.ToEmails) == 0 {
			warnings = append(warnings, "TO_EMAILS is not set - email sending will fail")
		}
		if cfg.MailgunUser == "" {
			warnings = append(warnings, "MAILGUN_USER is not set - email sending may fail")
		}
		if cfg.MailgunPass == "" {
			warnings = append(warnings, "MAILGUN_PASS is not set - email sending may fail")
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
