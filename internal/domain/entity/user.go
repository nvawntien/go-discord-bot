package entity

import "time"

// User represents a registered Discord user linked to their LeetCode account.
type User struct {
	ID               int64
	DiscordID        string
	GuildID          string
	LeetCodeUsername string
	CreatedAt        time.Time
}
