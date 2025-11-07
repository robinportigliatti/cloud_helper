package ovh

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	ovhPkg "github.com/robinportigliatti/cloud_helper/internal/ovh"
)

// Déclaration de la commande psql
var psqlCmd = &cobra.Command{
	Use:   "psql",
	Short: "Generate psql connection string for an OVHcloud PostgreSQL database",
	RunE:  runPsql,
}

// Fonction d'exécution de la commande psql
func runPsql(cmd *cobra.Command, args []string) error {
	// Récupération des flags
	serviceName := viper.GetString("service-name")
	clusterID := viper.GetString("cluster-id")
	endpoint := viper.GetString("endpoint")
	username := viper.GetString("username")
	database := viper.GetString("database")

	if serviceName == "" {
		return fmt.Errorf("le flag --service-name est obligatoire")
	}

	if clusterID == "" {
		return fmt.Errorf("le flag --cluster-id est obligatoire")
	}

	// Initialisation de la connexion à OVH
	var ovhClient ovhPkg.OVHClient
	err := ovhClient.Init(serviceName, clusterID, endpoint)
	if err != nil {
		return fmt.Errorf("OVH: Init: %w", err)
	}

	// Récupérer les détails de la base de données
	instance, err := ovhClient.GetDatabase()
	if err != nil {
		return fmt.Errorf("OVH: GetDatabase: %w", err)
	}

	// Valeurs par défaut si non spécifiées
	if username == "" {
		username = "avnadmin" // Utilisateur par défaut pour OVHcloud Database
	}
	if database == "" {
		database = "defaultdb" // Base de données par défaut
	}

	// Générer la chaîne de connexion psql
	connectionString, err := ovhClient.GenPsql(instance, username, database)
	if err != nil {
		return fmt.Errorf("OVH: GenPsql: %w", err)
	}

	fmt.Println("Chaîne de connexion psql pour OVHcloud PostgreSQL:")
	fmt.Println(connectionString)

	return nil
}

func PsqlCmd() *cobra.Command {
	psqlCmd.Flags().String("username", "avnadmin", "Nom d'utilisateur PostgreSQL")
	psqlCmd.Flags().String("database", "defaultdb", "Nom de la base de données")

	err := viper.BindPFlag("username", psqlCmd.Flags().Lookup("username"))
	if err != nil {
		slog.Error("Erreur lors du binding du flag username", slog.Any("error", err))
	}

	err = viper.BindPFlag("database", psqlCmd.Flags().Lookup("database"))
	if err != nil {
		slog.Error("Erreur lors du binding du flag database", slog.Any("error", err))
	}

	return psqlCmd
}
