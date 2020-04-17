package main

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/op/go-logging"
)

var (
	logger       *logging.Logger
	discord      *discordgo.Session
	discordToken string
	discordBotID string
)

func main() {
	// Initialize logging using the "go-logging" library
	// http://godoc.org/github.com/op/go-logging#Formatter
	logger = logging.MustGetLogger("hanabi-live")
	loggingBackend := logging.NewLogBackend(os.Stdout, "", 0)
	logFormat := logging.MustStringFormatter( // https://golang.org/pkg/time/#Time.Format
		`%{time:Mon Jan 02 15:04:05 MST 2006} - %{level:.4s} - %{shortfile} - %{message}`,
	)
	loggingBackendFormatted := logging.NewBackendFormatter(loggingBackend, logFormat)
	logging.SetBackend(loggingBackendFormatted)

	// Get the project path
	// https://stackoverflow.com/questions/18537257/
	var projectPath string
	if v, err := os.Executable(); err != nil {
		logger.Fatal("Failed to get the path of the currently running executable:", err)
	} else {
		projectPath = filepath.Dir(v)
	}

	// Check to see if the ".env" file exists
	envPath := path.Join(projectPath, ".env")
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		logger.Fatal("The \"" + envPath + "\" file does not exist. " +
			"Did you run the \"install_dependencies.sh\" script before running the server? " +
			"This file should automatically be created when running this script.")
		return
	}

	// Load the ".env" file which contains environment variables with secret values
	if err := godotenv.Load(envPath); err != nil {
		logger.Fatal("Failed to load the \".env\" file:", err)
		return
	}

	// Read some configuration values from environment variables
	discordToken = os.Getenv("DISCORD_TOKEN")
	if len(discordToken) == 0 {
		logger.Fatal("The \"DISCORD_TOKEN\" environment variable is blank.")
		return
	}

	// Bot accounts must be prefixed with "Bot"
	if v, err := discordgo.New("Bot " + discordToken); err != nil {
		logger.Error("Failed to create a Discord session:", err)
		return
	} else {
		discord = v
	}

	// Register function handlers for various events
	discord.AddHandler(discordReady)
	discord.AddHandler(discordMessageCreate)

	// Open the websocket and begin listening
	if err := discord.Open(); err != nil {
		logger.Fatal("Failed to open the Discord session:", err)
		return
	}
}

func discordReady(s *discordgo.Session, event *discordgo.Ready) {
	logger.Info("Discord bot connected with username: " + event.User.Username)
	discordBotID = event.User.ID
}

func discordMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == discordBotID {
		return
	}

	// Get the channel
	var channel *discordgo.Channel
	if v, err := discord.Channel(m.ChannelID); err != nil {
		logger.Error("Failed to get the Discord channel of \""+m.ChannelID+"\":", err)
		return
	} else {
		channel = v
	}

	// Log the message
	logger.Info("[#" + channel.Name + "] " +
		"<" + m.Author.Username + "#" + m.Author.Discriminator + "> " + m.Content)

	args := strings.Split(m.Content, " ")
	command := args[0]

	// Commands will start with a "/", so we can ignore everything else
	if !strings.HasPrefix(command, "/") {
		return
	}
	command = strings.TrimPrefix(command, "/")
	command = strings.ToLower(command) // Commands are case-insensitive

	if command == "badquestion" {
		msg := "Your question is not specific enough. In order to properly answer it, we need to know the amount of players, all of the cards in all of the hands, the amount of clues, and so forth. Please type out a full Alice and Bob story in the style of the reference document. (e.g. <https://github.com/Zamiell/hanabi-conventions/blob/master/Reference.md#the-reverse-finesse>)"
		discordSend(m.ChannelID, msg)
	} else if command == "badhere" {
		msg := "It looks like there are already some Hyphen-ated members playing in one of the voice channels. Did you already ask to join them in their next game? Unless there is a specific reason (e.g. experts playing a variant that is too difficult for a beginner), then you should make an effort to join existing players **before** pinging the rest of the server. The \"/here\" command is generally only used when there is no-one around and playing a game already."
		discordSend(m.ChannelID, msg)
	}
}

func discordSend(to string, msg string) {
	if _, err := discord.ChannelMessageSend(to, msg); err != nil {
		// Occasionally, sending messages to Discord can time out; if this occurs,
		// do not bother retrying, since losing a single message is fairly meaningless
		logger.Info("Failed to send \""+msg+"\" to Discord:", err)
		return
	}
}
