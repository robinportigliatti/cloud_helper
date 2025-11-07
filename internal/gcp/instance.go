package gcp

import (
	"time"
)

type DatabaseInstance struct {
	Name                         string              `json:"name"`
	Project                      string              `json:"project"`
	DatabaseVersion              string              `json:"databaseVersion"`
	Region                       string              `json:"region"`
	State                        string              `json:"state"`
	ConnectionName               string              `json:"connectionName"`
	IPAddresses                  []IPMapping         `json:"ipAddresses"`
	Settings                     Settings            `json:"settings"`
	BackupConfiguration          BackupConfiguration `json:"backupConfiguration,omitempty"`
	CreateTime                   time.Time           `json:"createTime"`
	InstanceType                 string              `json:"instanceType"`
	MaintenanceVersion           string              `json:"maintenanceVersion,omitempty"`
	AvailableMaintenanceVersions []string            `json:"availableMaintenanceVersions,omitempty"`
	GceZone                      string              `json:"gceZone,omitempty"`
	SelfLink                     string              `json:"selfLink"`
}

type IPMapping struct {
	Type      string `json:"type"`
	IPAddress string `json:"ipAddress"`
}

type Settings struct {
	Tier                   string              `json:"tier"`
	ActivationPolicy       string              `json:"activationPolicy"`
	DataDiskSizeGb         int64               `json:"dataDiskSizeGb,omitempty"`
	DataDiskType           string              `json:"dataDiskType,omitempty"`
	DatabaseFlags          []DatabaseFlag      `json:"databaseFlags,omitempty"`
	BackupConfiguration    BackupConfiguration `json:"backupConfiguration,omitempty"`
	MaintenanceWindow      MaintenanceWindow   `json:"maintenanceWindow,omitempty"`
	IPConfiguration        IPConfiguration     `json:"ipConfiguration,omitempty"`
	PricingPlan            string              `json:"pricingPlan"`
	StorageAutoResize      bool                `json:"storageAutoResize,omitempty"`
	StorageAutoResizeLimit int64               `json:"storageAutoResizeLimit,omitempty"`
	AvailabilityType       string              `json:"availabilityType,omitempty"`
	LocationPreference     LocationPreference  `json:"locationPreference,omitempty"`
}

type DatabaseFlag struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type BackupConfiguration struct {
	Enabled                     bool                    `json:"enabled"`
	StartTime                   string                  `json:"startTime,omitempty"`
	PointInTimeRecoveryEnabled  bool                    `json:"pointInTimeRecoveryEnabled,omitempty"`
	TransactionLogRetentionDays int                     `json:"transactionLogRetentionDays,omitempty"`
	BackupRetentionSettings     BackupRetentionSettings `json:"backupRetentionSettings,omitempty"`
}

type BackupRetentionSettings struct {
	RetentionUnit   string `json:"retentionUnit"`
	RetainedBackups int    `json:"retainedBackups"`
}

type MaintenanceWindow struct {
	Day         int    `json:"day"`
	Hour        int    `json:"hour"`
	UpdateTrack string `json:"updateTrack,omitempty"`
}

type IPConfiguration struct {
	AuthorizedNetworks []AclEntry `json:"authorizedNetworks,omitempty"`
	Ipv4Enabled        bool       `json:"ipv4Enabled"`
	RequireSsl         bool       `json:"requireSsl,omitempty"`
	PrivateNetwork     string     `json:"privateNetwork,omitempty"`
}

type AclEntry struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type LocationPreference struct {
	Zone                 string `json:"zone,omitempty"`
	SecondaryZone        string `json:"secondaryZone,omitempty"`
	FollowGaeApplication string `json:"followGaeApplication,omitempty"`
}

type DescribeInstanceResult struct {
	Items []DatabaseInstance `json:"items"`
}
