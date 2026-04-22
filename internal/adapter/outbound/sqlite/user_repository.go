package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/tien/go-discord-bot/internal/domain/entity"
)

// UserRepository implements outbound.UserRepository using SQLite.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new SQLite-backed user repository.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	query := `INSERT INTO users (discord_id, guild_id, leetcode_username) VALUES (?, ?, ?)`
	result, err := r.db.ExecContext(ctx, query, user.DiscordID, user.GuildID, user.LeetCodeUsername)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}
	user.ID = id

	return nil
}

func (r *UserRepository) DeleteByUsername(ctx context.Context, leetcodeUsername, guildID string) error {
	query := `DELETE FROM users WHERE leetcode_username = ? AND guild_id = ?`
	result, err := r.db.ExecContext(ctx, query, leetcodeUsername, guildID)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *UserRepository) GetByDiscordID(ctx context.Context, discordID, guildID string) ([]*entity.User, error) {
	query := `SELECT id, discord_id, guild_id, leetcode_username, created_at FROM users WHERE discord_id = ? AND guild_id = ?`
	rows, err := r.db.QueryContext(ctx, query, discordID, guildID)
	if err != nil {
		return nil, fmt.Errorf("query users: %w", err)
	}
	defer rows.Close()

	var users []*entity.User
	for rows.Next() {
		var user entity.User
		if err := rows.Scan(&user.ID, &user.DiscordID, &user.GuildID, &user.LeetCodeUsername, &user.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, &user)
	}

	return users, rows.Err()
}

func (r *UserRepository) GetByLeetCodeUsername(ctx context.Context, leetcodeUsername, guildID string) (*entity.User, error) {
	query := `SELECT id, discord_id, guild_id, leetcode_username, created_at FROM users WHERE leetcode_username = ? AND guild_id = ?`
	row := r.db.QueryRowContext(ctx, query, leetcodeUsername, guildID)

	var user entity.User
	if err := row.Scan(&user.ID, &user.DiscordID, &user.GuildID, &user.LeetCodeUsername, &user.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan user: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetByGuildID(ctx context.Context, guildID string) ([]*entity.User, error) {
	query := `SELECT id, discord_id, guild_id, leetcode_username, created_at FROM users WHERE guild_id = ?`
	rows, err := r.db.QueryContext(ctx, query, guildID)
	if err != nil {
		return nil, fmt.Errorf("query users: %w", err)
	}
	defer rows.Close()

	var users []*entity.User
	for rows.Next() {
		var user entity.User
		if err := rows.Scan(&user.ID, &user.DiscordID, &user.GuildID, &user.LeetCodeUsername, &user.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, &user)
	}

	return users, rows.Err()
}

func (r *UserRepository) GetAll(ctx context.Context) ([]*entity.User, error) {
	query := `SELECT id, discord_id, guild_id, leetcode_username, created_at FROM users`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query all users: %w", err)
	}
	defer rows.Close()

	var users []*entity.User
	for rows.Next() {
		var user entity.User
		if err := rows.Scan(&user.ID, &user.DiscordID, &user.GuildID, &user.LeetCodeUsername, &user.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, &user)
	}

	return users, rows.Err()
}
