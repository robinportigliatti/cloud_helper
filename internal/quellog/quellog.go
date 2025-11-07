package quellog

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Alain-L/quellog/analysis"
	"github.com/Alain-L/quellog/output"
	"github.com/Alain-L/quellog/parser"
)

type Quellog struct {
	input    string
	output   string
	files    []string
	filters  parser.LogFilters
	sections []string

	// Flags pour le mode de sortie
	jsonFormat bool
	mdFormat   bool

	// Flags pour l'analyse SQL
	sqlSummary bool
	sqlDetails []string
}

func (q *Quellog) Init(input string, outputDir string) error {
	q.input = input
	q.output = outputDir

	// Vérifier si l'input est un fichier ou un répertoire
	fileInfo, err := os.Stat(q.input)
	if err != nil {
		return fmt.Errorf("stat: %w", err)
	}

	if fileInfo.IsDir() {
		// Lire les fichiers du répertoire d'entrée
		entries, err := os.ReadDir(q.input)
		if err != nil {
			return fmt.Errorf("ReadDir: %w", err)
		}

		// Collecter tous les fichiers .log
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if filepath.Ext(entry.Name()) == ".log" {
				q.files = append(q.files, filepath.Join(q.input, entry.Name()))
			}
		}
	} else {
		// C'est un fichier unique
		q.files = append(q.files, q.input)
	}

	if len(q.files) == 0 {
		return fmt.Errorf("aucun fichier .log trouvé dans %s", q.input)
	}

	return nil
}

// SetFilters configure les filtres de logs
func (q *Quellog) SetFilters(beginTime, endTime time.Time, dbNames, userNames, excludeUsers, appNames []string) {
	q.filters = parser.LogFilters{
		BeginT:      beginTime,
		EndT:        endTime,
		DbFilter:    dbNames,
		UserFilter:  userNames,
		ExcludeUser: excludeUsers,
		AppFilter:   appNames,
	}
}

// SetSections configure les sections à afficher
func (q *Quellog) SetSections(sections []string) {
	if len(sections) == 0 {
		q.sections = []string{"all"}
	} else {
		q.sections = sections
	}
}

// SetOutputFormat configure le format de sortie
func (q *Quellog) SetOutputFormat(jsonFormat, mdFormat bool) {
	q.jsonFormat = jsonFormat
	q.mdFormat = mdFormat
}

// SetSQLOptions configure les options d'analyse SQL
func (q *Quellog) SetSQLOptions(sqlSummary bool, sqlDetails []string) {
	q.sqlSummary = sqlSummary
	q.sqlDetails = sqlDetails
}

// Generate exécute l'analyse et génère le rapport
func (q *Quellog) Generate() error {
	startTime := time.Now()

	// Calculer la taille totale des fichiers
	totalFileSize := q.calculateTotalFileSize()

	// Configurer le pipeline de parsing
	rawLogs := make(chan parser.LogEntry, 65536)
	filteredLogs := make(chan parser.LogEntry, 65536)

	// Lancer le parsing en parallèle
	go q.parseFilesAsync(rawLogs)

	// Appliquer les filtres
	go parser.FilterStream(rawLogs, filteredLogs, q.filters)

	// Traiter et afficher les résultats
	return q.processAndOutput(filteredLogs, startTime, totalFileSize)
}

// parseFilesAsync lit les fichiers en parallèle
func (q *Quellog) parseFilesAsync(out chan<- parser.LogEntry) {
	defer close(out)

	if len(q.files) == 1 {
		// Un seul fichier : pas besoin de pool de workers
		if err := parser.ParseFile(q.files[0], out); err != nil {
			log.Printf("[ERROR] Failed to parse file %s: %v", q.files[0], err)
		}
		return
	}

	// Plusieurs fichiers : utiliser un pool de workers
	fileChan := make(chan string, len(q.files))
	for _, file := range q.files {
		fileChan <- file
	}
	close(fileChan)

	// Déterminer le nombre de workers optimal
	numWorkers := len(q.files)
	if numWorkers > 4 {
		numWorkers = 4
	}

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for file := range fileChan {
				if err := parser.ParseFile(file, out); err != nil {
					log.Printf("[ERROR] Failed to parse file %s: %v", file, err)
				}
			}
		}()
	}
	wg.Wait()
}

// processAndOutput analyse les logs filtrés et génère la sortie
func (q *Quellog) processAndOutput(filteredLogs <-chan parser.LogEntry, startTime time.Time, totalFileSize int64) error {
	// Cas spécial : détails d'une requête SQL ou résumé SQL
	if len(q.sqlDetails) > 0 || q.sqlSummary {
		analyzer := analysis.NewSQLAnalyzer()
		for entry := range filteredLogs {
			analyzer.Process(&entry)
		}
		sqlMetrics := analyzer.Finalize()

		processingDuration := time.Since(startTime)
		q.printProcessingSummary(sqlMetrics.TotalQueries, processingDuration, totalFileSize)

		if len(q.sqlDetails) > 0 {
			output.PrintSqlDetails(sqlMetrics, q.sqlDetails)
		} else {
			output.PrintSQLSummary(sqlMetrics, false)
		}
		return nil
	}

	// Analyse complète
	metrics := analysis.AggregateMetrics(filteredLogs)
	processingDuration := time.Since(startTime)

	// Afficher selon le format demandé
	if q.jsonFormat {
		output.ExportJSON(metrics, q.sections)
		return nil
	}

	if q.mdFormat {
		output.ExportMarkdown(metrics, q.sections)
		return nil
	}

	// Format texte par défaut
	q.printProcessingSummary(metrics.Global.Count, processingDuration, totalFileSize)
	output.PrintMetrics(metrics, q.sections)

	return nil
}

// calculateTotalFileSize calcule la taille totale des fichiers
func (q *Quellog) calculateTotalFileSize() int64 {
	var total int64
	for _, file := range q.files {
		if fi, err := os.Stat(file); err == nil {
			total += fi.Size()
		}
	}
	return total
}

// printProcessingSummary affiche le résumé du traitement
func (q *Quellog) printProcessingSummary(numEntries int, duration time.Duration, fileSize int64) {
	fmt.Printf("quellog – %d entries processed in %.2f s (%s)\n",
		numEntries, duration.Seconds(), formatBytes(fileSize))
}

// formatBytes convertit un nombre d'octets en format lisible
func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}

	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(b)/float64(div), "kMGTPE"[exp])
}
