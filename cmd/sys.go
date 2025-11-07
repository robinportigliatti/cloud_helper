package cmd

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/robinportigliatti/cloud_helper/internal/aws/rds"
)

// Déclaration de la commande sys
var sysCmd = &cobra.Command{
	Use:   "sys",
	Short: "Affiche des informations système",
	RunE:  runSys,
}

// Fonction d'exécution de la commande sys
func runSys(cmd *cobra.Command, args []string) error {
	// Récupération des flags
	freeFlag, _ := cmd.Flags().GetBool("free")

	// Récupération des variables globales
	dbInstanceIdentifier := viper.GetString("db-instance-identifier")
	profile := viper.GetString("profile")

	// Initialisation de la connexion à RDS
	var rdsInstance rds.RDS
	err := rdsInstance.Init(dbInstanceIdentifier, profile)
	if err != nil {
		return fmt.Errorf("RDS: Init: %w", err)
	}
	// Exécution en fonction des options
	if freeFlag {
		output, err := rdsInstance.Free_m()
		if err != nil {
			return fmt.Errorf("RDS: Free_m: %w", err)
		}
		fmt.Println(output)
		slog.Info("Affichage de la mémoire disponible (free -m)")
	}

	return nil
}

func init() {
	// Définition des flags pour la commande sys
	sysCmd.Flags().Bool("free", false, "Afficher la mémoire disponible (free -m)")

	// Ajout de la commande au CLI principal
	rootCmd.AddCommand(sysCmd)
}
