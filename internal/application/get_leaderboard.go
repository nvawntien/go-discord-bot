package application

import (
	"context"
	"fmt"
	"log/slog"
	"sort"

	"github.com/tien/go-discord-bot/internal/domain/entity"
	"github.com/tien/go-discord-bot/internal/domain/port/outbound"
)

// LeaderboardEntry represents a single entry in the leaderboard.
type LeaderboardEntry struct {
	DiscordID        string
	LeetCodeUsername string
	Stats            *entity.UserStats
}

// GetLeaderboardUseCase handles fetching the leaderboard for a guild.
type GetLeaderboardUseCase struct {
	userRepo outbound.UserRepository
	lc       outbound.LeetCodeClient
	logger   *slog.Logger
}

// NewGetLeaderboardUseCase creates a new GetLeaderboardUseCase.
func NewGetLeaderboardUseCase(userRepo outbound.UserRepository, lc outbound.LeetCodeClient, logger *slog.Logger) *GetLeaderboardUseCase {
	return &GetLeaderboardUseCase{
		userRepo: userRepo,
		lc:       lc,
		logger:   logger,
	}
}

// GetLeaderboard fetches stats for all registered users in a guild and ranks them.
func (uc *GetLeaderboardUseCase) GetLeaderboard(ctx context.Context, guildID string) ([]LeaderboardEntry, error) {
	users, err := uc.userRepo.GetByGuildID(ctx, guildID)
	if err != nil {
		return nil, fmt.Errorf("fetch users: %w", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("chưa có ai đăng ký trong server này. Dùng `/register` để đăng ký")
	}

	entries := make([]LeaderboardEntry, 0, len(users))
	for _, user := range users {
		stats, err := uc.lc.GetUserProfile(ctx, user.LeetCodeUsername)
		if err != nil {
			uc.logger.Warn("Failed to fetch stats for leaderboard",
				"user", user.LeetCodeUsername,
				"error", err,
			)
			continue
		}

		entries = append(entries, LeaderboardEntry{
			DiscordID:        user.DiscordID,
			LeetCodeUsername: user.LeetCodeUsername,
			Stats:            stats,
		})
	}

	// Sort by total solved (descending)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Stats.TotalSolved > entries[j].Stats.TotalSolved
	})

	return entries, nil
}
