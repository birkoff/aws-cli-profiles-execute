package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	accountProfiles := map[string]string{
		"sandbox": "XXXXXXXXXXXX",
		"dev":     "XXXXXXXXXXXX",
		"qa":      "XXXXXXXXXXXX",
		"uat":     "XXXXXXXXXXXX",
		"staging": "XXXXXXXXXXXX",
		"prod":    "XXXXXXXXXXXX",
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: aws-cli-profiles-execute [AWS CLI command]")
		fmt.Println("Example: go run main.go \"aws s3 ls\"")
		return
	}

	awsCommand := os.Args[1]

	// Iterate over the profiles
	for profile, accountId := range accountProfiles {
		os.Setenv("AWS_PROFILE", profile)
		fmt.Println("Executing command for profile: " + profile)
		cmdArgs := strings.Fields(awsCommand)

		// The first item in cmdArgs must be "aws"
		if cmdArgs[0] != "aws" {
			fmt.Println("Invalid command. The first item in the command must be \"aws\"")
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
			fmt.Println(err)
			continue
		}

		fmt.Printf("Output for profile %s (Account ID: %s):\n%s\n", profile, accountId, string(output))
	}
}
