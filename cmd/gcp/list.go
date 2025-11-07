package gcp

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	gcpPkg "github.com/robinportigliatti/cloud_helper/internal/gcp"
)

// Déclaration de la commande list-instances
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all instances in the specified project",
	RunE:  runListInstances,
}

// Fonction d'exécution de la commande list-instances
func runListInstances(cmd *cobra.Command, args []string) error {
	var gcp gcpPkg.GCP
	projectID := viper.GetString("project-id")
	credentialsFile := viper.GetString("credentials-file")

	err := gcp.Init("", projectID, credentialsFile)
	if err != nil {
		return fmt.Errorf("GCP: Init: %w", err)
	}

	instances, err := gcp.ListInstances()
	if err != nil {
		return fmt.Errorf("GCP: ListInstances: %w", err)
	}

	// Format the output in a table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
	_, _ = fmt.Fprintln(w, "ConnectionName |\t IP |\t Port |\t DatabaseVersion |\t Status |\t Region |")

	// Remplissage du tableau avec les données
	for _, instance := range instances {
		connectionName := fmt.Sprintf("%s:%s:%s", projectID, instance.Region, instance.Name)
		_, _ = fmt.Fprintf(w, "%s |\t %s |\t %s |\t %s |\t %s |\t %s |\n",
			connectionName, instance.IP, instance.Port, instance.Version, instance.Status, instance.Region)
	}

	_ = w.Flush()
	return nil
}

func ListCmd() *cobra.Command {
	return listCmd
}
