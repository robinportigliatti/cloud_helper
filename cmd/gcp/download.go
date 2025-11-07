package gcp

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/robinportigliatti/cloud_helper/internal/gcp"
)

// Fonction d'exécution de la commande download
func runDownload(cmd *cobra.Command, args []string) error {
	// Récupération des flags
	typeFlag, _ := cmd.Flags().GetString("type")
	startFlag, _ := cmd.Flags().GetString("start")
	endFlag, _ := cmd.Flags().GetString("end")
	dirFlag, _ := cmd.Flags().GetString("directory")

	// Récupération des variables globales
	instanceName := viper.GetString("instance-name")
	projectID := viper.GetString("project-id")
	credentialsFile := viper.GetString("credentials-file")

	// Initialisation de la connexion à GCP
	var gcpInstance gcp.GCP
	err := gcpInstance.Init(instanceName, projectID, credentialsFile)
	if err != nil {
		return fmt.Errorf("GCP: Init: %w", err)
	}

	// Exécution en fonction du type de téléchargement
	switch typeFlag {
	case "logs":
		err = gcpInstance.DownloadLogs(startFlag, dirFlag, endFlag)
		if err != nil {
			return fmt.Errorf("GCP: DownloadLogs: %w", err)
		}
	case "metrics":
		// TODO: Implement metrics download for GCP
		slog.Info("Metrics download for GCP is not yet implemented")
		return fmt.Errorf("metrics download for GCP is not yet implemented")
	case "all":
		err = gcpInstance.DownloadLogs(startFlag, dirFlag, endFlag)
		if err != nil {
			return fmt.Errorf("GCP: DownloadLogs: %w", err)
		}
		// TODO: Implement metrics download for GCP
		slog.Info("Metrics download for GCP is not yet implemented (only logs downloaded)")
	}

	slog.Info("Téléchargement terminé",
		slog.String("type", typeFlag),
		slog.String("start", startFlag),
		slog.String("end", endFlag),
		slog.String("directory", dirFlag),
	)

	return nil
}

func DownloadCmd() *cobra.Command {
	downloadCmd := &cobra.Command{
		Use:   "download",
		Short: "Download GCP Cloud SQL logs or metrics",
		Args:  cobra.MinimumNArgs(0),
		RunE:  runDownload,
	}
	downloadCmd.Flags().String("type", "logs", "Type de fichier à télécharger (logs, metrics, all)")
	downloadCmd.Flags().String("start", time.Now().AddDate(0, 0, -1).Format("2006/01/02 15:04:00"), "Date de début")
	downloadCmd.Flags().String("end", time.Now().Format("2006/01/02 15:04:00"), "Date de fin")
	downloadCmd.Flags().String("directory", "./", "Répertoire de destination")
	return downloadCmd
}
