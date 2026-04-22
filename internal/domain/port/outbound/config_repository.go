package outbound

import "context"

// ConfigRepository defines the outbound operations for guild configuration persistence.
type ConfigRepository interface {
	SetNotificationChannel(ctx context.Context, guildID, channelID string) error
	GetNotificationChannel(ctx context.Context, guildID string) (string, error)
}
