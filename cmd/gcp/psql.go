package gcp

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	gcpPkg "github.com/robinportigliatti/cloud_helper/internal/gcp"
)

var show bool

// Déclaration de la commande psql
var psqlCmd = &cobra.Command{
	Use:   "psql",
	Short: "Generate psql connection command for Cloud SQL instance",
	RunE:  runPsql,
}

// Fonction d'exécution de la commande psql
func runPsql(cmd *cobra.Command, args []string) error {
	projectID := viper.GetString("project-id")
	instanceName := viper.GetString("instance-name")
	credentialsFile := viper.GetString("credentials-file")
	show := viper.GetBool("show")

	// Initialisation de la connexion à GCP
	var gcp gcpPkg.GCP
	err := gcp.Init(instanceName, projectID, credentialsFile)
	if err != nil {
		return fmt.Errorf("GCP: Init: %w", err)
	}

	if show {
		// Générer la commande psql
		psqlCmd, err := gcp.GenPsql()
		if err != nil {
			return fmt.Errorf("GCP: GenPsql: %w", err)
		}
		fmt.Print(psqlCmd)
		return nil
	}

	// Par défaut, afficher les informations de connexion
	instance := gcp.GetInstance()
	if instance.Name == "" {
		return fmt.Errorf("no instance found")
	}

	connectionName := instance.ConnectionName
	if connectionName == "" {
		connectionName = fmt.Sprintf("%s:%s:%s", projectID, instance.Region, instance.Name)
	}

	fmt.Println("Cloud SQL Instance Information:")
	fmt.Printf("  Name: %s\n", instance.Name)
	fmt.Printf("  Connection Name: %s\n", connectionName)
	fmt.Printf("  Database Version: %s\n", instance.DatabaseVersion)
	fmt.Printf("  Region: %s\n", instance.Region)
	fmt.Printf("  State: %s\n", instance.State)

	if len(instance.IPAddresses) > 0 {
		fmt.Println("  IP Addresses:")
		for _, ip := range instance.IPAddresses {
			fmt.Printf("    - %s (%s)\n", ip.IPAddress, ip.Type)
		}
	}

	fmt.Println("\nTo connect using gcloud:")
	fmt.Printf("  gcloud sql connect %s --project=%s --database=<dbname> --user=<username>\n",
		instance.Name, projectID)

	return nil
}

func PsqlCmd() *cobra.Command {
	psqlCmd.Flags().BoolVar(&show, "show", false, "Show psql connection command")

	err := viper.BindPFlag("show", psqlCmd.Flags().Lookup("show"))
	if err != nil {
		slog.Error("Erreur lors du binding du flag show", slog.Any("error", err))
	}

	return psqlCmd
}
