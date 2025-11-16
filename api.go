package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// Global API cache instance (5-minute TTL for preview generation)
var apiCache = NewAPICache()

// Cache TTL for API responses (makes previews instant if called within 5 minutes)
const cacheTTL = 5 * time.Minute

// Generate cache key from endpoint and parameters
func getCacheKey(endpoint string, params ...interface{}) string {
	key := fmt.Sprintf("%s:%v", endpoint, params)
	hash := sha256.Sum256([]byte(key))
	return fmt.Sprintf("%x", hash[:16]) // Use first 16 bytes of hash
}

// Generic retry wrapper - reduces code duplication from 67 lines to 20 lines
func retryWithBackoff[T any](operation func() (T, error), operationName string, maxRetries int) (T, error) {
	var result T
	var err error

	for i := 0; i < maxRetries; i++ {
		result, err = operation()
		if err == nil {
			return result, nil
		}

		if i < maxRetries-1 {
			wait := time.Duration(i+1) * time.Second
			log.Printf("⏳ Retrying %s in %v... (attempt %d/%d)", operationName, wait, i+2, maxRetries)
			time.Sleep(wait)
		}
	}

	return result, err
}

// Retry wrappers for API calls
func fetchSonarrHistoryWithRetry(ctx context.Context, cfg *Config, since time.Time, maxRetries int) ([]Episode, error) {
	return retryWithBackoff(func() ([]Episode, error) {
		return fetchSonarrHistory(ctx, cfg, since)
	}, "Sonarr history", maxRetries)
}

func fetchSonarrCalendarWithRetry(ctx context.Context, cfg *Config, start, end time.Time, maxRetries int) ([]Episode, error) {
	return retryWithBackoff(func() ([]Episode, error) {
		return fetchSonarrCalendar(ctx, cfg, start, end)
	}, "Sonarr calendar", maxRetries)
}

func fetchRadarrHistoryWithRetry(ctx context.Context, cfg *Config, since time.Time, maxRetries int) ([]Movie, error) {
	return retryWithBackoff(func() ([]Movie, error) {
		return fetchRadarrHistory(ctx, cfg, since)
	}, "Radarr history", maxRetries)
}

func fetchRadarrCalendarWithRetry(ctx context.Context, cfg *Config, start, end time.Time, maxRetries int) ([]Movie, error) {
	return retryWithBackoff(func() ([]Movie, error) {
		return fetchRadarrCalendar(ctx, cfg, start, end)
	}, "Radarr calendar", maxRetries)
}

func fetchSonarrHistory(ctx context.Context, cfg *Config, since time.Time) ([]Episode, error) {
	if cfg.SonarrURL == "" || cfg.SonarrAPIKey == "" {
		return nil, fmt.Errorf("Sonarr not configured")
	}

	// Check cache first
	cacheKey := getCacheKey("sonarr_history", cfg.SonarrURL, since.Unix())
	if cached, found := apiCache.Get(cacheKey); found {
		log.Printf("📦 Using cached Sonarr history")
		return cached.([]Episode), nil
	}

	url := fmt.Sprintf("%s/api/v3/history?pageSize=%d&sortKey=date&sortDirection=descending&includeEpisode=true&includeSeries=true", cfg.SonarrURL, cfg.APIPageSize)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Api-Key", cfg.SonarrAPIKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("HTTP %d (failed to read error body: %v)", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	// Stream JSON decoding (faster, less memory)
	var result struct {
		Records []struct {
			Date      time.Time `json:"date"`
			EventType string    `json:"eventType"`
			Series    struct {
				Title     string `json:"title"`
				TvdbID    int    `json:"tvdbId"`
				ImdbID    string `json:"imdbId"`
				Overview  string `json:"overview"`
				Monitored bool   `json:"monitored"`
				Images    []struct {
					CoverType string `json:"coverType"`
					RemoteURL string `json:"remoteUrl"`
				} `json:"images"`
			} `json:"series"`
			Episode struct {
				SeasonNumber  int    `json:"seasonNumber"`
				EpisodeNumber int    `json:"episodeNumber"`
				Title         string `json:"title"`
				AirDate       string `json:"airDate"`
				Overview      string `json:"overview"`
			} `json:"episode"`
		} `json:"records"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	episodes := []Episode{}
	for _, record := range result.Records {
		// Only include download events
		if record.EventType != "downloadFolderImported" && record.EventType != "downloadImported" {
			continue
		}

		// Filter by date
		if record.Date.Before(since) {
			continue
		}

		posterURL := ""
		for _, img := range record.Series.Images {
			if img.CoverType == "poster" {
				posterURL = img.RemoteURL
				break
			}
		}

		episodes = append(episodes, Episode{
			SeriesTitle:    record.Series.Title,
			SeasonNum:      record.Episode.SeasonNumber,
			EpisodeNum:     record.Episode.EpisodeNumber,
			Title:          record.Episode.Title,
			AirDate:        record.Episode.AirDate,
			Downloaded:     true,
			PosterURL:      posterURL,
			IMDBID:         record.Series.ImdbID,
			TvdbID:         record.Series.TvdbID,
			Overview:       record.Episode.Overview,
			SeriesOverview: record.Series.Overview,
			Monitored:      record.Series.Monitored,
		})
	}

	// Store in cache (reuse the cacheKey variable from above)
	apiCache.Set(cacheKey, episodes, cacheTTL)

	return episodes, nil
}

func fetchSonarrCalendar(ctx context.Context, cfg *Config, start, end time.Time) ([]Episode, error) {
	// Check cache first
	cacheKey := getCacheKey("sonarr_calendar", cfg.SonarrURL, start.Unix(), end.Unix())
	if cached, found := apiCache.Get(cacheKey); found {
		log.Printf("📦 Using cached Sonarr calendar")
		return cached.([]Episode), nil
	}

	url := fmt.Sprintf("%s/api/v3/calendar?unmonitored=true&includeSeries=true&includeEpisodeImages=true&start=%s&end=%s",
		cfg.SonarrURL, start.Format("2006-01-02"), end.Format("2006-01-02"))

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Api-Key", cfg.SonarrAPIKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	// Stream-decode JSON to save memory
	var calendar []CalendarEpisode
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&calendar); err != nil {
		return nil, err
	}

	// Map to Episode struct
	var episodes []Episode
	for _, entry := range calendar {
		posterURL := ""
		for _, img := range entry.Series.Images {
			if img.CoverType == "poster" {
				if img.Url != "" {
					posterURL = img.Url
				} else if img.RemoteUrl != "" {
					posterURL = img.RemoteUrl
				}
				break
			}
		}

		ep := Episode{
			SeriesTitle:    entry.Series.Title,
			SeasonNum:      entry.SeasonNumber,
			EpisodeNum:     entry.EpisodeNumber,
			Title:          entry.Title,
			AirDate:        entry.AirDate,
			PosterURL:      posterURL,
			IMDBID:         entry.Series.ImdbId,
			TvdbID:         entry.Series.TvdbId,
			Overview:       entry.Overview,
			SeriesOverview: entry.Series.Overview,
			Monitored:      entry.Series.Monitored,
			Rating:         entry.Series.Ratings.Value,
		}

		if ep.AirDate != "" {
			airDate, err := time.Parse("2006-01-02", ep.AirDate)
			if err == nil {
				ep.AirDate = airDate.Format("2006-01-02")
			}
			// If parsing fails, keep original date string
		}

		episodes = append(episodes, ep)
	}

	// Store in cache (reuse the cacheKey variable from above)
	apiCache.Set(cacheKey, episodes, cacheTTL)

	return episodes, nil
}

func fetchRadarrHistory(ctx context.Context, cfg *Config, since time.Time) ([]Movie, error) {
	if cfg.RadarrURL == "" || cfg.RadarrAPIKey == "" {
		return nil, fmt.Errorf("Radarr not configured")
	}

	// Check cache first
	cacheKey := getCacheKey("radarr_history", cfg.RadarrURL, since.Unix())
	if cached, found := apiCache.Get(cacheKey); found {
		log.Printf("📦 Using cached Radarr history")
		return cached.([]Movie), nil
	}

	url := fmt.Sprintf("%s/api/v3/history?pageSize=%d&sortKey=date&sortDirection=descending&includeMovie=true", cfg.RadarrURL, cfg.APIPageSize)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Api-Key", cfg.RadarrAPIKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("HTTP %d (failed to read error body: %v)", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Records []struct {
			Date      time.Time `json:"date"`
			EventType string    `json:"eventType"`
			Movie     struct {
				Title     string `json:"title"`
				Year      int    `json:"year"`
				TmdbID    int    `json:"tmdbId"`
				ImdbID    string `json:"imdbId"`
				InCinemas string `json:"inCinemas"`
				Overview  string `json:"overview"`
				Monitored bool   `json:"monitored"`
				Images    []struct {
					CoverType string `json:"coverType"`
					RemoteURL string `json:"remoteUrl"`
				} `json:"images"`
			} `json:"movie"`
		} `json:"records"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	movies := []Movie{}
	for _, record := range result.Records {
		// Only include download events
		if record.EventType != "downloadFolderImported" && record.EventType != "downloadImported" {
			continue
		}

		// Filter by date
		if record.Date.Before(since) {
			continue
		}

		posterURL := ""
		for _, img := range record.Movie.Images {
			if img.CoverType == "poster" {
				posterURL = img.RemoteURL
				break
			}
		}

		movies = append(movies, Movie{
			Title:       record.Movie.Title,
			Year:        record.Movie.Year,
			ReleaseDate: record.Movie.InCinemas,
			Downloaded:  true,
			PosterURL:   posterURL,
			IMDBID:      record.Movie.ImdbID,
			TmdbID:      record.Movie.TmdbID,
			Overview:    record.Movie.Overview,
			Monitored:   record.Movie.Monitored,
		})
	}

	// Store in cache (reuse the cacheKey variable from above)
	apiCache.Set(cacheKey, movies, cacheTTL)

	return movies, nil
}

func fetchRadarrCalendar(ctx context.Context, cfg *Config, start, end time.Time) ([]Movie, error) {
	// Check cache first
	cacheKey := getCacheKey("radarr_calendar", cfg.RadarrURL, start.Unix(), end.Unix())
	if cached, found := apiCache.Get(cacheKey); found {
		log.Printf("📦 Using cached Radarr calendar")
		return cached.([]Movie), nil
	}

	url := fmt.Sprintf("%s/api/v3/calendar?unmonitored=true&includeMovie=true&start=%s&end=%s",
		cfg.RadarrURL, start.Format("2006-01-02"), end.Format("2006-01-02"))

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Api-Key", cfg.RadarrAPIKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	// Stream-decode JSON to save memory
	var calendar []CalendarMovie
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&calendar); err != nil {
		return nil, err
	}

	// Map to Movie struct
	var movies []Movie
	for _, entry := range calendar {
		posterURL := ""
		for _, img := range entry.Images {
			if img.CoverType == "poster" {
				if img.Url != "" {
					posterURL = img.Url
				} else if img.RemoteUrl != "" {
					posterURL = img.RemoteUrl
				}
				break
			}
		}

		// Prefer IMDB rating, fallback to TMDB rating
		rating := entry.Ratings.Imdb.Value
		if rating == 0 {
			rating = entry.Ratings.Tmdb.Value
		}

		mv := Movie{
			Title:       entry.Title,
			Year:        entry.Year,
			ReleaseDate: entry.PhysicalRelease,
			PosterURL:   posterURL,
			IMDBID:      entry.ImdbId,
			TmdbID:      entry.TmdbId,
			Overview:    entry.Overview,
			Monitored:   entry.Monitored,
			Rating:      rating,
		}

		if mv.ReleaseDate != "" {
			releaseDate, err := time.Parse("2006-01-02", mv.ReleaseDate)
			if err == nil {
				mv.ReleaseDate = releaseDate.Format("2006-01-02")
			}
			// If parsing fails, keep original date string
		}

		movies = append(movies, mv)
	}

	// Store in cache (reuse the cacheKey variable from above)
	apiCache.Set(cacheKey, movies, cacheTTL)

	return movies, nil
}
