package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/GwentAPI/gwentapi/manipulator/models"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
	"sync"
	"time"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()
		if addrs == nil {
			addrs = []string{"localhost"}
		}
		result, err := parseData()
		if err != nil {
			return fmt.Errorf("Error while parsing the data: %s", err)
		}
		dataContainer = result
		elapsed := time.Since(start)
		log.Printf("Finished in %s", elapsed)
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

func parseData() (*DataContainer, error) {
	log.Println("Reading file...")
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var definition map[string]models.GwentCard
	groups := make(map[string]struct{})
	rarities := make(map[string]struct{})
	factions := make(map[string]struct{})
	categories := make(map[string]struct{})
	dec := json.NewDecoder(file)

	if err := dec.Decode(&definition); err != nil {
		return nil, err
	}

	for k := range definition {
		if deleteUnreleased(definition, k) {
			continue
		}
		renameScoiatael(definition, k)
		collectGroup(groups, definition[k])
		collectRarity(rarities, definition[k])
		collectFaction(factions, definition[k])
		collectCategories(categories, definition[k])
	}
	container := &DataContainer{
		Cards:      definition,
		Groups:     groups,
		Categories: categories,
		Rarities:   rarities,
		Factions:   factions,
	}

	return container, nil
}

func deleteUnreleased(input map[string]models.GwentCard, key string) bool {
	// Released is an artificial field that was added to the file
	// We do check it but we still verify each individual variation.
	if !input[key].Released {
		delete(input, key)
		return true
	}
	for variationKey, variationValue := range input[key].Variations {
		// Not BaseSet or not collectible, therefore we want to delete it
		// since the API only support BaseSet atm.
		if variationValue.Availability != "BaseSet" || !variationValue.Collectible {
			// Safety to check if a new set is found that wasn't released yet.
			// We don't data-mine and we are only interest in the already released card, however they should
			// Be reported in the case that something changed in the file and it needs to be corrected in the code
			// Or to warn of an incoming new set to prepare for it.
			if variationValue.Availability != "NonOwnable" && variationValue.Availability != "Tutorial" {
				log.Println("Variation not from BaseSet reported.")
				log.Println("Card: ", input[key].Name["en-US"])
			}
			delete(input[key].Variations, variationKey)
		}
	}

	// If a variation was deleted and that it resulted in a card having no variation, we delete the card as well
	// since there's no associated art, rarity, etc.
	if len(input[key].Variations) == 0 {
		delete(input, key)
		return true
	}
	return false
}

func collectGroup(groups map[string]struct{}, input models.GwentCard) {
	if _, ok := groups[input.Group]; !ok {
		groups[input.Group] = struct{}{}
	}
}

func collectRarity(rarities map[string]struct{}, input models.GwentCard) {
	for _, variation := range input.Variations {
		if _, ok := rarities[variation.Rarity]; !ok {
			rarities[variation.Rarity] = struct{}{}
		}
	}
}

func collectFaction(factions map[string]struct{}, input models.GwentCard) {
	if _, ok := factions[input.Faction]; !ok {
		factions[input.Faction] = struct{}{}
	}
}

func collectCategories(categories map[string]struct{}, input models.GwentCard) {
	for _, category := range input.Categories {
		if _, ok := categories[category]; !ok {
			categories[category] = struct{}{}
		}
	}
}

// The data file doesn't have Scoia'tael spelled correctly, so we rename it.
func renameScoiatael(input map[string]models.GwentCard, k string) {
	if input[k].Faction == "Scoiatael" {
		tmp := input[k]
		tmp.Faction = "Scoia'tael"
		input[k] = tmp
	}
}
