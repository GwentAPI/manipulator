package cmd

import (
	db "github.com/GwentAPI/manipulator/database"
	"github.com/spf13/cobra"
	"strings"
	"time"
)

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup the monboDB databases.",
	Long: `Backup the monboDB databases.
mongodump will be use to backup the databases found on the system.
The content will be gziped and archived.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(mongoDBAuthentication.Username) > 0 {
			flattenedHost := strings.Join(mongoDBAuthentication.Host[:], ",")
			if err := backupWithAuthentication(flattenedHost, mongoDBAuthentication.Username, mongoDBAuthentication.Password, mongoDBAuthentication.AuthenticationDatabase, mongoDBAuthentication.UseSSL); err != nil {
				return err
			}
			return nil
		}
		if err := backupDb(); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(backupCmd)
	mongoDBAuthentication = db.MongoConnectionSettings{
		Timeout: 15 * time.Second,
	}
	backupCmd.Flags().StringVar(&mongoDBAuthentication.Username, "u", "", "Username used for mongoDB authentication.")
	backupCmd.Flags().StringVar(&mongoDBAuthentication.AuthenticationDatabase, "authenticationDatabase", "", "Authentication database for mongoDB.")
	backupCmd.Flags().StringVar(&mongoDBAuthentication.Password, "p", "", "User password for mongoDB authentication.")
	backupCmd.Flags().StringSliceVar(&mongoDBAuthentication.Host, "host", []string{"localhost"}, "Address to a remote mongos.")
	// TODO: Don't use global var
	backupCmd.Flags().BoolVar(&mongoDBAuthentication.UseSSL, "ssl", false, "Set to true if you require SSL to connect to the database")
}
