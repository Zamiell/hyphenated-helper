package main

import (
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"
)

func commandDelete(m *discordgo.MessageCreate, ruleNum int) {
	// Only certain people can use this command
	if !stringInSlice(m.Author.ID, allowedDiscordIDs) {
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
