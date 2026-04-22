package inbound

import (
	"context"

	"github.com/tien/go-discord-bot/internal/domain/entity"
)

// UserUseCase defines the inbound operations for user management.
type UserUseCase interface {
	Register(ctx context.Context, discordID, guildID, leetcodeUsername string) error
	Unregister(ctx context.Context, discordID, guildID string) error
}

// StatsUseCase defines the inbound operations for viewing LeetCode statistics.
type StatsUseCase interface {
	GetUserStats(ctx context.Context, leetcodeUsername string) (*entity.UserStats, error)
}

// DailyUseCase defines the inbound operations for the daily challenge.
type DailyUseCase interface {
	GetDailyQuestion(ctx context.Context) (*entity.DailyQuestion, error)
}

// ConfigUseCase defines the inbound operations for guild configuration.
type ConfigUseCase interface {
	SetNotificationChannel(ctx context.Context, guildID, channelID string) error
}
