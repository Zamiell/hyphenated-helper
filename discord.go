package main

import (
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

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
		ruleNum := 0 // Default to not showing any particular rule
		if len(args) > 0 {
			if v, err := strconv.Atoi(args[0]); err == nil {
				ruleNum = v
			}
		}
		commandDelete(m, ruleNum)

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
		msg := "It looks like you are asking a question about the Hyphen-ated conventions or the Hyphen-ated group. Please put all such questions in the #questions-and-help channel - that's what it is for."
		discordSend(m.ChannelID, msg)
	}
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
