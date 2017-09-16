package cmd

import (
	"fmt"
	db "github.com/GwentAPI/gwentapi/manipulator/database"
	"github.com/spf13/cobra"
	"log"
	"time"
)

var addrs []string
var databaseName string
var useSSL bool
var repo db.ReposClient
var timeout time.Duration = 15 * time.Second

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the GwentAPI mongoDB database.",
	Long: `Update the GwentAPI mongoDB database.

It will override all data already present and will
ensure that indexes are valid.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("update called")
		err := RootCmd.RunE(cmd, args)
		if err != nil {
			return err
		}
		updateDb(dataContainer)
		return nil
	},
}

func init() {
	RootCmd.AddCommand(updateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	updateCmd.PersistentFlags().StringSliceVar(&addrs, "addrs", []string{"localhost"}, "List of mongoDB server addresses on standard mongoDB port.")
	updateCmd.PersistentFlags().BoolVar(&useSSL, "ssl", false, "Set to true if you require SSL to connect to the database")
	updateCmd.PersistentFlags().StringVar(&databaseName, "db", "", "Use default mongoDb database if not specified (test).")
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// updateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func updateDb(container *DataContainer) error {
	log.Println("Attempting to establish mongoDB session...")
	session, err := repo.CreateSession(addrs, databaseName, db.Authentication{}, useSSL, timeout)
	defer session.Close()
	if err != nil {
		log.Fatal("Failed to establish mongoDB connection: ", err)
	}
	database := session.DB("")
	log.Println("Upserting a bunch of collections...")
	repo.InsertGenericCollection(database, "groups", container.Groups)
	repo.InsertGenericCollection(database, "rarities", container.Rarities)
	repo.InsertGenericCollection(database, "factions", container.Factions)
	repo.InsertGenericCollection(database, "categories", container.Categories)
	log.Println("Upserting cards...")
	repo.InsertCard(database, "cards", container.Cards)
	log.Println("Upserting variations...")
	repo.InsertVariation(database, "variations", container.Cards)
	log.Println("Done")
	return nil
}
