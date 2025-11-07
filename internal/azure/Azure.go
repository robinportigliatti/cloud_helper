package azure

import (
	"compress/gzip"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Azure structure pour gérer les blobs Azure et les traitements
type Azure struct {
	AccountName   string
	ContainerName string
}

// NewAzure crée une nouvelle instance d'Azure avec les paramètres fournis
func NewAzure(accountName string, containerName string) (*Azure, error) {
	return &Azure{
		AccountName:   accountName,
		ContainerName: containerName,
	}, nil
}

// ListBlobs retourne la liste des blobs filtrés par date
func (a *Azure) ListBlobs(startTime, endTime string) ([]string, error) {
	cmd := exec.Command("sh", "-c",
		fmt.Sprintf(`az storage blob list --account-name=%s --container-name=%s | jq -r '.[] | "\(.name);\(.properties.creationTime);\(.properties.lastModified)"'`,
			a.AccountName, a.ContainerName))

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("échec de la récupération des blobs: %w", err)
	}

	startTimeParsed, err := time.Parse("2006-01-02T15:04:05", startTime)
	if err != nil {
		return nil, fmt.Errorf("format de startTime invalide: %w", err)
	}
	startTimestamp := startTimeParsed.Unix()

	endTimeParsed, err := time.Parse("2006-01-02T15:04:05", endTime)
	if err != nil {
		return nil, fmt.Errorf("format de endTime invalide: %w", err)
	}
	endTimestamp := endTimeParsed.Unix()

	var blobs []string
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		parts := strings.Split(line, ";")
		if len(parts) != 3 {
			continue
		}

		blobName := parts[0]
		lastModified := parts[2]

		lastModifiedTime, err := time.Parse(time.RFC3339, lastModified)
		if err != nil {
			continue
		}
		lastModifiedTimestamp := lastModifiedTime.Unix()

		if lastModifiedTimestamp >= startTimestamp && lastModifiedTimestamp <= endTimestamp {
			blobs = append(blobs, blobName)
		}
	}

	return blobs, nil
}

// DownloadBlob télécharge un blob spécifique dans ./logs/account-name/container-name/
func (a *Azure) DownloadBlob(blobName string, counter int) (string, error) {
	// Définir le chemin du dossier de téléchargement
	baseDir := filepath.Join("logs", a.AccountName, a.ContainerName)
	err := os.MkdirAll(baseDir, os.ModePerm)
	var str []byte
	if err != nil {
		return "", fmt.Errorf("échec de la création du répertoire %s: %w", baseDir, err)
	}

	// Chemin du fichier à télécharger
	filePath := filepath.Join(baseDir, fmt.Sprintf("PT1H.%d.json", counter))

	// Exécute la commande `az storage blob download`
	cmd := exec.Command("az", "storage", "blob", "download",
		"--account-name", a.AccountName,
		"--container-name", a.ContainerName,
		"--name", blobName,
		"--file", filePath,
	)
	str, err = cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("échec du téléchargement du blob %s: %s resulted in %s %w", blobName, cmd, str, err)
	}

	return filePath, nil
}

// ConvertToAnalyzable transforme un fichier JSON téléchargé en fichier analysable
func (a *Azure) ConvertToAnalyzable(inputFile string) (string, error) {
	outputFile := strings.Replace(inputFile, ".json", ".analysable.json", 1)

	cmd := exec.Command("jq", "-r", `
	select(.properties? and .properties.message?)
	| (.properties.errorLevel+":") as $lvl
	| (.properties.message | split($lvl)[0]) as $prefix
	| if $prefix then
		  $prefix + " " + $lvl + (.properties.message | split($lvl)[1])
		else
		  ""
		end,
		if .properties.detail? then
		  $prefix + " DETAIL: " + .properties.detail
		else
		  ""
		end`, inputFile)

	outFile, err := os.Create(outputFile)
	if err != nil {
		return "", fmt.Errorf("erreur lors de la création du fichier analysable %s: %w", outputFile, err)
	}
	defer func() { _ = outFile.Close() }()

	cmd.Stdout = outFile
	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("échec de la conversion de %s: %w", inputFile, err)
	}
	return outputFile, nil
}

// CompressFile compresse un fichier JSON analysable
func (a *Azure) CompressFile(inputFile string) (string, error) {
	outputFile := inputFile + ".gz"

	inFile, err := os.Open(inputFile)
	if err != nil {
		return "", fmt.Errorf("échec de l'ouverture du fichier %s: %w", inputFile, err)
	}
	defer func() { _ = inFile.Close() }()

	outFile, err := os.Create(outputFile)
	if err != nil {
		return "", fmt.Errorf("échec de la création du fichier compressé %s: %w", outputFile, err)
	}
	defer func() { _ = outFile.Close() }()

	gzWriter := gzip.NewWriter(outFile)
	defer func() { _ = gzWriter.Close() }()

	_, err = gzWriter.Write([]byte(inputFile))
	if err != nil {
		return "", fmt.Errorf("échec de la compression: %w", err)
	}

	return outputFile, nil
}

// DownloadFiles télécharge, convertit et compresse les fichiers
func (a *Azure) DownloadFiles(startTime, endTime string) error {
	blobs, err := a.ListBlobs(startTime, endTime)
	if err != nil {
		return err
	}
	i := 0
	for _, blob := range blobs {
		i = i + 1
		downloadedFile, err := a.DownloadBlob(blob, i)
		if err != nil {
			return err
		}

		analyzableFile, err := a.ConvertToAnalyzable(downloadedFile)
		if err != nil {
			return err
		}

		_, err = a.CompressFile(analyzableFile)
		if err != nil {
			return err
		}
	}

	return nil
}
