package main

import "time"

// Configuration defaults
const (
	// SMTP defaults (Mailgun as reasonable default, but works with any SMTP provider)
	DefaultSMTPHost                   = "smtp.mailgun.org"
	DefaultSMTPPort                   = "587"
	DefaultFromName                   = "Newslettar"
	DefaultTimezone                   = "UTC"
	DefaultScheduleDay                = "Sun"
	DefaultScheduleTime               = "09:00"
	DefaultShowPosters                = "true"
	DefaultShowDownloaded             = "true"
	DefaultShowSeriesOverview         = "false"
	DefaultShowEpisodeOverview        = "false"
	DefaultShowUnmonitored            = "false"
	DefaultShowSeriesRatings          = "false"
	DefaultShowEpisodeRatings         = "false"
	DefaultDarkMode                   = "true"
	DefaultShowTraktAnticipatedSeries = "false"
	DefaultShowTraktWatchedSeries     = "false"
	DefaultShowTraktAnticipatedMovies = "false"
	DefaultShowTraktWatchedMovies     = "false"
)

// API and performance defaults
const (
	DefaultAPIPageSize         = 1000
	DefaultMaxRetries          = 3
	DefaultPreviewRetries      = 2
	DefaultAPITimeout          = 30 * time.Second
	DefaultHTTPTimeout         = 15 * time.Second
	DefaultMaxIdleConns        = 10
	DefaultMaxIdleConnsPerHost = 5
	DefaultIdleConnTimeout     = 90 * time.Second
	DefaultEmailBatchSize      = 10
	DefaultEmailBatchDelay     = 1 * time.Second
)

// Log configuration
const (
	DefaultMaxLogLines = 500
	DefaultLogLevel    = "info" // debug, info, warn, error
)

// Server configuration
const (
	DefaultWebUIPort = "8080"
)
