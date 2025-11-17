package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"
)

// Internal scheduler
var scheduler *cron.Cron

// Setup internal cron scheduler (replaces systemd timer)
func setupScheduler(cfg *Config) {
	scheduler = cron.New(cron.WithLocation(getTimezone(cfg.Timezone)))

	// Convert day/time to cron expression
	cronExpr := convertToCronExpression(cfg.ScheduleDay, cfg.ScheduleTime)
	log.Printf("📅 Setting up scheduler: %s (cron: %s)", cfg.ScheduleDay+" "+cfg.ScheduleTime, cronExpr)

	_, err := scheduler.AddFunc(cronExpr, func() {
		log.Println("⏰ Scheduled newsletter triggered")
		runNewsletter()
	})

	if err != nil {
		log.Printf("⚠️  Failed to setup scheduler: %v", err)
		return
	}

	scheduler.Start()
	log.Println("✅ Internal scheduler started")
}

// Restart scheduler when config changes
func restartScheduler() {
	if scheduler != nil {
		scheduler.Stop()
	}
	cfg := getConfig()
	setupScheduler(cfg)
	log.Println("🔄 Scheduler restarted")
}

// Start and run web server with graceful shutdown
func startWebServer() {
	cfg := getConfig()

	// Setup internal scheduler
	setupScheduler(cfg)

	// Register HTTP handlers
	registerHandlers()

	// Graceful shutdown
	server := &http.Server{
		Addr: ":" + cfg.WebUIPort,
	}

	go func() {
		log.Printf("🌐 Web UI started on port %s", cfg.WebUIPort)
		log.Printf("📅 Scheduler: %s at %s (%s)", cfg.ScheduleDay, cfg.ScheduleTime, cfg.Timezone)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if scheduler != nil {
		scheduler.Stop()
	}

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("✅ Server stopped")
}
