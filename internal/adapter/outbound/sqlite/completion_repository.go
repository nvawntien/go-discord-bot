package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/tien/go-discord-bot/internal/domain/entity"
)

// CompletionRepository implements outbound.CompletionRepository using SQLite.
type CompletionRepository struct {
	db *sql.DB
}

// NewCompletionRepository creates a new SQLite-backed completion repository.
func NewCompletionRepository(db *sql.DB) *CompletionRepository {
	return &CompletionRepository{db: db}
}

func (r *CompletionRepository) HasCompleted(ctx context.Context, userID int64, date string) (bool, error) {
	query := `SELECT COUNT(1) FROM daily_completions WHERE user_id = ? AND date = ?`
	row := r.db.QueryRowContext(ctx, query, userID, date)

	var count int
	if err := row.Scan(&count); err != nil {
		return false, fmt.Errorf("check completion: %w", err)
	}

	return count > 0, nil
}

func (r *CompletionRepository) MarkCompleted(ctx context.Context, completion *entity.DailyCompletion) error {
	query := `INSERT OR IGNORE INTO daily_completions (user_id, date, question_slug) VALUES (?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, completion.UserID, completion.Date, completion.QuestionSlug)
	if err != nil {
		return fmt.Errorf("mark completed: %w", err)
	}

	return nil
}
