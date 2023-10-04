package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"git.rpjosh.de/RPJosh/go-logger"
	"gitea.hama.de/LFS/infoniqa-scripts/internal/infoniqa"
	"gitea.hama.de/LFS/infoniqa-scripts/internal/models"
)

func main() {
	defer logger.CloseFile()

	// Configure logger
	logger.SetGlobalLogger(logger.GetLoggerFromEnv(&logger.Logger{
		ColoredOutput: true,
		Level:         logger.LevelInfo,
		PrintSource:   true,
		File:          &logger.FileLogger{},
	}))

	// Check if the first argument is --help
	if len(os.Args) == 1 || os.Args[1] == "--help" || os.Args[1] == "-h" || os.Args[1] == "?" {
		printHelp()
	}

	// Get the configuration of the app
	config := models.GetConfig()
	logger.Info("Program started")

	// Initialize infoniqa client
	inf, err := infoniqa.NewInfoniqa(config.Url, config.Username, config.Password)
	if err != nil {
		logger.Fatal("Initialization of infoniqa client was not successfull: %s", err)
	}

	switch strings.ToLower(os.Args[1]) {
	case "kommen":
		if err := inf.Kommen(); err != nil {
			logger.Fatal("Failed to book 'kommen': %s", err)
		}
	case "gehen":
		if err := inf.Gehen(); err != nil {
			logger.Fatal("Failed to book 'gehen': %s", err)
		}
	case "abwesend":
		if len(os.Args) <= 2 {
			logger.Fatal("Missing required parameter for option 'abwesend'")
		}

		// Parse the second argument to an int (amount of minutes)
		minutes, err := strconv.Atoi(os.Args[2])
		if err != nil {
			logger.Fatal("Failed to convert the argument %q to a number: %s", os.Args[2], err)
		}

		// Buche kommen und dann gehen
		if err := inf.Gehen(); err != nil {
			logger.Fatal("Failed to book 'kommen': %s", err)
		}
		logger.Info("Waiting %d minutes....", minutes)
		time.Sleep(time.Duration(minutes * int(time.Minute)))
		if err := inf.Kommen(); err != nil {
			logger.Fatal("Failed to book 'kommen': %s", err)
		}

	default:
		logger.Fatal("Invalid argument given: %q", os.Args[0])
	}

	logger.Info("Program executed successfull")
}

// printHelp prints a help for the usage of this program and exists the program afterwards
func printHelp() {
	fmt.Println(`
Command line arguments:
  kommen               Books "kommen"
  gehen                Books "gehen"
  abwesend [minutes]   Books "gehen" and waits the given amount of minutes for booking "kommen"

  --help            Prints this help

Environment variables:
  INFONIQA_CONFIG        File path of the configuration file to use (defaulting to ./config.yaml)`)

	os.Exit(0)
}
