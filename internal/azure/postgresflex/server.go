package postgresflex

import (
	"time"
)

type Server struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	Location   string            `json:"location"`
	Tags       map[string]string `json:"tags,omitempty"`
	Sku        ServerSku         `json:"sku"`
	Properties ServerProperties  `json:"properties"`
}

type ServerSku struct {
	Name string `json:"name"`
	Tier string `json:"tier"`
}

type ServerProperties struct {
	AdministratorLogin       string            `json:"administratorLogin"`
	Version                  string            `json:"version"`
	State                    string            `json:"state"`
	FullyQualifiedDomainName string            `json:"fullyQualifiedDomainName"`
	Storage                  Storage           `json:"storage"`
	Backup                   Backup            `json:"backup"`
	Network                  Network           `json:"network,omitempty"`
	HighAvailability         HighAvailability  `json:"highAvailability,omitempty"`
	MaintenanceWindow        MaintenanceWindow `json:"maintenanceWindow,omitempty"`
	AvailabilityZone         string            `json:"availabilityZone,omitempty"`
	CreateMode               string            `json:"createMode,omitempty"`
}

type Storage struct {
	StorageSizeGB int  `json:"storageSizeGB"`
	AutoGrow      bool `json:"autoGrow,omitempty"`
	Iops          int  `json:"iops,omitempty"`
}

type Backup struct {
	BackupRetentionDays int       `json:"backupRetentionDays"`
	GeoRedundantBackup  string    `json:"geoRedundantBackup,omitempty"`
	EarliestRestoreDate time.Time `json:"earliestRestoreDate,omitempty"`
}

type Network struct {
	PublicNetworkAccess         string `json:"publicNetworkAccess,omitempty"`
	DelegatedSubnetResourceId   string `json:"delegatedSubnetResourceId,omitempty"`
	PrivateDnsZoneArmResourceId string `json:"privateDnsZoneArmResourceId,omitempty"`
}

type HighAvailability struct {
	Mode                    string `json:"mode,omitempty"`
	State                   string `json:"state,omitempty"`
	StandbyAvailabilityZone string `json:"standbyAvailabilityZone,omitempty"`
}

type MaintenanceWindow struct {
	CustomWindow string `json:"customWindow,omitempty"`
	DayOfWeek    int    `json:"dayOfWeek,omitempty"`
	StartHour    int    `json:"startHour,omitempty"`
	StartMinute  int    `json:"startMinute,omitempty"`
}

type ServerListResult struct {
	Value []Server `json:"value"`
}
