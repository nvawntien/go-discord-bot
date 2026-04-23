package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

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

func (r *CompletionRepository) GetStreakByUserIDs(ctx context.Context, userIDs []int64) (map[int64]int, error) {
	if len(userIDs) == 0 {
		return map[int64]int{}, nil
	}

	// Build IN clause placeholders
	placeholders := make([]string, len(userIDs))
	args := make([]any, len(userIDs))
	for i, id := range userIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	// Fetch all completion dates per user, sorted descending
	query := fmt.Sprintf(
		`SELECT user_id, date FROM daily_completions WHERE user_id IN (%s) ORDER BY user_id, date DESC`,
		strings.Join(placeholders, ","),
	)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("fetch completions: %w", err)
	}
	defer rows.Close()

	// Group dates by user
	userDates := make(map[int64][]string)
	for rows.Next() {
		var userID int64
		var date string
		if err := rows.Scan(&userID, &date); err != nil {
			return nil, fmt.Errorf("scan completion: %w", err)
		}
		userDates[userID] = append(userDates[userID], date)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Calculate streak for each user
	today := time.Now().Format("2006-01-02")
	streaks := make(map[int64]int)

	for userID, dates := range userDates {
		streak := 0
		expected := today

		for _, date := range dates {
			if date == expected {
				streak++
				// Move expected to previous day
				t, _ := time.Parse("2006-01-02", expected)
				expected = t.AddDate(0, 0, -1).Format("2006-01-02")
			} else if date < expected {
				// Gap found — streak is broken
				break
			}
			// If date > expected, skip (shouldn't happen with DESC sort)
		}

		streaks[userID] = streak
	}

	return streaks, nil
}
