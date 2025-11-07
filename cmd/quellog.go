package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/robinportigliatti/cloud_helper/internal/quellog"
)

// Déclaration de la commande quellog
var quellogCmd = &cobra.Command{
	Use:   "quellog",
	Short: "Génère des rapports Quellog à partir des logs PostgreSQL",
	Long: `Quellog est un outil d'analyse de logs PostgreSQL qui génère des rapports
détaillés sur les performances, les connexions, les checkpoints, et plus encore.

Exemples d'utilisation:
  # Analyse complète
  cloud_helper quellog --input ./logs/

  # Résumé uniquement
  cloud_helper quellog --input ./logs/ --summary

  # Analyse des checkpoints et connexions
  cloud_helper quellog --input ./logs/ --checkpoints --connections

  # Filtrer par base de données
  cloud_helper quellog --input ./logs/ --dbname mydb

  # Export en JSON
  cloud_helper quellog --input ./logs/ --json`,
	RunE: runQuellog,
}

// Fonction d'exécution de la commande quellog
func runQuellog(cmd *cobra.Command, args []string) error {
	// Récupération des flags
	inputFlag, _ := cmd.Flags().GetString("input")
	outputFlag, _ := cmd.Flags().GetString("output")

	// Flags booléens pour les sections
	summary, _ := cmd.Flags().GetBool("summary")
	checkpoints, _ := cmd.Flags().GetBool("checkpoints")
	connections, _ := cmd.Flags().GetBool("connections")
	sqlSummaryFlag, _ := cmd.Flags().GetBool("sql-summary")
	sqlPerformance, _ := cmd.Flags().GetBool("sql-performance")
	tempFiles, _ := cmd.Flags().GetBool("tempfiles")
	maintenance, _ := cmd.Flags().GetBool("maintenance")
	eventsFlag, _ := cmd.Flags().GetBool("events")
	clients, _ := cmd.Flags().GetBool("clients")

	// Formats de sortie
	jsonFormat, _ := cmd.Flags().GetBool("json")
	mdFormat, _ := cmd.Flags().GetBool("md")

	// Filtres
	dbname, _ := cmd.Flags().GetStringSlice("dbname")
	dbuser, _ := cmd.Flags().GetStringSlice("dbuser")
	excludeUser, _ := cmd.Flags().GetStringSlice("exclude-user")
	appname, _ := cmd.Flags().GetStringSlice("appname")
	begin, _ := cmd.Flags().GetString("begin")
	end, _ := cmd.Flags().GetString("end")

	// SQL options
	sqlDetails, _ := cmd.Flags().GetStringSlice("sql-detail")

	// Vérification que l'input est fourni
	if inputFlag == "" {
		return fmt.Errorf("l'option --input est obligatoire. Spécifiez un fichier ou un répertoire contenant les logs")
	}

	// Vérification de l'existence du fichier ou du dossier
	if _, err := os.Stat(inputFlag); os.IsNotExist(err) {
		return fmt.Errorf("le fichier ou dossier spécifié pour --input n'existe pas : %s", inputFlag)
	}

	var q quellog.Quellog

	// Initialiser avec le répertoire d'entrée
	err := q.Init(inputFlag, outputFlag)
	if err != nil {
		return fmt.Errorf("init: %w", err)
	}

	// Parser les filtres de temps
	var beginTime, endTime time.Time
	if begin != "" {
		beginTime, err = time.Parse("2006-01-02 15:04:05", begin)
		if err != nil {
			return fmt.Errorf("format de date invalide pour --begin (attendu: YYYY-MM-DD HH:MM:SS): %w", err)
		}
	}
	if end != "" {
		endTime, err = time.Parse("2006-01-02 15:04:05", end)
		if err != nil {
			return fmt.Errorf("format de date invalide pour --end (attendu: YYYY-MM-DD HH:MM:SS): %w", err)
		}
	}

	// Configurer les filtres
	q.SetFilters(beginTime, endTime, dbname, dbuser, excludeUser, appname)

	// Construire la liste des sections
	var sections []string
	if summary {
		sections = append(sections, "summary")
	}
	if checkpoints {
		sections = append(sections, "checkpoints")
	}
	if connections {
		sections = append(sections, "connections")
	}
	if sqlPerformance {
		sections = append(sections, "sql_performance")
	}
	if tempFiles {
		sections = append(sections, "tempfiles")
	}
	if maintenance {
		sections = append(sections, "maintenance")
	}
	if eventsFlag {
		sections = append(sections, "events")
	}
	if clients {
		sections = append(sections, "clients")
	}

	q.SetSections(sections)
	q.SetOutputFormat(jsonFormat, mdFormat)
	q.SetSQLOptions(sqlSummaryFlag, sqlDetails)

	err = q.Generate()
	if err != nil {
		return fmt.Errorf("generate: %w", err)
	}

	slog.Info("Rapport Quellog généré",
		slog.String("input", inputFlag),
		slog.String("output", outputFlag),
	)

	return nil
}

func init() {
	// Définition des flags pour la commande quellog
	quellogCmd.Flags().String("input", "", "Fichiers ou répertoire d'entrée (OBLIGATOIRE)")
	quellogCmd.Flags().String("output", "", "Répertoire de sortie (vide = stdout)")

	// Sections à afficher
	quellogCmd.Flags().Bool("summary", false, "Afficher uniquement le résumé")
	quellogCmd.Flags().Bool("events", false, "Afficher uniquement les événements")
	quellogCmd.Flags().Bool("sql-performance", false, "Afficher les performances SQL")
	quellogCmd.Flags().Bool("tempfiles", false, "Afficher les fichiers temporaires")
	quellogCmd.Flags().Bool("maintenance", false, "Afficher les opérations de maintenance")
	quellogCmd.Flags().Bool("checkpoints", false, "Afficher les checkpoints")
	quellogCmd.Flags().Bool("connections", false, "Afficher les connexions")
	quellogCmd.Flags().Bool("clients", false, "Afficher les clients")

	// Options SQL
	quellogCmd.Flags().Bool("sql-summary", false, "Afficher le résumé des performances SQL")
	quellogCmd.Flags().StringSlice("sql-detail", []string{}, "Afficher les détails d'une requête SQL spécifique")

	// Formats de sortie
	quellogCmd.Flags().Bool("json", false, "Export au format JSON")
	quellogCmd.Flags().Bool("md", false, "Export au format Markdown")

	// Filtres
	quellogCmd.Flags().StringSlice("dbname", []string{}, "Filtrer par nom de base de données (peut être répété)")
	quellogCmd.Flags().StringSlice("dbuser", []string{}, "Filtrer par utilisateur (peut être répété)")
	quellogCmd.Flags().StringSlice("exclude-user", []string{}, "Exclure certains utilisateurs (peut être répété)")
	quellogCmd.Flags().StringSlice("appname", []string{}, "Filtrer par nom d'application (peut être répété)")
	quellogCmd.Flags().String("begin", "", "Date de début (format: 2006-01-02 15:04:05)")
	quellogCmd.Flags().String("end", "", "Date de fin (format: 2006-01-02 15:04:05)")

	// Marquer l'option `--input` comme obligatoire
	err := quellogCmd.MarkFlagRequired("input")
	if err != nil {
		slog.Error("Erreur lors du marquage de input comme required", slog.Any("error", err))
		os.Exit(1)
	}

	// Ajout de la commande au CLI principal
	rootCmd.AddCommand(quellogCmd)
}
