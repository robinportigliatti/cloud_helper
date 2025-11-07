package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/robinportigliatti/cloud_helper/internal/aws/rds"
)

// Déclaration de la structure DBInstance
type DBInstance struct {
	DBInstance                   rds.DBInstance
	DefaultVCpus                 int
	DBParameters                 rds.DescribeDBParametersResult
	InstanceType                 rds.InstanceType
	ValidDBInstanceModifications rds.ValidDBInstanceModificationsMessage
}

// Déclaration de la commande generate
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Génère des fichiers à partir de la configuration RDS",
	RunE:  runGenerate,
}

// Fonction d'exécution de la commande generate
func runGenerate(cmd *cobra.Command, args []string) error {
	// Récupération des flags
	fileFlag, _ := cmd.Flags().GetString("file")
	templateFlag, _ := cmd.Flags().GetString("template")
	templateName, _ := cmd.Flags().GetString("template-name")

	// Récupération des variables globales
	dbInstanceIdentifier := viper.GetString("db-instance-identifier")
	profile := viper.GetString("profile")
	var err error
	var str string
	// Initialisation de la connexion à RDS
	r := rds.RDS{}
	err = r.Init(dbInstanceIdentifier, profile)
	if err != nil {
		return fmt.Errorf("RDS: Init: %w", err)
	}

	// Exécution en fonction du fichier demandé
	switch fileFlag {
	case "postgresql.conf":
		str, err = r.GenPostgreSQLConf()
		if err != nil {
			return fmt.Errorf("RDS: GenPostgreSQLConf: %w", err)
		}

		fmt.Println(str)
	case "pg_hba.conf":
		err = r.GetDBParameterGroupInformations()
		if err != nil {
			return fmt.Errorf("RDS: GetDBParameterGroupInformations: %w", err)
		}
	case ".pgpass":
		str, err := r.GenPgPass()
		if err != nil {
			return fmt.Errorf("RDS: GenPgPass: %w", err)
		}
		fmt.Println(str)
	case "psql":
		str, err = r.GenPsql()
		if err != nil {
			return fmt.Errorf("RDS: GenPsql: %w", err)
		}

		fmt.Println(str)
	case "audit":
		p := &DBInstance{
			DBInstance:                   r.GetdbInstance(),
			DBParameters:                 r.GetDBParameters(),
			InstanceType:                 r.GetInstanceType(),
			DefaultVCpus:                 r.GetDefaultVCpus(),
			ValidDBInstanceModifications: r.GetValidDBInstanceModifications(),
		}

		funcMap := template.FuncMap{
			"GetParameterValueByParameterName": func(parameters rds.DescribeDBParametersResult, parameterName string) (string, error) {
				return parameters.GetParameterValueByParameterName(parameterName)
			},
			"GetSupportsStorageAutoscalingByStorageType": func(validDBInstanceModificationsMessage rds.ValidDBInstanceModificationsMessage, storageType string) (bool, error) {
				return validDBInstanceModificationsMessage.GetSupportsStorageAutoscalingByStorageType(storageType)
			},
		}

		tmpl := template.Must(template.New("page.html").Funcs(funcMap).ParseFiles(templateFlag))
		err := tmpl.ExecuteTemplate(os.Stdout, templateName, p)
		if err != nil {
			return fmt.Errorf("ExecuteTemplate: %w", err)
		}

	case "all":
		str, err = r.GenPostgreSQLConf()
		if err != nil {
			return fmt.Errorf("RDS: GenPostgreSQLConf: %w", err)
		}

		fmt.Println(str)
		err = r.GetDBParameterGroupInformations()
		if err != nil {
			return fmt.Errorf("RDS: GetDBParameterGroupInformations: %w", err)
		}

		str, err = r.GenPgPass()
		if err != nil {
			return fmt.Errorf("RDS: GenPgPass: %w", err)
		}

		fmt.Println(str)
	}

	slog.Info("Génération terminée",
		slog.String("file", fileFlag),
		slog.String("template", templateFlag),
		slog.String("template-name", templateName),
	)

	return nil
}

func init() {
	// Définition des flags pour la commande generate
	generateCmd.Flags().String("file", "", "Fichier à générer (postgresql.conf, pg_hba.conf, .pgpass, audit, all)")
	generateCmd.Flags().String("template", "", "Template à utiliser pour la génération")
	generateCmd.Flags().String("template-name", "audit.md", "Nom du template à utiliser pour la génération")

	// Ajout de la commande au CLI principal
	rootCmd.AddCommand(generateCmd)
}
