package main

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/op/go-logging"
)

const (
	projectName     = "hyphenated-helper"
	discordZamielID = "71242588694249472"
)

var (
	logger         *logging.Logger
	discordSession *discordgo.Session
	discordToken   string
	discordBotID   string
)

func main() {
	// Initialize logging using the "go-logging" library
	// http://godoc.org/github.com/op/go-logging#Formatter
	logger = logging.MustGetLogger(projectName)
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
			"Copy the \".env_template\" file to \".env\".")
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
		discordSession = v
	}

	// Register function handlers for various events
	discordSession.AddHandler(discordReady)
	discordSession.AddHandler(discordMessageCreate)

	// Open the websocket and begin listening
	if err := discordSession.Open(); err != nil {
		logger.Fatal("Failed to open the Discord session:", err)
		return
	}

	// Block until a terminal signal is received
	logger.Info("Hyphen-ated helper is now running.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session
	discordSession.Close()
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
	if v, err := discordSession.Channel(m.ChannelID); err != nil {
		logger.Error("Failed to get the Discord channel of \""+m.ChannelID+"\":", err)
		return
	} else {
		channel = v
	}

	// Log the message
	logger.Info("[#" + channel.Name + "] " +
		"<" + m.Author.Username + "#" + m.Author.Discriminator + "> " + m.Content)

	discordCheckCommand(m)
}

func discordCheckCommand(m *discordgo.MessageCreate) {
	args := strings.Split(m.Content, " ")
	command := args[0]

	if !strings.HasPrefix(command, "/") {
		return
	}

	command = strings.TrimPrefix(command, "/")
	command = strings.ToLower(command) // Commands are case-insensitive

	switch command {
	case "d":
		commandDelete(m, 0)
	case "d1":
		commandDelete(m, 1)
	case "d2":
		commandDelete(m, 2)
	case "d3":
		commandDelete(m, 3)
	case "d4":
		commandDelete(m, 4)
	case "d5":
		commandDelete(m, 5)
	case "wrongchannel":
		msg := "It looks like you are asking a question about the Hyphen-ated conventions or the Hyphen-ated group. Please put all such questions in the #questions-and-help channel, as that's what it is for."
		discordSend(m.ChannelID, msg)
	}
}

func commandDelete(m *discordgo.MessageCreate, ruleNum int) {
	// Only certain people can use this command
	if m.Author.ID != discordZamielID {
		return
	}

	// Delete the "/d" message
	discordDelete(m.ChannelID, m.Message.ID)

	// Find the last message in the channel
	var messages []*discordgo.Message
	if v, err := discordSession.ChannelMessages(m.ChannelID, 1, "", "", ""); err != nil {
		logger.Errorf(
			"Failed to get the last message from channel %v: %v",
			m.ChannelID,
			err,
		)
		return
	} else {
		messages = v
	}

	if len(messages) == 0 {
		logger.Errorf("Failed to get any messages from channel %v.", m.ChannelID)
		return
	}

	lastMessage := messages[0]

	// Delete the last message
	discordDelete(m.ChannelID, lastMessage.ID)

	// Send them a message explaining why their message was deleted
	var ruleText string
	if ruleNum == 0 {
		ruleText = "one of the 5 rules"
	} else {
		ruleText = "rule #" + strconv.Itoa(ruleNum)
	}
	msg := "You asked the following question in the `#convention-questions` channel:\n"
	msg += "```\n"
	msg += lastMessage.Content
	msg += "```\n"
	msg += fmt.Sprintf(
		"An administrator thinks that this message might have broken %v, so it has been deleted.\n",
		ruleText,
	)
	msg += "Please make sure that your message follows the rules: <https://github.com/hanabi/hanabi.github.io/blob/main/misc/Convention_Questions.md>"
	discordSendPM(lastMessage.Author.ID, msg)
}

func discordSend(to string, msg string) {
	if _, err := discordSession.ChannelMessageSend(to, msg); err != nil {
		// Occasionally, sending messages to Discord can time out; if this occurs,
		// do not bother retrying, since losing a single message is fairly meaningless
		logger.Infof("Failed to send \"%v\" to Discord: %v", msg, err)
		return
	}
}

func discordSendPM(to string, msg string) {
	var PMChannel *discordgo.Channel
	if v, err := discordSession.UserChannelCreate(to); err != nil {
		logger.Errorf("Failed to get the PM channel for user %v: %v", to, err)
		return
	} else {
		PMChannel = v
	}

	discordSend(PMChannel.ID, msg)
}

func discordDelete(channelID string, messageID string) {
	if err := discordSession.ChannelMessageDelete(channelID, messageID); err != nil {
		logger.Errorf(
			"Failed to delete message %v from channel %v: %v",
			messageID,
			channelID,
			err,
		)
		return
	}
}
