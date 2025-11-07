package gcp

type FlagMetadata struct {
	Name            string   `json:"name"`
	Type            string   `json:"type"`
	AppliesTo       []string `json:"appliesTo"`
	AllowedValues   []string `json:"allowedValues,omitempty"`
	MinValue        *int64   `json:"minValue,omitempty"`
	MaxValue        *int64   `json:"maxValue,omitempty"`
	RequiresRestart bool     `json:"requiresRestart"`
}

type DescribeFlagsResult struct {
	Items []FlagMetadata `json:"items"`
}

func (d *DescribeFlagsResult) GetFlagValueByName(flagName string, instance *DatabaseInstance) (string, error) {
	if instance == nil || instance.Settings.DatabaseFlags == nil {
		return "", nil
	}

	for _, flag := range instance.Settings.DatabaseFlags {
		if flag.Name == flagName {
			return flag.Value, nil
		}
	}

	return "", nil
}
