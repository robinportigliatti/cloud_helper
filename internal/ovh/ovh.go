package ovh

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// OVHClient structure principale pour gérer OVHcloud Database Services
type OVHClient struct {
	serviceName string
	clusterID   string
	endpoint    string // eu, ca, us
}

// DatabaseInstance représente une instance PostgreSQL sur OVHcloud
type DatabaseInstance struct {
	ID         string     `json:"id"`
	Name       string     `json:"description"`
	Engine     string     `json:"engine"`
	Version    string     `json:"version"`
	Status     string     `json:"status"`
	Endpoints  []Endpoint `json:"endpoints"`
	NodeNumber int        `json:"nodeNumber"`
	Plan       string     `json:"plan"`
}

// Endpoint représente un point d'accès à la base de données
type Endpoint struct {
	Domain string `json:"domain"`
	Port   int    `json:"port"`
	Scheme string `json:"scheme"`
}

// Init initialise le client OVH avec un service name et un cluster ID
func (o *OVHClient) Init(serviceName string, clusterID string, endpoint string) error {
	o.serviceName = serviceName
	o.clusterID = clusterID

	// Par défaut, utiliser l'endpoint EU
	if endpoint == "" {
		endpoint = "ovh-eu"
	}
	o.endpoint = endpoint

	return nil
}

// Execute exécute une commande via le CLI ovhcloud
func (o *OVHClient) Execute(resource string, action string, additionalArgs string) (string, error) {
	var cmdStr string

	if additionalArgs != "" {
		cmdStr = fmt.Sprintf("ovhcloud-cli %s %s %s --endpoint %s --output json",
			resource, action, additionalArgs, o.endpoint)
	} else {
		cmdStr = fmt.Sprintf("ovhcloud-cli %s %s --endpoint %s --output json",
			resource, action, o.endpoint)
	}

	cmd := exec.Command("bash", "-c", cmdStr)
	output, err := cmd.CombinedOutput()
	result := string(output)

	if err != nil {
		return "", fmt.Errorf("%s: %s %w", cmdStr, result, err)
	}

	return result, nil
}

// ListDatabases retourne toutes les instances PostgreSQL du service
func (o *OVHClient) ListDatabases() ([]DatabaseInstance, error) {
	// Construire la commande pour lister les databases
	args := fmt.Sprintf("--service-name %s", o.serviceName)
	output, err := o.Execute("publiccloud", "database list", args)
	if err != nil {
		return nil, fmt.Errorf("OVH: ListDatabases: %w", err)
	}

	var clusterIDs []string
	err = json.Unmarshal([]byte(output), &clusterIDs)
	if err != nil {
		return nil, fmt.Errorf("OVH: Unmarshal cluster IDs: %w", err)
	}

	var instances []DatabaseInstance

	// Pour chaque cluster ID, récupérer les détails
	for _, id := range clusterIDs {
		args := fmt.Sprintf("--service-name %s --cluster-id %s", o.serviceName, id)
		output, err := o.Execute("publiccloud", "database info", args)
		if err != nil {
			continue
		}

		var instance DatabaseInstance
		err = json.Unmarshal([]byte(output), &instance)
		if err != nil {
			continue
		}

		// Filtrer pour ne garder que les instances PostgreSQL
		if strings.Contains(strings.ToLower(instance.Engine), "postgres") {
			instances = append(instances, instance)
		}
	}

	return instances, nil
}

// GetDatabase récupère les détails d'une instance spécifique
func (o *OVHClient) GetDatabase() (*DatabaseInstance, error) {
	if o.clusterID == "" {
		return nil, fmt.Errorf("OVH: cluster ID non spécifié")
	}

	args := fmt.Sprintf("--service-name %s --cluster-id %s", o.serviceName, o.clusterID)
	output, err := o.Execute("publiccloud", "database info", args)
	if err != nil {
		return nil, fmt.Errorf("OVH: GetDatabase: %w", err)
	}

	var instance DatabaseInstance
	err = json.Unmarshal([]byte(output), &instance)
	if err != nil {
		return nil, fmt.Errorf("OVH: Unmarshal: %w", err)
	}

	return &instance, nil
}

// GenPsql génère une chaîne de connexion psql
func (o *OVHClient) GenPsql(instance *DatabaseInstance, username string, database string) (string, error) {
	if len(instance.Endpoints) == 0 {
		return "", fmt.Errorf("OVH: aucun endpoint trouvé pour cette instance")
	}

	endpoint := instance.Endpoints[0]

	var result strings.Builder
	_, err := result.WriteString(fmt.Sprintf("psql -h %s -p %d -U %s -d %s\n",
		endpoint.Domain,
		endpoint.Port,
		username,
		database))
	if err != nil {
		return "", fmt.Errorf("WriteString: %w", err)
	}

	return result.String(), nil
}

// GetConnectionString retourne les informations de connexion
func (o *OVHClient) GetConnectionString() (string, int, error) {
	instance, err := o.GetDatabase()
	if err != nil {
		return "", 0, fmt.Errorf("OVH: GetDatabase: %w", err)
	}

	if len(instance.Endpoints) == 0 {
		return "", 0, fmt.Errorf("OVH: aucun endpoint trouvé")
	}

	endpoint := instance.Endpoints[0]
	return endpoint.Domain, endpoint.Port, nil
}
