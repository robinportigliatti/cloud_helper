package gcp

type MachineType struct {
	Name                         string `json:"name"`
	GuestCpus                    int    `json:"guestCpus"`
	MemoryMb                     int    `json:"memoryMb"`
	ImageSpaceGb                 int    `json:"imageSpaceGb,omitempty"`
	MaximumPersistentDisks       int    `json:"maximumPersistentDisks,omitempty"`
	MaximumPersistentDisksSizeGb string `json:"maximumPersistentDisksSizeGb,omitempty"`
	Zone                         string `json:"zone,omitempty"`
	SelfLink                     string `json:"selfLink"`
	IsSharedCpu                  bool   `json:"isSharedCpu,omitempty"`
	Description                  string `json:"description,omitempty"`
}

type MachineTypeList struct {
	Items []MachineType `json:"items"`
}
