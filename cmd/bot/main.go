package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/tien/go-discord-bot/internal/adapter/inbound/discord"
	"github.com/tien/go-discord-bot/internal/adapter/outbound/leetcode"
	"github.com/tien/go-discord-bot/internal/adapter/outbound/sqlite"
	"github.com/tien/go-discord-bot/internal/application"
	"github.com/tien/go-discord-bot/pkg/tracker"
)

func main() {
	// Load .env file (ignore error if not present)
	_ = godotenv.Load()

	// Setup logger
	logLevel := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	// Read config from env
	botToken := os.Getenv("DISCORD_BOT_TOKEN")
	appID := os.Getenv("DISCORD_APP_ID")
	dbPath := os.Getenv("DATABASE_PATH")
	pollIntervalStr := os.Getenv("POLL_INTERVAL")

	if botToken == "" || appID == "" {
		logger.Error("DISCORD_BOT_TOKEN and DISCORD_APP_ID are required")
		os.Exit(1)
	}

	if dbPath == "" {
		dbPath = "./data/bot.db"
	}

	pollInterval := 2 * time.Minute
	if pollIntervalStr != "" {
		d, err := time.ParseDuration(pollIntervalStr)
		if err != nil {
			logger.Error("Invalid POLL_INTERVAL", "value", pollIntervalStr, "error", err)
			os.Exit(1)
		}
		pollInterval = d
	}

	// --- Initialize outbound adapters ---

	// Database
	db, err := sqlite.NewDB(dbPath, logger)
	if err != nil {
		logger.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Repositories
	userRepo := sqlite.NewUserRepository(db)
	configRepo := sqlite.NewConfigRepository(db)
	compRepo := sqlite.NewCompletionRepository(db)

	// LeetCode client
	lcClient := leetcode.NewClient(logger)

	// --- Initialize application layer (use cases) ---

	registerUC := application.NewRegisterUserUseCase(userRepo, lcClient, logger)
	unregisterUC := application.NewUnregisterUserUseCase(userRepo, logger)
	statsUC := application.NewGetStatsUseCase(userRepo, lcClient, logger)
	dailyUC := application.NewGetDailyUseCase(lcClient)
	setChannelUC := application.NewSetChannelUseCase(configRepo, logger)

	// --- Initialize inbound adapter (Discord bot) ---

	handler := discord.NewHandler(registerUC, unregisterUC, statsUC, dailyUC, setChannelUC, logger)

	bot, err := discord.NewBot(botToken, appID, handler, logger)
	if err != nil {
		logger.Error("Failed to create Discord bot", "error", err)
		os.Exit(1)
	}

	if err := bot.Start(); err != nil {
		logger.Error("Failed to start Discord bot", "error", err)
		os.Exit(1)
	}
	defer bot.Stop()

	// --- Start background tracker ---

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dailyTracker := tracker.NewTracker(
		bot.Session(),
		userRepo, configRepo, compRepo,
		lcClient,
		pollInterval,
		logger,
	)

	go dailyTracker.Start(ctx)

	// --- Wait for shutdown signal ---

	logger.Info("Bot is running. Press Ctrl+C to stop.")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	logger.Info("Shutting down...")
	cancel()
}
