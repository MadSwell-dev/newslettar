package main

import (
	"sync"
	"time"
)

// Cache structures for API responses
type CacheEntry struct {
	Data      interface{}
	ExpiresAt time.Time
}

type APICache struct {
	mu    sync.RWMutex
	cache map[string]CacheEntry
}

func NewAPICache() *APICache {
	return &APICache{
		cache: make(map[string]CacheEntry),
	}
}

func (c *APICache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.cache[key]
	if !exists || time.Now().After(entry.ExpiresAt) {
		return nil, false
	}
	return entry.Data, true
}

func (c *APICache) Set(key string, data interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[key] = CacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(ttl),
	}
}

func (c *APICache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]CacheEntry)
}

// Config structures
type Config struct {
	SonarrURL                   string
	SonarrAPIKey                string
	RadarrURL                   string
	RadarrAPIKey                string
	TraktClientID               string
	SMTPHost                    string
	SMTPPort                    string
	SMTPUser                    string
	SMTPPass                    string
	FromEmail                   string
	FromName                    string
	ToEmails                    []string
	Timezone                    string
	ScheduleDay                 string
	ScheduleTime                string
	ShowPosters                 bool
	ShowDownloaded              bool
	ShowSeriesOverview          bool
	ShowEpisodeOverview         bool
	ShowUnmonitored             bool
	ShowSeriesRatings           bool
	DarkMode                    bool
	ShowTraktAnticipatedSeries  bool
	ShowTraktWatchedSeries      bool
	ShowTraktAnticipatedMovies  bool
	ShowTraktWatchedMovies      bool
	TraktAnticipatedSeriesLimit int
	TraktWatchedSeriesLimit     int
	TraktAnticipatedMoviesLimit int
	TraktWatchedMoviesLimit     int
	// Performance tuning
	APIPageSize     int
	MaxRetries      int
	PreviewRetries  int
	APITimeout      int // in seconds
	WebUIPort       string
	EmailBatchSize  int    // Number of recipients per batch
	EmailBatchDelay int    // Delay between batches in seconds
	LogLevel        string // debug, info, warn, error
	// Customizable email strings
	EmailTitle                string
	EmailIntro                string
	WeekRangePrefix           string
	ComingThisWeekHeading     string
	TVShowsHeading            string
	MoviesHeading             string
	NoShowsMessage            string
	NoMoviesMessage           string
	DownloadedSectionHeading  string
	NoDownloadedShowsMessage  string
	NoDownloadedMoviesMessage string
	TrendingSectionHeading    string
	AnticipatedSeriesHeading  string
	WatchedSeriesHeading      string
	AnticipatedMoviesHeading  string
	WatchedMoviesHeading      string
	FooterText                string
}

// Minimal structs - only fields we actually need (reduces memory & JSON parsing time)
type Episode struct {
	SeriesTitle    string
	SeasonNum      int
	EpisodeNum     int
	Title          string
	AirDate        string
	Downloaded     bool
	PosterURL      string
	IMDBID         string
	TvdbID         int
	Overview       string
	SeriesOverview string
	Monitored      bool
	Rating         float64
}

type Movie struct {
	Title       string
	Year        int
	ReleaseDate string
	Downloaded  bool
	PosterURL   string
	IMDBID      string
	TmdbID      int
	Overview    string
	Monitored   bool
	Rating      float64
}

// For Sonarr calendar response (nested series data)
type CalendarEpisode struct {
	SeasonNumber  int    `json:"seasonNumber"`
	EpisodeNumber int    `json:"episodeNumber"`
	Title         string `json:"title"`
	AirDate       string `json:"airDate"`
	Overview      string `json:"overview"`
	Series        struct {
		Title     string `json:"title"`
		TvdbId    int    `json:"tvdbId"`
		ImdbId    string `json:"imdbId"`
		Overview  string `json:"overview"`
		Monitored bool   `json:"monitored"`
		Ratings   struct {
			Value float64 `json:"value"`
		} `json:"ratings"`
		Images []struct {
			CoverType string `json:"coverType"`
			Url       string `json:"url"`       // Local URL if available
			RemoteUrl string `json:"remoteUrl"` // Fallback remote URL
		} `json:"images"`
	} `json:"series"`
}

// For Radarr calendar response (direct fields + images array)
type CalendarMovie struct {
	Title           string `json:"title"`
	Year            int    `json:"year"`
	PhysicalRelease string `json:"physicalRelease"` // Assuming you want physical release; adjust if needed (e.g., to "digitalRelease" or "inCinemas")
	ImdbId          string `json:"imdbId"`
	TmdbId          int    `json:"tmdbId"`
	Overview        string `json:"overview"`
	Monitored       bool   `json:"monitored"`
	Ratings         struct {
		Imdb struct {
			Value float64 `json:"value"`
		} `json:"imdb"`
		Tmdb struct {
			Value float64 `json:"value"`
		} `json:"tmdb"`
	} `json:"ratings"`
	Images []struct {
		CoverType string `json:"coverType"`
		Url       string `json:"url"`       // Local URL if available
		RemoteUrl string `json:"remoteUrl"` // Fallback remote URL
	} `json:"images"`
}

type SeriesGroup struct {
	SeriesTitle  string
	PosterURL    string
	Episodes     []Episode
	IMDBID       string
	TvdbID       int
	Overview     string
	SeriesRating float64
}

type TraktShow struct {
	Title       string
	Year        int
	ImageURL    string
	Overview    string
	ReleaseDate string
	Network     string
	IMDBID      string
	Rating      float64
	InLibrary   bool
}

type TraktMovie struct {
	Title       string
	Year        int
	ImageURL    string
	Overview    string
	ReleaseDate string
	IMDBID      string
	Rating      float64
	InLibrary   bool
}

type NewsletterData struct {
	WeekStart              string
	WeekEnd                string
	UpcomingSeriesGroups   []SeriesGroup
	UpcomingMovies         []Movie
	DownloadedSeriesGroups []SeriesGroup
	DownloadedMovies       []Movie
	TraktAnticipatedSeries []TraktShow
	TraktWatchedSeries     []TraktShow
	TraktAnticipatedMovies []TraktMovie
	TraktWatchedMovies     []TraktMovie
	// Customizable strings
	EmailTitle                string
	EmailIntro                string
	WeekRangePrefix           string
	ComingThisWeekHeading     string
	TVShowsHeading            string
	MoviesHeading             string
	NoShowsMessage            string
	NoMoviesMessage           string
	DownloadedSectionHeading  string
	NoDownloadedShowsMessage  string
	NoDownloadedMoviesMessage string
	TrendingSectionHeading    string
	AnticipatedSeriesHeading  string
	WatchedSeriesHeading      string
	AnticipatedMoviesHeading  string
	WatchedMoviesHeading      string
	FooterText                string
	// Template display options (needed for template rendering)
	ShowPosters                bool
	ShowDownloaded             bool
	ShowSeriesOverview         bool
	ShowEpisodeOverview        bool
	ShowSeriesRatings          bool
	DarkMode                   bool
	ShowTraktAnticipatedSeries bool
	ShowTraktWatchedSeries     bool
	ShowTraktAnticipatedMovies bool
	ShowTraktWatchedMovies     bool
}

type WebConfig struct {
	SonarrURL                   string `json:"sonarr_url"`
	SonarrAPIKey                string `json:"sonarr_api_key"`
	RadarrURL                   string `json:"radarr_url"`
	RadarrAPIKey                string `json:"radarr_api_key"`
	TraktClientID               string `json:"trakt_client_id"`
	SMTPHost                    string `json:"smtp_host"`
	SMTPPort                    string `json:"smtp_port"`
	SMTPUser                    string `json:"smtp_user"`
	SMTPPass                    string `json:"smtp_pass"`
	FromEmail                   string `json:"from_email"`
	FromName                    string `json:"from_name"`
	ToEmails                    string `json:"to_emails"`
	Timezone                    string `json:"timezone"`
	ScheduleDay                 string `json:"schedule_day"`
	ScheduleTime                string `json:"schedule_time"`
	ShowPosters                 string `json:"show_posters"`
	ShowDownloaded              string `json:"show_downloaded"`
	ShowSeriesOverview          string `json:"show_series_overview"`
	ShowEpisodeOverview         string `json:"show_episode_overview"`
	ShowUnmonitored             string `json:"show_unmonitored"`
	ShowSeriesRatings           string `json:"show_series_ratings"`
	DarkMode                    string `json:"dark_mode"`
	ShowTraktAnticipatedSeries  string `json:"show_trakt_anticipated_series"`
	ShowTraktWatchedSeries      string `json:"show_trakt_watched_series"`
	ShowTraktAnticipatedMovies  string `json:"show_trakt_anticipated_movies"`
	ShowTraktWatchedMovies      string `json:"show_trakt_watched_movies"`
	TraktAnticipatedSeriesLimit string `json:"trakt_anticipated_series_limit"`
	TraktWatchedSeriesLimit     string `json:"trakt_watched_series_limit"`
	TraktAnticipatedMoviesLimit string `json:"trakt_anticipated_movies_limit"`
	TraktWatchedMoviesLimit     string `json:"trakt_watched_movies_limit"`
	// Customizable email strings
	EmailTitle                string `json:"email_title"`
	EmailIntro                string `json:"email_intro"`
	WeekRangePrefix           string `json:"week_range_prefix"`
	ComingThisWeekHeading     string `json:"coming_this_week_heading"`
	TVShowsHeading            string `json:"tv_shows_heading"`
	MoviesHeading             string `json:"movies_heading"`
	NoShowsMessage            string `json:"no_shows_message"`
	NoMoviesMessage           string `json:"no_movies_message"`
	DownloadedSectionHeading  string `json:"downloaded_section_heading"`
	NoDownloadedShowsMessage  string `json:"no_downloaded_shows_message"`
	NoDownloadedMoviesMessage string `json:"no_downloaded_movies_message"`
	TrendingSectionHeading    string `json:"trending_section_heading"`
	AnticipatedSeriesHeading  string `json:"anticipated_series_heading"`
	WatchedSeriesHeading      string `json:"watched_series_heading"`
	AnticipatedMoviesHeading  string `json:"anticipated_movies_heading"`
	WatchedMoviesHeading      string `json:"watched_movies_heading"`
	FooterText                string `json:"footer_text"`
}

// Statistics for dashboard
type Statistics struct {
	mu              sync.RWMutex
	TotalEmailsSent int       `json:"total_emails_sent"`
	LastSentDate    time.Time `json:"last_sent_date"`
	LastSentDateStr string    `json:"last_sent_date_str"`
}

// Dashboard data
type DashboardData struct {
	Version          string            `json:"version"`
	Uptime           string            `json:"uptime"`
	UptimeSeconds    int64             `json:"uptime_seconds"`
	MemoryUsageMB    float64           `json:"memory_usage_mb"`
	Port             string            `json:"port"`
	TotalEmailsSent  int               `json:"total_emails_sent"`
	LastSentDate     string            `json:"last_sent_date"`
	NextScheduledRun string            `json:"next_scheduled_run"`
	Timezone         string            `json:"timezone"`
	ServiceStatus    map[string]string `json:"service_status"`
}
