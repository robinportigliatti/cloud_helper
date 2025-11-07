package rds

import (
	"context"
	"encoding/csv"
	"fmt"
	"image/color"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cwTypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	awsRds "github.com/aws/aws-sdk-go-v2/service/rds"
	rdsTypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

type RDS struct {
	dbInstanceIdentifier string
	profile              string
	awsConfig            aws.Config
	rdsClient            *awsRds.Client
	ec2Client            *ec2.Client
	cloudwatchClient     *cloudwatch.Client
	ctx                  context.Context

	// Cached data
	dbInstances                               DescribeDBInstanceResult
	dbParameterGroups                         DescribeDBParametersResult
	describeInstanceTypes                     DescribeInstanceTypes
	validDBInstanceModificationsMessageResult ValidDBInstanceModificationsMessageResult
}

func (rds RDS) GetDbInstances() ([]DBInstance, error) {
	return rds.dbInstances.DBInstances, nil
}

type DBParameter struct {
	ParameterName  string
	ParameterValue string
}

func (rds *RDS) Init(dbInstanceIdentifier string, profile string) error {
	var err error
	rds.dbInstanceIdentifier = dbInstanceIdentifier
	rds.profile = profile
	rds.ctx = context.Background()

	// Load AWS configuration with optional profile
	var cfgOptions []func(*config.LoadOptions) error
	if profile != "" && profile != "none" {
		cfgOptions = append(cfgOptions, config.WithSharedConfigProfile(profile))
	}

	rds.awsConfig, err = config.LoadDefaultConfig(rds.ctx, cfgOptions...)
	if err != nil {
		return fmt.Errorf("RDS: failed to load AWS config: %w", err)
	}

	// Initialize AWS service clients
	rds.rdsClient = awsRds.NewFromConfig(rds.awsConfig)
	rds.ec2Client = ec2.NewFromConfig(rds.awsConfig)
	rds.cloudwatchClient = cloudwatch.NewFromConfig(rds.awsConfig)

	// Fetch initial data
	rds.dbInstances, err = rds.DescribeDbInstances()
	if err != nil {
		return fmt.Errorf("RDS: DescribeDbInstances: %w", err)
	}

	err = rds.GetDBParameterGroupInformations()
	if err != nil {
		return fmt.Errorf("RDS: GetDBParameterGroupInformations: %w", err)
	}

	err = rds.GetEC2Information()
	if err != nil {
		return fmt.Errorf("RDS: GetEC2Information: %w", err)
	}

	err = rds.DescribeValidDBInstanceModifications()
	if err != nil {
		return fmt.Errorf("RDS: DescribeValidDBInstanceModifications: %w", err)
	}

	return nil
}

func (rds RDS) GetdbInstance() DBInstance {
	return rds.dbInstances.DBInstances[0]
}

func (rds RDS) GetDBParameters() DescribeDBParametersResult {
	return rds.dbParameterGroups
}

func (rds RDS) GetInstanceType() InstanceType {
	return rds.describeInstanceTypes.InstanceTypes[0]
}

func (rds *RDS) DescribeDbInstances() (DescribeDBInstanceResult, error) {
	var dbInstances DescribeDBInstanceResult

	input := &awsRds.DescribeDBInstancesInput{}
	if rds.dbInstanceIdentifier != "" {
		input.DBInstanceIdentifier = aws.String(rds.dbInstanceIdentifier)
	}

	result, err := rds.rdsClient.DescribeDBInstances(rds.ctx, input)
	if err != nil {
		return dbInstances, fmt.Errorf("RDS: DescribeDBInstances SDK call: %w", err)
	}

	// Convert AWS SDK types to our internal types
	dbInstances.DBInstances = make([]DBInstance, len(result.DBInstances))
	for i, sdkInstance := range result.DBInstances {
		dbInstances.DBInstances[i] = convertSDKDBInstanceToInternal(sdkInstance)
	}

	return dbInstances, nil
}

// Helper function to convert AWS SDK DBInstance to internal DBInstance
func convertSDKDBInstanceToInternal(sdkInstance rdsTypes.DBInstance) DBInstance {
	instance := DBInstance{}

	if sdkInstance.DBInstanceIdentifier != nil {
		instance.DBInstanceIdentifier = *sdkInstance.DBInstanceIdentifier
	}
	if sdkInstance.DBInstanceClass != nil {
		instance.DBInstanceClass = *sdkInstance.DBInstanceClass
	}
	if sdkInstance.Engine != nil {
		instance.Engine = *sdkInstance.Engine
	}
	if sdkInstance.DBInstanceStatus != nil {
		instance.DBInstanceStatus = *sdkInstance.DBInstanceStatus
	}
	if sdkInstance.MasterUsername != nil {
		instance.MasterUsername = *sdkInstance.MasterUsername
	}

	// Convert Endpoint
	if sdkInstance.Endpoint != nil {
		if sdkInstance.Endpoint.Address != nil {
			instance.Endpoint.Address = *sdkInstance.Endpoint.Address
		}
		if sdkInstance.Endpoint.Port != nil {
			instance.Endpoint.Port = int(*sdkInstance.Endpoint.Port)
		}
		if sdkInstance.Endpoint.HostedZoneId != nil {
			instance.Endpoint.HostedZoneID = *sdkInstance.Endpoint.HostedZoneId
		}
	}

	if sdkInstance.AllocatedStorage != nil {
		instance.AllocatedStorage = int(*sdkInstance.AllocatedStorage)
	}
	if sdkInstance.InstanceCreateTime != nil {
		instance.InstanceCreateTime = *sdkInstance.InstanceCreateTime
	}
	if sdkInstance.PreferredBackupWindow != nil {
		instance.PreferredBackupWindow = *sdkInstance.PreferredBackupWindow
	}
	if sdkInstance.BackupRetentionPeriod != nil {
		instance.BackupRetentionPeriod = int(*sdkInstance.BackupRetentionPeriod)
	}

	// Convert DBParameterGroups
	instance.DBParameterGroups = make([]struct {
		DBParameterGroupName string `json:"DBParameterGroupName"`
		ParameterApplyStatus string `json:"ParameterApplyStatus"`
	}, len(sdkInstance.DBParameterGroups))
	for i, pg := range sdkInstance.DBParameterGroups {
		if pg.DBParameterGroupName != nil {
			instance.DBParameterGroups[i].DBParameterGroupName = *pg.DBParameterGroupName
		}
		if pg.ParameterApplyStatus != nil {
			instance.DBParameterGroups[i].ParameterApplyStatus = *pg.ParameterApplyStatus
		}
	}

	if sdkInstance.AvailabilityZone != nil {
		instance.AvailabilityZone = *sdkInstance.AvailabilityZone
	}
	if sdkInstance.PreferredMaintenanceWindow != nil {
		instance.PreferredMaintenanceWindow = *sdkInstance.PreferredMaintenanceWindow
	}
	if sdkInstance.LatestRestorableTime != nil {
		instance.LatestRestorableTime = *sdkInstance.LatestRestorableTime
	}
	if sdkInstance.MultiAZ != nil {
		instance.MultiAZ = *sdkInstance.MultiAZ
	}
	if sdkInstance.EngineVersion != nil {
		instance.EngineVersion = *sdkInstance.EngineVersion
	}
	if sdkInstance.AutoMinorVersionUpgrade != nil {
		instance.AutoMinorVersionUpgrade = *sdkInstance.AutoMinorVersionUpgrade
	}
	if sdkInstance.LicenseModel != nil {
		instance.LicenseModel = *sdkInstance.LicenseModel
	}
	if sdkInstance.PubliclyAccessible != nil {
		instance.PubliclyAccessible = *sdkInstance.PubliclyAccessible
	}
	if sdkInstance.StorageType != nil {
		instance.StorageType = *sdkInstance.StorageType
	}
	if sdkInstance.DbInstancePort != nil {
		instance.DbInstancePort = int(*sdkInstance.DbInstancePort)
	}
	if sdkInstance.StorageEncrypted != nil {
		instance.StorageEncrypted = *sdkInstance.StorageEncrypted
	}
	if sdkInstance.DbiResourceId != nil {
		instance.DbiResourceID = *sdkInstance.DbiResourceId
	}
	if sdkInstance.CACertificateIdentifier != nil {
		instance.CACertificateIdentifier = *sdkInstance.CACertificateIdentifier
	}
	if sdkInstance.CopyTagsToSnapshot != nil {
		instance.CopyTagsToSnapshot = *sdkInstance.CopyTagsToSnapshot
	}
	if sdkInstance.MonitoringInterval != nil {
		instance.MonitoringInterval = int(*sdkInstance.MonitoringInterval)
	}
	if sdkInstance.DBInstanceArn != nil {
		instance.DBInstanceArn = *sdkInstance.DBInstanceArn
	}
	if sdkInstance.IAMDatabaseAuthenticationEnabled != nil {
		instance.IAMDatabaseAuthenticationEnabled = *sdkInstance.IAMDatabaseAuthenticationEnabled
	}
	if sdkInstance.PerformanceInsightsEnabled != nil {
		instance.PerformanceInsightsEnabled = *sdkInstance.PerformanceInsightsEnabled
	}
	if sdkInstance.PerformanceInsightsKMSKeyId != nil {
		instance.PerformanceInsightsKMSKeyID = *sdkInstance.PerformanceInsightsKMSKeyId
	}
	if sdkInstance.PerformanceInsightsRetentionPeriod != nil {
		instance.PerformanceInsightsRetentionPeriod = int(*sdkInstance.PerformanceInsightsRetentionPeriod)
	}
	if sdkInstance.DeletionProtection != nil {
		instance.DeletionProtection = *sdkInstance.DeletionProtection
	}
	if sdkInstance.MaxAllocatedStorage != nil {
		instance.MaxAllocatedStorage = int(*sdkInstance.MaxAllocatedStorage)
	}
	if sdkInstance.CustomerOwnedIpEnabled != nil {
		instance.CustomerOwnedIPEnabled = *sdkInstance.CustomerOwnedIpEnabled
	}
	if sdkInstance.ActivityStreamStatus != "" {
		instance.ActivityStreamStatus = string(sdkInstance.ActivityStreamStatus)
	}
	if sdkInstance.BackupTarget != nil {
		instance.BackupTarget = *sdkInstance.BackupTarget
	}

	return instance
}

func (rds *RDS) GenPsql() (string, error) {
	var result strings.Builder
	for i := 0; i < len(rds.dbInstances.DBInstances); i++ {
		_, err := result.WriteString(fmt.Sprintf("psql -h %s -p %d\n", rds.dbInstances.DBInstances[i].Endpoint.Address, rds.dbInstances.DBInstances[i].Endpoint.Port))
		if err != nil {
			return "", fmt.Errorf("WriteString: %w", err)
		}
	}
	return result.String(), nil
}

func (rds *RDS) GetDBInstanceClass() string {
	return rds.dbInstances.DBInstances[0].DBInstanceClass
}

func (rds *RDS) GetDefaultVCpus() int {
	return rds.describeInstanceTypes.InstanceTypes[0].VCPUInfo.DefaultVCpus
}

func (rds RDS) GetMemoryInfo() MemoryInfo {
	return rds.describeInstanceTypes.InstanceTypes[0].MemoryInfo
}

func (rds RDS) GetValidDBInstanceModifications() ValidDBInstanceModificationsMessage {
	return rds.validDBInstanceModificationsMessageResult.ValidDBInstanceModificationsMessage
}

func (rds *RDS) GetEC2Information() error {
	if rds.dbInstanceIdentifier != "" && len(rds.dbInstances.DBInstances) > 0 {
		// RDS instance classes have "db." prefix, EC2 instance types don't
		instanceType := strings.Replace(rds.dbInstances.DBInstances[0].DBInstanceClass, "db.", "", 1)

		input := &ec2.DescribeInstanceTypesInput{
			InstanceTypes: []ec2Types.InstanceType{ec2Types.InstanceType(instanceType)},
		}

		result, err := rds.ec2Client.DescribeInstanceTypes(rds.ctx, input)
		if err != nil {
			return fmt.Errorf("EC2: DescribeInstanceTypes SDK call: %w", err)
		}

		if len(result.InstanceTypes) == 0 {
			return fmt.Errorf("EC2: no instance type found for %s", instanceType)
		}

		// Convert AWS SDK types to our internal types
		// Only converting the fields we actually use in the code
		rds.describeInstanceTypes.InstanceTypes = make([]InstanceType, len(result.InstanceTypes))
		for i, sdkType := range result.InstanceTypes {
			it := InstanceType{}

			// Basic fields
			if sdkType.InstanceType != "" {
				it.InstanceType = string(sdkType.InstanceType)
			}
			if sdkType.CurrentGeneration != nil {
				it.CurrentGeneration = *sdkType.CurrentGeneration
			}

			// VCPUInfo - used by GetDefaultVCpus()
			if sdkType.VCpuInfo != nil {
				if sdkType.VCpuInfo.DefaultVCpus != nil {
					it.VCPUInfo.DefaultVCpus = int(*sdkType.VCpuInfo.DefaultVCpus)
				}
				if sdkType.VCpuInfo.DefaultCores != nil {
					it.VCPUInfo.DefaultCores = int(*sdkType.VCpuInfo.DefaultCores)
				}
				if sdkType.VCpuInfo.DefaultThreadsPerCore != nil {
					it.VCPUInfo.DefaultThreadsPerCore = int(*sdkType.VCpuInfo.DefaultThreadsPerCore)
				}
			}

			// MemoryInfo - used by GetMemoryInfo() and Free_m()
			if sdkType.MemoryInfo != nil && sdkType.MemoryInfo.SizeInMiB != nil {
				it.MemoryInfo.SizeInMiB = int(*sdkType.MemoryInfo.SizeInMiB)
			}

			rds.describeInstanceTypes.InstanceTypes[i] = it
		}
	}

	return nil
}

func (rds *RDS) GetDBParameterGroupInformations() error {
	if rds.dbInstanceIdentifier != "" && len(rds.dbInstances.DBInstances) > 0 {
		if len(rds.dbInstances.DBInstances[0].DBParameterGroups) == 0 {
			return nil
		}

		parameterGroupName := rds.dbInstances.DBInstances[0].DBParameterGroups[0].DBParameterGroupName

		input := &awsRds.DescribeDBParametersInput{
			DBParameterGroupName: aws.String(parameterGroupName),
		}

		result, err := rds.rdsClient.DescribeDBParameters(rds.ctx, input)
		if err != nil {
			return fmt.Errorf("RDS: DescribeDBParameters SDK call: %w", err)
		}

		// Convert AWS SDK types to our internal types
		rds.dbParameterGroups.Parameters = make([]Parameter, len(result.Parameters))
		for i, sdkParam := range result.Parameters {
			param := Parameter{}
			if sdkParam.ParameterName != nil {
				param.ParameterName = *sdkParam.ParameterName
			}
			if sdkParam.Description != nil {
				param.Description = *sdkParam.Description
			}
			if sdkParam.Source != nil {
				param.Source = *sdkParam.Source
			}
			if sdkParam.ApplyType != nil {
				param.ApplyType = *sdkParam.ApplyType
			}
			if sdkParam.DataType != nil {
				param.DataType = *sdkParam.DataType
			}
			if sdkParam.IsModifiable != nil {
				param.IsModifiable = *sdkParam.IsModifiable
			}
			if sdkParam.ApplyMethod != "" {
				param.ApplyMethod = string(sdkParam.ApplyMethod)
			}
			if sdkParam.ParameterValue != nil {
				param.ParameterValue = *sdkParam.ParameterValue
			}
			if sdkParam.AllowedValues != nil {
				param.AllowedValues = *sdkParam.AllowedValues
			}
			if sdkParam.MinimumEngineVersion != nil {
				param.MinimumEngineVersion = *sdkParam.MinimumEngineVersion
			}
			rds.dbParameterGroups.Parameters[i] = param
		}
	}

	return nil
}

func (rds *RDS) DescribeValidDBInstanceModifications() error {
	if rds.dbInstanceIdentifier != "" {
		input := &awsRds.DescribeValidDBInstanceModificationsInput{
			DBInstanceIdentifier: aws.String(rds.dbInstanceIdentifier),
		}

		result, err := rds.rdsClient.DescribeValidDBInstanceModifications(rds.ctx, input)
		if err != nil {
			return fmt.Errorf("RDS: DescribeValidDBInstanceModifications SDK call: %w", err)
		}

		// Convert AWS SDK types to our internal types
		if result.ValidDBInstanceModificationsMessage != nil {
			msg := &rds.validDBInstanceModificationsMessageResult.ValidDBInstanceModificationsMessage

			// Convert Storage array
			if len(result.ValidDBInstanceModificationsMessage.Storage) > 0 {
				msg.Storage = make([]Storage, len(result.ValidDBInstanceModificationsMessage.Storage))
				for i, sdkStorage := range result.ValidDBInstanceModificationsMessage.Storage {
					storage := Storage{}
					if sdkStorage.StorageType != nil {
						storage.StorageType = *sdkStorage.StorageType
					}
					if sdkStorage.SupportsStorageAutoscaling != nil {
						storage.SupportsStorageAutoscaling = *sdkStorage.SupportsStorageAutoscaling
					}

					// Convert StorageSize ranges
					if len(sdkStorage.StorageSize) > 0 {
						storage.StorageSize = make([]struct {
							From int `json:"From"`
							To   int `json:"To"`
							Step int `json:"Step"`
						}, len(sdkStorage.StorageSize))
						for j, sizeRange := range sdkStorage.StorageSize {
							if sizeRange.From != nil {
								storage.StorageSize[j].From = int(*sizeRange.From)
							}
							if sizeRange.To != nil {
								storage.StorageSize[j].To = int(*sizeRange.To)
							}
							if sizeRange.Step != nil {
								storage.StorageSize[j].Step = int(*sizeRange.Step)
							}
						}
					}

					// Convert ProvisionedIops ranges
					if len(sdkStorage.ProvisionedIops) > 0 {
						storage.ProvisionedIops = make([]struct {
							From int `json:"From"`
							To   int `json:"To"`
							Step int `json:"Step"`
						}, len(sdkStorage.ProvisionedIops))
						for j, iopsRange := range sdkStorage.ProvisionedIops {
							if iopsRange.From != nil {
								storage.ProvisionedIops[j].From = int(*iopsRange.From)
							}
							if iopsRange.To != nil {
								storage.ProvisionedIops[j].To = int(*iopsRange.To)
							}
							if iopsRange.Step != nil {
								storage.ProvisionedIops[j].Step = int(*iopsRange.Step)
							}
						}
					}

					// Convert IopsToStorageRatio ranges
					if len(sdkStorage.IopsToStorageRatio) > 0 {
						storage.IopsToStorageRatio = make([]struct {
							From float64 `json:"From"`
							To   float64 `json:"To"`
						}, len(sdkStorage.IopsToStorageRatio))
						for j, ratioRange := range sdkStorage.IopsToStorageRatio {
							if ratioRange.From != nil {
								storage.IopsToStorageRatio[j].From = *ratioRange.From
							}
							if ratioRange.To != nil {
								storage.IopsToStorageRatio[j].To = *ratioRange.To
							}
						}
					}

					msg.Storage[i] = storage
				}
			}
		}
	}

	return nil
}

func DownloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("http.Get: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("os.Create: %w", err)
	}
	defer func() { _ = out.Close() }()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}

	return err
}

func (rds RDS) DownloadLogs(start string, directory string, end string) error {
	str := ""
	var err error
	var unixEndMilli int64 = 0 // Initialisation correcte en int64

	if start == "" {
		str, err = rds.Execute("rds", "describe-db-log-files", fmt.Sprintf("--db-instance-identifier=%s", rds.dbInstanceIdentifier))
		if err != nil {
			return fmt.Errorf("RDS: Execute: %w", err)
		}
	} else {
		unixStart, errParse := time.Parse("2006/01/02 15:04:00", start)
		if errParse != nil {
			return fmt.Errorf("time.Parse: %w", errParse)
		}
		str, err = rds.Execute("rds", "describe-db-log-files", fmt.Sprintf("--db-instance-identifier=%s --file-last-written=%d", rds.dbInstanceIdentifier, unixStart.UnixMilli()))
		if err != nil {
			return fmt.Errorf("RDS: Execute: %w", err)
		}
		if end != "" {
			unixEnd, errParse := time.Parse("2006/01/02 15:04:00", end)
			if errParse != nil {
				return fmt.Errorf("time.Parse: %w", errParse)
			}
			unixEndMilli = unixEnd.UnixMilli()
		}
	}

	logFiles := new(DescribeDBLogFilesResult)
	err = json.Unmarshal([]byte(str), &logFiles)
	if err != nil {
		return fmt.Errorf("json.Unmarshal) %w", err)
	}

	logPath := ""
	if directory == "./" {
		logPath = fmt.Sprintf("%slogs/%s", directory, rds.dbInstanceIdentifier)
	} else {
		logPath = directory
	}

	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		err = os.MkdirAll(logPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("os.MkdirAll: %w", err)
		}
	}

	for i := 0; i < len(logFiles.DescribeDBLogFiles); i++ {
		if logFiles.DescribeDBLogFiles[i].LastWritten <= unixEndMilli {
			filePath := fmt.Sprintf("%s/%s", logPath, strings.ReplaceAll(logFiles.DescribeDBLogFiles[i].LogFileName, "error/", ""))
			var file, err2 = os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
			if err2 != nil {
				fmt.Println(err2.Error())
			}

			_, err = file.WriteString(str)
			if err != nil {
				return fmt.Errorf("file.WriteString: %w", err)
			}

			defer func() { _ = file.Close() }()
		}
	}

	return nil
}

func (rds RDS) DownloadMetrics(start string, end string, directory string) error {
	str, err := rds.Execute("cloudwatch", "list-metrics", "--namespace AWS/RDS")
	if err != nil {
		return fmt.Errorf("RDS: Execute: %w", err)
	}
	metrics := new(ListMetricsResult)

	var pngFiles []string
	err = json.Unmarshal([]byte(str), &metrics)
	if err != nil {
		return fmt.Errorf("json.Unmarshal: %w", err)
	}

	lastDayStr := ""
	currentTimeStr := ""
	if start == "" && end == "" {
		currentTime := time.Now()
		lastDay := currentTime.AddDate(0, 0, -1)
		lastDayStr = lastDay.Format("2006/01/02 15:04:05.000000000")
		currentTimeStr = currentTime.Format("2006/01/02 15:04:05.000000000")
	} else {
		lastDayStr = start
		currentTimeStr = end
	}

	metricsPath := ""
	if directory == "./" {
		metricsPath = "./metrics/"
	} else {
		metricsPath = fmt.Sprintf("%s/metrics", directory)
	}

	err = os.MkdirAll(metricsPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("os.MkdirAll: %w", err)
	}

	for i := 0; i < len(metrics.Metrics); i++ {
		metricName := metrics.Metrics[i].MetricName
		for j := 0; j < len(metrics.Metrics[i].Dimensions); j++ {
			dimensionName := metrics.Metrics[i].Dimensions[j].Name
			dimensionValue := metrics.Metrics[i].Dimensions[j].Value
			dimensionParam := fmt.Sprintf("'Name=%s,Value=%s'", dimensionName, dimensionValue)
			if dimensionName != "DBInstanceIdentifier" || dimensionValue != rds.dbInstanceIdentifier {
				continue
			}

			dimensionPath := fmt.Sprintf("%s/%s", metricsPath, dimensionName)
			err = os.MkdirAll(dimensionPath, os.ModePerm)
			if err != nil {
				return fmt.Errorf("os.MkdirAll: %w", err)
			}

			dimensionValuePath := fmt.Sprintf("%s/%s", dimensionPath, dimensionValue)
			err = os.MkdirAll(dimensionValuePath, os.ModePerm)
			if err != nil {
				return fmt.Errorf("os.MkdirAll: %w", err)
			}

			metricPath := fmt.Sprintf("%s/%s", dimensionValuePath, metricName)
			err = os.MkdirAll(metricPath, os.ModePerm)
			if err != nil {
				return fmt.Errorf("os.MkdirAll: %w", err)
			}

			statistics := [5]string{"SampleCount", "Average", "Sum", "Minimum", "Maximum"}
			for k := 0; k < len(statistics); k++ {
				str, err := rds.Execute("cloudwatch", "get-metric-statistics", fmt.Sprintf("--namespace AWS/RDS --metric-name %s --start-time '%s' --end-time '%s' --period 60 --statistics %s --dimensions=%s", metricName, lastDayStr, currentTimeStr, statistics[k], dimensionParam))
				if err != nil {
					return fmt.Errorf("RDS: Execute: %w", err)
				}
				datapoints := new(GetMetricsStatisticsResult)
				err = json.Unmarshal([]byte(str), &datapoints)
				if err != nil {
					return fmt.Errorf("json.Unmarshal: %w", err)
				}
				filePath := fmt.Sprintf("%s/%s.%s.%s.%s.csv", metricPath, dimensionName, dimensionValue, metricName, statistics[k])
				var file, err2 = os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
				if err2 != nil {
					return fmt.Errorf("os.OpenFile: %w", err2)
				}

				title := fmt.Sprintf("\"%s\";\"%s\";\"%s\";\r\n", "Timestamp", statistics[k], "Unit")
				_, err = file.WriteString(title)
				if err != nil {
					return fmt.Errorf("file.WriteString: %w", err)
				}

				for l := 0; l < len(datapoints.Datapoints); l++ {
					data := fmt.Sprintf("\"%s\";\"%s\";\"%s\";\r\n", datapoints.Datapoints[l].Timestamp, datapoints.Datapoints[l].GetField(statistics[k]), datapoints.Datapoints[l].Unit)
					_, err := file.WriteString(data)
					if err != nil {
						return fmt.Errorf("file.WriteString: %w", err)
					}

				}

				func() { _ = file.Close() }()
				file, err = os.Open(filePath)
				if err != nil {
					return fmt.Errorf("os.Open: %w", err)
				}
				func() { _ = file.Close() }()
				file, err = os.Open(filePath)
				if err != nil {
					return fmt.Errorf("os.Open: %w", err)
				}
				defer func() { _ = file.Close() }()

				data, yAxisLabel, err := readCSV(file)
				if err != nil {
					return fmt.Errorf("readCSV: %w", err)
				}

				baseFilename := filepath.Base(filePath)
				outputFilename := strings.Replace(baseFilename, ".csv", ".png", 1)
				outputDir := filepath.Dir(filePath)
				outputPath := filepath.Join(outputDir, outputFilename)

				err = createPNGGraph(data, yAxisLabel, outputPath)

				if err != nil {
					return fmt.Errorf("createPNGGraph: %w", err)
				}

				pngFiles = append(pngFiles, outputFilename)
			}

			// Création du fichier HTML
			err = createMetricsHTML(pngFiles, rds.dbInstanceIdentifier)
			if err != nil {
				return fmt.Errorf("createMetricsHTML: %w", err)
			}
		}
	}

	return nil
}

func (rds RDS) Free_m() (string, error) {
	currentTime := time.Now()
	fiveMinutesBefore := currentTime.AddDate(0, 0, -1)

	startTime := fiveMinutesBefore.Format("2006/01/02 15:04:05.000000000")
	endTime := currentTime.Format("2006/01/02 15:04:05.000000000")

	dimensionParam := fmt.Sprintf("'Name=%s,Value=%s'", "DBInstanceIdentifier", rds.dbInstanceIdentifier)
	statistic := "Maximum"
	metric := "FreeableMemory"

	str, err := rds.Execute("cloudwatch", "get-metric-statistics",
		fmt.Sprintf("--namespace AWS/RDS --metric-name %s --start-time '%s' --end-time '%s' --period 86400 --unit Bytes --statistics %s --dimensions=%s",
			metric, startTime, endTime, statistic, dimensionParam))
	if err != nil {
		return "", fmt.Errorf("RDS: Execute: %w", err)
	}

	datapoints := new(GetMetricsStatisticsResult)
	err = json.Unmarshal([]byte(str), &datapoints)
	if err != nil {
		return "", fmt.Errorf("json.Unmarshal: %w", err)
	}

	// Vérification qu'il y a bien un datapoint
	if len(datapoints.Datapoints) == 0 {
		return "", fmt.Errorf("no datapoints for %s", metric)
	}

	freeableMemory := datapoints.Datapoints[0].GetField(statistic)
	freeableMemoryFloat, err := strconv.ParseFloat(freeableMemory, 64)
	if err != nil {
		return "", fmt.Errorf("strconv.ParseFloat: %w", err)
	}

	freeableMemory_MB := freeableMemoryFloat / 1024 / 1024
	total_memory_MB := rds.GetMemoryInfo().SizeInMiB
	usedMemoryMB := float64(total_memory_MB) - freeableMemory_MB

	// Concaténation du résultat dans une chaîne de caractères
	result := fmt.Sprintf(
		"%15s %12s %12s\n%15d %12.2f %12.2f\n",
		"total", "utilisé", "disponible",
		total_memory_MB, usedMemoryMB, freeableMemory_MB,
	)

	return result, nil
}

func (rds RDS) GenPgPass() (string, error) {
	// FIXME : needs improvement
	// FIXME : il faudrait faire une commande qui récupère le endpoint et le port directement
	str := fmt.Sprintf("%s:%d:%s:%s", rds.dbInstances.DBInstances[0].Endpoint.Address, rds.dbInstances.DBInstances[0].Endpoint.Port, "postgres", "<TODO>")
	return str, nil
}
func (rds RDS) GenPostgreSQLConf() (string, error) {
	var result strings.Builder
	_, err := result.WriteString("ParameterName;ParameterValue\n")
	if err != nil {
		return "", fmt.Errorf("WriteString: %w", err)
	}
	for i := 0; i < len(rds.dbParameterGroups.Parameters); i++ {
		if rds.dbParameterGroups.Parameters[i].ParameterValue != "" {
			str := fmt.Sprintf("%s;%s\n", rds.dbParameterGroups.Parameters[i].ParameterName, rds.dbParameterGroups.Parameters[i].ParameterValue)
			_, err := result.WriteString(str)
			if err != nil {
				return "", fmt.Errorf("WriteString: %w", err)
			}
		}
	}
	return result.String(), nil
}

func (rds RDS) GetParameterValueByParameterName(parameterName string) (string, error) {
	value, err := rds.dbParameterGroups.GetParameterValueByParameterName(parameterName)
	if err != nil {
		return "", fmt.Errorf("GetParameterValueByParameterName: %w", err)
	}

	return value, nil
}

func (rds RDS) CheckParameter(parameterName string, parameterValue string) error {
	value, err := rds.GetParameterValueByParameterName(parameterName)
	if err != nil {
		return fmt.Errorf("GetParameterValueByParameterName: %w", err)
	}
	str := ""
	if value != parameterValue {
		str = fmt.Sprintf("%s should be at %s", parameterName, parameterValue)
	}

	if value == parameterValue {
		str = fmt.Sprintf("%s: OK", parameterName)
	}

	fmt.Println(str)

	return nil
}

func (rds RDS) CheckPgbadger() error {
	// FIXME : needs improvement
	err := rds.CheckParameter("log_connections", "on")
	if err != nil {
		return fmt.Errorf("CheckParameter: %w", err)
	}
	err = rds.CheckParameter("log_disconnections", "on")
	if err != nil {
		return fmt.Errorf("CheckParameter: %w", err)
	}
	err = rds.CheckParameter("log_rotation_size", "1GO")
	if err != nil {
		return fmt.Errorf("CheckParameter: %w", err)
	}
	err = rds.CheckParameter("lc_messages", "C")
	if err != nil {
		return fmt.Errorf("CheckParameter: %w", err)
	}
	err = rds.CheckParameter("log_line_prefix", "%m [%p]: [%l-1] xact=%x,user=%u,db=%d,client=%h, app=%a ")
	if err != nil {
		return fmt.Errorf("CheckParameter: %w", err)
	}
	err = rds.CheckParameter("log_lock_waits", "on")
	if err != nil {
		return fmt.Errorf("CheckParameter: %w", err)
	}
	err = rds.CheckParameter("log_temp_files", "0")
	if err != nil {
		return fmt.Errorf("CheckParameter: %w", err)
	}
	err = rds.CheckParameter("log_autovacuum_min_duration", "0")
	if err != nil {
		return fmt.Errorf("CheckParameter: %w", err)
	}
	err = rds.CheckParameter("log_min_duration_statement", "0")
	if err != nil {
		return fmt.Errorf("CheckParameter: %w", err)
	}
	err = rds.CheckParameter("log_duration", "on")
	if err != nil {
		return fmt.Errorf("CheckParameter: %w", err)
	}
	err = rds.CheckParameter("log_checkpoints", "on")
	if err != nil {
		return fmt.Errorf("CheckParameter: %w", err)
	}
	err = rds.CheckParameter("log_statement", "on")
	if err != nil {
		return fmt.Errorf("CheckParameter: %w", err)
	}
	err = rds.CheckParameter("track_io_timing", "on")
	if err != nil {
		return fmt.Errorf("CheckParameter: %w", err)
	}

	_, err = rds.dbParameterGroups.GetParameterValueByParameterName("track_activity_query_size")
	if err != nil {
		return fmt.Errorf("GetParameterValueByParameterName: %w", err)
	}

	return nil
}

// Creating graphs
func readCSV(file io.Reader) (plotter.XYs, string, error) {
	reader := csv.NewReader(file)
	reader.Comma = ';'
	reader.TrimLeadingSpace = true

	lines, err := reader.ReadAll()
	if err != nil {
		return nil, "", err
	}

	yAxisLabel := lines[0][1]

	// Trier les lignes en fonction de la colonne Timestamp
	sort.Slice(lines[1:], func(i, j int) bool {
		return lines[1+i][0] < lines[1+j][0]
	})

	layout := "2006-01-02 15:04:05 -0700 MST"

	interval := 5 * time.Minute

	avgData := make(map[int64]float64)
	avgCount := make(map[int64]int)

	for _, line := range lines[1:] {
		t, err := time.Parse(layout, line[0])
		if err != nil {
			return nil, "", err
		}
		y, err := strconv.ParseFloat(strings.TrimSpace(line[1]), 64)
		if err != nil {
			return nil, "", err
		}

		intervalStart := t.Truncate(interval).Unix()

		avgData[intervalStart] += y
		avgCount[intervalStart]++
	}

	data := make(plotter.XYs, len(avgData))
	i := 0
	for k, v := range avgData {
		data[i].X = float64(k)
		data[i].Y = v / float64(avgCount[k])
		i++
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].X < data[j].X
	})

	return data, yAxisLabel, nil
}

func writeHTML(file *os.File, content string, step string) error {
	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("file.WriteString: %w", err)
	}

	return nil
}

func createMetricsHTML(outputFilenames []string, instanceIdentifier string) error {
	htmlFile, err := os.Create(fmt.Sprintf("%s.html", instanceIdentifier))
	if err != nil {
		return err
	}
	defer func() { _ = htmlFile.Close() }()

	// Écrire l'en-tête HTML
	err = writeHTML(htmlFile, "<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n", "début HTML")
	if err != nil {
		return fmt.Errorf("writeHTML: %w", err)
	}

	err = writeHTML(htmlFile, "<meta charset=\"UTF-8\">\n<meta name=\"viewport\" content=\"width=device-width, initial-scale=1, shrink-to-fit=no\">\n", "meta tags")
	if err != nil {
		return fmt.Errorf("writeHTML: %w", err)
	}

	err = writeHTML(htmlFile, "<link rel=\"stylesheet\" href=\"https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/css/bootstrap.min.css\" integrity=\"sha384-ggOyR0iXCbMQv3Xipma34MD+dH/1fQ784/j6cY/iJTQUOhcWr7x9JvoRxT2MZw1T\" crossorigin=\"anonymous\">\n", "Bootstrap CSS")
	if err != nil {
		return fmt.Errorf("writeHTML: %w", err)
	}

	err = writeHTML(htmlFile, "<title>Metrics Viewer</title>\n</head>\n<body>\n", "titre HTML")
	if err != nil {
		return fmt.Errorf("writeHTML: %w", err)
	}

	err = writeHTML(htmlFile, "<div class=\"container mt-4\">\n<h2>Metrics</h2>\n", "conteneur principal")
	if err != nil {
		return fmt.Errorf("writeHTML: %w", err)
	}

	// Parcourir les catégories de métriques
	categories := make(map[string]bool)
	for _, filename := range outputFilenames {
		category := getCategory(filename)
		categories[category] = true
	}

	// Générer le contenu pour chaque catégorie
	for category := range categories {
		err = writeHTML(htmlFile, fmt.Sprintf("<div class=\"mt-3\" id=\"%s\">\n", category), "début catégorie")
		if err != nil {
			return fmt.Errorf("writeHTML: %w", err)
		}

		err = writeHTML(htmlFile, "<div class=\"card card-body\">\n", "conteneur carte")
		if err != nil {
			return fmt.Errorf("writeHTML: %w", err)
		}

		err = writeHTML(htmlFile, fmt.Sprintf("<h5 class=\"card-title\">%s</h5>\n", getCategoryTitle(category)), "titre catégorie")
		if err != nil {
			return fmt.Errorf("writeHTML: %w", err)
		}

		// Afficher toutes les images pour la catégorie
		for _, filename := range outputFilenames {
			if strings.Contains(filename, "Average") {
				err = writeHTML(htmlFile, fmt.Sprintf("<img src=\"%s\" class=\"img-fluid\" alt=\"%s\">\n", filename, filepath.Base(filename)), "image catégorie")
				if err != nil {
					return fmt.Errorf("writeHTML: %w", err)
				}
			}
		}

		err = writeHTML(htmlFile, "</div>\n</div>\n", "fermeture div catégorie")
		if err != nil {
			return fmt.Errorf("writeHTML: %w", err)
		}
	}

	// Bouton Collapse pour les autres metrics
	err = writeHTML(htmlFile, "<button class=\"btn btn-secondary mt-3\" type=\"button\" data-toggle=\"collapse\" data-target=\"#otherMetrics\" aria-expanded=\"false\" aria-controls=\"otherMetrics\">\n", "bouton collapse")
	if err != nil {
		return fmt.Errorf("writeHTML: %w", err)
	}

	err = writeHTML(htmlFile, "Afficher les autres Metrics\n", "texte bouton collapse")
	if err != nil {
		return fmt.Errorf("writeHTML: %w", err)
	}

	err = writeHTML(htmlFile, "</button>\n", "fermeture bouton collapse")
	if err != nil {
		return fmt.Errorf("writeHTML: %w", err)
	}

	// Contenu collapse pour les autres Metrics
	err = writeHTML(htmlFile, "<div class=\"collapse mt-3\" id=\"otherMetrics\">\n", "début collapse")
	if err != nil {
		return fmt.Errorf("writeHTML: %w", err)
	}

	err = writeHTML(htmlFile, "<div class=\"card card-body\">\n", "conteneur collapse")
	if err != nil {
		return fmt.Errorf("writeHTML: %w", err)
	}

	// Afficher toutes les images ici
	for _, filename := range outputFilenames {
		err = writeHTML(htmlFile, fmt.Sprintf("<img src=\"%s\" class=\"img-fluid\" alt=\"%s\">\n", filename, filepath.Base(filename)), "image collapse")
		if err != nil {
			return fmt.Errorf("writeHTML: %w", err)
		}
	}

	err = writeHTML(htmlFile, "</div>\n</div>\n", "fermeture collapse")
	if err != nil {
		return fmt.Errorf("writeHTML: %w", err)
	}

	// Fermer le corps et l'HTML
	err = writeHTML(htmlFile, "<script src=\"https://code.jquery.com/jquery-3.3.1.slim.min.js\" integrity=\"sha384-q8i/X+965DzO0rT7abK41JStQIAqVgRVzpbzo5smXKp4YfRvH+8abtTE1Pi6jizo\" crossorigin=\"anonymous\"></script>\n", "script jQuery")
	if err != nil {
		return fmt.Errorf("writeHTML: %w", err)
	}

	err = writeHTML(htmlFile, "<script src=\"https://cdn.jsdelivr.net/npm/@popperjs/core@2.10.2/dist/umd/popper.min.js\" integrity=\"sha384-ggOyR0iXCbMQv3Xipma34MD+dH/1fQ784/j6cY/iJTQUOhcWr7x9JvoRxT2MZw1T\" crossorigin=\"anonymous\"></script>\n", "script Popper.js")
	if err != nil {
		return fmt.Errorf("writeHTML: %w", err)
	}

	err = writeHTML(htmlFile, "<script src=\"https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/js/bootstrap.min.js\" integrity=\"sha384-JjSmVgyd0p3pXB1rRibZUAYoIIy6OrQ6VrjIEaFf/nJGzIxFDsf4x0xIM+B07jRM\" crossorigin=\"anonymous\"></script>\n", "script Bootstrap")
	if err != nil {
		return fmt.Errorf("writeHTML: %w", err)
	}

	err = writeHTML(htmlFile, "</body>\n</html>\n", "fermeture HTML")
	if err != nil {
		return fmt.Errorf("writeHTML: %w", err)
	}

	return nil
}

// getCategory extrait la catégorie à partir du nom du fichier
func getCategory(filename string) string {
	parts := strings.Split(filename, "/")
	if len(parts) >= 4 {
		return parts[3]
	}
	return ""
}

// getCategoryTitle génère un titre lisible à partir de la catégorie
func getCategoryTitle(category string) string {
	// Vous pouvez personnaliser cette fonction pour obtenir un titre plus convivial
	// En fonction de la catégorie fournie
	return category + " Metrics"
}

func createPNGGraph(data plotter.XYs, yAxisLabel, outputFilename string) error {
	p := plot.New()

	baseFilename := filepath.Base(outputFilename)
	p.Title.Text = "Graphe des données CSV : " + baseFilename
	p.X.Label.Text = "Timestamp"
	p.Y.Label.Text = yAxisLabel

	plotter.DefaultGlyphStyle.Radius = vg.Points(2)

	line, err := plotter.NewLine(data)
	if err != nil {
		return err
	}
	//  points.Shape = draw.CrossGlyph{}
	line.Color = color.RGBA{R: 255, B: 255, A: 255}

	p.Add(line)

	p.X.Tick.Marker = plot.TimeTicks{Format: "2006-01-02\n15:04:05"}
	p.X.Tick.Label.Rotation = -math.Pi / 2
	p.X.Tick.Label.XAlign = draw.XRight
	p.X.Tick.Label.YAlign = draw.YCenter

	err = p.Save(10*vg.Inch, 5*vg.Inch, outputFilename)
	if err != nil {
		return err
	}

	return nil
}

func (rds RDS) GetAllParameterNames() ([]string, error) {
	var parameterNames []string
	for _, parameter := range rds.dbParameterGroups.Parameters {
		parameterNames = append(parameterNames, parameter.ParameterName)
	}
	return parameterNames, nil
}
