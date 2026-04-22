package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tien/go-discord-bot/internal/domain/port/outbound"
)

// UnregisterUserUseCase handles user unregistration.
type UnregisterUserUseCase struct {
	userRepo outbound.UserRepository
	logger   *slog.Logger
}

// NewUnregisterUserUseCase creates a new UnregisterUserUseCase.
func NewUnregisterUserUseCase(userRepo outbound.UserRepository, logger *slog.Logger) *UnregisterUserUseCase {
	return &UnregisterUserUseCase{
		userRepo: userRepo,
		logger:   logger,
	}
}

// Unregister removes the calling Discord user's registration.
func (uc *UnregisterUserUseCase) Unregister(ctx context.Context, discordID, guildID string) error {
	// Check if user is registered
	existing, err := uc.userRepo.GetByDiscordID(ctx, discordID, guildID)
	if err != nil {
		return fmt.Errorf("lookup user: %w", err)
	}
	if len(existing) == 0 {
		return fmt.Errorf("bạn chưa đăng ký. Dùng `/register` để đăng ký")
	}

	if err := uc.userRepo.DeleteByUsername(ctx, existing[0].LeetCodeUsername, guildID); err != nil {
		return fmt.Errorf("unregister: %w", err)
	}

	uc.logger.Info("User unregistered",
		"discord_id", discordID,
		"guild_id", guildID,
		"leetcode_username", existing[0].LeetCodeUsername,
	)

	return nil
}
