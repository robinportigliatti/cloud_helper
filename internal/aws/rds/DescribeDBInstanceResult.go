package rds

import (
	"time"
)

type DBInstance struct {
	DBInstanceIdentifier string `json:"DBInstanceIdentifier"`
	DBInstanceClass      string `json:"DBInstanceClass"`
	Engine               string `json:"Engine"`
	DBInstanceStatus     string `json:"DBInstanceStatus"`
	MasterUsername       string `json:"MasterUsername"`
	Endpoint             struct {
		Address      string `json:"Address"`
		Port         int    `json:"Port"`
		HostedZoneID string `json:"HostedZoneId"`
	} `json:"Endpoint"`
	AllocatedStorage      int           `json:"AllocatedStorage"`
	InstanceCreateTime    time.Time     `json:"InstanceCreateTime"`
	PreferredBackupWindow string        `json:"PreferredBackupWindow"`
	BackupRetentionPeriod int           `json:"BackupRetentionPeriod"`
	DBSecurityGroups      []interface{} `json:"DBSecurityGroups"`
	VpcSecurityGroups     []struct {
		VpcSecurityGroupID string `json:"VpcSecurityGroupId"`
		Status             string `json:"Status"`
	} `json:"VpcSecurityGroups"`
	DBParameterGroups []struct {
		DBParameterGroupName string `json:"DBParameterGroupName"`
		ParameterApplyStatus string `json:"ParameterApplyStatus"`
	} `json:"DBParameterGroups"`
	AvailabilityZone string `json:"AvailabilityZone"`
	DBSubnetGroup    struct {
		DBSubnetGroupName        string `json:"DBSubnetGroupName"`
		DBSubnetGroupDescription string `json:"DBSubnetGroupDescription"`
		VpcID                    string `json:"VpcId"`
		SubnetGroupStatus        string `json:"SubnetGroupStatus"`
		Subnets                  []struct {
			SubnetIdentifier       string `json:"SubnetIdentifier"`
			SubnetAvailabilityZone struct {
				Name string `json:"Name"`
			} `json:"SubnetAvailabilityZone"`
			SubnetOutpost struct {
			} `json:"SubnetOutpost"`
			SubnetStatus string `json:"SubnetStatus"`
		} `json:"Subnets"`
	} `json:"DBSubnetGroup"`
	PreferredMaintenanceWindow string `json:"PreferredMaintenanceWindow"`
	PendingModifiedValues      struct {
	} `json:"PendingModifiedValues"`
	LatestRestorableTime             time.Time     `json:"LatestRestorableTime"`
	MultiAZ                          bool          `json:"MultiAZ"`
	EngineVersion                    string        `json:"EngineVersion"`
	AutoMinorVersionUpgrade          bool          `json:"AutoMinorVersionUpgrade"`
	ReadReplicaDBInstanceIdentifiers []interface{} `json:"ReadReplicaDBInstanceIdentifiers"`
	LicenseModel                     string        `json:"LicenseModel"`
	OptionGroupMemberships           []struct {
		OptionGroupName string `json:"OptionGroupName"`
		Status          string `json:"Status"`
	} `json:"OptionGroupMemberships"`
	PubliclyAccessible                 bool          `json:"PubliclyAccessible"`
	StorageType                        string        `json:"StorageType"`
	DbInstancePort                     int           `json:"DbInstancePort"`
	StorageEncrypted                   bool          `json:"StorageEncrypted"`
	DbiResourceID                      string        `json:"DbiResourceId"`
	CACertificateIdentifier            string        `json:"CACertificateIdentifier"`
	DomainMemberships                  []interface{} `json:"DomainMemberships"`
	CopyTagsToSnapshot                 bool          `json:"CopyTagsToSnapshot"`
	MonitoringInterval                 int           `json:"MonitoringInterval"`
	DBInstanceArn                      string        `json:"DBInstanceArn"`
	IAMDatabaseAuthenticationEnabled   bool          `json:"IAMDatabaseAuthenticationEnabled"`
	PerformanceInsightsEnabled         bool          `json:"PerformanceInsightsEnabled"`
	PerformanceInsightsKMSKeyID        string        `json:"PerformanceInsightsKMSKeyId"`
	PerformanceInsightsRetentionPeriod int           `json:"PerformanceInsightsRetentionPeriod"`
	DeletionProtection                 bool          `json:"DeletionProtection"`
	AssociatedRoles                    []interface{} `json:"AssociatedRoles"`
	MaxAllocatedStorage                int           `json:"MaxAllocatedStorage,omitempty"`
	TagList                            []interface{} `json:"TagList"`
	CustomerOwnedIPEnabled             bool          `json:"CustomerOwnedIpEnabled"`
	ActivityStreamStatus               string        `json:"ActivityStreamStatus"`
	BackupTarget                       string        `json:"BackupTarget"`
}

type DescribeDBInstanceResult struct {
	DBInstances []DBInstance `json:"DBInstances"`
}
