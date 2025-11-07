package gcp

import (
	"time"
)

type MetricDescriptor struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	Unit        string `json:"unit"`
	ValueType   string `json:"valueType"`
	MetricKind  string `json:"metricKind"`
}

type TimeSeriesPoint struct {
	Interval TimeInterval `json:"interval"`
	Value    PointValue   `json:"value"`
}

type TimeInterval struct {
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

type PointValue struct {
	DoubleValue *float64 `json:"doubleValue,omitempty"`
	Int64Value  *int64   `json:"int64Value,omitempty"`
	BoolValue   *bool    `json:"boolValue,omitempty"`
	StringValue *string  `json:"stringValue,omitempty"`
}

type TimeSeries struct {
	Metric     Metric            `json:"metric"`
	Resource   MonitoredResource `json:"resource"`
	MetricKind string            `json:"metricKind"`
	ValueType  string            `json:"valueType"`
	Points     []TimeSeriesPoint `json:"points"`
	Unit       string            `json:"unit"`
}

type Metric struct {
	Type   string            `json:"type"`
	Labels map[string]string `json:"labels,omitempty"`
}

type MonitoredResource struct {
	Type   string            `json:"type"`
	Labels map[string]string `json:"labels"`
}

type ListTimeSeriesResult struct {
	TimeSeries      []TimeSeries `json:"timeSeries"`
	NextPageToken   string       `json:"nextPageToken,omitempty"`
	ExecutionErrors []string     `json:"executionErrors,omitempty"`
}

// GetField récupère la valeur d'un point de données
func (p *TimeSeriesPoint) GetValue() float64 {
	if p.Value.DoubleValue != nil {
		return *p.Value.DoubleValue
	}
	if p.Value.Int64Value != nil {
		return float64(*p.Value.Int64Value)
	}
	return 0.0
}
