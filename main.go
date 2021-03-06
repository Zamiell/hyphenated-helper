package main

import (
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/op/go-logging"
)

const (
	projectName = "hyphenated-helper"
)

var (
	logger            *logging.Logger
	discordSession    *discordgo.Session
	discordToken      string
	discordBotID      string
	allowedDiscordIDs = []string{
		"71242588694249472",  // Zamiel
		"248637602624700417", // Padi
	}
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
