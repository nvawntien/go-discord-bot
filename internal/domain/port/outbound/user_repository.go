package outbound

import (
	"context"

	"github.com/tien/go-discord-bot/internal/domain/entity"
)

// UserRepository defines the outbound operations for user persistence.
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	DeleteByUsername(ctx context.Context, leetcodeUsername, guildID string) error
	GetByDiscordID(ctx context.Context, discordID, guildID string) ([]*entity.User, error)
	GetByLeetCodeUsername(ctx context.Context, leetcodeUsername, guildID string) (*entity.User, error)
	GetByGuildID(ctx context.Context, guildID string) ([]*entity.User, error)
	GetAll(ctx context.Context) ([]*entity.User, error)
}

