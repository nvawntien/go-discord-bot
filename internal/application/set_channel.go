package application

import (
	"context"
	"log/slog"

	"github.com/tien/go-discord-bot/internal/domain/port/outbound"
)

// SetChannelUseCase handles setting the notification channel for a guild.
type SetChannelUseCase struct {
	configRepo outbound.ConfigRepository
	logger     *slog.Logger
}

// NewSetChannelUseCase creates a new SetChannelUseCase.
func NewSetChannelUseCase(configRepo outbound.ConfigRepository, logger *slog.Logger) *SetChannelUseCase {
	return &SetChannelUseCase{
		configRepo: configRepo,
		logger:     logger,
	}
}

// SetNotificationChannel sets the notification channel for the given guild.
func (uc *SetChannelUseCase) SetNotificationChannel(ctx context.Context, guildID, channelID string) error {
	if err := uc.configRepo.SetNotificationChannel(ctx, guildID, channelID); err != nil {
		return err
	}

	uc.logger.Info("Notification channel set",
		"guild_id", guildID,
		"channel_id", channelID,
	)

	return nil
}
