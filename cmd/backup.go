package cmd

import (
	"github.com/spf13/cobra"
)

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup the monboDB databases.",
	Long: `Backup the monboDB databases.
mongodump will be use to backup the databases found on the system.
The content will be gziped and archived.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := backupDb(); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(backupCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// backupCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// backupCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
