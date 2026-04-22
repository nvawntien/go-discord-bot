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

// Unregister removes a user's registration.
func (uc *UnregisterUserUseCase) Unregister(ctx context.Context, discordID, guildID string) error {
	if err := uc.userRepo.Delete(ctx, discordID, guildID); err != nil {
		return fmt.Errorf("bạn chưa đăng ký. Dùng `/register` để đăng ký")
	}

	uc.logger.Info("User unregistered",
		"discord_id", discordID,
		"guild_id", guildID,
	)

	return nil
}
