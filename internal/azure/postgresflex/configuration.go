package postgresflex

type Configuration struct {
	ID         string                  `json:"id"`
	Name       string                  `json:"name"`
	Type       string                  `json:"type"`
	Properties ConfigurationProperties `json:"properties"`
}

type ConfigurationProperties struct {
	Value                  string `json:"value"`
	Description            string `json:"description,omitempty"`
	DefaultValue           string `json:"defaultValue,omitempty"`
	DataType               string `json:"dataType,omitempty"`
	AllowedValues          string `json:"allowedValues,omitempty"`
	Source                 string `json:"source,omitempty"`
	IsDynamicConfig        bool   `json:"isDynamicConfig,omitempty"`
	IsReadOnly             bool   `json:"isReadOnly,omitempty"`
	IsConfigPendingRestart bool   `json:"isConfigPendingRestart,omitempty"`
}

type ConfigurationListResult struct {
	Value []Configuration `json:"value"`
}

func (c *ConfigurationListResult) GetValueByName(name string) (string, error) {
	for _, config := range c.Value {
		if config.Name == name {
			return config.Properties.Value, nil
		}
	}
	return "", nil
}
