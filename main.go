package main

import (
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

func main() {
	logFile, err := os.OpenFile("aws_cli_profiles_execution.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	multiWriter := io.MultiWriter(os.Stdout, logFile)

	infoLogger := log.New(multiWriter, "[INFO] ", log.Ldate|log.Ltime)
	errorLogger := log.New(multiWriter, "[ERROR] ", log.Ldate|log.Ltime)

	accountProfiles := map[string]string{
		"sandbox": "XXXXXXXXXXXX",
		"dev":     "XXXXXXXXXXXX",
		"qa":      "XXXXXXXXXXXX",
		"uat":     "XXXXXXXXXXXX",
		"staging": "XXXXXXXXXXXX",
		"prod":    "XXXXXXXXXXXX",
	}

	if len(os.Args) < 2 {
		errorLogger.Println("Usage: aws-cli-profiles-execute [AWS CLI command]")
		errorLogger.Println("Example: go run main.go \"aws s3 ls\"")
		return
	}

	awsCommand := os.Args[1]

	// Iterate over the profiles
	for profile, accountId := range accountProfiles {
		os.Setenv("AWS_PROFILE", profile)
		infoLogger.Println("Executing command for profile: " + profile)
		cmdArgs := strings.Fields(awsCommand)

		// The first item in cmdArgs must be "aws"
		if cmdArgs[0] != "aws" {
			errorLogger.Println("Invalid command. The first item in the command must be \"aws\"")
			return
		}

		// Checks if any item in cmdArgs contains the string "{account_id}" and replaces it with the accountId
		for i, _ := range cmdArgs {
			if strings.Contains(cmdArgs[i], "{account_id}") {
				cmdArgs[i] = strings.Replace(cmdArgs[i], "{account_id}", accountId, -1)
			}

			if strings.Contains(cmdArgs[i], "{profile}") {
				cmdArgs[i] = strings.Replace(cmdArgs[i], "{profile}", profile, -1)
			}
		}

		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		output, err := cmd.CombinedOutput()

		if err != nil {
			errorLogger.Println(err)
			continue
		}

		infoLogger.Printf("Output for profile %s (Account ID: %s):\n%s\n", profile, accountId, string(output))
	}
}
