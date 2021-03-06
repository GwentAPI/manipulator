package cmd

import (
	"fmt"
	db "github.com/GwentAPI/manipulator/database"
	"github.com/spf13/cobra"
	"log"
	"strings"
	"time"
)

var repo db.ReposClient

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the GwentAPI mongoDB database.",
	Long: `Update the GwentAPI mongoDB database.

It will override all data already present and will
ensure that indexes are valid.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()
		if mongoDBAuthentication.Host == nil {
			mongoDBAuthentication.Host = []string{"localhost"}
		}
		result, err := parseData()
		if err != nil {
			return fmt.Errorf("Error while parsing the data: %s", err)
		}
		flattenedHost := strings.Join(mongoDBAuthentication.Host[:], ",")
		if len(mongoDBAuthentication.Username) > 0 {
			if err := backupWithAuthentication(flattenedHost, mongoDBAuthentication.Username, mongoDBAuthentication.Password, mongoDBAuthentication.AuthenticationDatabase, mongoDBAuthentication.UseSSL); err != nil {
				return err
			}
		} else {
			if err := backupDb(flattenedHost); err != nil {
				return err
			}
		}
		dataContainer = result
		if err := updateDb(dataContainer); err != nil {
			return err
		}
		elapsed := time.Since(start)
		log.Printf("Finished in %s", elapsed)
		return nil
	},
}

func init() {
	RootCmd.AddCommand(updateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	mongoDBAuthentication = db.MongoConnectionSettings{
		Timeout: 15 * time.Second,
	}

	updateCmd.Flags().StringVar(&mongoDBAuthentication.Username, "u", "", "Username used for mongoDB authentication.")
	updateCmd.Flags().StringVar(&mongoDBAuthentication.AuthenticationDatabase, "authenticationDatabase", "", "Authentication database for mongoDB.")
	updateCmd.Flags().StringVar(&mongoDBAuthentication.Password, "p", "", "User password for mongoDB authentication.")
	updateCmd.PersistentFlags().StringSliceVar(&mongoDBAuthentication.Host, "host", []string{"localhost"}, "List of mongoDB server addresses on standard mongoDB port.")
	updateCmd.PersistentFlags().BoolVar(&mongoDBAuthentication.UseSSL, "ssl", false, "Set to true if you require SSL to connect to the database")
	updateCmd.PersistentFlags().StringVar(&mongoDBAuthentication.Db, "db", "", "Use default mongoDb database if not specified (default test).")
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// updateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func updateDb(container *DataContainer) error {
	log.Println("Attempting to establish mongoDB session...")
	session, err := repo.CreateSession(mongoDBAuthentication)
	if err != nil {
		return fmt.Errorf("Failed to establish mongoDB connection:: %s", err)
	}
	defer session.Close()
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
