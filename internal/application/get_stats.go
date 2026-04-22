package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tien/go-discord-bot/internal/domain/entity"
	"github.com/tien/go-discord-bot/internal/domain/port/outbound"
)

// GetStatsUseCase handles fetching LeetCode user statistics.
type GetStatsUseCase struct {
	userRepo outbound.UserRepository
	lc       outbound.LeetCodeClient
	logger   *slog.Logger
}

// NewGetStatsUseCase creates a new GetStatsUseCase.
func NewGetStatsUseCase(userRepo outbound.UserRepository, lc outbound.LeetCodeClient, logger *slog.Logger) *GetStatsUseCase {
	return &GetStatsUseCase{
		userRepo: userRepo,
		lc:       lc,
		logger:   logger,
	}
}

// GetUserStats fetches stats for a given LeetCode username.
// If username is empty, it looks up the registered username for the Discord user.
func (uc *GetStatsUseCase) GetUserStats(ctx context.Context, discordID, guildID, username string) (*entity.UserStats, error) {
	if username == "" {
		// Look up registered user
		user, err := uc.userRepo.GetByDiscordID(ctx, discordID, guildID)
		if err != nil {
			return nil, fmt.Errorf("lookup user: %w", err)
		}
		if user == nil {
			return nil, fmt.Errorf("bạn chưa đăng ký. Dùng `/register` để đăng ký hoặc nhập username trực tiếp")
		}
		username = user.LeetCodeUsername
	}

	stats, err := uc.lc.GetUserProfile(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("không tìm thấy user **%s** trên LeetCode", username)
	}

	return stats, nil
}
