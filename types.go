package main

// Config structures
type Config struct {
	SonarrURL            string
	SonarrAPIKey         string
	RadarrURL            string
	RadarrAPIKey         string
	MailgunSMTP          string
	MailgunPort          string
	MailgunUser          string
	MailgunPass          string
	FromEmail            string
	FromName             string
	ToEmails             []string
	Timezone             string
	ScheduleDay          string
	ScheduleTime         string
	ShowPosters          bool
	ShowDownloaded       bool
	ShowQualityProfiles  bool
	ShowSeriesOverview   bool
	ShowEpisodeOverview  bool
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
	QualityProfile string
}

type Movie struct {
	Title          string
	Year           int
	ReleaseDate    string
	Downloaded     bool
	PosterURL      string
	IMDBID         string
	TmdbID         int
	Overview       string
	QualityProfile string
}

// For Sonarr calendar response (nested series data)
type CalendarEpisode struct {
	SeasonNumber  int    `json:"seasonNumber"`
	EpisodeNumber int    `json:"episodeNumber"`
	Title         string `json:"title"`
	AirDate       string `json:"airDate"`
	Overview      string `json:"overview"`
	Series        struct {
		Title          string `json:"title"`
		TvdbId         int    `json:"tvdbId"`
		ImdbId         string `json:"imdbId"`
		Overview       string `json:"overview"`
		QualityProfile struct {
			Name string `json:"name"`
		} `json:"qualityProfile"`
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
	QualityProfile  struct {
		Name string `json:"name"`
	} `json:"qualityProfile"`
	Images []struct {
		CoverType string `json:"coverType"`
		Url       string `json:"url"`       // Local URL if available
		RemoteUrl string `json:"remoteUrl"` // Fallback remote URL
	} `json:"images"`
}

type SeriesGroup struct {
	SeriesTitle string
	PosterURL   string
	Episodes    []Episode
	IMDBID      string
	TvdbID      int
	Overview    string
}

type NewsletterData struct {
	WeekStart              string
	WeekEnd                string
	UpcomingSeriesGroups   []SeriesGroup
	UpcomingMovies         []Movie
	DownloadedSeriesGroups []SeriesGroup
	DownloadedMovies       []Movie
}

type WebConfig struct {
	SonarrURL           string `json:"sonarr_url"`
	SonarrAPIKey        string `json:"sonarr_api_key"`
	RadarrURL           string `json:"radarr_url"`
	RadarrAPIKey        string `json:"radarr_api_key"`
	MailgunSMTP         string `json:"mailgun_smtp"`
	MailgunPort         string `json:"mailgun_port"`
	MailgunUser         string `json:"mailgun_user"`
	MailgunPass         string `json:"mailgun_pass"`
	FromEmail           string `json:"from_email"`
	FromName            string `json:"from_name"`
	ToEmails            string `json:"to_emails"`
	Timezone            string `json:"timezone"`
	ScheduleDay         string `json:"schedule_day"`
	ScheduleTime        string `json:"schedule_time"`
	ShowPosters         string `json:"show_posters"`
	ShowDownloaded      string `json:"show_downloaded"`
	ShowQualityProfiles string `json:"show_quality_profiles"`
	ShowSeriesOverview  string `json:"show_series_overview"`
	ShowEpisodeOverview string `json:"show_episode_overview"`
}
