package discord

import (
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

// Bot manages the Discord session and command registration.
type Bot struct {
	session *discordgo.Session
	appID   string
	handler *Handler
	logger  *slog.Logger
}

// NewBot creates a new Discord bot instance.
func NewBot(token, appID string, handler *Handler, logger *slog.Logger) (*Bot, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("create discord session: %w", err)
	}

	session.Identify.Intents = discordgo.IntentsGuilds

	return &Bot{
		session: session,
		appID:   appID,
		handler: handler,
		logger:  logger,
	}, nil
}

// Start opens the Discord WebSocket connection and registers slash commands.
func (b *Bot) Start() error {
	// Register the interaction handler
	b.session.AddHandler(b.handler.OnInteraction)

	// Open WebSocket connection
	if err := b.session.Open(); err != nil {
		return fmt.Errorf("open discord session: %w", err)
	}

	// Register slash commands globally
	_, err := b.session.ApplicationCommandBulkOverwrite(b.appID, "", commands)
	if err != nil {
		return fmt.Errorf("register slash commands: %w", err)
	}

	b.logger.Info("Discord bot started",
		"user", b.session.State.User.Username,
		"commands_registered", len(commands),
	)

	return nil
}

// Stop gracefully closes the Discord session.
func (b *Bot) Stop() error {
	b.logger.Info("Stopping Discord bot...")
	return b.session.Close()
}

// Session returns the underlying discordgo session (used by the tracker to send messages).
func (b *Bot) Session() *discordgo.Session {
	return b.session
}
