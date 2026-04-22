package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tien/go-discord-bot/internal/domain/entity"
	"github.com/tien/go-discord-bot/internal/domain/port/outbound"
)

// RegisterUserUseCase handles user registration.
type RegisterUserUseCase struct {
	userRepo outbound.UserRepository
	lc       outbound.LeetCodeClient
	logger   *slog.Logger
}

// NewRegisterUserUseCase creates a new RegisterUserUseCase.
func NewRegisterUserUseCase(userRepo outbound.UserRepository, lc outbound.LeetCodeClient, logger *slog.Logger) *RegisterUserUseCase {
	return &RegisterUserUseCase{
		userRepo: userRepo,
		lc:       lc,
		logger:   logger,
	}
}

// Register validates the LeetCode username and creates a new user registration.
// Each Discord user can only register one LeetCode username per guild.
func (uc *RegisterUserUseCase) Register(ctx context.Context, discordID, guildID, leetcodeUsername string) error {
	// Check if this Discord user already registered
	existing, err := uc.userRepo.GetByDiscordID(ctx, discordID, guildID)
	if err != nil {
		return fmt.Errorf("check existing user: %w", err)
	}
	if len(existing) > 0 {
		return fmt.Errorf("bạn đã đăng ký với username **%s** rồi. Dùng `/unregister` để hủy trước khi đăng ký username mới", existing[0].LeetCodeUsername)
	}

	// Check if this LeetCode username is already registered in this guild
	existingLC, err := uc.userRepo.GetByLeetCodeUsername(ctx, leetcodeUsername, guildID)
	if err != nil {
		return fmt.Errorf("check existing username: %w", err)
	}
	if existingLC != nil {
		return fmt.Errorf("username **%s** đã được đăng ký bởi người khác trong server này", leetcodeUsername)
	}

	// Validate that the LeetCode username exists on LeetCode
	_, err = uc.lc.GetUserProfile(ctx, leetcodeUsername)
	if err != nil {
		return fmt.Errorf("không tìm thấy user **%s** trên LeetCode. Kiểm tra lại username", leetcodeUsername)
	}

	user := &entity.User{
		DiscordID:        discordID,
		GuildID:          guildID,
		LeetCodeUsername: leetcodeUsername,
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	uc.logger.Info("User registered",
		"discord_id", discordID,
		"guild_id", guildID,
		"leetcode_username", leetcodeUsername,
	)

	return nil
}
