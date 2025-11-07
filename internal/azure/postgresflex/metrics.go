package postgresflex

import (
	"fmt"
	"time"
)

type MetricValue struct {
	Timestamp time.Time `json:"timeStamp"`
	Average   *float64  `json:"average,omitempty"`
	Minimum   *float64  `json:"minimum,omitempty"`
	Maximum   *float64  `json:"maximum,omitempty"`
	Total     *float64  `json:"total,omitempty"`
	Count     *float64  `json:"count,omitempty"`
}

type Timeseries struct {
	Data []MetricValue `json:"data"`
}

type Metric struct {
	ID         string       `json:"id"`
	Type       string       `json:"type"`
	Name       MetricName   `json:"name"`
	Unit       string       `json:"unit"`
	Timeseries []Timeseries `json:"timeseries"`
}

type MetricName struct {
	Value          string `json:"value"`
	LocalizedValue string `json:"localizedValue"`
}

type MetricsResult struct {
	Cost           float64  `json:"cost,omitempty"`
	Timespan       string   `json:"timespan"`
	Interval       string   `json:"interval,omitempty"`
	Value          []Metric `json:"value"`
	Namespace      string   `json:"namespace,omitempty"`
	ResourceRegion string   `json:"resourceregion,omitempty"`
}

// GetField récupère une valeur spécifique d'un MetricValue
func (m *MetricValue) GetField(field string) string {
	switch field {
	case "Average":
		if m.Average != nil {
			return fmt.Sprintf("%f", *m.Average)
		}
	case "Minimum":
		if m.Minimum != nil {
			return fmt.Sprintf("%f", *m.Minimum)
		}
	case "Maximum":
		if m.Maximum != nil {
			return fmt.Sprintf("%f", *m.Maximum)
		}
	case "Total":
		if m.Total != nil {
			return fmt.Sprintf("%f", *m.Total)
		}
	case "Count":
		if m.Count != nil {
			return fmt.Sprintf("%f", *m.Count)
		}
	}
	return "0"
}
