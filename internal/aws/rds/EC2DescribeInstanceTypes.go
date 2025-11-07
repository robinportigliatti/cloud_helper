package rds

type MemoryInfo struct {
	SizeInMiB int `json:"SizeInMiB"`
}

type InstanceType struct {
	InstanceType                 string   `json:"InstanceType"`
	CurrentGeneration            bool     `json:"CurrentGeneration"`
	FreeTierEligible             bool     `json:"FreeTierEligible"`
	SupportedUsageClasses        []string `json:"SupportedUsageClasses"`
	SupportedRootDeviceTypes     []string `json:"SupportedRootDeviceTypes"`
	SupportedVirtualizationTypes []string `json:"SupportedVirtualizationTypes"`
	BareMetal                    bool     `json:"BareMetal"`
	Hypervisor                   string   `json:"Hypervisor"`
	ProcessorInfo                struct {
		SupportedArchitectures   []string `json:"SupportedArchitectures"`
		SustainedClockSpeedInGhz float64  `json:"SustainedClockSpeedInGhz"`
	} `json:"ProcessorInfo"`
	VCPUInfo struct {
		DefaultVCpus          int `json:"DefaultVCpus"`
		DefaultCores          int `json:"DefaultCores"`
		DefaultThreadsPerCore int `json:"DefaultThreadsPerCore"`
	} `json:"VCpuInfo"`
	MemoryInfo               MemoryInfo `json:"MemoryInfo"`
	InstanceStorageSupported bool       `json:"InstanceStorageSupported"`
	EbsInfo                  struct {
		EbsOptimizedSupport string `json:"EbsOptimizedSupport"`
		EncryptionSupport   string `json:"EncryptionSupport"`
		NvmeSupport         string `json:"NvmeSupport"`
	} `json:"EbsInfo"`
	NetworkInfo struct {
		NetworkPerformance       string `json:"NetworkPerformance"`
		MaximumNetworkInterfaces int    `json:"MaximumNetworkInterfaces"`
		MaximumNetworkCards      int    `json:"MaximumNetworkCards"`
		DefaultNetworkCardIndex  int    `json:"DefaultNetworkCardIndex"`
		NetworkCards             []struct {
			NetworkCardIndex         int    `json:"NetworkCardIndex"`
			NetworkPerformance       string `json:"NetworkPerformance"`
			MaximumNetworkInterfaces int    `json:"MaximumNetworkInterfaces"`
		} `json:"NetworkCards"`
		Ipv4AddressesPerInterface    int    `json:"Ipv4AddressesPerInterface"`
		Ipv6AddressesPerInterface    int    `json:"Ipv6AddressesPerInterface"`
		Ipv6Supported                bool   `json:"Ipv6Supported"`
		EnaSupport                   string `json:"EnaSupport"`
		EfaSupported                 bool   `json:"EfaSupported"`
		EncryptionInTransitSupported bool   `json:"EncryptionInTransitSupported"`
	} `json:"NetworkInfo"`
	PlacementGroupInfo struct {
		SupportedStrategies []string `json:"SupportedStrategies"`
	} `json:"PlacementGroupInfo"`
	HibernationSupported          bool     `json:"HibernationSupported"`
	BurstablePerformanceSupported bool     `json:"BurstablePerformanceSupported"`
	DedicatedHostsSupported       bool     `json:"DedicatedHostsSupported"`
	AutoRecoverySupported         bool     `json:"AutoRecoverySupported"`
	SupportedBootModes            []string `json:"SupportedBootModes"`
}

type DescribeInstanceTypes struct {
	InstanceTypes []InstanceType `json:"InstanceTypes"`
}
