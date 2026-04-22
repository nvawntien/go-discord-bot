package application

import (
	"context"

	"github.com/tien/go-discord-bot/internal/domain/entity"
	"github.com/tien/go-discord-bot/internal/domain/port/outbound"
)

// GetDailyUseCase handles fetching the daily LeetCode challenge.
type GetDailyUseCase struct {
	lc outbound.LeetCodeClient
}

// NewGetDailyUseCase creates a new GetDailyUseCase.
func NewGetDailyUseCase(lc outbound.LeetCodeClient) *GetDailyUseCase {
	return &GetDailyUseCase{lc: lc}
}

// GetDailyQuestion fetches today's daily coding challenge.
func (uc *GetDailyUseCase) GetDailyQuestion(ctx context.Context) (*entity.DailyQuestion, error) {
	return uc.lc.GetDailyQuestion(ctx)
}
