package outbound

import (
	"context"

	"github.com/tien/go-discord-bot/internal/domain/entity"
)

// LeetCodeClient defines the outbound operations for interacting with the LeetCode API.
type LeetCodeClient interface {
	GetDailyQuestion(ctx context.Context) (*entity.DailyQuestion, error)
	GetUserProfile(ctx context.Context, username string) (*entity.UserStats, error)
	GetRecentAcceptedSubmissions(ctx context.Context, username string, limit int) ([]entity.Submission, error)
}
