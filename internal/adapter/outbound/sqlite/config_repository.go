package sqlite

import (
	"context"
	"database/sql"
	"fmt"
)

// ConfigRepository implements outbound.ConfigRepository using SQLite.
type ConfigRepository struct {
	db *sql.DB
}

// NewConfigRepository creates a new SQLite-backed config repository.
func NewConfigRepository(db *sql.DB) *ConfigRepository {
	return &ConfigRepository{db: db}
}

func (r *ConfigRepository) SetNotificationChannel(ctx context.Context, guildID, channelID string) error {
	query := `INSERT INTO guild_configs (guild_id, notification_channel_id) VALUES (?, ?)
               ON CONFLICT(guild_id) DO UPDATE SET notification_channel_id = excluded.notification_channel_id`
	_, err := r.db.ExecContext(ctx, query, guildID, channelID)
	if err != nil {
		return fmt.Errorf("set notification channel: %w", err)
	}

	return nil
}

func (r *ConfigRepository) GetNotificationChannel(ctx context.Context, guildID string) (string, error) {
	query := `SELECT notification_channel_id FROM guild_configs WHERE guild_id = ?`
	row := r.db.QueryRowContext(ctx, query, guildID)

	var channelID string
	if err := row.Scan(&channelID); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("get notification channel: %w", err)
	}

	return channelID, nil
}
