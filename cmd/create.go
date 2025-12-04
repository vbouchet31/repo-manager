package cmd

import (
	"fmt"
	"strings"

	"meo-repo-manager/config"
	"meo-repo-manager/gh"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new repository",
	Run:   runCreate,
}

func init() {
	rootCmd.AddCommand(createCmd)
}

func runCreate(cmd *cobra.Command, args []string) {
	// 1. Repository Name
	var name string
	prompt := &survey.Input{
		Message: "Repository Name:",
	}
	if err := survey.AskOne(prompt, &name); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	prefix := config.AppConfig.Prefix
	if prefix != "" {
		if strings.HasPrefix(name, prefix) {
			fmt.Printf("\033[33m!\033[0m Warning: Prefix '%s' is already present in the name. It will be stripped and re-added.\n", prefix)
			name = strings.TrimPrefix(name, prefix)
		}
		name = prefix + name
	}

	// 2. Select Users
	var selectedUsers []string
	if len(config.AppConfig.Users) > 0 {
		userPrompt := &survey.MultiSelect{
			Message: "Select users to add:",
			Options: config.AppConfig.Users,
			Default: config.AppConfig.Users,
		}
		if err := survey.AskOne(userPrompt, &selectedUsers); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
	}

	// 3. Add More Users
	var addMore bool
	confirmMore := &survey.Confirm{
		Message: "Add more users?",
		Default: false,
	}
	if err := survey.AskOne(confirmMore, &addMore); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if addMore {
		var moreUsersStr string
		moreUsersPrompt := &survey.Input{
			Message: "Enter usernames (comma separated):",
		}
		if err := survey.AskOne(moreUsersPrompt, &moreUsersStr); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		if moreUsersStr != "" {
			parts := strings.Split(moreUsersStr, ",")
			for _, p := range parts {
				trimmed := strings.TrimSpace(p)
				if trimmed != "" {
					selectedUsers = append(selectedUsers, trimmed)
				}
			}
		}
	}

	// 4. Recap
	fmt.Println("\n--- Recap ---")
	fmt.Printf("Repository: %s/%s\n", config.AppConfig.Organization, name)
	fmt.Printf("Users to add: %v\n", selectedUsers)
	fmt.Println("-------------")

	// 5. Confirm
	var confirm bool
	confirmPrompt := &survey.Confirm{
		Message: "Proceed with creation?",
		Default: false,
	}
	if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if !confirm {
		fmt.Println("Aborted.")
		return
	}

	// 6. Execution
	client := gh.NewClient()
	fmt.Printf("Creating repository %s/%s...\n", config.AppConfig.Organization, name)
	if err := client.CreateRepository(config.AppConfig.Organization, name); err != nil {
		fmt.Printf("Error creating repository: %v\n", err)
		return
	}
	fmt.Println("Repository created successfully.")

	for _, user := range selectedUsers {
		fmt.Printf("Adding user %s...\n", user)
		if err := client.AddCollaborator(config.AppConfig.Organization, name, user, "push"); err != nil {
			fmt.Printf("Error adding user %s: %v\n", user, err)
		} else {
			fmt.Printf("User %s added.\n", user)
		}
	}

	fmt.Println("Done.")
}
