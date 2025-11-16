package main

import (
	"log"
	"os"
	"strings"
	"sync"
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
		MailgunSMTP:         getEnvFromFile(envMap, "MAILGUN_SMTP", "smtp.mailgun.org"),
		MailgunPort:         getEnvFromFile(envMap, "MAILGUN_PORT", "587"),
		MailgunUser:         getEnvFromFile(envMap, "MAILGUN_USER", ""),
		MailgunPass:         getEnvFromFile(envMap, "MAILGUN_PASS", ""),
		FromEmail:           getEnvFromFile(envMap, "FROM_EMAIL", ""),
		FromName:            getEnvFromFile(envMap, "FROM_NAME", "Newslettar"),
		ToEmails:            toEmails,
		Timezone:            getEnvFromFile(envMap, "TIMEZONE", "UTC"),
		ScheduleDay:         getEnvFromFile(envMap, "SCHEDULE_DAY", "Sun"),
		ScheduleTime:        getEnvFromFile(envMap, "SCHEDULE_TIME", "09:00"),
		ShowPosters:         getEnvFromFile(envMap, "SHOW_POSTERS", "true") != "false",
		ShowDownloaded:      getEnvFromFile(envMap, "SHOW_DOWNLOADED", "true") != "false",
		ShowQualityProfiles: getEnvFromFile(envMap, "SHOW_QUALITY_PROFILES", "false") != "false",
		ShowSeriesOverview:  getEnvFromFile(envMap, "SHOW_SERIES_OVERVIEW", "false") != "false",
		ShowEpisodeOverview: getEnvFromFile(envMap, "SHOW_EPISODE_OVERVIEW", "false") != "false",
		ShowUnmonitored:     getEnvFromFile(envMap, "SHOW_UNMONITORED", "false") != "false",
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
