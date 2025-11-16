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

const version = "1.1.3"

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
