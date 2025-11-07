package azure

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/robinportigliatti/cloud_helper/internal/azure" // Adapter selon ton chemin d'importation
)

// DownloadCmd retourne une commande "download" avec des arguments
func DownloadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "download",
		Short: "Download files from an Azure container within a time range",
		RunE:  RunDownloadCmd,
	}

	// Ajout des flags avec des valeurs par défaut
	cmd.Flags().String("container-name", "", "Azure Blob container name (obligatoire)")
	cmd.Flags().String("begin-time", "", "Start time for filtering files (obligatoire, format: YYYY-MM-DD HH:MM:SS)")
	cmd.Flags().String("end-time", "", "End time for filtering files (obligatoire, format: YYYY-MM-DD HH:MM:SS)")

	// Bind avec Viper pour permettre l'utilisation de fichiers de config
	err := viper.BindPFlag("container-name", cmd.Flags().Lookup("container-name"))
	if err != nil {
		slog.Error("Erreur lors du marquage de container-name comme required", slog.Any("error", err))
		os.Exit(1)
	}

	err = viper.BindPFlag("begin-time", cmd.Flags().Lookup("begin-time"))
	if err != nil {
		slog.Error("Erreur lors du marquage de begin-time comme required", slog.Any("error", err))
		os.Exit(1)
	}

	err = viper.BindPFlag("end-time", cmd.Flags().Lookup("end-time"))
	if err != nil {
		slog.Error("Erreur lors du marquage de end-time comme required", slog.Any("error", err))
		os.Exit(1)
	}
	return cmd
}

func RunDownloadCmd(cmd *cobra.Command, args []string) error {
	// Récupération des arguments
	accountName := viper.GetString("account-name")
	containerName := viper.GetString("container-name")
	beginTimeStr := viper.GetString("begin-time")
	endTimeStr := viper.GetString("end-time")

	// Vérification des paramètres obligatoires
	if accountName == "" || containerName == "" || beginTimeStr == "" || endTimeStr == "" {
		return fmt.Errorf("les paramètres --account-name, --container-name, --begin-time et --end-time sont obligatoires")
	}

	// Conversion des dates
	beginTime, err := time.Parse("2006-01-02T15:04:05", beginTimeStr)
	if err != nil {
		return fmt.Errorf("format de date invalide pour --begin-time (attendu: YYYY-MM-DDTHH:MM:SS)")
	}

	endTime, err := time.Parse("2006-01-02T15:04:05", endTimeStr)
	if err != nil {
		return fmt.Errorf("format de date invalide pour --end-time (attendu: YYYY-MM-DDTHH:MM:SS)")
	}

	// Vérification de la cohérence des dates
	if endTime.Before(beginTime) {
		return fmt.Errorf("end-time doit être postérieur à begin-time")
	}

	// Affichage des paramètres récupérés
	slog.Info("Download parameters",
		"Account Name", accountName,
		"Container Name", containerName,
		"Begin Time", beginTimeStr,
		"End Time", endTimeStr,
	)

	// Initialisation de l'objet Azure
	azureClient, err := azure.NewAzure(accountName, containerName)
	if err != nil {
		return fmt.Errorf("erreur d'initialisation d'Azure: %w", err)
	}

	// Exécuter le processus de téléchargement et traitement
	slog.Info("Démarrage du téléchargement et traitement des fichiers...")
	err = azureClient.DownloadFiles(beginTimeStr, endTimeStr)
	if err != nil {
		return fmt.Errorf("échec du téléchargement et traitement des fichiers: %w", err)
	}

	slog.Info("Téléchargement et traitement terminés avec succès!")
	return nil
}
