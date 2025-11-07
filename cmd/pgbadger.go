package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/robinportigliatti/cloud_helper/internal/pgdg/pgbadger"
)

// Déclaration de la commande pgbadger
var pgbadgerCmd = &cobra.Command{
	Use:   "pgbadger",
	Short: "Génère des rapports PGBadger à partir des logs PostgreSQL",
	RunE:  runPgBadger,
}

// Fonction d'exécution de la commande pgbadger
func runPgBadger(cmd *cobra.Command, args []string) error {
	// Récupération des flags
	inputFlag, _ := cmd.Flags().GetString("input")
	outputFlag, _ := cmd.Flags().GetString("output")
	logLinePrefix, _ := cmd.Flags().GetString("log-line-prefix")

	// Vérification que l'input est fourni
	if inputFlag == "" {
		return fmt.Errorf("l'option --input est obligatoire. Spécifiez un fichier ou un répertoire contenant les logs")
	}

	// Vérification de l'existence du fichier ou du dossier
	if _, err := os.Stat(inputFlag); os.IsNotExist(err) {
		return fmt.Errorf("le fichier ou dossier spécifié pour --input n'existe pas : %s", inputFlag)
	}

	var pgb pgbadger.PGBADGER

	// Initialiser avec le répertoire d'entrée
	err := pgb.Init(inputFlag, outputFlag, logLinePrefix)
	if err != nil {
		return fmt.Errorf("init: %w", err)
	}

	err = pgb.Generate()
	if err != nil {
		return fmt.Errorf("generate: %w", err)
	}

	slog.Info("Rapport PGBadger généré",
		slog.String("input", inputFlag),
		slog.String("output", outputFlag),
	)

	return nil
}

func init() {
	// Récupération du répertoire actuel pour le dossier de sortie par défaut
	mydir, err := os.Getwd()
	if err != nil {
		slog.Error("Erreur critique, arrêt du programme", slog.Any("error", err))
		os.Exit(1)
	}

	// Définition des flags pour la commande pgbadger
	pgbadgerCmd.Flags().String("input", "", "Fichiers ou répertoire d'entrée (OBLIGATOIRE)")
	pgbadgerCmd.Flags().String("output", fmt.Sprintf("%s/%s", mydir, "pgbadgers"), "Répertoire de sortie")
	pgbadgerCmd.Flags().String("log-line-prefix", "", "Log line prefix (auto-détecté si non spécifié)")

	// Marquer l'option `--input` comme obligatoire
	err = pgbadgerCmd.MarkFlagRequired("input")
	if err != nil {
		slog.Error("Erreur lors du marquage de input comme required", slog.Any("error", err))
		os.Exit(1)
	}
	// Ajout de la commande au CLI principal
	rootCmd.AddCommand(pgbadgerCmd)
}
