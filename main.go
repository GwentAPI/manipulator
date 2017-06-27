package main

import (
	"encoding/json"
	"log"
	"os"
)

func main() {
	var definition map[string]GwentCard
	fullReadPrint(&definition)
	work(definition)
}

func fullReadPrint(definition *map[string]GwentCard) {
	file, err := os.Open("1498322680.240085.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	dec := json.NewDecoder(file)

	if err := dec.Decode(definition); err != nil {
		log.Println(err)
		return
	}
	/*
		for k, v := range definition {
			fmt.Println("ID: ", k)
			fmt.Println(v)
			break
		}
	*/

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

	for k := range definition {
		if !definition[k].Released {
			log.Println(k)
		}
	}

	for k := range categories {
		log.Println(k)
	}

	for k := range factions {
		log.Println(k)
	}

}

func deleteUnreleased(input map[string]GwentCard, key string) bool {
	if !input[key].Released {
		delete(input, key)
		return true
	}
	for variationKey, variationValue := range input[key].Variations {
		if variationValue.Availability != "BaseSet" || !variationValue.Collectible {
			if variationValue.Availability != "NonOwnable" && variationValue.Availability != "Tutorial" {
				log.Println("Variation not from BaseSet reported.")
				log.Println("Card: ", input[key].Name["en-US"])
			}
			delete(input[key].Variations, variationKey)
		}
	}

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

func renameScoiatael(input map[string]GwentCard, k string) {
	if input[k].Faction == "Scoiatael" {
		tmp := input[k]
		tmp.Faction = "Scoia'tael"
		input[k] = tmp
	}
}
