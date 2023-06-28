package main

import (
	"bufio"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

type AwsProfile struct {
	Name         string
	SSOStartUrl  string
	SSORegion    string
	SSOAccountId string
	SSORoleName  string
	Region       string
	Output       string
}

func main() {
	logFile, err := os.OpenFile("aws_cli_profiles_execution.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	multiWriter := io.MultiWriter(os.Stdout, logFile)

	infoLogger := log.New(multiWriter, "[INFO] ", log.Ldate|log.Ltime)
	errorLogger := log.New(multiWriter, "[ERROR] ", log.Ldate|log.Ltime)
	
	awsConfigFile := os.Getenv("HOME") + "/.aws/config"
	profiles, err := parseAwsConfigFile(awsConfigFile)

	if err != nil {
		errorLogger.Println("Error reading AWS config file: %v\n", err)
		return
	}

	if len(os.Args) < 3 {
		errorLogger.Println("Usage: aws-cli-profiles-execute [AWS CLI command] [profile1,profile2,...]")
		errorLogger.Println("Example: go run main.go \"aws s3 ls\" \"dev,qa\"")
		errorLogger.Println("Available profiles:")
		for _, profile := range profiles {
			infoLogger.Printf("Profile: %s\n", profile.Name)
			infoLogger.Printf("SSOStartUrl: %s\n", profile.SSOStartUrl)
			infoLogger.Printf("SSORegion: %s\n", profile.SSORegion)
			infoLogger.Printf("SSOAccountId: %s\n", profile.SSOAccountId)
			infoLogger.Printf("SSORoleName: %s\n", profile.SSORoleName)
			infoLogger.Printf("Region: %s\n", profile.Region)
			infoLogger.Printf("Output: %s\n", profile.Output)
			infoLogger.Println("--------------------------------------------------")
		}
		return
	}

	awsCommand := os.Args[1]
	profilesScope := os.Args[2]

	//context := []string{"dev", "qa", "uat", "prod"}
	context := strings.Split(profilesScope, ",")

	infoLogger.Println("--------------------------------------------------")
	infoLogger.Println("AWS CLI Profiles Execution")
	infoLogger.Println("AWS CLI Command: ", awsCommand)
	infoLogger.Println("Current Context: ", context)
	infoLogger.Println("--------------------------------------------------")
	for _, profile := range profiles {
		if !isInContext(context, profile.Name) {
			continue
		}
		infoLogger.Printf("Profile: %s\n", profile.Name)
		infoLogger.Printf("SSOStartUrl: %s\n", profile.SSOStartUrl)
		infoLogger.Printf("SSORegion: %s\n", profile.SSORegion)
		infoLogger.Printf("SSOAccountId: %s\n", profile.SSOAccountId)
		infoLogger.Printf("SSORoleName: %s\n", profile.SSORoleName)
		infoLogger.Printf("Region: %s\n", profile.Region)
		infoLogger.Printf("Output: %s\n", profile.Output)
		infoLogger.Println("--------------------------------------------------")
	}

	cmdArgs := strings.Fields(awsCommand)

	// The first item in cmdArgs must be "aws"
	if cmdArgs[0] != "aws" {
		errorLogger.Println("Invalid command. The first item in the command must be \"aws\"")
		return
	}

	// Iterate over the profiles
	for _, profile := range profiles {
		// Checks if the profile is in the context
		if !isInContext(context, profile.Name) {
			continue
		}

		// Set the original command to be executed
		execCmd := cmdArgs

		// Replace account_id and profile in the command with the actual values
		// Checks if any item in execCmd contains the string "{account_id}" and replaces it with the accountId
		for i, _ := range execCmd {
			if strings.Contains(execCmd[i], "{account_id}") {
				execCmd[i] = strings.Replace(execCmd[i], "{account_id}", profile.SSOAccountId, -1)
			}

			if strings.Contains(execCmd[i], "{profile}") {
				execCmd[i] = strings.Replace(execCmd[i], "{profile}", profile.Name, -1)
			}
		}

		infoLogger.Println("Executing command for profile: " + profile.Name)
		os.Setenv("AWS_PROFILE", profile.Name)
		cmd := exec.Command(execCmd[0], execCmd[1:]...)
		output, err := cmd.CombinedOutput()

		if err != nil {
			errorLogger.Println(err)
			continue
		}

		infoLogger.Printf("Output for profile %s (Account ID: %s):\n%s\n", profile.Name, profile.SSOAccountId, string(output))
	}
}

func isInContext(context []string, profile string) bool {
	for _, c := range context {
		if c == profile {
			return true
		}
	}
	return false
}
func parseAwsConfigFile(awsConfigFile string) ([]AwsProfile, error) {
	file, err := os.Open(awsConfigFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var profiles []AwsProfile
	var currProfile *AwsProfile

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		// Continue if line is empty or a comment
		if len(line) == 0 || line[0] == '#' || line[0] == ';' {
			continue
		}

		if strings.HasPrefix(line, "[profile ") {
			if currProfile != nil {
				profiles = append(profiles, *currProfile)
			}
			profileName := strings.TrimSuffix(strings.TrimPrefix(line, "[profile "), "]")
			currProfile = &AwsProfile{Name: profileName}
		} else if currProfile != nil {
			lineParts := strings.SplitN(line, "=", 2)
			// Just in case
			if len(lineParts) != 2 {
				continue
			}

			key := strings.TrimSpace(lineParts[0])
			value := strings.TrimSpace(lineParts[1])

			switch key {
			case "sso_start_url":
				currProfile.SSOStartUrl = value
			case "sso_region":
				currProfile.SSORegion = value
			case "sso_account_id":
				currProfile.SSOAccountId = value
			case "sso_role_name":
				currProfile.SSORoleName = value
			case "region":
				currProfile.Region = value
			case "output":
				currProfile.Output = value
			}
		}
	}
	if currProfile != nil {
		profiles = append(profiles, *currProfile)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return profiles, nil
}
