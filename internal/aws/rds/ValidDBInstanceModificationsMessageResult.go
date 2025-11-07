package rds

import (
	"fmt"

	"golang.org/x/exp/slices"
)

type Storage struct {
	StorageType string `json:"StorageType"`
	StorageSize []struct {
		From int `json:"From"`
		To   int `json:"To"`
		Step int `json:"Step"`
	} `json:"StorageSize"`
	ProvisionedIops []struct {
		From int `json:"From"`
		To   int `json:"To"`
		Step int `json:"Step"`
	} `json:"ProvisionedIops"`
	IopsToStorageRatio []struct {
		From float64 `json:"From"`
		To   float64 `json:"To"`
	} `json:"IopsToStorageRatio"`
	SupportsStorageAutoscaling bool `json:"SupportsStorageAutoscaling"`
}

type ValidDBInstanceModificationsMessage struct {
	Storage                []Storage     `json:"Storage"`
	ValidProcessorFeatures []interface{} `json:"ValidProcessorFeatures"`
}

type ValidDBInstanceModificationsMessageResult struct {
	ValidDBInstanceModificationsMessage ValidDBInstanceModificationsMessage `json:"ValidDBInstanceModificationsMessage"`
}

func (d *ValidDBInstanceModificationsMessage) GetSupportsStorageAutoscalingByStorageType(storageType string) (bool, error) {
	idx := slices.IndexFunc(d.Storage, func(c Storage) bool { return c.StorageType == storageType })
	if idx == -1 {
		return false, fmt.Errorf("storage type %s not found", storageType)
	}
	return d.Storage[idx].SupportsStorageAutoscaling, nil
}
