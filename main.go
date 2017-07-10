package main

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

var filePath string = "0.9.8.json"

const DOMAIN string = "46bf3452-28e7-482c-9bbf-df053873b021"

func main() {
	var definition map[string]GwentCard
	readData(&definition)
	work(definition)
}

func readData(definition *map[string]GwentCard) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	dec := json.NewDecoder(file)

	if err := dec.Decode(definition); err != nil {
		log.Println(err)
		return
	}

}

func work(definition map[string]GwentCard) {
	groups := make(map[string]struct{})
	rarities := make(map[string]struct{})
	factions := make(map[string]struct{})
	categories := make(map[string]struct{})

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

	session, err := CreateSession([]string{"localhost"}, "gwentTest", Authentication{}, false, 10*time.Second)
	defer session.Close()
	if err != nil {
		log.Fatal("Failed to establish mongoDB connection: ", err)
	}
	database := session.DB("")

	InsertGenericCollection(database, "groups", groups)
	InsertGenericCollection(database, "rarities", rarities)
	InsertGenericCollection(database, "factions", factions)
	InsertGenericCollection(database, "categories", categories)

	InsertCard(database, "cards", definition)
	InsertVariation(database, "variations", definition)
}

func deleteUnreleased(input map[string]GwentCard, key string) bool {
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

func collectGroup(groups map[string]struct{}, input GwentCard) {
	if _, ok := groups[input.Group]; !ok {
		groups[input.Group] = struct{}{}
	}
}

func collectRarity(rarities map[string]struct{}, input GwentCard) {
	for _, variation := range input.Variations {
		if _, ok := rarities[variation.Rarity]; !ok {
			rarities[variation.Rarity] = struct{}{}
		}
	}
}

func collectFaction(factions map[string]struct{}, input GwentCard) {
	if _, ok := factions[input.Faction]; !ok {
		factions[input.Faction] = struct{}{}
	}
}

func collectCategories(categories map[string]struct{}, input GwentCard) {
	for _, category := range input.Categories {
		if _, ok := categories[category]; !ok {
			categories[category] = struct{}{}
		}
	}
}

// The data file doesn't have Scoia'tael spelled correctly, so we rename it.
func renameScoiatael(input map[string]GwentCard, k string) {
	if input[k].Faction == "Scoiatael" {
		tmp := input[k]
		tmp.Faction = "Scoia'tael"
		input[k] = tmp
	}
}
