package gcp

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	gcpPkg "github.com/robinportigliatti/cloud_helper/internal/gcp"
)

// Variables globales pour les flags
var projectID string
var instanceName string
var credentialsFile string
var list bool

var GcpCmd = &cobra.Command{
	Use:   "gcp",
	Short: "Interact with Google Cloud Platform",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Lier les flags à Viper pour pouvoir les récupérer partout
		viper.Set("project-id", projectID)
		viper.Set("instance-name", instanceName)
		viper.Set("credentials-file", credentialsFile)
		viper.Set("list", list)
	},
	RunE: RunGcpCmd,
}

func RunGcpCmd(cmd *cobra.Command, args []string) error {
	list := viper.GetBool("list")
	projectID := viper.GetString("project-id")
	credentialsFile := viper.GetString("credentials-file")

	if list {
		var gcp gcpPkg.GCP
		err := gcp.Init("", projectID, credentialsFile)
		if err != nil {
			return fmt.Errorf("GCP: Init: %w", err)
		}

		instances, err := gcp.ListInstances()
		if err != nil {
			return fmt.Errorf("GCP: ListInstances: %w", err)
		}

		// Afficher les instances
		fmt.Printf("%-30s %-20s %-15s %-15s %-10s %-15s\n",
			"Name", "Version", "IP", "Port", "Status", "Region")
		for _, instance := range instances {
			fmt.Printf("%-30s %-20s %-15s %-15s %-10s %-15s\n",
				instance.Name, instance.Version, instance.IP,
				instance.Port, instance.Status, instance.Region)
		}
		return nil
	}

	// Récupération des variables globales
	instanceName := viper.GetString("instance-name")

	// Initialisation de la connexion à GCP
	var gcpInstance gcpPkg.GCP
	err := gcpInstance.Init(instanceName, projectID, credentialsFile)
	if err != nil {
		return fmt.Errorf("GCP: Init: %w", err)
	}

	instances, err := gcpInstance.GetInstances()
	if err != nil {
		return fmt.Errorf("GCP: GetInstances: %w", err)
	}

	if len(instances) == 0 {
		return fmt.Errorf("GCP: No instance found")
	}

	// Implémenter la logique pour la commande gcp
	slog.Info("Could find GCP instance", slog.String("instance", instanceName))
	return nil
}

func init() {
	GcpCmd.PersistentFlags().StringVar(&projectID, "project-id", "", "Google Cloud Project ID")
	GcpCmd.PersistentFlags().StringVar(&instanceName, "instance-name", "", "Cloud SQL Instance Name")
	GcpCmd.PersistentFlags().StringVar(&credentialsFile, "credentials-file", "", "Path to credentials JSON file")
	GcpCmd.PersistentFlags().BoolVar(&list, "list", false, "List all Cloud SQL instances")

	err := viper.BindPFlag("project-id", GcpCmd.PersistentFlags().Lookup("project-id"))
	if err != nil {
		slog.Error("Erreur lors du binding du flag project-id", slog.Any("error", err))
		os.Exit(1)
	}

	err = viper.BindPFlag("instance-name", GcpCmd.PersistentFlags().Lookup("instance-name"))
	if err != nil {
		slog.Error("Erreur lors du binding du flag instance-name", slog.Any("error", err))
		os.Exit(1)
	}

	err = viper.BindPFlag("credentials-file", GcpCmd.PersistentFlags().Lookup("credentials-file"))
	if err != nil {
		slog.Error("Erreur lors du binding du flag credentials-file", slog.Any("error", err))
		os.Exit(1)
	}

	GcpCmd.AddCommand(ListCmd())
	GcpCmd.AddCommand(PsqlCmd())
	GcpCmd.AddCommand(DownloadCmd())
}
