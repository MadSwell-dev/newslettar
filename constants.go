package main

import "time"

// Configuration defaults
const (
	DefaultSMTPHost      = "smtp.mailgun.org"
	DefaultSMTPPort      = "587"
	DefaultFromName      = "Newslettar"
	DefaultTimezone      = "UTC"
	DefaultScheduleDay   = "Sun"
	DefaultScheduleTime  = "09:00"
	DefaultShowPosters   = "true"
	DefaultShowDownloaded = "true"
	DefaultShowSeriesOverview = "false"
	DefaultShowEpisodeOverview = "false"
	DefaultShowUnmonitored = "false"
)

// API and performance defaults
const (
	DefaultAPIPageSize      = 1000
	DefaultMaxRetries       = 3
	DefaultPreviewRetries   = 2
	DefaultAPITimeout       = 30 * time.Second
	DefaultHTTPTimeout      = 15 * time.Second
	DefaultMaxIdleConns     = 10
	DefaultMaxIdleConnsPerHost = 5
	DefaultIdleConnTimeout  = 90 * time.Second
)

// Log configuration
const (
	DefaultMaxLogLines = 500
)

// Server configuration
const (
	DefaultWebUIPort = "8080"
)
