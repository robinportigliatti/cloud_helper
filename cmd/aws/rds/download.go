package rds

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/robinportigliatti/cloud_helper/internal/aws/rds"
)

// Fonction d'exécution de la commande download
func runDownload(cmd *cobra.Command, args []string) error {
	// Récupération des flags
	typeFlag, _ := cmd.Flags().GetString("type")
	startFlag, _ := cmd.Flags().GetString("start")
	endFlag, _ := cmd.Flags().GetString("end")
	dirFlag, _ := cmd.Flags().GetString("directory")

	// Récupération des variables globales
	dbInstanceIdentifier := viper.GetString("db-instance-identifier")
	profile := viper.GetString("profile")

	// Initialisation de la connexion à RDS
	var rdsInstance rds.RDS
	err := rdsInstance.Init(dbInstanceIdentifier, profile)
	if err != nil {
		return fmt.Errorf("RDS: Init: %w", err)
	}
	// Exécution en fonction du type de téléchargement
	switch typeFlag {
	case "logs":
		err = rdsInstance.DownloadLogs(startFlag, dirFlag, endFlag)
		if err != nil {
			return fmt.Errorf("RDS: DownloadLogs: %w", err)
		}
	case "metrics":
		err = rdsInstance.DownloadMetrics(startFlag, endFlag, dirFlag)
		if err != nil {
			return fmt.Errorf("RDS: DownloadMetrics: %w", err)
		}
	case "all":
		err = rdsInstance.DownloadLogs(startFlag, dirFlag, endFlag)
		if err != nil {
			return fmt.Errorf("RDS: DownloadLogs: %w", err)
		}

		err = rdsInstance.DownloadMetrics(startFlag, endFlag, dirFlag)
		if err != nil {
			return fmt.Errorf("RDS: DownloadMetrics: %w", err)
		}
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
		Short: "Download RDS logs or metrics",
		Args:  cobra.MinimumNArgs(0),
		RunE:  runDownload,
	}
	downloadCmd.Flags().String("type", "logs", "Type de fichier à télécharger (logs, metrics)")
	downloadCmd.Flags().String("start", time.Now().AddDate(0, 0, -1).Format("2006/01/02 15:04:00"), "Date de début")
	downloadCmd.Flags().String("end", time.Now().Format("2006/01/02 15:04:00"), "Date de fin")
	downloadCmd.Flags().String("directory", "./", "Répertoire de destination")
	return downloadCmd
}
