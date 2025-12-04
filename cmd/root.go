package cmd

import (
	"fmt"
	"os"

	"meo-repo-manager/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "meo-repo-manager",
	Short: "A CLI to manage GitHub repositories",
	Long:  `A CLI to automate GitHub repository creation within a specific organization, enforcing naming conventions and managing user access.`,
}

var cfgFile string

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")
	rootCmd.PersistentFlags().String("github-token", "", "GitHub token (overrides GITHUB_TOKEN env var)")
	if err := viper.BindPFlag("github-token", rootCmd.PersistentFlags().Lookup("github-token")); err != nil {
		fmt.Printf("Error binding flag: %v\n", err)
		os.Exit(1)
	}
}

func initConfig() {
	if err := config.LoadConfig(cfgFile); err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}
}
