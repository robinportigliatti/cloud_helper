package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/robinportigliatti/cloud_helper/cmd/aws/rds"
	"github.com/robinportigliatti/cloud_helper/cmd/azure"
	"github.com/robinportigliatti/cloud_helper/cmd/gcp"
	"github.com/robinportigliatti/cloud_helper/cmd/ovh"
)

// Commande racine
var rootCmd = &cobra.Command{
	Use:           "cloud_helper",
	Short:         "RDS Helper CLI",
	Long:          "Une application CLI pour gérer RDS avec plusieurs sous-commandes.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute exécute la commande racine
func Execute() {
	cmd, err := rootCmd.ExecuteC()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cloud_helper: %s: %s\n", cmd.CommandPath(), err.Error())
		os.Exit(1)
	}
}

// Initialisation de la configuration
func init() {
	cobra.OnInitialize(initConfig)

	// Déclaration des flags globaux
	rootCmd.AddCommand(azure.AzureCmd)
	rootCmd.AddCommand(rds.RdsCmd)
	rootCmd.AddCommand(gcp.GcpCmd) // Ensure GCP command is added here
	rootCmd.AddCommand(ovh.OvhCmd)
}

func initConfig() {
	viper.AutomaticEnv() // Lire depuis les variables d'environnement
}
