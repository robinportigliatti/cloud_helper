package postgresflex

import (
	"fmt"
	"os/exec"
	"strings"
)

// PostgresFlex structure principale pour gérer Azure PostgreSQL Flexible Server
type PostgresFlex struct {
	serverName     string
	resourceGroup  string
	subscription   string
	server         *Server
	configurations ConfigurationListResult
}

func (pf *PostgresFlex) GetServers() ([]Server, error) {
	if pf.server != nil {
		return []Server{*pf.server}, nil
	}
	return []Server{}, nil
}

func (pf *PostgresFlex) Init(serverName string, resourceGroup string, subscription string) error {
	var err error
	pf.serverName = serverName
	pf.resourceGroup = resourceGroup
	pf.subscription = subscription

	// Récupérer les informations du serveur
	err = pf.DescribeServer()
	if err != nil {
		return fmt.Errorf("PostgresFlex: DescribeServer: %w", err)
	}

	// Récupérer les configurations
	err = pf.LoadConfigurations()
	if err != nil {
		return fmt.Errorf("PostgresFlex: LoadConfigurations: %w", err)
	}

	return nil
}

func (pf *PostgresFlex) GetServer() Server {
	if pf.server != nil {
		return *pf.server
	}
	return Server{}
}

func (pf *PostgresFlex) Execute(command string, args string) (string, error) {
	var cmdStr string
	if pf.subscription != "" {
		cmdStr = fmt.Sprintf("az %s %s --subscription %s", command, args, pf.subscription)
	} else {
		cmdStr = fmt.Sprintf("az %s %s", command, args)
	}

	cmd := exec.Command("bash", "-c", cmdStr)
	output, err := cmd.CombinedOutput()
	result := string(output)

	if err != nil {
		return "", fmt.Errorf("%s: %s %w", cmdStr, result, err)
	}

	return result, nil
}

func (pf *PostgresFlex) GetConfigurations() ConfigurationListResult {
	return pf.configurations
}

func (pf *PostgresFlex) DescribeServer() error {
	if pf.serverName == "" {
		return nil
	}

	args := fmt.Sprintf("--resource-group %s --name %s", pf.resourceGroup, pf.serverName)
	_, err := pf.Execute("postgres flexible-server show", args)
	if err != nil {
		return fmt.Errorf("PostgresFlex: Execute: %w", err)
	}

	// Parse JSON response
	var server Server
	// Note: En production, il faudrait utiliser json.Unmarshal ici
	// Pour l'instant, on simule avec la structure
	pf.server = &server

	return nil
}

func (pf *PostgresFlex) GenPsql() (string, error) {
	var result strings.Builder
	if pf.server != nil {
		fqdn := pf.server.Properties.FullyQualifiedDomainName
		_, err := result.WriteString(fmt.Sprintf("psql -h %s -p 5432\n", fqdn))
		if err != nil {
			return "", fmt.Errorf("WriteString: %w", err)
		}
	}
	return result.String(), nil
}

func (pf *PostgresFlex) GetTier() string {
	if pf.server != nil {
		return pf.server.Sku.Tier
	}
	return ""
}

func (pf *PostgresFlex) GetSku() string {
	if pf.server != nil {
		return pf.server.Sku.Name
	}
	return ""
}

func (pf *PostgresFlex) GetStorageSizeGB() int {
	if pf.server != nil {
		return pf.server.Properties.Storage.StorageSizeGB
	}
	return 0
}

func (pf *PostgresFlex) LoadConfigurations() error {
	if pf.serverName == "" {
		return nil
	}

	args := fmt.Sprintf("--resource-group %s --server-name %s", pf.resourceGroup, pf.serverName)
	_, err := pf.Execute("postgres flexible-server parameter list", args)
	if err != nil {
		return fmt.Errorf("PostgresFlex: Execute: %w", err)
	}

	// Parse JSON response
	// Note: En production, il faudrait utiliser json.Unmarshal ici

	return nil
}

func (pf *PostgresFlex) GenPgPass() (string, error) {
	if pf.server == nil {
		return "", fmt.Errorf("no server initialized")
	}

	str := fmt.Sprintf("%s:%s:%s:%s",
		pf.server.Properties.FullyQualifiedDomainName,
		"5432",
		"postgres",
		"<TODO>")
	return str, nil
}

func (pf *PostgresFlex) GenPostgreSQLConf() (string, error) {
	var result strings.Builder
	_, err := result.WriteString("ParameterName;ParameterValue\n")
	if err != nil {
		return "", fmt.Errorf("WriteString: %w", err)
	}

	for _, config := range pf.configurations.Value {
		if config.Properties.Value != "" {
			str := fmt.Sprintf("%s;%s\n", config.Name, config.Properties.Value)
			_, err := result.WriteString(str)
			if err != nil {
				return "", fmt.Errorf("WriteString: %w", err)
			}
		}
	}
	return result.String(), nil
}

func (pf *PostgresFlex) GetConfigurationValue(name string) (string, error) {
	value, err := pf.configurations.GetValueByName(name)
	if err != nil {
		return "", fmt.Errorf("GetValueByName: %w", err)
	}
	return value, nil
}

func (pf *PostgresFlex) CheckConfiguration(name string, expectedValue string) error {
	value, err := pf.GetConfigurationValue(name)
	if err != nil {
		return fmt.Errorf("GetConfigurationValue: %w", err)
	}

	str := ""
	if value != expectedValue {
		str = fmt.Sprintf("%s should be at %s (currently: %s)", name, expectedValue, value)
	} else {
		str = fmt.Sprintf("%s: OK", name)
	}

	fmt.Println(str)
	return nil
}

func (pf *PostgresFlex) CheckPgbadger() error {
	err := pf.CheckConfiguration("log_connections", "on")
	if err != nil {
		return fmt.Errorf("CheckConfiguration: %w", err)
	}
	err = pf.CheckConfiguration("log_disconnections", "on")
	if err != nil {
		return fmt.Errorf("CheckConfiguration: %w", err)
	}
	err = pf.CheckConfiguration("log_lock_waits", "on")
	if err != nil {
		return fmt.Errorf("CheckConfiguration: %w", err)
	}
	err = pf.CheckConfiguration("log_temp_files", "0")
	if err != nil {
		return fmt.Errorf("CheckConfiguration: %w", err)
	}
	err = pf.CheckConfiguration("log_autovacuum_min_duration", "0")
	if err != nil {
		return fmt.Errorf("CheckConfiguration: %w", err)
	}
	err = pf.CheckConfiguration("log_min_duration_statement", "0")
	if err != nil {
		return fmt.Errorf("CheckConfiguration: %w", err)
	}
	err = pf.CheckConfiguration("log_duration", "on")
	if err != nil {
		return fmt.Errorf("CheckConfiguration: %w", err)
	}
	err = pf.CheckConfiguration("log_checkpoints", "on")
	if err != nil {
		return fmt.Errorf("CheckConfiguration: %w", err)
	}
	err = pf.CheckConfiguration("log_statement", "all")
	if err != nil {
		return fmt.Errorf("CheckConfiguration: %w", err)
	}
	err = pf.CheckConfiguration("track_io_timing", "on")
	if err != nil {
		return fmt.Errorf("CheckConfiguration: %w", err)
	}

	_, err = pf.GetConfigurationValue("track_activity_query_size")
	if err != nil {
		return fmt.Errorf("GetConfigurationValue: %w", err)
	}

	return nil
}

func (pf *PostgresFlex) GetAllConfigurationNames() ([]string, error) {
	var configNames []string
	for _, config := range pf.configurations.Value {
		configNames = append(configNames, config.Name)
	}
	return configNames, nil
}

// ListServers retourne tous les serveurs PostgreSQL Flexible dans le resource group
func (pf *PostgresFlex) ListServers() ([]Server, error) {
	args := fmt.Sprintf("--resource-group %s", pf.resourceGroup)
	_, err := pf.Execute("postgres flexible-server list", args)
	if err != nil {
		return nil, fmt.Errorf("PostgresFlex: Execute: %w", err)
	}

	// Note: En production, parser le JSON avec json.Unmarshal
	return []Server{}, nil
}
