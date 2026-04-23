package discord

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/tien/go-discord-bot/internal/application"
	"github.com/tien/go-discord-bot/internal/domain/entity"
)

// Handler routes Discord interactions to the appropriate use case.
type Handler struct {
	registerUC    *application.RegisterUserUseCase
	unregisterUC  *application.UnregisterUserUseCase
	statsUC       *application.GetStatsUseCase
	dailyUC       *application.GetDailyUseCase
	setChannelUC  *application.SetChannelUseCase
	leaderboardUC *application.GetLeaderboardUseCase
	logger        *slog.Logger
}

// NewHandler creates a new interaction handler.
func NewHandler(
	registerUC *application.RegisterUserUseCase,
	unregisterUC *application.UnregisterUserUseCase,
	statsUC *application.GetStatsUseCase,
	dailyUC *application.GetDailyUseCase,
	setChannelUC *application.SetChannelUseCase,
	leaderboardUC *application.GetLeaderboardUseCase,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		registerUC:    registerUC,
		unregisterUC:  unregisterUC,
		statsUC:       statsUC,
		dailyUC:       dailyUC,
		setChannelUC:  setChannelUC,
		leaderboardUC: leaderboardUC,
		logger:        logger,
	}
}

// OnInteraction handles incoming Discord interactions.
func (h *Handler) OnInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	data := i.ApplicationCommandData()
	ctx := context.Background()

	switch data.Name {
	case "register":
		h.handleRegister(ctx, s, i, data)
	case "unregister":
		h.handleUnregister(ctx, s, i)
	case "stats":
		h.handleStats(ctx, s, i, data)
	case "daily":
		h.handleDaily(ctx, s, i)
	case "setchannel":
		h.handleSetChannel(ctx, s, i)
	case "leaderboard":
		h.handleLeaderboard(ctx, s, i)
	}
}

func (h *Handler) handleRegister(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, data discordgo.ApplicationCommandInteractionData) {
	username := data.Options[0].StringValue()
	discordID := i.Member.User.ID
	guildID := i.GuildID

	// Defer response since LeetCode API call may take time
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	if err := h.registerUC.Register(ctx, discordID, guildID, username); err != nil {
		h.editErrorResponse(s, i, err.Error())
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "✅ Đăng ký thành công!",
		Description: fmt.Sprintf("Đã liên kết tài khoản LeetCode **%s** với <@%s>.\nBot sẽ theo dõi và thông báo khi bạn giải xong daily question!", username, discordID),
		Color:       0x00D166, // Green
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Dùng /stats để xem thống kê • /unregister để hủy đăng ký",
		},
	}

	_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})
}

func (h *Handler) handleUnregister(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	discordID := i.Member.User.ID
	guildID := i.GuildID

	if err := h.unregisterUC.Unregister(ctx, discordID, guildID); err != nil {
		h.respondError(s, i, err.Error())
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "👋 Đã hủy đăng ký",
		Description: "Tài khoản LeetCode của bạn đã được gỡ khỏi bot.",
		Color:       0xFEE75C, // Yellow
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

func (h *Handler) handleStats(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, data discordgo.ApplicationCommandInteractionData) {
	var username string
	if len(data.Options) > 0 {
		username = data.Options[0].StringValue()
	}
	discordID := i.Member.User.ID
	guildID := i.GuildID

	// Defer since API call may take time
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	stats, err := h.statsUC.GetUserStats(ctx, discordID, guildID, username)
	if err != nil {
		h.editErrorResponse(s, i, err.Error())
		return
	}

	embed := buildStatsEmbed(stats)
	_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})
}

func (h *Handler) handleDaily(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Defer since API call may take time
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	question, err := h.dailyUC.GetDailyQuestion(ctx)
	if err != nil {
		h.editErrorResponse(s, i, "Không thể lấy daily question. Thử lại sau!")
		return
	}

	embed := buildDailyEmbed(question)
	_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})
}

func (h *Handler) handleSetChannel(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	guildID := i.GuildID
	channelID := i.ChannelID

	if err := h.setChannelUC.SetNotificationChannel(ctx, guildID, channelID); err != nil {
		h.respondError(s, i, "Không thể set channel. Thử lại sau!")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "📢 Channel thông báo đã được set!",
		Description: fmt.Sprintf("Thông báo daily challenge sẽ được gửi vào <#%s>", channelID),
		Color:       0x5865F2, // Discord Blurple
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

func (h *Handler) handleLeaderboard(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	guildID := i.GuildID

	// Defer since multiple API calls needed
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	entries, err := h.leaderboardUC.GetLeaderboard(ctx, guildID)
	if err != nil {
		h.editErrorResponse(s, i, err.Error())
		return
	}

	embed := buildLeaderboardEmbed(entries)
	_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})
}

// respondError sends an ephemeral error response.
func (h *Handler) respondError(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) {
	embed := &discordgo.MessageEmbed{
		Title:       "❌ Lỗi",
		Description: msg,
		Color:       0xED4245, // Red
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}

// editErrorResponse edits a deferred response with an error message.
func (h *Handler) editErrorResponse(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) {
	embed := &discordgo.MessageEmbed{
		Title:       "❌ Lỗi",
		Description: msg,
		Color:       0xED4245, // Red
	}

	_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})
}

// buildStatsEmbed creates a rich embed for user statistics.
func buildStatsEmbed(stats *entity.UserStats) *discordgo.MessageEmbed {
	displayName := stats.Username
	if stats.RealName != "" {
		displayName = fmt.Sprintf("%s (%s)", stats.Username, stats.RealName)
	}

	return &discordgo.MessageEmbed{
		Title: fmt.Sprintf("📊 Stats — %s", displayName),
		Color: 0xFFA116, // LeetCode orange
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "🏆 Ranking",
				Value:  fmt.Sprintf("#%d", stats.Ranking),
				Inline: true,
			},
			{
				Name:   "📝 Total Solved",
				Value:  fmt.Sprintf("%d", stats.TotalSolved),
				Inline: true,
			},
			{
				Name:   "\u200B", // Zero-width space for spacing
				Value:  "\u200B",
				Inline: true,
			},
			{
				Name:   "🟢 Easy",
				Value:  fmt.Sprintf("%d", stats.EasySolved),
				Inline: true,
			},
			{
				Name:   "🟡 Medium",
				Value:  fmt.Sprintf("%d", stats.MediumSolved),
				Inline: true,
			},
			{
				Name:   "🔴 Hard",
				Value:  fmt.Sprintf("%d", stats.HardSolved),
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Data from LeetCode",
		},
	}
}

// buildDailyEmbed creates a rich embed for the daily challenge question.
func buildDailyEmbed(q *entity.DailyQuestion) *discordgo.MessageEmbed {
	color := 0x00D166 // Green (Easy)
	diffEmoji := "🟢"
	switch q.Difficulty {
	case "Medium":
		color = 0xFEE75C // Yellow
		diffEmoji = "🟡"
	case "Hard":
		color = 0xED4245 // Red
		diffEmoji = "🔴"
	}

	tags := "None"
	if len(q.TopicTags) > 0 {
		tags = strings.Join(q.TopicTags, ", ")
	}

	return &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("📅 Daily Challenge — %s", q.Date),
		Description: fmt.Sprintf("**[%s. %s](%s)**", q.QuestionID, q.Title, q.Link),
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
			Text: "Dùng /register để bot theo dõi tiến độ của bạn!",
		},
	}
}

// buildLeaderboardEmbed creates a rich embed for the leaderboard.
func buildLeaderboardEmbed(entries []application.LeaderboardEntry) *discordgo.MessageEmbed {
	var sb strings.Builder

	medals := []string{"🥇", "🥈", "🥉"}

	for i, entry := range entries {
		medal := fmt.Sprintf("`#%d`", i+1)
		if i < len(medals) {
			medal = medals[i]
		}

		sb.WriteString(fmt.Sprintf(
			"%s **%s** (<@%s>)\n┗ 📝 %d solved ┃ 🟢 %d ┃ 🟡 %d ┃ 🔴 %d\n\n",
			medal,
			entry.LeetCodeUsername,
			entry.DiscordID,
			entry.Stats.TotalSolved,
			entry.Stats.EasySolved,
			entry.Stats.MediumSolved,
			entry.Stats.HardSolved,
		))
	}

	return &discordgo.MessageEmbed{
		Title:       "🏆 Leaderboard",
		Description: sb.String(),
		Color:       0xFFA116, // LeetCode orange
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Tổng %d thành viên • Xếp theo Total Solved", len(entries)),
		},
	}
}
