package main

import (
	"embed"
	"flag"
	"html/template"
	"log"
	"net/http"
	"time"
)

// Embed static files to reduce memory and simplify deployment
//
//go:embed templates/*.html
var templateFS embed.FS

//go:embed assets/*
var assetsFS embed.FS

const version = "0.9.3"

// Track server start time for uptime monitoring
var startTime = time.Now()

// Global HTTP client (reused for all requests - 3-5x faster)
var httpClient = &http.Client{
	Timeout: 15 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 5,
		IdleConnTimeout:     90 * time.Second,
	},
}

// Precompiled templates (compiled once at startup)
var emailTemplate *template.Template

// Global statistics tracker
var stats = &Statistics{}

func init() {
	// Redirect log output to our ring buffer
	log.SetOutput(&logWriter{})
	log.SetFlags(log.Ldate | log.Ltime)
}

func main() {
	webMode := flag.Bool("web", false, "Run in web UI mode")
	flag.Parse()

	// Load config once at startup
	cachedConfig = loadConfig()

	// Load statistics from disk (persistent across restarts)
	if err := loadStats(); err != nil {
		log.Printf("⚠️  Could not load statistics: %v (starting fresh)", err)
	}

	// Start periodic cache cleanup to prevent unbounded memory growth
	// Clean up expired entries every 10 minutes (cache TTL is 5 minutes)
	apiCache.StartPeriodicCleanup(10 * time.Minute)

	// Validate configuration and display warnings
	warnings := validateConfig(cachedConfig)
	if len(warnings) > 0 {
		log.Println("⚠️  Configuration warnings:")
		for _, warning := range warnings {
			log.Printf("   - %s", warning)
		}
	}

	// Precompile email template with custom functions
	var err error
	emailTemplate, err = initEmailTemplate()
	if err != nil {
		log.Fatalf("❌ Failed to parse email template: %v", err)
	}

	if *webMode {
		startWebServer()
	} else {
		runNewsletter()
	}
}
