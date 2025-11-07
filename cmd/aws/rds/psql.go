package rds

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/robinportigliatti/cloud_helper/cmd/aws/rds/psql"
	"github.com/robinportigliatti/cloud_helper/internal/aws/rds"
)

func runPsql(cmd *cobra.Command, args []string) error {
	// Récupération des variables globales
	dbInstanceIdentifier := viper.GetString("db-instance-identifier")
	profile := viper.GetString("profile")

	// Initialisation de la connexion à RDS
	var rdsInstance rds.RDS
	err := rdsInstance.Init(dbInstanceIdentifier, profile)
	if err != nil {
		return fmt.Errorf("RDS: Init: %w", err)
	}

	// Implémenter la logique pour la commande psql
	fmt.Println("Connected to PostgreSQL instance:", rdsInstance.GetdbInstance().DBInstanceIdentifier)
	return nil
}

func PsqlCmd() *cobra.Command {
	psqlCmd := &cobra.Command{
		Use:   "psql",
		Short: "Execute subCommands or query",
		Args:  cobra.MinimumNArgs(0),
		RunE:  runPsql,
	}
	psqlCmd.AddCommand(psql.ShowCmd())
	return psqlCmd
}
