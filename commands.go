package main
import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"github.com/bwmarrin/discordgo"
)

func cmdHelp(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	embed := &discordgo.MessageEmbed{
		Title:       "Bot Commands",
		Description: "Here are all available commands:",
		Color:       0x3498DB,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "General",
				Value:  "`!help` - Show this help message\n`!ping` - Check bot latency\n`!info` - Bot information\n`!avatar [@user]` - Get user's avatar",
				Inline: false,
			},
			{
				Name:   "Information",
				Value:  "`!serverinfo` - Server information\n`!userinfo [@user]` - User information",
				Inline: false,
			},
			{
				Name:   "Moderation",
				Value:  "`!kick @user [reason]` - Kick a user\n`!ban @user [reason]` - Ban a user\n`!clear <amount>` - Clear messages (max 100)",
				Inline: false,
			},
			{
				Name:   "Fun",
				Value:  "`!8ball <question>` - Ask the magic 8-ball\n`!roll [max]` - Roll a dice\n`!poll <question>` - Create a yes/no poll",
				Inline: false,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Requested by %s", m.Author.Username),
		},
	}

	_, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
	if err != nil {
		sendError(s, m.ChannelID, "Failed to send help message.")
	}
}

func cmdPing(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	start := time.Now()
	msg, err := s.ChannelMessageSend(m.ChannelID, "Pinging...")
	if err != nil {
		sendError(s, m.ChannelID, "Failed to send ping message.")
		return
	}

	latency := time.Since(start).Milliseconds()

	embed := &discordgo.MessageEmbed{
		Title: "Pong!",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Bot Latency",
				Value:  fmt.Sprintf("%dms", latency),
				Inline: true,
			},
			{
				Name:   "API Latency",
				Value:  fmt.Sprintf("%dms", s.HeartbeatLatency().Milliseconds()),
				Inline: true,
			},
		},
		Color:     0x2ECC71,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	_, err = s.ChannelMessageEditEmbed(m.ChannelID, msg.ID, embed)
	if err != nil {
		sendError(s, m.ChannelID, "Failed to update ping message.")
	}
}

func cmdInfo(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	guilds := len(s.State.Guilds)

	embed := &discordgo.MessageEmbed{
		Title:       "ℹ️ Bot Information",
		Description: "A multipurpose Discord bot template built with Go",
		Color:       0x9B59B6,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Servers",
				Value:  fmt.Sprintf("%d", guilds),
				Inline: true,
			},
			{
				Name:   "Language",
				Value:  "Go (Golang)",
				Inline: true,
			},
			{
				Name:   "Library",
				Value:  "DiscordGo",
				Inline: true,
			},
			{
				Name:   "Prefix",
				Value:  fmt.Sprintf("`%s`", prefix),
				Inline: true,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: s.State.User.AvatarURL(""),
		},
	}

	_, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
	if err != nil {
		sendError(s, m.ChannelID, "Failed to send info message.")
	}
}

func cmdServerInfo(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	guild, err := s.Guild(m.GuildID)
	if err != nil {
		sendError(s, m.ChannelID, "Failed to retrieve server information.")
		return
	}

	owner, _ := s.User(guild.OwnerID)
	ownerName := "Unknown"
	if owner != nil {
		ownerName = owner.Username
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("📊 %s", guild.Name),
		Description: fmt.Sprintf("Server ID: `%s`", guild.ID),
		Color:       0xE67E22,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Owner",
				Value:  ownerName,
				Inline: true,
			},
			{
				Name:   "Members",
				Value:  fmt.Sprintf("%d", guild.MemberCount),
				Inline: true,
			},
			{
				Name:   "Created",
				Value:  fmt.Sprintf("<t:%d:R>", guild.ID>>22+1420070400000/1000),
				Inline: true,
			},
			{
				Name:   "Channels",
				Value:  fmt.Sprintf("%d", len(guild.Channels)),
				Inline: true,
			},
			{
				Name:   "Roles",
				Value:  fmt.Sprintf("%d", len(guild.Roles)),
				Inline: true,
			},
			{
				Name:   "Emojis",
				Value:  fmt.Sprintf("%d", len(guild.Emojis)),
				Inline: true,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: guild.IconURL(""),
		},
	}

	_, err = s.ChannelMessageSendEmbed(m.ChannelID, embed)
	if err != nil {
		sendError(s, m.ChannelID, "Failed to send server info.")
	}
}

func cmdUserInfo(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	var userID string
	if len(m.Mentions) > 0 {
		userID = m.Mentions[0].ID
	} else {
		userID = m.Author.ID
	}

	user, err := s.User(userID)
	if err != nil {
		sendError(s, m.ChannelID, "Failed to retrieve user information.")
		return
	}

	member, err := s.GuildMember(m.GuildID, userID)
	if err != nil {
		sendError(s, m.ChannelID, "Failed to retrieve member information.")
		return
	}

	roles := make([]string, 0)
	for _, roleID := range member.Roles {
		role, err := s.State.Role(m.GuildID, roleID)
		if err == nil {
			roles = append(roles, role.Mention())
		}
	}

	rolesText := "None"
	if len(roles) > 0 {
		rolesText = strings.Join(roles, ", ")
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("👤 %s", user.Username),
		Description: fmt.Sprintf("User ID: `%s`", user.ID),
		Color:       0x1ABC9C,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Account Created",
				Value:  fmt.Sprintf("<t:%d:R>", user.ID>>22+1420070400000/1000),
				Inline: true,
			},
			{
				Name:   "Joined Server",
				Value:  fmt.Sprintf("<t:%d:R>", member.JoinedAt.Unix()),
				Inline: true,
			},
			{
				Name:   "Roles",
				Value:  rolesText,
				Inline: false,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: user.AvatarURL(""),
		},
	}

	_, err = s.ChannelMessageSendEmbed(m.ChannelID, embed)
	if err != nil {
		sendError(s, m.ChannelID, "Failed to send user info.")
	}
}

func cmdKick(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if !hasPermission(s, m.GuildID, m.Author.ID, discordgo.PermissionKickMembers) {
		sendError(s, m.ChannelID, "You don't have permission to kick members.")
		return
	}

	if len(m.Mentions) == 0 {
		sendError(s, m.ChannelID, "Please mention a user to kick.")
		return
	}

	target := m.Mentions[0]
	reason := "No reason provided"
	if len(args) > 0 {
		reason = strings.Join(args, " ")
	}

	if target.ID == m.Author.ID {
		sendError(s, m.ChannelID, "You cannot kick yourself!")
		return
	}

	err := s.GuildMemberDeleteWithReason(m.GuildID, target.ID, reason)
	if err != nil {
		sendError(s, m.ChannelID, fmt.Sprintf("Failed to kick user: %v", err))
		return
	}

	sendSuccess(s, m.ChannelID, fmt.Sprintf("Kicked %s. Reason: %s", target.Mention(), reason))
}

func cmdBan(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if !hasPermission(s, m.GuildID, m.Author.ID, discordgo.PermissionBanMembers) {
		sendError(s, m.ChannelID, "You don't have permission to ban members.")
		return
	}

	if len(m.Mentions) == 0 {
		sendError(s, m.ChannelID, "Please mention a user to ban.")
		return
	}

	target := m.Mentions[0]
	reason := "No reason provided"
	if len(args) > 0 {
		reason = strings.Join(args, " ")
	}

	if target.ID == m.Author.ID {
		sendError(s, m.ChannelID, "You cannot ban yourself!")
		return
	}

	err := s.GuildBanCreateWithReason(m.GuildID, target.ID, reason, 0)
	if err != nil {
		sendError(s, m.ChannelID, fmt.Sprintf("Failed to ban user: %v", err))
		return
	}

	sendSuccess(s, m.ChannelID, fmt.Sprintf("Banned %s. Reason: %s", target.Mention(), reason))
}

func cmdClear(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if !hasPermission(s, m.GuildID, m.Author.ID, discordgo.PermissionManageMessages) {
		sendError(s, m.ChannelID, "You don't have permission to manage messages.")
		return
	}

	if len(args) == 0 {
		sendError(s, m.ChannelID, "Please specify the number of messages to delete (max 100).")
		return
	}

	amount, err := strconv.Atoi(args[0])
	if err != nil || amount < 1 || amount > 100 {
		sendError(s, m.ChannelID, "Please provide a valid number between 1 and 100.")
		return
	}

	messages, err := s.ChannelMessages(m.ChannelID, amount, "", "", "")
	if err != nil {
		sendError(s, m.ChannelID, "Failed to fetch messages.")
		return
	}

	messageIDs := make([]string, len(messages))
	for i, msg := range messages {
		messageIDs[i] = msg.ID
	}

	err = s.ChannelMessagesBulkDelete(m.ChannelID, messageIDs)
	if err != nil {
		sendError(s, m.ChannelID, "Failed to delete messages.")
		return
	}

	s.ChannelMessageDelete(m.ChannelID, m.ID)

	msg, _ := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Deleted %d messages.", len(messageIDs)))
	time.AfterFunc(3*time.Second, func() {
		s.ChannelMessageDelete(m.ChannelID, msg.ID)
	})
}

func cmdPoll(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) == 0 {
		sendError(s, m.ChannelID, "Please provide a question for the poll.")
		return
	}

	question := strings.Join(args, " ")

	embed := &discordgo.MessageEmbed{
		Title:       "Poll",
		Description: question,
		Color:       0x3498DB,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Poll by %s", m.Author.Username),
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	msg, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
	if err != nil {
		sendError(s, m.ChannelID, "Failed to create poll.")
		return
	}

	s.MessageReactionAdd(m.ChannelID, msg.ID, "👍")
	s.MessageReactionAdd(m.ChannelID, msg.ID, "👎")
}

func cmd8Ball(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) == 0 {
		sendError(s, m.ChannelID, "Please ask a question!")
		return
	}

	responses := []string{
		"It is certain.", "It is decidedly so.", "Without a doubt.",
		"Yes definitely.", "You may rely on it.", "As I see it, yes.",
		"Most likely.", "Outlook good.", "Yes.", "Signs point to yes.",
		"Reply hazy, try again.", "Ask again later.", "Better not tell you now.",
		"Cannot predict now.", "Concentrate and ask again.",
		"Don't count on it.", "My reply is no.", "My sources say no.",
		"Outlook not so good.", "Very doubtful.",
	}

	rand.Seed(time.Now().UnixNano())
	response := responses[rand.Intn(len(responses))]

	embed := &discordgo.MessageEmbed{
		Title:       "Magic 8-Ball",
		Description: fmt.Sprintf("**Question:** %s\n\n**Answer:** %s", strings.Join(args, " "), response),
		Color:       0x9B59B6,
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	_, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
	if err != nil {
		sendError(s, m.ChannelID, "Failed to send 8-ball response.")
	}
}

func cmdRoll(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	max := 6
	if len(args) > 0 {
		val, err := strconv.Atoi(args[0])
		if err == nil && val > 0 && val <= 1000 {
			max = val
		}
	}

	rand.Seed(time.Now().UnixNano())
	result := rand.Intn(max) + 1

	embed := &discordgo.MessageEmbed{
		Title:       "Dice Roll",
		Description: fmt.Sprintf("You rolled a **%d** (1-%d)", result, max),
		Color:       0xE74C3C,
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	_, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
	if err != nil {
		sendError(s, m.ChannelID, "Failed to roll dice.")
	}
}

func cmdAvatar(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	var user *discordgo.User
	if len(m.Mentions) > 0 {
		user = m.Mentions[0]
	} else {
		user = m.Author
	}

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("%s's Avatar", user.Username),
		Color: 0x1ABC9C,
		Image: &discordgo.MessageEmbedImage{
			URL: user.AvatarURL("1024"),
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	_, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
	if err != nil {
		sendError(s, m.ChannelID, "Failed to fetch avatar.")
	}
}
