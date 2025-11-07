package ovh

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	ovhPkg "github.com/robinportigliatti/cloud_helper/internal/ovh"
)

// Déclaration de la commande download
var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download PostgreSQL logs from OVHcloud Database",
	Long: `Download PostgreSQL logs from OVHcloud Database service.
Note: OVHcloud databases primarily use logs forwarding to Logs Data Platform.
This command provides an interface for downloading logs when available via API.`,
	RunE: runDownload,
}

// Fonction d'exécution de la commande download
func runDownload(cmd *cobra.Command, args []string) error {
	// Récupération des flags
	serviceName := viper.GetString("service-name")
	clusterID := viper.GetString("cluster-id")
	endpoint := viper.GetString("endpoint")
	startFlag := viper.GetString("start-time")
	endFlag := viper.GetString("end-time")
	dirFlag := viper.GetString("directory")

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

	slog.Info("Téléchargement des logs OVH",
		slog.String("serviceName", serviceName),
		slog.String("clusterID", clusterID),
		slog.String("start", startFlag),
		slog.String("end", endFlag),
		slog.String("directory", dirFlag),
	)

	// TODO: Implémenter le téléchargement des logs via l'API OVHcloud
	// OVHcloud utilise principalement un système de logs forwarding vers Logs Data Platform
	// L'endpoint exact pour télécharger les logs directement doit être déterminé
	// Endpoint probable: GET /cloud/project/{serviceName}/database/postgresql/{clusterId}/logs

	slog.Warn("Le téléchargement de logs OVHcloud n'est pas encore complètement implémenté")
	slog.Info("OVHcloud recommande d'utiliser le système de logs forwarding vers Logs Data Platform")
	slog.Info("Consultez la documentation: https://help.ovhcloud.com/csm/en-public-cloud-databases-logs-to-customers")

	return nil
}

func DownloadCmd() *cobra.Command {
	downloadCmd.Flags().String("start-time", time.Now().AddDate(0, 0, -1).Format("2006-01-02T15:04:05"), "Date de début (format: YYYY-MM-DDTHH:MM:SS)")
	downloadCmd.Flags().String("end-time", time.Now().Format("2006-01-02T15:04:05"), "Date de fin (format: YYYY-MM-DDTHH:MM:SS)")
	downloadCmd.Flags().String("directory", "./", "Répertoire de destination")

	err := viper.BindPFlag("start-time", downloadCmd.Flags().Lookup("start-time"))
	if err != nil {
		slog.Error("Erreur lors du binding du flag start-time", slog.Any("error", err))
	}

	err = viper.BindPFlag("end-time", downloadCmd.Flags().Lookup("end-time"))
	if err != nil {
		slog.Error("Erreur lors du binding du flag end-time", slog.Any("error", err))
	}

	err = viper.BindPFlag("directory", downloadCmd.Flags().Lookup("directory"))
	if err != nil {
		slog.Error("Erreur lors du binding du flag directory", slog.Any("error", err))
	}

	return downloadCmd
}
