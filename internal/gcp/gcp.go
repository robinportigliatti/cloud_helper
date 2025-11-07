package gcp

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq" // Driver PostgreSQL pour compatibilité si nécessaire
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	"google.golang.org/api/sqladmin/v1"
)

// GCP structure principale pour gérer Cloud SQL sur GCP
type GCP struct {
	instanceName    string
	projectID       string
	credentialsFile string
	service         *sqladmin.Service
	computeService  *compute.Service
	instance        *DatabaseInstance
	machineType     *MachineType
	databaseFlags   DescribeFlagsResult
}

func (g *GCP) GetInstances() ([]DatabaseInstance, error) {
	if g.instance != nil {
		return []DatabaseInstance{*g.instance}, nil
	}
	return []DatabaseInstance{}, nil
}

func (g *GCP) Init(instanceName string, projectID string, credentialsFile string) error {
	var err error
	g.instanceName = instanceName
	g.projectID = projectID
	g.credentialsFile = credentialsFile

	// Initialiser le service SQL Admin
	ctx := context.Background()
	var opts []option.ClientOption
	if credentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(credentialsFile))
	}
	opts = append(opts, option.WithScopes(sqladmin.SqlserviceAdminScope))

	g.service, err = sqladmin.NewService(ctx, opts...)
	if err != nil {
		return fmt.Errorf("GCP: Init service SQL Admin: %w", err)
	}

	// Initialiser le service Compute
	g.computeService, err = compute.NewService(ctx, opts...)
	if err != nil {
		return fmt.Errorf("GCP: Init service Compute: %w", err)
	}

	// Récupérer les informations de l'instance
	err = g.DescribeInstance()
	if err != nil {
		return fmt.Errorf("GCP: DescribeInstance: %w", err)
	}

	// Récupérer les informations du type de machine
	err = g.GetMachineTypeInformation()
	if err != nil {
		return fmt.Errorf("GCP: GetMachineTypeInformation: %w", err)
	}

	// Récupérer les flags de la base de données
	err = g.GetDatabaseFlagsInformation()
	if err != nil {
		return fmt.Errorf("GCP: GetDatabaseFlagsInformation: %w", err)
	}

	return nil
}

func (g *GCP) GetInstance() DatabaseInstance {
	if g.instance != nil {
		return *g.instance
	}
	return DatabaseInstance{}
}

func (g *GCP) Execute(command string, args string) (string, error) {
	var cmdStr string
	if g.credentialsFile != "" {
		cmdStr = fmt.Sprintf("GOOGLE_APPLICATION_CREDENTIALS=%s gcloud %s %s --project=%s",
			g.credentialsFile, command, args, g.projectID)
	} else {
		cmdStr = fmt.Sprintf("gcloud %s %s --project=%s", command, args, g.projectID)
	}

	cmd := exec.Command("bash", "-c", cmdStr)
	output, err := cmd.CombinedOutput()
	result := string(output)

	if err != nil {
		return "", fmt.Errorf("%s: %s %w", cmdStr, result, err)
	}

	return result, nil
}

func (g *GCP) GetDatabaseFlags() DescribeFlagsResult {
	return g.databaseFlags
}

func (g *GCP) GetMachineType() *MachineType {
	return g.machineType
}

func (g *GCP) DescribeInstance() error {
	if g.instanceName == "" {
		return nil
	}

	ctx := context.Background()
	instanceResp, err := g.service.Instances.Get(g.projectID, g.instanceName).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("GCP: Instances.Get: %w", err)
	}

	// Convertir en notre structure
	g.instance = &DatabaseInstance{
		Name:            instanceResp.Name,
		Project:         instanceResp.Project,
		DatabaseVersion: instanceResp.DatabaseVersion,
		Region:          instanceResp.Region,
		State:           instanceResp.State,
		ConnectionName:  instanceResp.ConnectionName,
		InstanceType:    instanceResp.InstanceType,
		SelfLink:        instanceResp.SelfLink,
		GceZone:         instanceResp.GceZone,
	}

	// Copier les IP addresses
	for _, ip := range instanceResp.IpAddresses {
		g.instance.IPAddresses = append(g.instance.IPAddresses, IPMapping{
			Type:      ip.Type,
			IPAddress: ip.IpAddress,
		})
	}

	// Copier les settings
	if instanceResp.Settings != nil {
		g.instance.Settings = Settings{
			Tier:                   instanceResp.Settings.Tier,
			ActivationPolicy:       instanceResp.Settings.ActivationPolicy,
			DataDiskSizeGb:         instanceResp.Settings.DataDiskSizeGb,
			DataDiskType:           instanceResp.Settings.DataDiskType,
			PricingPlan:            instanceResp.Settings.PricingPlan,
			StorageAutoResize:      instanceResp.Settings.StorageAutoResize != nil && *instanceResp.Settings.StorageAutoResize,
			StorageAutoResizeLimit: instanceResp.Settings.StorageAutoResizeLimit,
			AvailabilityType:       instanceResp.Settings.AvailabilityType,
		}

		// Copier les database flags
		for _, flag := range instanceResp.Settings.DatabaseFlags {
			g.instance.Settings.DatabaseFlags = append(g.instance.Settings.DatabaseFlags, DatabaseFlag{
				Name:  flag.Name,
				Value: flag.Value,
			})
		}

		// Copier la configuration de backup
		if instanceResp.Settings.BackupConfiguration != nil {
			g.instance.Settings.BackupConfiguration = BackupConfiguration{
				Enabled:                    instanceResp.Settings.BackupConfiguration.Enabled,
				StartTime:                  instanceResp.Settings.BackupConfiguration.StartTime,
				PointInTimeRecoveryEnabled: instanceResp.Settings.BackupConfiguration.PointInTimeRecoveryEnabled,
			}
		}
	}

	return nil
}

func (g *GCP) GenPsql() (string, error) {
	var result strings.Builder
	if g.instance != nil && len(g.instance.IPAddresses) > 0 {
		_, err := result.WriteString(fmt.Sprintf("psql -h %s -p 5432\n", g.instance.IPAddresses[0].IPAddress))
		if err != nil {
			return "", fmt.Errorf("WriteString: %w", err)
		}
	}
	return result.String(), nil
}

func (g *GCP) GetTier() string {
	if g.instance != nil {
		return g.instance.Settings.Tier
	}
	return ""
}

func (g *GCP) GetVCpus() int {
	if g.machineType != nil {
		return g.machineType.GuestCpus
	}
	return 0
}

func (g *GCP) GetMemoryMb() int {
	if g.machineType != nil {
		return g.machineType.MemoryMb
	}
	return 0
}

func (g *GCP) GetMachineTypeInformation() error {
	if g.instanceName == "" || g.instance == nil {
		return nil
	}

	tier := g.instance.Settings.Tier
	zone := g.instance.GceZone
	if zone == "" {
		zone = g.instance.Region + "-a" // Default zone
	}

	ctx := context.Background()
	machineTypeResp, err := g.computeService.MachineTypes.Get(g.projectID, zone, tier).Context(ctx).Do()
	if err != nil {
		// Si le tier n'est pas trouvé directement, essayer d'extraire le type de machine du tier
		// Les tiers Cloud SQL ont un format comme "db-custom-2-7680" ou "db-n1-standard-1"
		parts := strings.Split(tier, "-")
		if len(parts) >= 3 {
			machineTypeStr := strings.Join(parts[1:], "-")
			machineTypeResp, err = g.computeService.MachineTypes.Get(g.projectID, zone, machineTypeStr).Context(ctx).Do()
			if err != nil {
				return nil // Not a critical error
			}
		} else {
			return nil
		}
	}

	g.machineType = &MachineType{
		Name:        machineTypeResp.Name,
		GuestCpus:   int(machineTypeResp.GuestCpus),
		MemoryMb:    int(machineTypeResp.MemoryMb),
		Zone:        machineTypeResp.Zone,
		SelfLink:    machineTypeResp.SelfLink,
		IsSharedCpu: machineTypeResp.IsSharedCpu,
		Description: machineTypeResp.Description,
	}

	return nil
}

func (g *GCP) GetDatabaseFlagsInformation() error {
	ctx := context.Background()
	flagsResp, err := g.service.Flags.List().Context(ctx).Do()
	if err != nil {
		return nil // Not critical
	}

	for _, flag := range flagsResp.Items {
		// Filtrer uniquement les flags PostgreSQL
		if contains(flag.AppliesTo, "POSTGRES") {
			g.databaseFlags.Items = append(g.databaseFlags.Items, FlagMetadata{
				Name:            flag.Name,
				Type:            flag.Type,
				AppliesTo:       flag.AppliesTo,
				AllowedValues:   flag.AllowedStringValues,
				RequiresRestart: flag.RequiresRestart,
			})
		}
	}

	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.Contains(s, item) {
			return true
		}
	}
	return false
}

func (g *GCP) Free_m() (string, error) {
	if g.instance == nil {
		return "", fmt.Errorf("no instance initialized")
	}

	// TODO: Implement metrics retrieval via Cloud Monitoring API
	// For now, return placeholder values
	totalMemoryMB := float64(g.GetMemoryMb())

	result := fmt.Sprintf(
		"%15s %12s %12s\n%15.0f %12s %12s\n",
		"total", "utilisé", "disponible",
		totalMemoryMB, "N/A", "N/A",
	)

	return result, nil
}

func (g *GCP) GenPgPass() (string, error) {
	if g.instance == nil || len(g.instance.IPAddresses) == 0 {
		return "", fmt.Errorf("no instance or IP address")
	}

	str := fmt.Sprintf("%s:%s:%s:%s",
		g.instance.IPAddresses[0].IPAddress,
		"5432",
		"postgres",
		"<TODO>")
	return str, nil
}

func (g *GCP) GenPostgreSQLConf() (string, error) {
	var result strings.Builder
	_, err := result.WriteString("FlagName;FlagValue\n")
	if err != nil {
		return "", fmt.Errorf("WriteString: %w", err)
	}

	if g.instance != nil {
		for _, flag := range g.instance.Settings.DatabaseFlags {
			if flag.Value != "" {
				str := fmt.Sprintf("%s;%s\n", flag.Name, flag.Value)
				_, err := result.WriteString(str)
				if err != nil {
					return "", fmt.Errorf("WriteString: %w", err)
				}
			}
		}
	}
	return result.String(), nil
}

func (g *GCP) GetFlagValueByName(flagName string) (string, error) {
	value, err := g.databaseFlags.GetFlagValueByName(flagName, g.instance)
	if err != nil {
		return "", fmt.Errorf("GetFlagValueByName: %w", err)
	}
	return value, nil
}

func (g *GCP) CheckFlag(flagName string, flagValue string) error {
	value, err := g.GetFlagValueByName(flagName)
	if err != nil {
		return fmt.Errorf("GetFlagValueByName: %w", err)
	}

	str := ""
	if value != flagValue {
		str = fmt.Sprintf("%s should be at %s", flagName, flagValue)
	} else {
		str = fmt.Sprintf("%s: OK", flagName)
	}

	fmt.Println(str)
	return nil
}

func (g *GCP) CheckPgbadger() error {
	err := g.CheckFlag("log_connections", "on")
	if err != nil {
		return fmt.Errorf("CheckFlag: %w", err)
	}
	err = g.CheckFlag("log_disconnections", "on")
	if err != nil {
		return fmt.Errorf("CheckFlag: %w", err)
	}
	err = g.CheckFlag("log_lock_waits", "on")
	if err != nil {
		return fmt.Errorf("CheckFlag: %w", err)
	}
	err = g.CheckFlag("log_temp_files", "0")
	if err != nil {
		return fmt.Errorf("CheckFlag: %w", err)
	}
	err = g.CheckFlag("log_autovacuum_min_duration", "0")
	if err != nil {
		return fmt.Errorf("CheckFlag: %w", err)
	}
	err = g.CheckFlag("log_min_duration_statement", "0")
	if err != nil {
		return fmt.Errorf("CheckFlag: %w", err)
	}
	err = g.CheckFlag("log_duration", "on")
	if err != nil {
		return fmt.Errorf("CheckFlag: %w", err)
	}
	err = g.CheckFlag("log_checkpoints", "on")
	if err != nil {
		return fmt.Errorf("CheckFlag: %w", err)
	}
	err = g.CheckFlag("log_statement", "all")
	if err != nil {
		return fmt.Errorf("CheckFlag: %w", err)
	}
	err = g.CheckFlag("track_io_timing", "on")
	if err != nil {
		return fmt.Errorf("CheckFlag: %w", err)
	}

	_, err = g.GetFlagValueByName("track_activity_query_size")
	if err != nil {
		return fmt.Errorf("GetFlagValue: %w", err)
	}

	return nil
}

func (g *GCP) GetAllFlagNames() ([]string, error) {
	var flagNames []string
	if g.instance != nil {
		for _, flag := range g.instance.Settings.DatabaseFlags {
			flagNames = append(flagNames, flag.Name)
		}
	}
	return flagNames, nil
}

// Instance représente une instance Cloud SQL PostgreSQL (legacy structure for compatibility)
type Instance struct {
	Name    string
	Version string
	IP      string
	Port    string
	Status  string
	Region  string
}

// getPostgresPort récupère le port de l'instance Cloud SQL (sinon 5432 par défaut)
func getPostgresPort(instance *sqladmin.DatabaseInstance) string {
	defaultPort := "5432"
	if instance.Settings != nil {
		for _, flag := range instance.Settings.DatabaseFlags {
			if flag.Name == "cloudsql.ports" {
				return flag.Value
			}
		}
	}
	return defaultPort
}

// ListInstances retourne les instances PostgreSQL du projet (legacy method)
func (g *GCP) ListInstances() ([]Instance, error) {
	if g.service == nil {
		return nil, fmt.Errorf("le client GCP n'a pas été initialisé")
	}

	var instances []Instance
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := g.service.Instances.List(g.projectID)

	err := req.Pages(ctx, func(page *sqladmin.InstancesListResponse) error {
		for _, instance := range page.Items {
			if instance.DatabaseVersion != "" && strings.HasPrefix(instance.DatabaseVersion, "POSTGRES") {
				ip := "N/A"
				if len(instance.IpAddresses) > 0 {
					ip = instance.IpAddresses[0].IpAddress
				}
				instances = append(instances, Instance{
					Name:    instance.Name,
					Version: instance.DatabaseVersion,
					IP:      ip,
					Port:    getPostgresPort(instance),
					Status:  instance.State,
					Region:  instance.Region,
				})
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des instances : %w", err)
	}

	return instances, nil
}

// ConnectToCloudSQLWithIAM établit une connexion sécurisée à Cloud SQL PostgreSQL via IAM
func ConnectToCloudSQLWithIAM(projectID, instanceConnectionName, dbName, user string) (*sql.DB, error) {
	ctx := context.Background()

	if projectID == "" || instanceConnectionName == "" {
		return nil, fmt.Errorf("le projectID et l'instanceConnectionName sont obligatoires")
	}

	dialer, err := cloudsqlconn.NewDialer(ctx, cloudsqlconn.WithIAMAuthN())
	if err != nil {
		return nil, fmt.Errorf("échec de la création du dialer Cloud SQL : %w", err)
	}

	config, err := pgx.ParseConfig(fmt.Sprintf(
		"user=%s dbname=%s sslmode=disable host=/cloudsql/%s",
		user, dbName, instanceConnectionName,
	))
	if err != nil {
		return nil, fmt.Errorf("échec de la configuration de pgx : %w", err)
	}

	config.DialFunc = func(ctx context.Context, network, _ string) (net.Conn, error) {
		return dialer.Dial(ctx, instanceConnectionName)
	}

	db := stdlib.OpenDB(*config)

	if err = db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("échec de la connexion à Cloud SQL : %w", err)
	}

	return db, nil
}

// ExecuteGcloudSQLConnect exécute la commande gcloud sql connect
func ExecuteGcloudSQLConnect(projectID, instanceConnectionName string) (string, error) {
	gcloudCmd := fmt.Sprintf("gcloud sql connect %s --project=%s", instanceConnectionName, projectID)
	cmdExec := exec.Command("bash", "-c", gcloudCmd)

	output, err := cmdExec.CombinedOutput()
	result := string(output)

	if err != nil {
		return "", fmt.Errorf("%s: %s %w", gcloudCmd, result, err)
	}

	return result, nil
}

// DownloadLogs télécharge les logs PostgreSQL depuis Cloud Logging
func (g *GCP) DownloadLogs(start string, directory string, end string) error {
	if g.instanceName == "" {
		return fmt.Errorf("instance name is required")
	}

	// Parse des dates de début et fin
	var startTime, endTime time.Time
	var err error

	if start != "" {
		startTime, err = time.Parse("2006/01/02 15:04:00", start)
		if err != nil {
			return fmt.Errorf("time.Parse start: %w", err)
		}
	} else {
		startTime = time.Now().AddDate(0, 0, -1) // Hier par défaut
	}

	if end != "" {
		endTime, err = time.Parse("2006/01/02 15:04:00", end)
		if err != nil {
			return fmt.Errorf("time.Parse end: %w", err)
		}
	} else {
		endTime = time.Now()
	}

	// Créer le répertoire de destination
	logPath := ""
	if directory == "./" {
		logPath = fmt.Sprintf("%slogs/%s", directory, g.instanceName)
	} else {
		logPath = directory
	}

	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		err = os.MkdirAll(logPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("os.MkdirAll: %w", err)
		}
	}

	// Format des timestamps pour gcloud logging
	startTimestamp := startTime.Format(time.RFC3339)
	endTimestamp := endTime.Format(time.RFC3339)

	// Requête pour récupérer les logs PostgreSQL
	filter := fmt.Sprintf(
		`resource.type="cloudsql_database" AND resource.labels.database_id="%s:%s" AND timestamp>="%s" AND timestamp<="%s"`,
		g.projectID, g.instanceName, startTimestamp, endTimestamp,
	)

	// Exécuter la commande gcloud logging read
	args := fmt.Sprintf(`read '%s' --format=json --order=asc`, filter)
	result, err := g.Execute("logging", args)
	if err != nil {
		return fmt.Errorf("Execute gcloud logging: %w", err)
	}

	// Parser le JSON et extraire les textPayload
	var logs []map[string]interface{}
	err = json.Unmarshal([]byte(result), &logs)
	if err != nil {
		return fmt.Errorf("json.Unmarshal: %w", err)
	}

	// Créer un fichier texte avec les logs formatés pour pgbadger
	textFilePath := fmt.Sprintf("%s/postgres_%s_%s.log", logPath,
		startTime.Format("20060102_150405"),
		endTime.Format("20060102_150405"))

	textFile, err := os.OpenFile(textFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("os.OpenFile text: %w", err)
	}
	defer func() { _ = textFile.Close() }()

	// Écrire chaque log dans le fichier texte
	for _, log := range logs {
		if textPayload, ok := log["textPayload"].(string); ok {
			_, err = textFile.WriteString(textPayload + "\n")
			if err != nil {
				return fmt.Errorf("WriteString: %w", err)
			}
		} else if jsonPayload, ok := log["jsonPayload"].(map[string]interface{}); ok {
			// Si c'est un jsonPayload, essayer d'extraire le message
			if message, ok := jsonPayload["message"].(string); ok {
				_, err = textFile.WriteString(message + "\n")
				if err != nil {
					return fmt.Errorf("WriteString json message: %w", err)
				}
			}
		}
	}

	fmt.Printf("Logs téléchargés:\n")
	fmt.Printf("  - Fichier: %s\n", textFilePath)
	fmt.Printf("  - Nombre d'entrées: %d\n", len(logs))

	return nil
}

// Helper functions for graphs and metrics (similar to RDS)
