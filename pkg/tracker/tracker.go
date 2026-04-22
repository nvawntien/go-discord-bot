package tracker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/tien/go-discord-bot/internal/domain/entity"
	"github.com/tien/go-discord-bot/internal/domain/port/outbound"
)

const maxConcurrent = 5

// Tracker is a background worker that polls LeetCode for daily challenge completions
// and sends Discord notifications.
type Tracker struct {
	session    *discordgo.Session
	userRepo   outbound.UserRepository
	configRepo outbound.ConfigRepository
	compRepo   outbound.CompletionRepository
	lc         outbound.LeetCodeClient
	interval   time.Duration
	logger     *slog.Logger
}

// NewTracker creates a new background daily tracker.
func NewTracker(
	session *discordgo.Session,
	userRepo outbound.UserRepository,
	configRepo outbound.ConfigRepository,
	compRepo outbound.CompletionRepository,
	lc outbound.LeetCodeClient,
	interval time.Duration,
	logger *slog.Logger,
) *Tracker {
	return &Tracker{
		session:    session,
		userRepo:   userRepo,
		configRepo: configRepo,
		compRepo:   compRepo,
		lc:         lc,
		interval:   interval,
		logger:     logger,
	}
}

// Start begins the background polling loop. It blocks until the context is cancelled.
func (t *Tracker) Start(ctx context.Context) {
	t.logger.Info("Daily tracker started", "interval", t.interval)

	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()

	// Run immediately on start, then on each tick
	t.poll(ctx)

	for {
		select {
		case <-ctx.Done():
			t.logger.Info("Daily tracker stopped")
			return
		case <-ticker.C:
			t.poll(ctx)
		}
	}
}

// poll performs a single check cycle for all registered users.
func (t *Tracker) poll(ctx context.Context) {
	t.logger.Debug("Starting poll cycle")

	// 1. Fetch today's daily question
	daily, err := t.lc.GetDailyQuestion(ctx)
	if err != nil {
		t.logger.Error("Failed to fetch daily question", "error", err)
		return
	}

	t.logger.Debug("Daily question fetched",
		"title", daily.Title,
		"slug", daily.TitleSlug,
		"date", daily.Date,
	)

	// 2. Get all registered users
	users, err := t.userRepo.GetAll(ctx)
	if err != nil {
		t.logger.Error("Failed to fetch users", "error", err)
		return
	}

	if len(users) == 0 {
		return
	}

	// 3. Check each user concurrently with a semaphore
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	for _, user := range users {
		wg.Add(1)
		go func(u *entity.User) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			t.checkUser(ctx, u, daily)
		}(user)
	}

	wg.Wait()
	t.logger.Debug("Poll cycle complete", "users_checked", len(users))
}

// checkUser checks if a single user has completed the daily challenge and sends a notification if so.
func (t *Tracker) checkUser(ctx context.Context, user *entity.User, daily *entity.DailyQuestion) {
	// Check if already notified
	completed, err := t.compRepo.HasCompleted(ctx, user.ID, daily.Date)
	if err != nil {
		t.logger.Error("Failed to check completion",
			"user", user.LeetCodeUsername,
			"error", err,
		)
		return
	}
	if completed {
		return // Already notified
	}

	// Fetch recent accepted submissions
	submissions, err := t.lc.GetRecentAcceptedSubmissions(ctx, user.LeetCodeUsername, 20)
	if err != nil {
		t.logger.Warn("Failed to fetch submissions",
			"user", user.LeetCodeUsername,
			"error", err,
		)
		return
	}

	// Check if the daily question is in the recent submissions
	solved := false
	for _, sub := range submissions {
		if sub.TitleSlug == daily.TitleSlug {
			solved = true
			break
		}
	}

	if !solved {
		return
	}

	// Mark as completed
	completion := &entity.DailyCompletion{
		UserID:       user.ID,
		Date:         daily.Date,
		QuestionSlug: daily.TitleSlug,
	}

	if err := t.compRepo.MarkCompleted(ctx, completion); err != nil {
		t.logger.Error("Failed to mark completion",
			"user", user.LeetCodeUsername,
			"error", err,
		)
		return
	}

	// Send notification
	t.notify(user, daily)
}

// notify sends a Discord embed notification for a daily challenge completion.
func (t *Tracker) notify(user *entity.User, daily *entity.DailyQuestion) {
	channelID, err := t.configRepo.GetNotificationChannel(context.Background(), user.GuildID)
	if err != nil || channelID == "" {
		t.logger.Warn("No notification channel set for guild",
			"guild_id", user.GuildID,
		)
		return
	}

	color := 0x00D166 // Green (Easy)
	diffEmoji := "🟢"
	switch daily.Difficulty {
	case "Medium":
		color = 0xFEE75C
		diffEmoji = "🟡"
	case "Hard":
		color = 0xED4245
		diffEmoji = "🔴"
	}

	tags := "None"
	if len(daily.TopicTags) > 0 {
		tags = strings.Join(daily.TopicTags, ", ")
	}

	embed := &discordgo.MessageEmbed{
		Title: "🎉 Daily Challenge Completed!",
		Description: fmt.Sprintf(
			"<@%s> đã giải xong daily question hôm nay!\n\n**[%s. %s](%s)**",
			user.DiscordID,
			daily.QuestionID,
			daily.Title,
			daily.Link,
		),
		Color: color,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Difficulty",
				Value:  fmt.Sprintf("%s %s", diffEmoji, daily.Difficulty),
				Inline: true,
			},
			{
				Name:   "Topics",
				Value:  tags,
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("LeetCode: %s", user.LeetCodeUsername),
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	_, err = t.session.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		t.logger.Error("Failed to send notification",
			"channel_id", channelID,
			"user", user.LeetCodeUsername,
			"error", err,
		)
		return
	}

	t.logger.Info("Notification sent",
		"user", user.LeetCodeUsername,
		"question", daily.Title,
		"guild_id", user.GuildID,
	)
}
