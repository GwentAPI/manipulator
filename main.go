package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/GwentAPI/gwentapi/manipulator/common"
	db "github.com/GwentAPI/gwentapi/manipulator/database"
	"github.com/GwentAPI/gwentapi/manipulator/models"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type stringslice []string

func (i *stringslice) String() string {
	return fmt.Sprintf("%s", *i)
}
func (i *stringslice) Set(value string) error {
	*i = strings.Split(value, ",")
	return nil
}

var repo db.ReposClient
var filePath string = ""
var addrs stringslice
var databaseName string
var useSSL bool
var downloadImage bool
var downloadOnly bool
var timeout time.Duration = 15 * time.Second

const MAX_PARALLEL_DOWNLOAD int = 50

func init() {
	flag.StringVar(&filePath, "input", "", "Path to input file")
	flag.Var(&addrs, "addrs", "List of mongoDB server addresses. Default to localhost on standard mongoDB port")
	flag.BoolVar(&useSSL, "ssl", false, "Set to true if you require SSL to connect to the database")
	flag.StringVar(&databaseName, "db", "", "Set the database to use. 'test' is used if not specified")
	flag.BoolVar(&downloadImage, "download", false, "Set to download the artworks images")
	flag.BoolVar(&downloadOnly, "downloadOnly", false, "Set to only download the artworks image without running db insertion")

	flag.Parse()
}

func main() {
	var wg sync.WaitGroup
	start := time.Now()
	if addrs == nil {
		addrs = []string{"localhost"}
	}
	if filePath == "" {
		fmt.Println("Input file not provided")
		os.Exit(0)
	}

	var definition map[string]models.GwentCard
	readData(&definition)
	work(definition, &wg)

	wg.Wait()
	elapsed := time.Since(start)
	log.Printf("Finished in %s", elapsed)
}

func readData(definition *map[string]models.GwentCard) {
	log.Println("Reading file...")
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
	log.Println("Done")
}

func work(definition map[string]models.GwentCard, wg *sync.WaitGroup) {
	downloadQueue := make(chan models.GwentCard)
	wg.Add(1)
	go startDownload(downloadQueue, wg)

	groups := make(map[string]struct{})
	rarities := make(map[string]struct{})
	factions := make(map[string]struct{})
	categories := make(map[string]struct{})

	for k := range definition {
		if deleteUnreleased(definition, k) {
			continue
		}
		renameScoiatael(definition, k)
		if downloadImage || downloadOnly {
			downloadQueue <- definition[k]
		}
		collectGroup(groups, definition[k])
		collectRarity(rarities, definition[k])
		collectFaction(factions, definition[k])
		collectCategories(categories, definition[k])
	}
	close(downloadQueue)
	if !downloadOnly {
		log.Println("Attempting to establish mongoDB session...")
		session, err := repo.CreateSession(addrs, databaseName, db.Authentication{}, useSSL, timeout)
		defer session.Close()
		if err != nil {
			log.Fatal("Failed to establish mongoDB connection: ", err)
		}
		database := session.DB("")

		log.Println("Upserting a bunch of collections...")
		repo.InsertGenericCollection(database, "groups", groups)
		repo.InsertGenericCollection(database, "rarities", rarities)
		repo.InsertGenericCollection(database, "factions", factions)
		repo.InsertGenericCollection(database, "categories", categories)
		log.Println("Upserting cards...")
		repo.InsertCard(database, "cards", definition)
		log.Println("Upserting variations...")
		repo.InsertVariation(database, "variations", definition)
		log.Println("Done")
	}
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

func startDownload(queue <-chan models.GwentCard, wg *sync.WaitGroup) {
	downloadGuard := make(chan struct{}, MAX_PARALLEL_DOWNLOAD)
	for card := range queue {
		baseFileName := common.GetArtUrl(card.Name["en-US"])
		var noVariation int = 0
		for _, variation := range card.Variations {
			wg.Add(2)
			noVariation++
			downloadGuard <- struct{}{}
			downloadGuard <- struct{}{}
			go func(variation models.GwentVariation, noVariation int) {
				download_file(variation.Art.Thumbnail, baseFileName+"-"+strconv.Itoa(noVariation)+"-thumbnail.png", wg)
				download_file(variation.Art.Medium, baseFileName+"-"+strconv.Itoa(noVariation)+"-medium.png", wg)
				<-downloadGuard
				<-downloadGuard
			}(variation, noVariation)
		}
	}
	wg.Done()
}

func download_file(url string, fileName string, wg *sync.WaitGroup) {
	var retry int = 0
	const MAX_RETRY int = 3

	for retry < MAX_RETRY {
		if retry != 0 {
			log.Println("Retrying ", fileName)
		}
		response, e := http.Get(url)

		if e != nil {
			log.Println("Error downloading file: ", fileName, e)
			retry++
			if retry == MAX_RETRY {
				log.Println("Skipping: ", fileName, " Reason: failed too many time.")
			}
		} else {
			dir, _ := filepath.Abs("./artworks/")
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				os.Mkdir(dir, os.ModeDir)
			}
			path, _ := filepath.Abs("./artworks/" + fileName)
			file, err := os.Create(path)
			if err != nil {
				log.Println("Error creating file: ", path, err)
			} else {
				_, err = io.Copy(file, response.Body)
				if err != nil {
					log.Println("Error creating file: ", err)
				}
			}
			err = file.Close()
			if err != nil {
				log.Fatal(err)
			}
			response.Body.Close()
			break
		}
	}
	wg.Done()
}
