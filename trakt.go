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
		Title string `json:"title"`
		Year  int    `json:"year"`
		IDs   struct {
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
		Title string `json:"title"`
		Year  int    `json:"year"`
		IDs   struct {
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
	shows, err := fetchTraktShows(ctx, cfg, url)
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
	shows, err := fetchTraktShows(ctx, cfg, url)
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
	movies, err := fetchTraktMovies(ctx, cfg, url)
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
	movies, err := fetchTraktMovies(ctx, cfg, url)
	if err != nil {
		return nil, err
	}

	// Cache for 5 minutes (same as other API calls)
	apiCache.Set(cacheKey, movies, cacheTTL)
	return movies, nil
}

// fetchTraktShows is a helper function to fetch shows from Trakt API
func fetchTraktShows(ctx context.Context, cfg *Config, url string) ([]TraktShow, error) {
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

	// Convert to our format and fetch images
	shows := make([]TraktShow, 0, len(responses))
	for i, resp := range responses {
		if i >= 10 { // Limit to top 10
			break
		}

		show := TraktShow{
			Title:    resp.Show.Title,
			Year:     resp.Show.Year,
			Overview: resp.Show.Overview,
		}

		// Try to get poster from TMDB if available
		if resp.Show.IDs.TMDB != 0 {
			show.ImageURL = fmt.Sprintf("https://image.tmdb.org/t/p/w500/%d", resp.Show.IDs.TMDB)
		}

		shows = append(shows, show)
	}

	log.Printf("✅ Fetched %d shows from Trakt", len(shows))
	return shows, nil
}

// fetchTraktMovies is a helper function to fetch movies from Trakt API
func fetchTraktMovies(ctx context.Context, cfg *Config, url string) ([]TraktMovie, error) {
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

	// Convert to our format and fetch images
	movies := make([]TraktMovie, 0, len(responses))
	for i, resp := range responses {
		if i >= 10 { // Limit to top 10
			break
		}

		movie := TraktMovie{
			Title:    resp.Movie.Title,
			Year:     resp.Movie.Year,
			Overview: resp.Movie.Overview,
		}

		// Try to get poster from TMDB if available
		if resp.Movie.IDs.TMDB != 0 {
			movie.ImageURL = fmt.Sprintf("https://image.tmdb.org/t/p/w500/%d", resp.Movie.IDs.TMDB)
		}

		movies = append(movies, movie)
	}

	log.Printf("✅ Fetched %d movies from Trakt", len(movies))
	return movies, nil
}
