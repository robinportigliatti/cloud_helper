package psql

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/robinportigliatti/cloud_helper/internal/aws/rds"
)

var outputFormat string
var force bool

func runShow(cmd *cobra.Command, args []string) error {
	// Récupération des variables globales
	dbInstanceIdentifier := viper.GetString("db-instance-identifier")
	profile := viper.GetString("profile")

	// Initialisation de la connexion à RDS
	var rdsInstance rds.RDS
	err := rdsInstance.Init(dbInstanceIdentifier, profile)
	if err != nil {
		return fmt.Errorf("RDS: Init: %w", err)
	}

	if len(args) == 0 {
		if !force {
			fmt.Print("Voulez-vous récupérer tous les paramètres ? (y/n): ")
			var response string
			_, err = fmt.Scanln(&response)
			if err != nil {
				return fmt.Errorf("scanln: %w", err)
			}

			if strings.ToLower(response) != "y" {
				return nil
			}
		}
		// Récupérer tous les paramètres
		args, err = rdsInstance.GetAllParameterNames()
		if err != nil {
			return fmt.Errorf("RDS: GetAllParameterNames: %w", err)
		}
	}

	parameters := make([]map[string]string, 0)
	for _, setting := range args {
		value, err := rdsInstance.GetParameterValueByParameterName(setting)
		if err != nil {
			return fmt.Errorf("RDS: GetParameterValueByParameterName: %w", err)
		}
		parameters = append(parameters, map[string]string{"Name": setting, "Value": value})
	}

	switch outputFormat {
	case "csv":
		w := csv.NewWriter(os.Stdout)
		for _, param := range parameters {
			if err := w.Write([]string{param["Name"], param["Value"]}); err != nil {
				return fmt.Errorf("CSV Write: %w", err)
			}
		}
		w.Flush()
	case "json":
		data, err := json.MarshalIndent(parameters, "", "  ")
		if err != nil {
			return fmt.Errorf("JSON Marshal: %w", err)
		}
		fmt.Println(string(data))
	default:
		if len(parameters) == 1 {
			for _, param := range parameters {
				fmt.Println(param["Value"])
			}
		} else {
			for _, param := range parameters {
				fmt.Printf("%s = %s\n", param["Name"], param["Value"])
			}
		}
	}
	return nil
}

func ShowCmd() *cobra.Command {
	showCmd := &cobra.Command{
		Use:   "show [settings...]",
		Short: "Show PostgreSQL settings",
		Args:  cobra.MinimumNArgs(0),
		RunE:  runShow,
	}
	showCmd.Flags().StringVarP(&outputFormat, "format", "F", "default", "Format de sortie (default, csv, json)")
	showCmd.Flags().BoolVarP(&force, "force", "f", false, "Ne pas demander confirmation pour récupérer tous les paramètres")
	return showCmd
}
