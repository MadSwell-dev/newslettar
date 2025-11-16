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

// Retry wrappers for API calls
func fetchSonarrHistoryWithRetry(ctx context.Context, cfg *Config, since time.Time, maxRetries int) ([]Episode, error) {
	var episodes []Episode
	var err error
	for i := 0; i < maxRetries; i++ {
		episodes, err = fetchSonarrHistory(ctx, cfg, since)
		if err == nil {
			return episodes, nil
		}
		if i < maxRetries-1 {
			wait := time.Duration(i+1) * time.Second
			log.Printf("⏳ Retrying Sonarr history in %v... (attempt %d/%d)", wait, i+2, maxRetries)
			time.Sleep(wait)
		}
	}
	return episodes, err
}

func fetchSonarrCalendarWithRetry(ctx context.Context, cfg *Config, start, end time.Time, maxRetries int) ([]Episode, error) {
	var episodes []Episode
	var err error
	for i := 0; i < maxRetries; i++ {
		episodes, err = fetchSonarrCalendar(ctx, cfg, start, end)
		if err == nil {
			return episodes, nil
		}
		if i < maxRetries-1 {
			wait := time.Duration(i+1) * time.Second
			log.Printf("⏳ Retrying Sonarr calendar in %v... (attempt %d/%d)", wait, i+2, maxRetries)
			time.Sleep(wait)
		}
	}
	return episodes, err
}

func fetchRadarrHistoryWithRetry(ctx context.Context, cfg *Config, since time.Time, maxRetries int) ([]Movie, error) {
	var movies []Movie
	var err error
	for i := 0; i < maxRetries; i++ {
		movies, err = fetchRadarrHistory(ctx, cfg, since)
		if err == nil {
			return movies, nil
		}
		if i < maxRetries-1 {
			wait := time.Duration(i+1) * time.Second
			log.Printf("⏳ Retrying Radarr history in %v... (attempt %d/%d)", wait, i+2, maxRetries)
			time.Sleep(wait)
		}
	}
	return movies, err
}

func fetchRadarrCalendarWithRetry(ctx context.Context, cfg *Config, start, end time.Time, maxRetries int) ([]Movie, error) {
	var movies []Movie
	var err error
	for i := 0; i < maxRetries; i++ {
		movies, err = fetchRadarrCalendar(ctx, cfg, start, end)
		if err == nil {
			return movies, nil
		}
		if i < maxRetries-1 {
			wait := time.Duration(i+1) * time.Second
			log.Printf("⏳ Retrying Radarr calendar in %v... (attempt %d/%d)", wait, i+2, maxRetries)
			time.Sleep(wait)
		}
	}
	return movies, err
}

func fetchSonarrHistory(ctx context.Context, cfg *Config, since time.Time) ([]Episode, error) {
	if cfg.SonarrURL == "" || cfg.SonarrAPIKey == "" {
		return nil, fmt.Errorf("Sonarr not configured")
	}

	url := fmt.Sprintf("%s/api/v3/history?pageSize=1000&sortKey=date&sortDirection=descending&includeEpisode=true&includeSeries=true", cfg.SonarrURL)
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
		body, _ := io.ReadAll(resp.Body)
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

	return episodes, nil
}

func fetchSonarrCalendar(ctx context.Context, cfg *Config, start, end time.Time) ([]Episode, error) {
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
		}

		if ep.AirDate != "" {
			airDate, _ := time.Parse("2006-01-02", ep.AirDate)
			ep.AirDate = airDate.Format("2006-01-02")
		}

		episodes = append(episodes, ep)
	}

	return episodes, nil
}

func fetchRadarrHistory(ctx context.Context, cfg *Config, since time.Time) ([]Movie, error) {
	if cfg.RadarrURL == "" || cfg.RadarrAPIKey == "" {
		return nil, fmt.Errorf("Radarr not configured")
	}

	url := fmt.Sprintf("%s/api/v3/history?pageSize=1000&sortKey=date&sortDirection=descending&includeMovie=true", cfg.RadarrURL)
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
		body, _ := io.ReadAll(resp.Body)
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

	return movies, nil
}

func fetchRadarrCalendar(ctx context.Context, cfg *Config, start, end time.Time) ([]Movie, error) {
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

		mv := Movie{
			Title:       entry.Title,
			Year:        entry.Year,
			ReleaseDate: entry.PhysicalRelease,
			PosterURL:   posterURL,
			IMDBID:      entry.ImdbId,
			TmdbID:      entry.TmdbId,
			Overview:    entry.Overview,
			Monitored:   entry.Monitored,
		}

		if mv.ReleaseDate != "" {
			releaseDate, _ := time.Parse("2006-01-02", mv.ReleaseDate)
			mv.ReleaseDate = releaseDate.Format("2006-01-02")
		}

		movies = append(movies, mv)
	}

	return movies, nil
}
