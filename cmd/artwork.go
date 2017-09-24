package cmd

import (
	"fmt"
	"github.com/GwentAPI/gwentapi/manipulator/common"
	"github.com/GwentAPI/gwentapi/manipulator/models"
	"github.com/spf13/cobra"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

const MAX_RETRY int = 3

var maxDownloadFlag int
var downloadPath string

// artworkCmd represents the artwork command
var artworkCmd = &cobra.Command{
	Use:   "artwork",
	Short: "Download the artwork of the cards.",
	Long:  `Download the artwork of the cards.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()
		result, err := parseData()
		if err != nil {
			return fmt.Errorf("Error while parsing the data: %s", err)
		}
		dataContainer = result
		downloadQueue := make(chan models.GwentCard)
		wg.Add(1)
		go startDownload(downloadQueue, &wg)
		for k := range dataContainer.Cards {
			downloadQueue <- dataContainer.Cards[k]
		}
		wg.Done()
		close(downloadQueue)
		wg.Wait()
		elapsed := time.Since(start)
		log.Printf("Finished in %s", elapsed)
		return nil
	},
}

func init() {
	RootCmd.AddCommand(artworkCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	artworkCmd.PersistentFlags().IntVar(&maxDownloadFlag, "maxConcurrent", 50, "Limit the number of concurrent artworks to be download.")
	artworkCmd.PersistentFlags().StringVar(&downloadPath, "out", "./artworks/", "Destination folder for the downloaded artworks.")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// artworkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func startDownload(queue <-chan models.GwentCard, wg *sync.WaitGroup) {
	downloadGuard := make(chan struct{}, maxDownloadFlag)
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
			dir, _ := filepath.Abs(downloadPath)
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				os.Mkdir(dir, os.ModeDir)
			}
			path, _ := filepath.Abs(downloadPath + fileName)
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
