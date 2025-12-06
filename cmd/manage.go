package cmd

import (
	"fmt"
	"os"
	"strings"

	"meo-repo-manager/config"
	"meo-repo-manager/gh"

	"github.com/google/go-github/v69/github"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var manageCmd = &cobra.Command{
	Use:   "manage",
	Short: "Manage an existing repository",
	Run:   runManage,
}

func init() {
	rootCmd.AddCommand(manageCmd)
	manageCmd.Flags().String("repo", "", "Repository name to manage")
}

func runManage(cmd *cobra.Command, args []string) {
	// 0. Validate Token
	token := viper.GetString("github-token")
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	if token == "" {
		fmt.Println("Error: GitHub token is required. Set GITHUB_TOKEN env var or use --github-token flag.")
		return
	}

	client := gh.NewClient(token)
	fmt.Print("Validating token... ")
	if err := client.Validate(); err != nil {
		fmt.Printf("Failed!\nError: %v\n", err)
		return
	}
	fmt.Println("OK!")

	// 1. Select Repository
	repoName, _ := cmd.Flags().GetString("repo")
	org := config.AppConfig.Organization

	if repoName != "" {
		// Validate existence
		fmt.Printf("Checking repository %s/%s...\n", org, repoName)
		if _, err := client.GetRepository(org, repoName); err != nil {
			fmt.Printf("Error: Repository '%s' not found in organization '%s' (or you don't have access).\n", repoName, org)
			return
		}
	} else {
		// List and select
		fmt.Printf("Fetching repositories in %s with prefix '%s'...\n", org, config.AppConfig.Prefix)
		repos, err := client.ListRepositories(org, config.AppConfig.Prefix)
		if err != nil {
			fmt.Printf("Error listing repositories: %v\n", err)
			return
		}
		if len(repos) == 0 {
			fmt.Println("No matching repositories found.")
			return
		}

		var repoNames []string
		for _, r := range repos {
			if r.Name != nil {
				repoNames = append(repoNames, *r.Name)
			}
		}

		prompt := &survey.Select{
			Message: "Select repository to manage:",
			Options: repoNames,
		}
		if err := survey.AskOne(prompt, &repoName); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
	}

	// 2. Manage Users
	fmt.Printf("Fetching collaborators for %s/%s...\n", org, repoName)
	currentCollabs, err := client.ListCollaborators(org, repoName)
	if err != nil {
		fmt.Printf("Error fetching collaborators: %v\n", err)
		return
	}

	currentCollabMap := make(map[string]bool)
	var allOptions []string
	var defaultSelection []string

	// 1. Add config users first (priority)
	configUserMap := make(map[string]bool)
	for _, u := range config.AppConfig.Users {
		configUserMap[u] = true

		// Check if user is a current collaborator to display role
		displayStr := u
		if currentCollabMap[u] {
			// Find role
			for _, cu := range currentCollabs {
				if cu.Login != nil && *cu.Login == u {
					if cu.Permissions != nil {
						role := "read"
						if cu.Permissions["admin"] {
							role = "admin"
						} else if cu.Permissions["push"] {
							role = "write"
						} else if cu.Permissions["maintain"] {
							role = "maintain"
						}
						displayStr = fmt.Sprintf("%s (%s)", u, role)
					}
					break
				}
			}
		} else {
			// User in config but NOT in collaborators list.
			// Check if they actually have access (e.g. owner/admin implicit access)
			// This avoids showing them as unchecked if they actually have access.
			perm, err := client.GetPermissionLevel(org, repoName, u)
			if err == nil && perm != nil && perm.Permission != nil {
				p := *perm.Permission
				if p == "admin" || p == "write" || p == "maintain" || p == "read" {
					// They have access!
					role := p
					displayStr = fmt.Sprintf("%s (%s)", u, role)
					// Mark them as present so they get selected by default
					currentCollabMap[u] = true

					// Add to currentCollabs list temporarily so we can check admin status later
					if p == "admin" {
						adminUser := &github.User{
							Login:       github.Ptr(u),
							Permissions: map[string]bool{"admin": true},
						}
						currentCollabs = append(currentCollabs, adminUser)
					}
				}
			}
		}

		allOptions = append(allOptions, displayStr)
	}

	// Build map of current collaborators
	for _, u := range currentCollabs {
		if u.Login != nil {
			currentCollabMap[*u.Login] = true
		}
	}

	// Now build default selection for config users who are already collaborators
	for _, u := range config.AppConfig.Users {
		if currentCollabMap[u] {
			// Reconstruct display string to match option
			displayStr := u
			for _, cu := range currentCollabs {
				if cu.Login != nil && *cu.Login == u {
					if cu.Permissions != nil {
						role := "read"
						if cu.Permissions["admin"] {
							role = "admin"
						} else if cu.Permissions["push"] {
							role = "write"
						} else if cu.Permissions["maintain"] {
							role = "maintain"
						}
						displayStr = fmt.Sprintf("%s (%s)", u, role)
					}
					break
				}
			}
			defaultSelection = append(defaultSelection, displayStr)
		}
	}

	// 2. Add remaining current collaborators (who are NOT in config)
	for _, u := range currentCollabs {
		if u.Login != nil {
			login := *u.Login
			if !configUserMap[login] {
				// Check if user is admin
				isAdmin := false
				if u.Permissions != nil && u.Permissions["admin"] {
					isAdmin = true
				}

				// Skip if admin and not in config
				if isAdmin {
					continue
				}

				displayStr := login
				if u.Permissions != nil {
					role := "read"
					if u.Permissions["admin"] {
						role = "admin"
					} else if u.Permissions["push"] {
						role = "write"
					} else if u.Permissions["maintain"] {
						role = "maintain"
					}
					displayStr = fmt.Sprintf("%s (%s)", login, role)
				}
				allOptions = append(allOptions, displayStr)
				defaultSelection = append(defaultSelection, displayStr)
			}
		}
	}

	var selectedUsers []string
	userPrompt := &survey.MultiSelect{
		Message: "Manage users (deselect to remove, select to add):",
		Options: allOptions,
		Default: defaultSelection,
	}
	if err := survey.AskOne(userPrompt, &selectedUsers); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Validation: Check for potential admin removals and warn
	preMap := make(map[string]bool)
	for _, u := range selectedUsers {
		cleanUser := strings.Split(u, " (")[0]
		preMap[cleanUser] = true
	}

	// Just warn about admins, don't finalize removal list yet
	for user := range currentCollabMap {
		if !preMap[user] {
			// Check if this user was hidden (admin not in config)
			isHiddenAdmin := false
			if !configUserMap[user] {
				for _, u := range currentCollabs {
					if u.Login != nil && *u.Login == user {
						if u.Permissions != nil && u.Permissions["admin"] {
							isHiddenAdmin = true
						}
						break
					}
				}
			}

			if !isHiddenAdmin {
				isAdmin := false
				for _, u := range currentCollabs {
					if u.Login != nil && *u.Login == user {
						if u.Permissions != nil && u.Permissions["admin"] {
							isAdmin = true
						}
						break
					}
				}

				if isAdmin {
					fmt.Printf("\033[33m!\033[0m Warning: User '%s' has admin rights (and might be the owner). They will NOT be removed to prevent accidental lockout.\n", user)
				}
			}
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
					// Check if already in selectedUsers to avoid duplicates
					exists := false
					for _, s := range selectedUsers {
						if s == trimmed {
							exists = true
							break
						}
					}
					if !exists {
						selectedUsers = append(selectedUsers, trimmed)
					}
				}
			}
		}
	}

	// Final Diff Calculation
	// Recalculate everything based on final selectedUsers
	selectedMap := make(map[string]bool)
	for _, u := range selectedUsers {
		cleanUser := strings.Split(u, " (")[0]
		selectedMap[cleanUser] = true
	}

	var toAdd []string
	var toRemove []string

	// Calculate toRemove
	for user := range currentCollabMap {
		if !selectedMap[user] {
			// Check if this user was hidden (admin not in config)
			isHiddenAdmin := false
			if !configUserMap[user] {
				for _, u := range currentCollabs {
					if u.Login != nil && *u.Login == user {
						if u.Permissions != nil && u.Permissions["admin"] {
							isHiddenAdmin = true
						}
						break
					}
				}
			}

			if !isHiddenAdmin {
				isAdmin := false
				for _, u := range currentCollabs {
					if u.Login != nil && *u.Login == user {
						if u.Permissions != nil && u.Permissions["admin"] {
							isAdmin = true
						}
						break
					}
				}

				// Admin Check (Safety)
				if isAdmin {
					// Skip removing admin (warning already shown above)
				} else {
					toRemove = append(toRemove, user)
				}
			}
		}
	}

	// Calculate toAdd
	for _, user := range selectedUsers {
		cleanUser := strings.Split(user, " (")[0]
		if !currentCollabMap[cleanUser] {
			toAdd = append(toAdd, cleanUser)
		}
	}

	// 4. Recap
	fmt.Println("\n--- Recap ---")
	fmt.Printf("Repository: %s/%s\n", org, repoName)
	if len(toRemove) > 0 {
		fmt.Printf("Users to REMOVE: %v\n", toRemove)
	} else {
		fmt.Println("Users to REMOVE: []")
	}
	if len(toAdd) > 0 {
		fmt.Printf("Users to ADD: %v\n", toAdd)
	} else {
		fmt.Println("Users to ADD: []")
	}
	fmt.Println("-------------")

	if len(toAdd) == 0 && len(toRemove) == 0 {
		fmt.Println("No changes detected.")
		return
	}

	// 5. Confirm
	var confirm bool
	confirmPrompt := &survey.Confirm{
		Message: "Proceed with changes?",
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
	for _, user := range toRemove {
		fmt.Printf("Removing user %s...\n", user)
		if err := client.RemoveCollaborator(org, repoName, user); err != nil {
			fmt.Printf("Error removing user %s: %v\n", user, err)
		} else {
			fmt.Printf("User %s removed.\n", user)
		}
	}

	for _, user := range toAdd {
		fmt.Printf("Adding user %s...\n", user)
		if err := client.AddCollaborator(org, repoName, user, "push"); err != nil {
			fmt.Printf("Error adding user %s: %v\n", user, err)
		} else {
			fmt.Printf("User %s added.\n", user)
		}
	}

	fmt.Println("Done.")
}
