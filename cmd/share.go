package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/GwentAPI/gwentapi/manipulator/models"
	"log"
	"os"
	"os/exec"
	"time"
)

const BACKUP_FOLDER string = "./backup/"

func backupDb() error {
	t := time.Now()
	format := "2006-01-02T15-04-05.000"
	cmd := "mongodump"
	args := []string{"--gzip", "--out", BACKUP_FOLDER + t.Format(format) + "/"}
	if err := exec.Command(cmd, args...).Run(); err != nil {
		return fmt.Errorf("Error while creating backup: %s", err)
	}
	log.Println("Database backup created.")
	return nil
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
