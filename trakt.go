package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// Trakt API response structures
type traktShowResponse struct {
	Show struct {
		Title      string `json:"title"`
		Year       int    `json:"year"`
		FirstAired string `json:"first_aired"`
		Network    string `json:"network"`
		IDs        struct {
			Slug string `json:"slug"`
			TVDB int    `json:"tvdb"`
			IMDB string `json:"imdb"`
			TMDB int    `json:"tmdb"`
		} `json:"ids"`
		Overview string `json:"overview"`
	} `json:"show"`
}

type traktMovieResponse struct {
	Movie struct {
		Title    string `json:"title"`
		Year     int    `json:"year"`
		Released string `json:"released"`
		IDs      struct {
			Slug string `json:"slug"`
			IMDB string `json:"imdb"`
			TMDB int    `json:"tmdb"`
		} `json:"ids"`
		Overview string `json:"overview"`
	} `json:"movie"`
}

// fetchTraktAnticipatedSeries fetches the most anticipated series of the coming week from Trakt
func fetchTraktAnticipatedSeries(ctx context.Context, cfg *Config) ([]TraktShow, error) {
	if cfg.TraktClientID == "" {
		return nil, nil
	}

	// Check cache first
	cacheKey := getCacheKey("trakt_anticipated_series", cfg.TraktClientID, 0)
	if cached, found := apiCache.Get(cacheKey); found {
		log.Printf("📦 Using cached Trakt anticipated series")
		return cached.([]TraktShow), nil
	}

	url := "https://api.trakt.tv/shows/anticipated"
	shows, err := fetchTraktShows(ctx, cfg, url, true) // Filter to next week only
	if err != nil {
		return nil, err
	}

	// Cache for 5 minutes (same as other API calls)
	apiCache.Set(cacheKey, shows, cacheTTL)
	return shows, nil
}

// fetchTraktWatchedSeries fetches the most watched series of the last week from Trakt
func fetchTraktWatchedSeries(ctx context.Context, cfg *Config) ([]TraktShow, error) {
	if cfg.TraktClientID == "" {
		return nil, nil
	}

	// Check cache first
	cacheKey := getCacheKey("trakt_watched_series", cfg.TraktClientID, 0)
	if cached, found := apiCache.Get(cacheKey); found {
		log.Printf("📦 Using cached Trakt watched series")
		return cached.([]TraktShow), nil
	}

	url := "https://api.trakt.tv/shows/watched/weekly"
	shows, err := fetchTraktShows(ctx, cfg, url, false) // No date filtering for watched
	if err != nil {
		return nil, err
	}

	// Cache for 5 minutes (same as other API calls)
	apiCache.Set(cacheKey, shows, cacheTTL)
	return shows, nil
}

// fetchTraktAnticipatedMovies fetches the most anticipated movies of the coming week from Trakt
func fetchTraktAnticipatedMovies(ctx context.Context, cfg *Config) ([]TraktMovie, error) {
	if cfg.TraktClientID == "" {
		return nil, nil
	}

	// Check cache first
	cacheKey := getCacheKey("trakt_anticipated_movies", cfg.TraktClientID, 0)
	if cached, found := apiCache.Get(cacheKey); found {
		log.Printf("📦 Using cached Trakt anticipated movies")
		return cached.([]TraktMovie), nil
	}

	url := "https://api.trakt.tv/movies/anticipated"
	movies, err := fetchTraktMovies(ctx, cfg, url, true) // Filter to next week only
	if err != nil {
		return nil, err
	}

	// Cache for 5 minutes (same as other API calls)
	apiCache.Set(cacheKey, movies, cacheTTL)
	return movies, nil
}

// fetchTraktWatchedMovies fetches the most watched movies of the last week from Trakt
func fetchTraktWatchedMovies(ctx context.Context, cfg *Config) ([]TraktMovie, error) {
	if cfg.TraktClientID == "" {
		return nil, nil
	}

	// Check cache first
	cacheKey := getCacheKey("trakt_watched_movies", cfg.TraktClientID, 0)
	if cached, found := apiCache.Get(cacheKey); found {
		log.Printf("📦 Using cached Trakt watched movies")
		return cached.([]TraktMovie), nil
	}

	url := "https://api.trakt.tv/movies/watched/weekly"
	movies, err := fetchTraktMovies(ctx, cfg, url, false) // No date filtering for watched
	if err != nil {
		return nil, err
	}

	// Cache for 5 minutes (same as other API calls)
	apiCache.Set(cacheKey, movies, cacheTTL)
	return movies, nil
}

// fetchTraktShows is a helper function to fetch shows from Trakt API
func fetchTraktShows(ctx context.Context, cfg *Config, url string, filterToNextWeek bool) ([]TraktShow, error) {
	// Add extended parameter to get full details
	if len(url) > 0 && url[len(url)-1] != '?' {
		url += "?extended=full"
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Trakt request: %w", err)
	}

	// Trakt requires these headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("trakt-api-version", "2")
	req.Header.Set("trakt-api-key", cfg.TraktClientID)

	client := &http.Client{
		Timeout: time.Duration(cfg.APITimeout) * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from Trakt: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("trakt API error (status %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Trakt response: %w", err)
	}

	var responses []traktShowResponse
	if err := json.Unmarshal(body, &responses); err != nil {
		return nil, fmt.Errorf("failed to parse Trakt response: %w", err)
	}

	// Calculate next week's date range for filtering
	now := time.Now()
	nextWeekStart := now
	nextWeekEnd := now.AddDate(0, 0, 7)

	// Convert to our format
	shows := make([]TraktShow, 0, 5)
	for _, resp := range responses {
		if len(shows) >= 5 { // Limit to top 5
			break
		}

		// If filtering to next week and we have a first_aired date, check it
		if filterToNextWeek && resp.Show.FirstAired != "" {
			firstAired, err := time.Parse(time.RFC3339, resp.Show.FirstAired)
			if err == nil {
				// Skip if not premiering in the next 7 days
				if firstAired.Before(nextWeekStart) || firstAired.After(nextWeekEnd) {
					continue
				}
			}
		}

		show := TraktShow{
			Title:       resp.Show.Title,
			Year:        resp.Show.Year,
			Overview:    resp.Show.Overview,
			ReleaseDate: resp.Show.FirstAired,
			Network:     resp.Show.Network,
			IMDBID:      resp.Show.IDs.IMDB,
		}

		// Images are not available from Trakt API directly
		// Would need TMDB/TVDB API integration for posters
		show.ImageURL = ""

		shows = append(shows, show)
	}

	log.Printf("✅ Fetched %d shows from Trakt", len(shows))
	return shows, nil
}

// fetchTraktMovies is a helper function to fetch movies from Trakt API
func fetchTraktMovies(ctx context.Context, cfg *Config, url string, filterToNextWeek bool) ([]TraktMovie, error) {
	// Add extended parameter to get full details
	if len(url) > 0 && url[len(url)-1] != '?' {
		url += "?extended=full"
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Trakt request: %w", err)
	}

	// Trakt requires these headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("trakt-api-version", "2")
	req.Header.Set("trakt-api-key", cfg.TraktClientID)

	client := &http.Client{
		Timeout: time.Duration(cfg.APITimeout) * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from Trakt: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("trakt API error (status %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Trakt response: %w", err)
	}

	var responses []traktMovieResponse
	if err := json.Unmarshal(body, &responses); err != nil {
		return nil, fmt.Errorf("failed to parse Trakt response: %w", err)
	}

	// Calculate next week's date range for filtering
	now := time.Now()
	nextWeekStart := now
	nextWeekEnd := now.AddDate(0, 0, 7)

	// Convert to our format
	movies := make([]TraktMovie, 0, 5)
	for _, resp := range responses {
		if len(movies) >= 5 { // Limit to top 5
			break
		}

		// If filtering to next week and we have a released date, check it
		if filterToNextWeek && resp.Movie.Released != "" {
			released, err := time.Parse("2006-01-02", resp.Movie.Released)
			if err == nil {
				// Skip if not releasing in the next 7 days
				if released.Before(nextWeekStart) || released.After(nextWeekEnd) {
					continue
				}
			}
		}

		movie := TraktMovie{
			Title:       resp.Movie.Title,
			Year:        resp.Movie.Year,
			Overview:    resp.Movie.Overview,
			ReleaseDate: resp.Movie.Released,
			IMDBID:      resp.Movie.IDs.IMDB,
		}

		// Images are not available from Trakt API directly
		// Would need TMDB/TVDB API integration for posters
		movie.ImageURL = ""

		movies = append(movies, movie)
	}

	log.Printf("✅ Fetched %d movies from Trakt", len(movies))
	return movies, nil
}
