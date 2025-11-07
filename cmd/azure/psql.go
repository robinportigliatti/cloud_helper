package azure

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	postgresflex "github.com/robinportigliatti/cloud_helper/internal/azure/postgresflex"
)

var show bool

// Déclaration de la commande psql
var psqlCmd = &cobra.Command{
	Use:   "psql",
	Short: "Generate psql connection command for Azure PostgreSQL Flexible Server",
	RunE:  runPsql,
}

// Fonction d'exécution de la commande psql
func runPsql(cmd *cobra.Command, args []string) error {
	resourceGroup := viper.GetString("resource-group")
	serverName := viper.GetString("server-name")
	subscription := viper.GetString("subscription")
	show := viper.GetBool("show")

	// Initialisation de la connexion à Azure PostgreSQL
	var pf postgresflex.PostgresFlex
	err := pf.Init(serverName, resourceGroup, subscription)
	if err != nil {
		return fmt.Errorf("PostgresFlex: Init: %w", err)
	}

	if show {
		// Générer la commande psql
		psqlCmd, err := pf.GenPsql()
		if err != nil {
			return fmt.Errorf("PostgresFlex: GenPsql: %w", err)
		}
		fmt.Print(psqlCmd)
		return nil
	}

	// Par défaut, afficher les informations de connexion
	server := pf.GetServer()
	if server.Name == "" {
		return fmt.Errorf("no server found")
	}

	fmt.Println("Azure PostgreSQL Flexible Server Information:")
	fmt.Printf("  Name: %s\n", server.Name)
	fmt.Printf("  Location: %s\n", server.Location)
	fmt.Printf("  Version: %s\n", server.Properties.Version)
	fmt.Printf("  State: %s\n", server.Properties.State)
	fmt.Printf("  Tier: %s (%s)\n", server.Sku.Tier, server.Sku.Name)
	fmt.Printf("  FQDN: %s\n", server.Properties.FullyQualifiedDomainName)
	fmt.Printf("  Storage: %d GB\n", server.Properties.Storage.StorageSizeGB)
	fmt.Printf("  Backup Retention: %d days\n", server.Properties.Backup.BackupRetentionDays)

	fmt.Println("\nTo connect using psql:")
	fmt.Printf("  psql -h %s -p 5432 -U %s -d <database>\n",
		server.Properties.FullyQualifiedDomainName,
		server.Properties.AdministratorLogin)

	fmt.Println("\nTo connect using az CLI:")
	fmt.Printf("  az postgres flexible-server connect --name %s --resource-group %s\n",
		server.Name, resourceGroup)

	return nil
}

func PsqlCmd() *cobra.Command {
	psqlCmd.Flags().BoolVar(&show, "show", false, "Show psql connection command")

	err := viper.BindPFlag("show", psqlCmd.Flags().Lookup("show"))
	if err != nil {
		slog.Error("Erreur lors du binding du flag show", slog.Any("error", err))
	}

	return psqlCmd
}
