package azure

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	postgresflex "github.com/robinportigliatti/cloud_helper/internal/azure/postgresflex"
)

// Déclaration de la commande list
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all PostgreSQL Flexible Servers in the resource group",
	RunE:  runListServers,
}

// Fonction d'exécution de la commande list
func runListServers(cmd *cobra.Command, args []string) error {
	var pf postgresflex.PostgresFlex
	resourceGroup := viper.GetString("resource-group")
	subscription := viper.GetString("subscription")

	err := pf.Init("", resourceGroup, subscription)
	if err != nil {
		return fmt.Errorf("PostgresFlex: Init: %w", err)
	}

	servers, err := pf.ListServers()
	if err != nil {
		return fmt.Errorf("PostgresFlex: ListServers: %w", err)
	}

	// Format the output in a table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
	_, _ = fmt.Fprintln(w, "ServerName |\t Location |\t Version |\t State |\t Tier |\t FQDN |")

	// Remplissage du tableau avec les données
	for _, server := range servers {
		_, _ = fmt.Fprintf(w, "%s |\t %s |\t %s |\t %s |\t %s |\t %s |\n",
			server.Name,
			server.Location,
			server.Properties.Version,
			server.Properties.State,
			server.Sku.Tier,
			server.Properties.FullyQualifiedDomainName)
	}

	_ = w.Flush()
	return nil
}

func ListCmd() *cobra.Command {
	return listCmd
}
