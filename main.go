package main
import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"github.com/bwmarrin/discordgo"
)

var (
	botToken = os.Getenv("DISCORD_BOT_TOKEN")
	prefix   = "!"
	session  *discordgo.Session
)

func main() {
	if botToken == "" {
		log.Fatal("DISCORD_BOT_TOKEN environment variable not set")
	}

	var err error
	session, err = discordgo.New("Bot " + botToken)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}

	session.Identify.Intents = discordgo.IntentsGuildMessages |
		discordgo.IntentsGuildMembers |
		discordgo.IntentsGuilds |
		discordgo.IntentsMessageContent

	session.AddHandler(onReady)
	session.AddHandler(onMessageCreate)
	session.AddHandler(onGuildCreate)
	session.AddHandler(onGuildMemberAdd)

	err = session.Open()
	if err != nil {
		log.Fatalf("Error opening connection: %v", err)
	}
	defer session.Close()

	fmt.Println("Bot is now running. Press CTRL+C to exit.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	
	fmt.Println("\nShutting down gracefully...")
}

func onReady(s *discordgo.Session, event *discordgo.Ready) {
	log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	
	err := s.UpdateGameStatus(0, "!help for commands")
	if err != nil {
		log.Printf("Error setting status: %v", err)
	}
}

func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}

	if !strings.HasPrefix(m.Content, prefix) {
		return
	}

	args := strings.Fields(m.Content[len(prefix):])
	if len(args) == 0 {
		return
	}

	command := strings.ToLower(args[0])
	cmdArgs := args[1:]

	handleCommand(s, m, command, cmdArgs)
}

func onGuildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	if event.Guild.Unavailable {
		return
	}

	log.Printf("Joined guild: %s (ID: %s)", event.Guild.Name, event.Guild.ID)
	
	if event.Guild.SystemChannelID != "" {
		welcomeMsg := fmt.Sprintf("Thanks for adding me! Use `%shelp` to see available commands.", prefix)
		_, err := s.ChannelMessageSend(event.Guild.SystemChannelID, welcomeMsg)
		if err != nil {
			log.Printf("Error sending welcome message: %v", err)
		}
	}
}

func onGuildMemberAdd(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	guild, err := s.Guild(m.GuildID)
	if err != nil {
		log.Printf("Error getting guild: %v", err)
		return
	}

	if guild.SystemChannelID != "" {
		welcomeMsg := fmt.Sprintf("Welcome to %s, %s!", guild.Name, m.User.Mention())
		_, err := s.ChannelMessageSend(guild.SystemChannelID, welcomeMsg)
		if err != nil {
			log.Printf("Error sending member welcome: %v", err)
		}
	}
}

func handleCommand(s *discordgo.Session, m *discordgo.MessageCreate, cmd string, args []string) {
	switch cmd {
	case "help":
		cmdHelp(s, m, args)
	case "ping":
		cmdPing(s, m, args)
	case "info":
		cmdInfo(s, m, args)
	case "serverinfo":
		cmdServerInfo(s, m, args)
	case "userinfo":
		cmdUserInfo(s, m, args)
	case "kick":
		cmdKick(s, m, args)
	case "ban":
		cmdBan(s, m, args)
	case "clear":
		cmdClear(s, m, args)
	case "poll":
		cmdPoll(s, m, args)
	case "8ball":
		cmd8Ball(s, m, args)
	case "roll":
		cmdRoll(s, m, args)
	case "avatar":
		cmdAvatar(s, m, args)
	default:
		sendError(s, m.ChannelID, "Unknown command. Use `!help` for available commands.")
	}
}

func sendError(s *discordgo.Session, channelID, message string) {
	embed := &discordgo.MessageEmbed{
		Title:       "Error",
		Description: message,
		Color:       0xFF0000,
		Timestamp:   time.Now().Format(time.RFC3339),
	}
	_, err := s.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		log.Printf("Error sending error message: %v", err)
	}
}

func sendSuccess(s *discordgo.Session, channelID, message string) {
	embed := &discordgo.MessageEmbed{
		Title:       "Success",
		Description: message,
		Color:       0x00FF00,
		Timestamp:   time.Now().Format(time.RFC3339),
	}
	_, err := s.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		log.Printf("Error sending success message: %v", err)
	}
}

func hasPermission(s *discordgo.Session, guildID, userID string, permission int64) bool {
	member, err := s.GuildMember(guildID, userID)
	if err != nil {
		return false
	}

	guild, err := s.Guild(guildID)
	if err == nil && guild.OwnerID == userID {
		return true
	}

	for _, roleID := range member.Roles {
		role, err := s.State.Role(guildID, roleID)
		if err != nil {
			continue
		}

		if role.Permissions&permission != 0 || role.Permissions&discordgo.PermissionAdministrator != 0 {
			return true
		}
	}

	return false
}
