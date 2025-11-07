package azure

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	postgresflex "github.com/robinportigliatti/cloud_helper/internal/azure/postgresflex"
)

var accountName string
var resourceGroup string
var serverName string
var subscription string
var list bool

// AzureCmd représente la commande principale "azure"
var AzureCmd = &cobra.Command{
	Use:   "azure",
	Short: "Interact with Azure PostgreSQL Flexible Server",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Lier les flags à Viper pour pouvoir les récupérer partout
		viper.Set("account-name", accountName)
		viper.Set("resource-group", resourceGroup)
		viper.Set("server-name", serverName)
		viper.Set("subscription", subscription)
		viper.Set("list", list)
	},
	RunE: RunAzureCmd,
}

func RunAzureCmd(cmd *cobra.Command, args []string) error {
	list := viper.GetBool("list")
	resourceGroup := viper.GetString("resource-group")
	subscription := viper.GetString("subscription")

	if list {
		var pf postgresflex.PostgresFlex
		err := pf.Init("", resourceGroup, subscription)
		if err != nil {
			return fmt.Errorf("PostgresFlex: Init: %w", err)
		}

		servers, err := pf.ListServers()
		if err != nil {
			return fmt.Errorf("PostgresFlex: ListServers: %w", err)
		}

		// Afficher les serveurs
		fmt.Printf("%-40s %-20s %-15s %-20s %-15s\n",
			"Name", "Location", "Version", "State", "Tier")
		for _, server := range servers {
			fmt.Printf("%-40s %-20s %-15s %-20s %-15s\n",
				server.Name, server.Location, server.Properties.Version,
				server.Properties.State, server.Sku.Tier)
		}
		return nil
	}

	// Récupération des variables globales
	serverName := viper.GetString("server-name")

	// Initialisation de la connexion à Azure PostgreSQL
	var pf postgresflex.PostgresFlex
	err := pf.Init(serverName, resourceGroup, subscription)
	if err != nil {
		return fmt.Errorf("PostgresFlex: Init: %w", err)
	}

	servers, err := pf.GetServers()
	if err != nil {
		return fmt.Errorf("PostgresFlex: GetServers: %w", err)
	}

	if len(servers) == 0 {
		return fmt.Errorf("PostgresFlex: No server found")
	}

	// Implémenter la logique pour la commande azure
	slog.Info("Could find Azure PostgreSQL Flexible Server", slog.String("server", serverName))
	return nil
}

func init() {
	AzureCmd.PersistentFlags().StringVar(&accountName, "account-name", "", "Azure Storage Account Name")
	AzureCmd.PersistentFlags().StringVar(&resourceGroup, "resource-group", "", "Azure Resource Group")
	AzureCmd.PersistentFlags().StringVar(&serverName, "server-name", "", "PostgreSQL Flexible Server Name")
	AzureCmd.PersistentFlags().StringVar(&subscription, "subscription", "", "Azure Subscription ID")
	AzureCmd.PersistentFlags().BoolVar(&list, "list", false, "List all PostgreSQL Flexible Servers")

	if err := viper.BindPFlag("account-name", AzureCmd.PersistentFlags().Lookup("account-name")); err != nil {
		slog.Error("Erreur lors du binding du flag account-name", slog.Any("error", err))
		os.Exit(1)
	}

	if err := viper.BindPFlag("resource-group", AzureCmd.PersistentFlags().Lookup("resource-group")); err != nil {
		slog.Error("Erreur lors du binding du flag resource-group", slog.Any("error", err))
		os.Exit(1)
	}

	if err := viper.BindPFlag("server-name", AzureCmd.PersistentFlags().Lookup("server-name")); err != nil {
		slog.Error("Erreur lors du binding du flag server-name", slog.Any("error", err))
		os.Exit(1)
	}

	if err := viper.BindPFlag("subscription", AzureCmd.PersistentFlags().Lookup("subscription")); err != nil {
		slog.Error("Erreur lors du binding du flag subscription", slog.Any("error", err))
		os.Exit(1)
	}

	AzureCmd.AddCommand(DownloadCmd()) // Ajout de la sous-commande "download"
	AzureCmd.AddCommand(ListCmd())
	AzureCmd.AddCommand(PsqlCmd())
}
