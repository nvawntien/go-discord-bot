package outbound

import (
	"context"

	"github.com/tien/go-discord-bot/internal/domain/entity"
)

// CompletionRepository defines the outbound operations for tracking daily completions.
type CompletionRepository interface {
	HasCompleted(ctx context.Context, userID int64, date string) (bool, error)
	MarkCompleted(ctx context.Context, completion *entity.DailyCompletion) error
}
