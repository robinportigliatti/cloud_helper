package rds

import (
	"fmt"
	"reflect"
	"time"
)

type GetMetricsStatisticsResult struct {
	Label      string      `json:"Label"`
	Datapoints []Datapoint `json:"Datapoints"`
}

type Datapoint struct {
	Timestamp   time.Time `json:"Timestamp"`
	SampleCount float64   `json:"SampleCount"`
	Average     float64   `json:"Average"`
	Sum         float64   `json:"Sum"`
	Minimum     float64   `json:"Minimum"`
	Maximum     float64   `json:"Maximum"`
	Unit        string    `json:"Unit"`
}

func (v *Datapoint) GetField(field string) string {
	r := reflect.ValueOf(v)
	f := reflect.Indirect(r).FieldByName(field)
	return fmt.Sprintf("%f", f.Float())
}
