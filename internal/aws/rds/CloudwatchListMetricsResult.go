package rds

type ListMetricsResult struct {
	Metrics []struct {
		Namespace  string `json:"Namespace"`
		MetricName string `json:"MetricName"`
		Dimensions []struct {
			Name  string `json:"Name"`
			Value string `json:"Value"`
		} `json:"Dimensions"`
	} `json:"Metrics"`
}
