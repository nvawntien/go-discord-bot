package discord

import "github.com/bwmarrin/discordgo"

// Slash command definitions for the bot.
var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "register",
		Description: "Đăng ký LeetCode username để bot theo dõi",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "username",
				Description: "LeetCode username của bạn",
				Required:    true,
			},
		},
	},
	{
		Name:        "unregister",
		Description: "Hủy đăng ký LeetCode username của bạn khỏi bot",
	},
	{
		Name:        "stats",
		Description: "Xem thống kê LeetCode của bạn hoặc người khác",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "username",
				Description: "LeetCode username (bỏ trống = xem stats của mình)",
				Required:    false,
			},
		},
	},
	{
		Name:        "daily",
		Description: "Xem daily challenge LeetCode hôm nay",
	},
	{
		Name:                     "setchannel",
		Description:              "Set channel hiện tại nhận thông báo daily (Admin only)",
		DefaultMemberPermissions: permPtr(discordgo.PermissionAdministrator),
	},
	{
		Name:        "leaderboard",
		Description: "Xem bảng xếp hạng LeetCode trong server",
	},
}

// permPtr returns a pointer to an int64, used for DefaultMemberPermissions.
func permPtr(p int64) *int64 {
	return &p
}
