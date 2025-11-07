package ovh

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	ovhPkg "github.com/robinportigliatti/cloud_helper/internal/ovh"
)

var serviceName string
var clusterID string
var endpoint string
var list bool

// OvhCmd représente la commande principale "ovh"
var OvhCmd = &cobra.Command{
	Use:   "ovh",
	Short: "Interact with OVHcloud Database PostgreSQL",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Lier les flags à Viper pour pouvoir les récupérer partout
		viper.Set("service-name", serviceName)
		viper.Set("cluster-id", clusterID)
		viper.Set("endpoint", endpoint)
		viper.Set("list", list)
	},
	RunE: RunOvhCmd,
}

func RunOvhCmd(cmd *cobra.Command, args []string) error {
	list := viper.GetBool("list")
	serviceName := viper.GetString("service-name")
	endpoint := viper.GetString("endpoint")

	if list {
		var ovhClient ovhPkg.OVHClient
		err := ovhClient.Init(serviceName, "", endpoint)
		if err != nil {
			return fmt.Errorf("OVH: Init: %w", err)
		}

		databases, err := ovhClient.ListDatabases()
		if err != nil {
			return fmt.Errorf("OVH: ListDatabases: %w", err)
		}

		// Afficher les bases de données
		fmt.Printf("%-40s %-20s %-15s %-20s %-10s\n",
			"Cluster ID", "Name", "Version", "Status", "Plan")
		for _, db := range databases {
			fmt.Printf("%-40s %-20s %-15s %-20s %-10s\n",
				db.ID, db.Name, db.Version, db.Status, db.Plan)
		}
		return nil
	}

	// Récupération des variables globales
	clusterID := viper.GetString("cluster-id")

	// Initialisation de la connexion à OVH
	var ovhClient ovhPkg.OVHClient
	err := ovhClient.Init(serviceName, clusterID, endpoint)
	if err != nil {
		return fmt.Errorf("OVH: Init: %w", err)
	}

	database, err := ovhClient.GetDatabase()
	if err != nil {
		return fmt.Errorf("OVH: GetDatabase: %w", err)
	}

	if database == nil {
		return fmt.Errorf("OVH: No database found")
	}

	// Implémenter la logique pour la commande ovh
	slog.Info("Could find OVH database", slog.String("clusterID", clusterID))
	return nil
}

func init() {
	OvhCmd.PersistentFlags().StringVar(&serviceName, "service-name", "", "OVHcloud Public Cloud Service Name (Project ID)")
	OvhCmd.PersistentFlags().StringVar(&clusterID, "cluster-id", "", "OVHcloud Database Cluster ID")
	OvhCmd.PersistentFlags().StringVar(&endpoint, "endpoint", "ovh-eu", "OVHcloud API endpoint (ovh-eu, ovh-ca, ovh-us)")
	OvhCmd.PersistentFlags().BoolVar(&list, "list", false, "List all PostgreSQL databases")

	if err := viper.BindPFlag("service-name", OvhCmd.PersistentFlags().Lookup("service-name")); err != nil {
		slog.Error("Erreur lors du binding du flag service-name", slog.Any("error", err))
		os.Exit(1)
	}

	if err := viper.BindPFlag("cluster-id", OvhCmd.PersistentFlags().Lookup("cluster-id")); err != nil {
		slog.Error("Erreur lors du binding du flag cluster-id", slog.Any("error", err))
		os.Exit(1)
	}

	if err := viper.BindPFlag("endpoint", OvhCmd.PersistentFlags().Lookup("endpoint")); err != nil {
		slog.Error("Erreur lors du binding du flag endpoint", slog.Any("error", err))
		os.Exit(1)
	}

	OvhCmd.AddCommand(ListCmd())
	OvhCmd.AddCommand(PsqlCmd())
	OvhCmd.AddCommand(DownloadCmd())
}
