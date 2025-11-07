package ovh

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	ovhPkg "github.com/robinportigliatti/cloud_helper/internal/ovh"
)

// Déclaration de la commande list
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all PostgreSQL databases in the specified OVHcloud project",
	RunE:  runListDatabases,
}

// Fonction d'exécution de la commande list
func runListDatabases(cmd *cobra.Command, args []string) error {
	var ovhClient ovhPkg.OVHClient
	serviceName := viper.GetString("service-name")
	endpoint := viper.GetString("endpoint")

	err := ovhClient.Init(serviceName, "", endpoint)
	if err != nil {
		return fmt.Errorf("OVH: Init: %w", err)
	}

	databases, err := ovhClient.ListDatabases()
	if err != nil {
		return fmt.Errorf("OVH: ListDatabases: %w", err)
	}

	// Format the output in a table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
	_, _ = fmt.Fprintln(w, "Cluster ID |\t Name |\t Engine |\t Version |\t Status |\t Nodes |\t Plan |")

	// Remplissage du tableau avec les données
	for _, db := range databases {
		endpoint := "N/A"
		port := 0
		if len(db.Endpoints) > 0 {
			endpoint = db.Endpoints[0].Domain
			port = db.Endpoints[0].Port
		}

		_, _ = fmt.Fprintf(w, "%s |\t %s |\t %s |\t %s |\t %s |\t %d |\t %s |\t %s:%d\n",
			db.ID,
			db.Name,
			db.Engine,
			db.Version,
			db.Status,
			db.NodeNumber,
			db.Plan,
			endpoint,
			port)
	}

	_ = w.Flush()
	return nil
}

func ListCmd() *cobra.Command {
	return listCmd
}
