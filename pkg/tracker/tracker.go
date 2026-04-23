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
	lastDate   string // Track current date to detect day change
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
	// 1. Fetch today's daily question
	daily, err := t.lc.GetDailyQuestion(ctx)
	if err != nil {
		t.logger.Error("Failed to fetch daily question", "error", err)
		return
	}

	t.logger.Info("Poll cycle",
		"daily_title", daily.Title,
		"daily_slug", daily.TitleSlug,
		"daily_date", daily.Date,
	)

	// 2. Check if day changed — auto-post daily question to all guilds
	if daily.Date != t.lastDate {
		if t.lastDate == "" {
			// First poll on startup — just record the date, don't broadcast
			t.logger.Info("Initial date set", "date", daily.Date)
		} else {
			// Day actually changed while bot is running — broadcast
			t.logger.Info("New day detected, posting daily question", "date", daily.Date)
			t.broadcastDailyQuestion(daily)
		}
		t.lastDate = daily.Date
	}

	// 3. Get all registered users
	users, err := t.userRepo.GetAll(ctx)
	if err != nil {
		t.logger.Error("Failed to fetch users", "error", err)
		return
	}

	if len(users) == 0 {
		t.logger.Info("No registered users to check")
		return
	}

	// 4. Check each user concurrently with a semaphore
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
	t.logger.Info("Poll cycle complete", "users_checked", len(users))
}

// broadcastDailyQuestion sends the daily question to all guilds that have a notification channel set.
func (t *Tracker) broadcastDailyQuestion(daily *entity.DailyQuestion) {
	users, err := t.userRepo.GetAll(context.Background())
	if err != nil {
		t.logger.Error("Failed to fetch users for broadcast", "error", err)
		return
	}

	guildsSeen := make(map[string]bool)
	for _, user := range users {
		if guildsSeen[user.GuildID] {
			continue
		}
		guildsSeen[user.GuildID] = true

		channelID, err := t.configRepo.GetNotificationChannel(context.Background(), user.GuildID)
		if err != nil || channelID == "" {
			continue
		}

		embed := buildDailyQuestionEmbed(daily)
		_, err = t.session.ChannelMessageSendEmbed(channelID, embed)
		if err != nil {
			t.logger.Error("Failed to broadcast daily question",
				"guild_id", user.GuildID,
				"error", err,
			)
			continue
		}

		t.logger.Info("Daily question broadcasted",
			"guild_id", user.GuildID,
			"question", daily.Title,
		)
	}
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

	t.logger.Info("Checking user submissions",
		"user", user.LeetCodeUsername,
		"daily_slug", daily.TitleSlug,
		"recent_submissions_count", len(submissions),
	)

	// Check if the daily question is in the recent submissions — capture the submission
	var matchedSub *entity.Submission
	for _, sub := range submissions {
		if sub.TitleSlug == daily.TitleSlug {
			matchedSub = &sub
			t.logger.Info("Daily question SOLVED!",
				"user", user.LeetCodeUsername,
				"question", daily.Title,
				"submission_id", sub.ID,
			)
			break
		}
	}

	if matchedSub == nil {
		return
	}

	// Send notification first — only mark completed if notification succeeds
	if !t.notifyCompletion(ctx, user, daily, matchedSub) {
		return
	}

	// Mark as completed only after successful notification
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
	}
}

// notifyCompletion sends a Discord embed notification for a daily challenge completion.
// Includes the solution link and tags users who haven't completed the daily yet.
// Returns true if the notification was sent successfully.
func (t *Tracker) notifyCompletion(ctx context.Context, user *entity.User, daily *entity.DailyQuestion, submission *entity.Submission) bool {
	channelID, err := t.configRepo.GetNotificationChannel(ctx, user.GuildID)
	if err != nil || channelID == "" {
		t.logger.Warn("No notification channel set for guild",
			"guild_id", user.GuildID,
		)
		return false
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

	// Build solution link
	solutionLink := fmt.Sprintf("https://leetcode.com/submissions/detail/%s/", submission.ID)

	// Build description with solution link
	description := fmt.Sprintf(
		"<@%s> đã giải xong daily question hôm nay!\n\n**[%s. %s](%s)**\n🔗 [Xem bài giải](%s)",
		user.DiscordID,
		daily.QuestionID,
		daily.Title,
		daily.Link,
		solutionLink,
	)

	embed := &discordgo.MessageEmbed{
		Title:       "🎉 Daily Challenge Completed!",
		Description: description,
		Color:       color,
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

	// Build message content with non-completer mentions
	messageContent := t.buildNonCompleterMentions(ctx, user, daily)

	_, err = t.session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content: messageContent,
		Embeds:  []*discordgo.MessageEmbed{embed},
	})
	if err != nil {
		t.logger.Error("Failed to send completion notification",
			"channel_id", channelID,
			"user", user.LeetCodeUsername,
			"error", err,
		)
		return false
	}

	t.logger.Info("Completion notification sent",
		"user", user.LeetCodeUsername,
		"question", daily.Title,
		"guild_id", user.GuildID,
	)

	return true
}

// buildNonCompleterMentions builds a message tagging users who haven't completed the daily yet.
func (t *Tracker) buildNonCompleterMentions(ctx context.Context, completedUser *entity.User, daily *entity.DailyQuestion) string {
	guildUsers, err := t.userRepo.GetByGuildID(ctx, completedUser.GuildID)
	if err != nil {
		return ""
	}

	var pending []string
	for _, u := range guildUsers {
		if u.ID == completedUser.ID {
			continue
		}

		completed, err := t.compRepo.HasCompleted(ctx, u.ID, daily.Date)
		if err != nil || completed {
			continue
		}

		pending = append(pending, fmt.Sprintf("<@%s>", u.DiscordID))
	}

	if len(pending) == 0 {
		return ""
	}

	return fmt.Sprintf("⏳ Còn %s chưa giải daily hôm nay, cố lên nào! 💪", strings.Join(pending, ", "))
}

// buildDailyQuestionEmbed creates a rich embed for the daily challenge auto-post.
func buildDailyQuestionEmbed(q *entity.DailyQuestion) *discordgo.MessageEmbed {
	color := 0x00D166 // Green (Easy)
	diffEmoji := "🟢"
	switch q.Difficulty {
	case "Medium":
		color = 0xFEE75C
		diffEmoji = "🟡"
	case "Hard":
		color = 0xED4245
		diffEmoji = "🔴"
	}

	tags := "None"
	if len(q.TopicTags) > 0 {
		tags = strings.Join(q.TopicTags, ", ")
	}

	return &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("📅 Daily Challenge — %s", q.Date),
		Description: fmt.Sprintf("**[%s. %s](%s)**\n\nHãy giải bài này hôm nay! Bot sẽ tự động thông báo khi bạn hoàn thành. 💪", q.QuestionID, q.Title, q.Link),
		Color:       color,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Difficulty",
				Value:  fmt.Sprintf("%s %s", diffEmoji, q.Difficulty),
				Inline: true,
			},
			{
				Name:   "Topics",
				Value:  tags,
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "LeetCode Daily Challenge",
		},
	}
}
