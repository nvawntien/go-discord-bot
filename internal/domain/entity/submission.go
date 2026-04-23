package entity

import "time"

// Submission represents a single accepted submission from LeetCode.
type Submission struct {
	ID        string
	Title     string
	TitleSlug string
	Timestamp int64 // Unix timestamp
}

// DailyCompletion tracks that a user has completed the daily challenge.
// Used to avoid duplicate notifications.
type DailyCompletion struct {
	ID           int64
	UserID       int64
	Date         string // e.g. "2026-04-22"
	QuestionSlug string
	CompletedAt  time.Time
}

// UserStats holds aggregated LeetCode profile statistics.
type UserStats struct {
	Username    string
	RealName    string
	Ranking     int
	TotalSolved int
	EasySolved  int
	MediumSolved int
	HardSolved  int
}
