package pgbadger

import (
	"fmt"
	"os"
	"os/exec"
)

type PGBADGER struct {
	input          string
	output         string
	logLinePrefix  string
	files          []os.DirEntry
}

func (pgbadger *PGBADGER) Init(input string, output string, logLinePrefix string) error {
	pgbadger.input = input
	pgbadger.output = output
	pgbadger.logLinePrefix = logLinePrefix

	// Vérifier si l'input est un fichier ou un répertoire
	fileInfo, err := os.Stat(pgbadger.input)
	if err != nil {
		return fmt.Errorf("stat: %w", err)
	}

	if fileInfo.IsDir() {
		// Lire les fichiers du répertoire d'entrée
		files, err := os.ReadDir(pgbadger.input)
		if err != nil {
			return fmt.Errorf("ReadDir: %w", err)
		}
		pgbadger.files = files
	} else {
		// C'est un fichier unique, on laisse files vide pour le traiter différemment
		pgbadger.files = nil
	}

	return nil
}

func (pgbadger PGBADGER) Execute(args string) (string, error) {
	cmdStr := fmt.Sprintf("%s %s", "pgbadger", args)

	// Exécuter la commande
	cmd := exec.Command("bash", "-c", cmdStr)

	// Capturer stdout + stderr ensemble
	output, err := cmd.CombinedOutput()
	result := string(output)

	if err != nil {
		return "", fmt.Errorf("%s: %s %w", cmdStr, result, err)
	}

	return result, nil
}

func (pgbadger *PGBADGER) Generate() error {
	// Vérifier si le répertoire de sortie existe, sinon le créer
	err := os.MkdirAll(pgbadger.output, os.ModePerm)
	if err != nil {
		return fmt.Errorf("report dir: mkdir: %w", err)
	}

	// Collecter tous les fichiers .log
	var logFiles []string

	if pgbadger.files == nil {
		// Cas où input est un fichier unique
		logFiles = append(logFiles, pgbadger.input)
	} else {
		// Cas où input est un répertoire
		for _, file := range pgbadger.files {
			if file.IsDir() {
				continue
			}
			// Accepter tous les fichiers .log
			if len(file.Name()) > 4 && file.Name()[len(file.Name())-4:] == ".log" {
				logFiles = append(logFiles, fmt.Sprintf("%s/%s", pgbadger.input, file.Name()))
			}
		}
	}

	if len(logFiles) == 0 {
		return fmt.Errorf("aucun fichier .log trouvé dans %s", pgbadger.input)
	}

	// Construire la commande pgbadger avec tous les fichiers
	filesArg := ""
	for _, logFile := range logFiles {
		filesArg += " " + logFile
	}

	// Créer le chemin du fichier de sortie
	outputFile := fmt.Sprintf("%s/out.html", pgbadger.output)

	// Mode incrémental avec -I : permet de générer des rapports jour par jour
	// et de les agréger dans un rapport global avec historique
	// Note: on retire le -p car pgbadger devrait auto-détecter le format
	// Si besoin, l'utilisateur peut le spécifier via --log-line-prefix
	cmd := fmt.Sprintf("-I -o %s%s", outputFile, filesArg)
	if pgbadger.logLinePrefix != "" {
		cmd = fmt.Sprintf("-p '%s' -I -o %s%s", pgbadger.logLinePrefix, outputFile, filesArg)
	}

	fmt.Printf("Exécution de pgbadger en mode incrémental sur %d fichier(s)...\n", len(logFiles))
	result, err := pgbadger.Execute(cmd)

	if err != nil {
		return fmt.Errorf("Execute: %w", err)
	}

	fmt.Printf("Résultat:\n%s\n", result)
	fmt.Printf("\nRapport généré: %s\n", outputFile)
	fmt.Printf("Les rapports incrémentaux sont dans: %s/\n", pgbadger.output)

	return nil
}
