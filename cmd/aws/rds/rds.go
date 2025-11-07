package rds

import (
	"fmt"
	"log/slog"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	rdsPkg "github.com/robinportigliatti/cloud_helper/internal/aws/rds"
)

// Variables globales pour les flags
var profile string
var dbInstanceIdentifier string
var list bool

var RdsCmd = &cobra.Command{
	Use:   "rds",
	Short: "Interact with PostgreSQL",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Lier les flags à Viper pour pouvoir les récupérer partout
		viper.Set("profile", profile)
		viper.Set("db-instance-identifier", dbInstanceIdentifier)
		viper.Set("list", list)
	},
	RunE: RunRdsCmd,
}

func RunRdsCmd(cmd *cobra.Command, args []string) error {
	list := viper.GetBool("list")

	if list {
		var rds rdsPkg.RDS
		profile := viper.GetString("profile")
		err := rds.Init("", profile)
		if err != nil {
			return fmt.Errorf("RDS: Init: %w", err)
		}

		dbInstances, err := rds.GetDbInstances()
		if err != nil {
			return fmt.Errorf("RDS: DescribeDbInstances: %w", err)
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
		_, _ = fmt.Fprintln(w, "DBInstanceIdentifier |\tDBInstanceStatus |\t Endpoint |\t Port")

		// Remplissage du tableau avec les données
		for _, dbInstance := range dbInstances {
			_, _ = fmt.Fprintf(w, "%s |\t %s |\t %s |\t %d\n", dbInstance.DBInstanceIdentifier, dbInstance.DBInstanceStatus, dbInstance.Endpoint.Address, dbInstance.Endpoint.Port)
		}

		_ = w.Flush()
		return nil
	}

	// Récupération des variables globales
	dbInstanceIdentifier := viper.GetString("db-instance-identifier")
	profile := viper.GetString("profile")
	// Initialisation de la connexion à RDS
	var rdsInstance rdsPkg.RDS
	err := rdsInstance.Init(dbInstanceIdentifier, profile)
	var instances []rdsPkg.DBInstance
	if err != nil {
		return fmt.Errorf("RDS: Init: %w", err)
	}

	instances, err = rdsInstance.GetDbInstances()
	if err != nil {
		return fmt.Errorf("RDS: GetDbInstances: %w", err)
	}

	if len(instances) == 0 {
		return fmt.Errorf("RDS: No instance found")
	}

	// Implémenter la logique pour la commande psql
	slog.Info("Could find rds informations")
	return nil
}

func init() {
	RdsCmd.PersistentFlags().StringVar(&profile, "profile", "default", "Profil AWS à utiliser")
	RdsCmd.PersistentFlags().StringVar(&dbInstanceIdentifier, "db-instance-identifier", "", "Identifiant de l'instance RDS")
	RdsCmd.PersistentFlags().BoolVar(&list, "list", false, "Identifiant de l'instance RDS")
	err := viper.BindPFlag("profile", RdsCmd.PersistentFlags().Lookup("profile"))
	if err != nil {
		slog.Error("Erreur lors du binding du flag profile", slog.Any("error", err))
		os.Exit(1)
	}

	err = viper.BindPFlag("db-instance-identifier", RdsCmd.PersistentFlags().Lookup("db-instance-identifier"))
	if err != nil {
		slog.Error("Erreur lors du binding du flag db-instance-identifier", slog.Any("error", err))
		os.Exit(1)
	}

	RdsCmd.AddCommand(PsqlCmd())
	RdsCmd.AddCommand(DownloadCmd())
}
