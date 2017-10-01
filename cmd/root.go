package cmd

import (
	"errors"
	"fmt"
	"github.com/GwentAPI/manipulator/models"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"sync"
)

type DataContainer struct {
	Cards      map[string]models.GwentCard
	Groups     map[string]struct{}
	Rarities   map[string]struct{}
	Factions   map[string]struct{}
	Categories map[string]struct{}
}

var wg sync.WaitGroup
var dataContainer *DataContainer
var cfgFile string

var filePath string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "manipulator",
	Short: "Manipulatpr is a CLI tool to help manage GwentAPI",
	Long: `Manipulator is a CLI tool to help manage GwentAPI

This application is a tool to quickly perform maintenance operation on GwentAPI database and application.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if len(filePath) == 0 {
			return errors.New("Input file not provided")
		}
		if _, err := os.Stat(filePath); err != nil {
			return fmt.Errorf("Invalid file path: %s", filePath)
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.test.yaml)")
	RootCmd.PersistentFlags().StringVar(&filePath, "input", "", "json file containing the cards data")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".test" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".test")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
