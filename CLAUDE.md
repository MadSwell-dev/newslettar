# CLAUDE.md - AI Assistant Guide for Newslettar

> **Last Updated:** 2025-11-17 (v0.6.1)
> **Purpose:** Comprehensive guide for AI assistants working on the Newslettar codebase

---

## 📋 Table of Contents

1. [Project Overview](#project-overview)
2. [Architecture & Design Patterns](#architecture--design-patterns)
3. [File Structure & Responsibilities](#file-structure--responsibilities)
4. [Development Workflow](#development-workflow)
5. [Coding Conventions & Standards](#coding-conventions--standards)
6. [Common Tasks](#common-tasks)
7. [Testing & Quality](#testing--quality)
8. [Important Considerations](#important-considerations)
9. [Git Workflow](#git-workflow)

---

## 🎯 Project Overview

**Newslettar** is an automated newsletter generator for Sonarr and Radarr media servers, written in Go.

### Key Characteristics

- **Language:** Go 1.23+
- **Architecture:** Monolithic single-binary design
- **Size:** ~4,400 lines of Go code across 11 files
- **Memory:** ~12MB RAM usage at runtime
- **Dependencies:** Minimal (only `github.com/robfig/cron/v3`)
- **Deployment:** Docker, Debian packages, or direct binary
- **License:** MIT

### Core Functionality

1. **Fetches media data** from Sonarr/Radarr APIs (parallel API calls)
2. **Integrates with Trakt.tv** for trending content (optional)
3. **Generates HTML newsletters** using Go templates
4. **Sends scheduled emails** via SMTP with internal cron
5. **Provides Web UI** for configuration and testing on port 8080
6. **Dashboard view** with system stats, newsletter metrics, service status, and live logs

### Design Philosophy

- **Performance-first:** Parallel API calls, connection pooling, caching
- **Memory-efficient:** Embedded assets, ring buffer logging, minimal allocations
- **Self-contained:** Single binary with no external dependencies at runtime
- **Graceful degradation:** Works with partial data if services unavailable

---

## 🏗️ Architecture & Design Patterns

### Overall Architecture

```
┌─────────────────────────────────────────────────────────┐
│                      main.go                             │
│              (Entry point & initialization)              │
└───────────────────┬─────────────────────────────────────┘
                    │
        ┌───────────┴──────────┐
        │                      │
        ▼                      ▼
┌──────────────┐      ┌──────────────┐
│  Web Mode    │      │  CLI Mode    │
│  (server.go) │      │ (newsletter) │
└──────┬───────┘      └──────┬───────┘
       │                     │
       │  HTTP Handler       │  Direct Call
       │  Trigger            │
       ▼                     ▼
┌────────────────────────────────────┐
│      newsletter.go                 │
│  (Newsletter generation logic)     │
└────────┬───────────────────────────┘
         │
         │  Parallel API Calls
         │
    ┌────┴─────┬──────────┬────────┐
    ▼          ▼          ▼        ▼
┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐
│Sonarr  │ │Radarr  │ │Trakt   │ │SMTP    │
│api.go  │ │api.go  │ │trakt.go│ │utils.go│
└────────┘ └────────┘ └────────┘ └────────┘
```

### Key Design Patterns

#### 1. **Goroutines & WaitGroup (Concurrency)**

**Location:** `newsletter.go:34-166`

All API calls run in parallel using goroutines with `sync.WaitGroup`:

```go
var wg sync.WaitGroup
wg.Add(4) // Base API calls

go func() {
    defer wg.Done()
    // Fetch Sonarr history
}()

go func() {
    defer wg.Done()
    // Fetch Sonarr calendar
}()

// ... similar for Radarr, Trakt

wg.Wait() // Block until all complete
```

**Performance Impact:** 3-4x faster than sequential API calls

#### 2. **Generic Retry with Exponential Backoff**

**Location:** `api.go:27-46`

Uses Go generics for type-safe retry logic:

```go
func retryWithBackoff[T any](
    operation func() (T, error),
    operationName string,
    maxRetries int
) (T, error)
```

**Backoff Schedule:** 1s, 2s, 4s (exponential)

#### 3. **Thread-Safe Caching with TTL**

**Location:** `types.go:8-51`, implemented in `api.go` and `trakt.go`

```go
type APICache struct {
    mu    sync.RWMutex
    cache map[string]CacheEntry
}

type CacheEntry struct {
    Value      interface{}
    Expiration time.Time
}
```

**Cache Keys:** SHA256 hash of (URL + params)
**TTL:** 5 minutes (for preview generation speed)

#### 4. **Singleton Pattern**

**Global Instances:**
- HTTP client: `main.go:22-30` (connection pooling)
- Config cache: `config.go:13-23` (thread-safe)
- API cache: `api.go:15` (single instance)
- Cron scheduler: `server.go:16` (shared across requests)

#### 5. **Generic Filtering**

**Location:** `newsletter.go:411-442`

Type-safe filtering using Go generics:

```go
type Monitorable interface {
    IsMonitored() bool
}

func filterMonitored[T Monitorable](items []T) []T
```

Eliminates code duplication for Episodes, Movies, Series filtering.

#### 6. **Ring Buffer Logging**

**Location:** `utils.go`

In-memory circular buffer for logs (no disk I/O):

```go
var logBuffer []string
var logBufferMu sync.Mutex
const maxLogLines = 500
```

**Benefits:**
- Zero disk writes
- Fast `/api/logs` endpoint
- Bounded memory usage

---

## 📁 File Structure & Responsibilities

### Core Go Files (4,396 total lines)

| File | Lines | Primary Responsibility |
|------|-------|------------------------|
| `main.go` | 69 | Entry point, initialization, global singletons |
| `server.go` | 91 | HTTP server, cron scheduler, graceful shutdown |
| `handlers.go` | 808 | HTTP request handlers, Web UI routes, dashboard API |
| `ui.go` | 1,365 | Embedded Web UI HTML templates |
| `newsletter.go` | 442 | Newsletter generation orchestration |
| `api.go` | 455 | Sonarr/Radarr API client with retry/cache |
| `trakt.go` | 498 | Trakt.tv API integration |
| `config.go` | 214 | Configuration loading, validation, reload |
| `types.go` | 251 | Data structures, cache implementation |
| `utils.go` | 246 | Email sending, logging, utility functions |
| `constants.go` | 50 | Application constants |

### File-by-File Details

#### `main.go` - Entry Point
**Key Functions:**
- `main()` - Parses flags, loads config, starts web/CLI mode
- `init()` - Redirects logging to ring buffer

**Global Variables:**
- `httpClient` - Reused HTTP client with connection pooling
- `emailTemplate` - Precompiled Go template for emails
- `startTime` - Server start time for uptime tracking
- `stats` - Global statistics tracker for dashboard
- `version` - Application version (1.6.0)

**Embedded Resources:**
```go
//go:embed templates/*.html
var templateFS embed.FS
```

#### `server.go` - Web Server & Scheduler
**Key Functions:**
- `startWebServer()` - HTTP server with graceful shutdown
- `startScheduler()` - Cron job setup with timezone support
- `restartScheduler()` - Dynamic schedule updates
- `convertToCronExpression()` - Day/time → cron format

**Routes Registration:**
```go
http.HandleFunc("/", serveWeb)
http.HandleFunc("/api/config", configHandler)
http.HandleFunc("/api/send", sendHandler)
// ... see handlers.go for full list
```

#### `handlers.go` - HTTP Request Handlers
**Key Handlers:**
- `serveWeb()` - Main UI with gzip compression
- `configHandler()` - GET/POST configuration
- `testAPIHandler()` - Generic API connection tester
- `previewHandler()` - Generate newsletter preview
- `sendHandler()` - Trigger immediate newsletter send
- `healthHandler()` - Health check endpoint
- `dashboardHandler()` - Dashboard stats API (system, newsletter, service status)

**JSON Response Pattern:**
```go
type APIResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
    Data    any    `json:"data,omitempty"`
}
```

#### `newsletter.go` - Newsletter Generation
**Key Functions:**
- `runNewsletter()` - Main orchestrator (parallel API calls)
- `sendNewsletter()` - Template rendering and email sending
- `generatePreview()` - HTML preview with retry logic
- `filterMonitored[T]()` - Generic monitored item filter
- `initEmailTemplate()` - Template compilation with custom functions

**Processing Flow:**
1. Parallel API calls (goroutines + WaitGroup)
2. Filter unmonitored items (if enabled)
3. Deduplicate episodes
4. Validate content exists
5. Render template
6. Send emails (batch support)

#### `api.go` - Sonarr/Radarr Client
**Key Functions:**
- `fetchSonarrHistory()` - Downloaded episodes (last 7 days)
- `fetchSonarrCalendar()` - Upcoming episodes (next 7 days)
- `fetchRadarrHistory()` - Downloaded movies
- `fetchRadarrCalendar()` - Upcoming movies
- `retryWithBackoff[T]()` - Generic retry wrapper

**API Patterns:**
- Context-aware with timeout
- Paginated requests (configurable page size)
- SHA256-based cache keys
- Connection pooling via global `httpClient`

#### `trakt.go` - Trakt.tv Integration
**Key Functions:**
- `fetchTraktAnticipatedSeries()` - Upcoming trending series
- `fetchTraktWatchedSeries()` - Most watched series
- `fetchTraktAnticipatedMovies()` - Upcoming trending movies
- `fetchTraktWatchedMovies()` - Most watched movies
- `getSonarrLibrary()`, `getRadarrLibrary()` - Library cross-reference

**Features:**
- Library membership checks (shows "In Library" badge)
- Date filtering for next week releases
- Configurable limits per category
- No poster images (Trakt API limitation)

#### `config.go` - Configuration Management
**Key Functions:**
- `loadConfig()` - Load from .env file + environment
- `getConfig()` - Thread-safe config access
- `reloadConfig()` - Hot reload configuration
- `validateConfig()` - Validation with warnings
- `convertWebConfigToConfig()` - String → typed conversion

**Two Config Types:**
- `Config` - Internal use (boolean, int types)
- `WebConfig` - API communication (string types)

**Loading Priority:** `.env` file > environment variables > defaults

#### `types.go` - Data Structures
**Key Types:**
- `Config` / `WebConfig` - Configuration structures
- `Episode`, `Movie`, `Series` - Media data
- `APICache`, `CacheEntry` - Caching infrastructure
- `APIResponse` - Standardized JSON responses
- `Statistics` - Email statistics tracking (thread-safe)
- `DashboardData` - Dashboard API response structure

**Memory Optimization:**
> Comment: "only fields we actually need (reduces memory & JSON parsing time)"

#### `utils.go` - Utilities
**Key Functions:**
- `sendEmail()` - SMTP email sending with batch support
- `testSMTP()` - Connection testing
- `logDebug()`, `logInfo()`, `logWarn()`, `logError()` - Leveled logging
- Ring buffer writer implementation

**Email Features:**
- Batch sending (configurable delay between batches)
- STARTTLS support
- Proper MIME headers for HTML

### Supporting Files

#### `templates/email.html` (24,301 bytes)
- Go template syntax with conditionals
- Dark/light mode support
- Responsive HTML design
- Custom functions: `formatDateWithDay`, `truncate`

#### `scripts/build-deb.sh` (5,258 bytes)
- Debian package builder
- Creates systemd service
- Installs to `/opt/newslettar/`
- Includes `newslettar-ctl` management script

#### `.env.example` / `.env`
- Configuration template
- All settings with documentation
- Works with Docker and native installs

#### `Makefile` (99 lines)
- Build automation
- Cross-platform compilation
- Docker image building
- Debian package creation

---

## 🔧 Development Workflow

### Prerequisites

```bash
# Check Go version (must be 1.23+)
go version

# Install optional tools
go install github.com/cosmtrek/air@latest        # Auto-reload
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest  # Linting
```

### Local Development Setup

```bash
# Clone repository
git clone https://github.com/agencefanfare/newslettar.git
cd newslettar

# Copy configuration
cp .env.example .env

# Edit configuration with your API keys
nano .env

# Build
make build

# Run locally (starts web UI on port 8080)
make run

# Or use auto-reload for development
make dev
```

### Common Development Tasks

#### Build Commands

```bash
make build          # Build for current platform → build/newslettar
make build-all      # Build for all platforms → dist/
make clean          # Remove build artifacts
```

#### Testing & Quality

```bash
make test           # Run tests with race detection
make coverage       # Generate coverage report
make fmt            # Format code with gofmt
make lint           # Run golangci-lint
```

#### Docker Development

```bash
make docker-build   # Build Docker image
make docker-run     # Build and run container
docker-compose up   # Run with docker-compose
```

#### Debian Package

```bash
make deb            # Build .deb package → dist/newslettar_*.deb
```

### Code Formatting

**Always run before committing:**
```bash
make fmt
```

This runs `gofmt -s -w` on all `.go` files.

### Hot Reload Development

```bash
# Install air if not already installed
go install github.com/cosmtrek/air@latest

# Start auto-reload server
make dev
```

Changes to `.go` files will automatically rebuild and restart.

---

## 📐 Coding Conventions & Standards

### Go Style Guidelines

**Follow:** [Effective Go](https://golang.org/doc/effective_go.html)

**Key Principles:**
- Descriptive names over abbreviations
- Handle all errors explicitly
- Add comments for exported functions
- Use `gofmt` for formatting
- Prefer table-driven tests

### Naming Conventions

#### Variables
```go
// Good
var cachedConfig *Config
var httpClient *http.Client

// Bad
var cfg *Config  // Avoid abbreviations in package scope
var c *http.Client
```

#### Functions
```go
// Exported (public)
func LoadConfig() *Config

// Unexported (private)
func convertToCronExpression(day, time string) string
```

#### Constants
```go
const (
    maxLogLines  = 500    // Unexported
    version      = "1.5.1" // Unexported but referenced
)
```

### Error Handling

**Always handle errors explicitly:**

```go
// Good
data, err := fetchAPI(url)
if err != nil {
    log.Printf("❌ Failed to fetch: %v", err)
    return err
}

// Bad - Never ignore errors
data, _ := fetchAPI(url)
```

**Graceful Degradation Pattern:**

```go
failedServices := []string{}
workingServices := []string{}

if err != nil {
    failedServices = append(failedServices, "Sonarr")
    // Continue with other services
} else {
    workingServices = append(workingServices, "Sonarr")
}
```

### Concurrency Patterns

**Use WaitGroup for parallel operations:**

```go
var wg sync.WaitGroup
wg.Add(taskCount)

go func() {
    defer wg.Done()
    // Do work
}()

wg.Wait()
```

**Use mutexes for shared state:**

```go
var configMu sync.RWMutex
var cachedConfig *Config

func getConfig() *Config {
    configMu.RLock()
    defer configMu.RUnlock()
    return cachedConfig
}

func updateConfig(cfg *Config) {
    configMu.Lock()
    defer configMu.Unlock()
    cachedConfig = cfg
}
```

### Context Usage

**Always accept context for API calls:**

```go
func fetchAPI(ctx context.Context, cfg *Config, url string) (Data, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    // ...
}
```

**Create context with timeout:**

```go
ctx, cancel := context.WithTimeout(
    context.Background(),
    time.Duration(cfg.APITimeout)*time.Second,
)
defer cancel()
```

### Logging Conventions

**Use emoji prefixes for visual clarity:**

```go
log.Printf("✓ Successfully connected")  // Success
log.Printf("⚠️  Warning: %s", msg)      // Warning
log.Printf("❌ Error: %v", err)         // Error
log.Printf("📺 Fetching from Sonarr")   // Info
```

**Respect log levels:**

```go
logDebug("Detailed debugging info")
logInfo("General information")
logWarn("Warning message")
logError("Error message")
```

### Memory Management

**Explicit nil-ing after large operations:**

```go
// After sending newsletter
downloadedEpisodes = nil
upcomingEpisodes = nil
downloadedMovies = nil
upcomingMovies = nil
```

**Use streaming JSON parsing:**

```go
// Good - stream parsing
err = json.NewDecoder(resp.Body).Decode(&result)

// Avoid - loads entire response into memory
body, _ := io.ReadAll(resp.Body)
json.Unmarshal(body, &result)
```

### Template Best Practices

**Use custom functions for reusable logic:**

```go
emailTemplate = template.Must(
    template.New("email").
        Funcs(template.FuncMap{
            "formatDateWithDay": formatDateWithDay,
            "truncate": truncate,
        }).
        ParseFS(templateFS, "templates/email.html"),
)
```

**Precompile templates at startup:**

```go
// In main.go init
emailTemplate, err = initEmailTemplate()
if err != nil {
    log.Fatalf("Failed to parse template: %v", err)
}
```

### Code Comments

**Comment complex logic:**

```go
// Check for duplicates using episode ID (handles multi-episode files)
// where the same episode might appear twice with different IDs
if _, exists := episodeIDs[episodeID]; exists {
    continue
}
```

**Document exported functions:**

```go
// fetchSonarrHistory retrieves episodes downloaded in the specified date range
// from the Sonarr history API endpoint. It handles pagination automatically
// and caches results for 5 minutes.
func fetchSonarrHistory(ctx context.Context, cfg *Config, ...) ([]Episode, error)
```

---

## 🛠️ Common Tasks

### Adding a New Configuration Option

1. **Update `.env.example`:**
```bash
# New Feature
NEW_FEATURE_ENABLED=true
```

2. **Add to `Config` struct in `types.go`:**
```go
type Config struct {
    // ... existing fields
    NewFeatureEnabled bool
}
```

3. **Add to `WebConfig` struct:**
```go
type WebConfig struct {
    // ... existing fields
    NewFeatureEnabled string `json:"newFeatureEnabled"`
}
```

4. **Load in `config.go`:**
```go
func loadConfig() *Config {
    // ... existing code
    cfg.NewFeatureEnabled = getEnvFromFile(envMap, "NEW_FEATURE_ENABLED", "false") == "true"
    return cfg
}
```

5. **Update conversion functions in `config.go`:**
```go
func convertWebConfigToConfig(wc WebConfig) *Config {
    // ... existing code
    NewFeatureEnabled: wc.NewFeatureEnabled == "true",
}

func convertConfigToWebConfig(c *Config) WebConfig {
    // ... existing code
    NewFeatureEnabled: strconv.FormatBool(c.NewFeatureEnabled),
}
```

### Adding a New API Endpoint

1. **Add route in `server.go`:**
```go
func startWebServer() {
    // ... existing routes
    http.HandleFunc("/api/new-endpoint", newEndpointHandler)
}
```

2. **Implement handler in `handlers.go`:**
```go
func newEndpointHandler(w http.ResponseWriter, r *http.Request) {
    // Set JSON content type
    w.Header().Set("Content-Type", "application/json")

    // Validate method
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Get config
    cfg := getConfig()

    // Do work
    result, err := doWork(cfg)
    if err != nil {
        json.NewEncoder(w).Encode(APIResponse{
            Success: false,
            Message: fmt.Sprintf("Error: %v", err),
        })
        return
    }

    // Return success
    json.NewEncoder(w).Encode(APIResponse{
        Success: true,
        Message: "Operation completed",
        Data: result,
    })
}
```

### Adding a New External API Integration

1. **Add API client function:**
```go
// In api.go or new file
func fetchFromNewAPI(ctx context.Context, cfg *Config) ([]Item, error) {
    // Use retryWithBackoff for reliability
    return retryWithBackoff(func() ([]Item, error) {
        // Check cache first
        cacheKey := generateCacheKey(cfg.NewAPIURL, params)
        if cached, found := apiCache.Get(cacheKey); found {
            return cached.([]Item), nil
        }

        // Make API request
        req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
        if err != nil {
            return nil, err
        }

        // Add authentication
        req.Header.Set("Authorization", cfg.NewAPIKey)

        // Execute request
        resp, err := httpClient.Do(req)
        if err != nil {
            return nil, err
        }
        defer resp.Body.Close()

        // Parse response
        var items []Item
        if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
            return nil, err
        }

        // Cache result
        apiCache.Set(cacheKey, items, 5*time.Minute)

        return items, nil
    }, "NewAPI", cfg.MaxRetries)
}
```

2. **Add to parallel fetching in `newsletter.go`:**
```go
var newAPIItems []Item
var errNewAPI error

if cfg.NewAPIEnabled {
    wg.Add(1)
    go func() {
        defer wg.Done()
        newAPIItems, errNewAPI = fetchFromNewAPI(ctx, cfg)
    }()
}

wg.Wait()

if errNewAPI != nil {
    log.Printf("⚠️  NewAPI failed: %v", errNewAPI)
}
```

### Modifying the Email Template

**Template Location:** `templates/email.html`

**Important:**
- Must rebuild binary after template changes (embedded at compile time)
- Use Go template syntax: `{{.Variable}}`, `{{if .Condition}}`, `{{range .Items}}`
- Available functions: `formatDateWithDay`, `truncate`

**Example modification:**
```html
{{if .NewFeatureEnabled}}
<div class="new-feature">
    {{range .NewItems}}
    <div class="item">
        <h3>{{.Title}}</h3>
    </div>
    {{end}}
</div>
{{end}}
```

**After changes:**
```bash
make build
make run
```

### Adding a Test

**Create test file:**
```go
// config_test.go
package main

import "testing"

func TestLoadConfig(t *testing.T) {
    tests := []struct {
        name     string
        envVars  map[string]string
        expected string
    }{
        {
            name:     "default timezone",
            envVars:  map[string]string{},
            expected: "UTC",
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

**Run tests:**
```bash
make test
```

### Debugging Tips

**Enable debug logging:**
```bash
LOG_LEVEL=debug make run
```

**Check logs via API:**
```bash
curl http://localhost:8080/api/logs
```

**Test individual services:**
```bash
# Test Sonarr
curl http://localhost:8080/api/test-sonarr

# Test Radarr
curl http://localhost:8080/api/test-radarr

# Test Email
curl -X POST http://localhost:8080/api/test-email
```

**Generate preview without sending:**
```bash
curl http://localhost:8080/api/preview > preview.html
open preview.html  # macOS
xdg-open preview.html  # Linux
```

### Dashboard Feature

**Overview:**
The Dashboard tab (v1.6.0+) is the default view when accessing the Web UI. It provides a comprehensive overview of the system status, newsletter statistics, service health, and recent logs.

**Key Components:**

1. **System Stats Card:**
   - Application version
   - Running port number
   - System uptime (days, hours, minutes)
   - Memory usage (~12MB baseline)

2. **Newsletter Stats Card:**
   - Total emails sent (lifetime counter)
   - Last sent date and time
   - Next scheduled run
   - Configured timezone

3. **Service Status Card:**
   - Real-time connection checks for:
     - Sonarr (🟢 Connected, 🔴 Error, ⚪ Not Configured)
     - Radarr (🟢 Connected, 🔴 Error, ⚪ Not Configured)
     - Email (🟡 Configured, ⚪ Not Configured)
     - Trakt (🟡 Configured, ⚪ Not Configured)

4. **Recent Logs Card:**
   - Last 20 log entries
   - Auto-scrolls to bottom
   - Updates with dashboard refresh

5. **Quick Actions:**
   - Preview Newsletter button
   - Send Now button
   - Go to Configuration button

**Technical Details:**

- Dashboard data fetched from `/api/dashboard` endpoint
- Auto-refreshes every 10 seconds when active
- Statistics tracked in global `stats` variable (thread-safe)
- Service status uses lightweight API calls with 5-second timeout
- Email statistics incremented in `newsletter.go` after successful send

**Accessing Dashboard Data:**
```bash
# Get dashboard JSON data
curl http://localhost:8080/api/dashboard

# Response includes:
# - version, uptime, memory_usage_mb, port
# - total_emails_sent, last_sent_date, next_scheduled_run
# - service_status (map of service names to status strings)
```

---

## ✅ Testing & Quality

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make coverage

# Run specific test
go test -v -run TestFunctionName

# Run tests with race detector
go test -v -race ./...
```

### Code Quality Checks

```bash
# Format code
make fmt

# Run linter
make lint

# Check for common issues
go vet ./...
```

### Manual Testing Checklist

Before submitting a PR, test:

- [ ] Web UI loads at `http://localhost:8080`
- [ ] Configuration can be loaded and saved
- [ ] Test buttons work (Sonarr, Radarr, Email)
- [ ] Preview generates successfully
- [ ] Test email sends successfully
- [ ] Logs appear in `/api/logs` endpoint
- [ ] Build succeeds: `make build`
- [ ] Tests pass: `make test`
- [ ] Code formatted: `make fmt`

### Docker Testing

```bash
# Build and test Docker image
make docker-build
make docker-run

# Check container logs
docker logs newslettar

# Test inside container
docker exec newslettar ./newslettar
```

### Load Testing

```bash
# Test parallel API calls
for i in {1..10}; do
  curl -s http://localhost:8080/api/preview > /dev/null &
done
wait
```

### Memory Testing

```bash
# Check memory usage
docker stats newslettar

# Or for native install
ps aux | grep newslettar
```

---

## ⚠️ Important Considerations

### Performance Guidelines

1. **Always use parallel API calls** - Use goroutines + WaitGroup
2. **Cache aggressively** - 5-minute TTL for preview generation
3. **Reuse HTTP client** - Global `httpClient` with connection pooling
4. **Stream JSON parsing** - Use `json.NewDecoder()` not `json.Unmarshal()`
5. **Precompile templates** - Done at startup, not per request

### Memory Constraints

**Target:** ~12MB RAM usage

**Memory-saving techniques:**
- Embedded assets (no file I/O)
- Ring buffer logging (bounded size)
- Minimal data structures (only needed fields)
- Explicit nil-ing after processing
- No global caches that grow unbounded

### Thread Safety

**Must use mutexes for:**
- Config access (`configMu sync.RWMutex`)
- Cache operations (`APICache.mu sync.RWMutex`)
- Log buffer writes (`logBufferMu sync.Mutex`)

**Safe patterns:**
```go
// Read lock for read-only access
mu.RLock()
defer mu.RUnlock()

// Write lock for modifications
mu.Lock()
defer mu.Unlock()
```

### Error Handling Philosophy

**Graceful degradation over hard failures:**

```go
// Good - continues with partial data
if errSonarr != nil {
    log.Printf("⚠️  Sonarr unavailable, skipping")
    // Continue with Radarr, Trakt, etc.
}

// Bad - fails completely
if errSonarr != nil {
    return fmt.Errorf("cannot proceed: %w", errSonarr)
}
```

**When to fail hard:**
- Configuration is completely invalid
- Template compilation fails
- SMTP configuration prevents any emails

### Security Considerations

1. **API keys in environment** - Never hardcode
2. **SMTP passwords** - Store in `.env`, never commit
3. **Input validation** - Validate all config values
4. **No SQL injection risk** - No database used
5. **Email headers** - Properly escaped in templates

### Timezone Handling

**Always use configured timezone:**

```go
loc, _ := time.LoadLocation(cfg.Timezone)
now := time.Now().In(loc)
```

**Cron expressions are timezone-aware:**
```go
scheduler.SetLocation(loc)
```

### Configuration Reload

**When config changes:**
```go
reloadConfig()
restartScheduler()
```

**Important:** Scheduler must restart to pick up new day/time.

### Logging Best Practices

**Avoid excessive logging:**
- Log important events (success, failure, warnings)
- Use debug level for detailed info
- Don't log inside tight loops
- Don't log sensitive data (API keys, passwords)

**Good logging:**
```go
log.Printf("✓ Fetched %d episodes from Sonarr", len(episodes))
```

**Bad logging:**
```go
for _, episode := range episodes {
    log.Printf("Episode: %+v", episode) // Too verbose
}
```

### Template Gotchas

**Templates are embedded at compile time:**
- Changes require rebuild
- Use `make run` to test changes
- Can't hot-reload templates

**Template data must be exported:**
```go
// Good
type TemplateData struct {
    Episodes []Episode  // Exported
}

// Bad
type TemplateData struct {
    episodes []Episode  // Unexported - won't work
}
```

### Docker Networking

**Accessing host services from Docker:**

Linux:
```
SONARR_URL=http://172.17.0.1:8989
```

Mac/Windows:
```
SONARR_URL=http://host.docker.internal:8989
```

### Common Pitfalls

1. **Forgetting to call `wg.Done()`** - Will deadlock
2. **Not using mutexes on shared state** - Race conditions
3. **Ignoring context cancellation** - Goroutine leaks
4. **Not closing response bodies** - Connection leaks
5. **Hardcoding timeouts** - Use `cfg.APITimeout`
6. **JSON struct tags mismatch** - API fields won't parse

---

## 🔄 Git Workflow

### Branch Strategy

**Main branch:** `main`
**Feature branches:** `feature/description`
**Bug fixes:** `fix/description`
**AI branches:** `claude/description-sessionid`

### Commit Message Format

Use conventional commits:

```
feat: add Plex integration
fix: correct timezone handling in scheduler
docs: update README with new features
style: format code with gofmt
refactor: simplify API retry logic
test: add tests for config validation
chore: update dependencies
```

### PR Checklist

Before submitting:

- [ ] Code formatted: `make fmt`
- [ ] Tests pass: `make test`
- [ ] No linter errors: `make lint`
- [ ] Manual testing completed
- [ ] Documentation updated (README, CONTRIBUTING)
- [ ] Commit messages follow convention
- [ ] Branch up to date with main

### Development Branch Workflow

```bash
# Create feature branch
git checkout -b feature/amazing-feature

# Make changes
# ... edit files ...

# Stage and commit
git add .
git commit -m "feat: add amazing feature"

# Push to origin
git push -u origin feature/amazing-feature

# Create PR via GitHub UI
```

### Testing Before Push

```bash
# Full test suite
make test

# Build check
make build

# Docker build check
make docker-build

# Format check
make fmt
git diff  # Should show no changes
```

### Updating Version

**When releasing:**

1. Update `version.json`:
```json
{
  "version": "1.6.0",
  "released": "2025-11-17",
  "changelog": [
    "Description of changes"
  ]
}
```

2. Update `main.go`:
```go
const version = "1.6.0"
```

3. Tag release:
```bash
git tag v1.6.0
git push origin v1.6.0
```

---

## 📚 Additional Resources

### Official Documentation

- **README.md** - User-facing documentation
- **CONTRIBUTING.md** - Contribution guidelines
- **LICENSE** - MIT License
- **.env.example** - Configuration reference

### External Documentation

- [Sonarr API](https://sonarr.tv/docs/api/)
- [Radarr API](https://radarr.video/docs/api/)
- [Trakt.tv API](https://trakt.docs.apiary.io/)
- [Go Documentation](https://golang.org/doc/)
- [Effective Go](https://golang.org/doc/effective_go.html)

### Tools

- [air](https://github.com/cosmtrek/air) - Auto-reload for development
- [golangci-lint](https://golangci-lint.run/) - Linting
- [Docker](https://www.docker.com/) - Containerization
- [Make](https://www.gnu.org/software/make/) - Build automation

---

## 🎓 Learning Path for New Contributors

### Beginner Tasks
1. Fix typos in documentation
2. Add configuration validation
3. Improve error messages
4. Add more template customization options

### Intermediate Tasks
1. Add new Trakt.tv endpoints
2. Implement additional email templates
3. Add Prometheus metrics
4. Create Discord/Slack webhooks

### Advanced Tasks
1. Add Plex/Jellyfin integration
2. Implement multi-user support
3. Create web-based template editor
4. Add database backend for analytics

---

## 📞 Getting Help

- **GitHub Issues:** https://github.com/agencefanfare/newslettar/issues
- **Discussions:** https://github.com/agencefanfare/newslettar/discussions
- **Email:** hello@agencefanfare.com

---

**Last Updated:** 2025-11-17
**Maintainer:** Agency Fanfare
**License:** MIT
