package collector

const (
	StateEnabled = "Enabled"
	StateAbsent  = "Absent"
)

// Session
type Session struct {
	Id          string `json:"Id,omitempty"`
	Name        string `json:"Name,omitempty"`
	Username    string `json:"UserName,omitempty"`
	Password    string `json:"Password,omitempty"`
	CreatedTime string `json:"CreatedTime,omitempty"`
	SessionType string `json:"SessionType,omitempty"`
	OdataId     string `json:"@odata.id,omitempty"`
}

// Odata is a common structure to unmarshal Open Data Protocol metadata
type Odata struct {
	OdataContext string `json:"@odata.context"`
	OdataId      string `json:"@odata.id"`
	OdataType    string `json:"@odata.type"`
}

type OdataSlice []Odata

func (m *OdataSlice) GetLinks() []string {
	list := []string{}
	seen := map[string]bool{}

	for _, c := range *m {
		s := c.OdataId
		if ok := seen[s]; !ok {
			seen[s] = true
			list = append(list, s)
		}
	}

	return list
}

// Status is a common structure used in any entity with a status
type Status struct {
	Health       string `json:"Health"`
	HealthRollup string `json:"HealthRollup"`
	State        string `json:"State"`
}

// V1Response represents structure of the response body from /redfish/v1
type V1Response struct {
	RedfishVersion     string `json:"RedfishVersion"`
	Name               string `json:"Name"`
	Product            string `json:"Product"`
	Vendor             string `json:"Vendor"`
	Description        string `json:"Description"`
	AccountService     Odata  `json:"AccountService"`
	CertificateService Odata  `json:"CertificateService"`
	Chassis            Odata  `json:"Chassis"`
	EventService       Odata  `json:"EventService"`
	Fabrics            Odata  `json:"Fabrics"`
	JobService         Odata  `json:"JobService"`
	JsonSchemas        Odata  `json:"JsonSchemas"`
	Managers           Odata  `json:"Managers"`
	Registries         Odata  `json:"Registries"`
	SessionService     Odata  `json:"SessionService"`
	Systems            Odata  `json:"Systems"`
	Tasks              Odata  `json:"Tasks"`
	TelemetryService   Odata  `json:"TelemetryService"`
	UpdateService      Odata  `json:"UpdateService"`
}

type GroupResponse struct {
	Name        string     `json:"Name"`
	Description string     `json:"Description"`
	Members     OdataSlice `json:"Members"`
}

type Processor struct {
	Id                    string  `json:"Id"`
	Name                  string  `json:"Name"`
	Description           string  `json:"Description"`
	InstructionSet        xstring `json:"InstructionSet"`
	Manufacturer          string  `json:"Manufacturer"`
	MaxSpeedMHz           *int    `json:"MaxSpeedMHz"`
	Model                 string  `json:"Model"`
	Family                string  `json:"Family"`
	OperatingSpeedMHz     *int    `json:"OperatingSpeedMHz"`
	PartNumber            string  `json:"PartNumber"`
	ProcessorArchitecture xstring `json:"ProcessorArchitecture"`
	ProcessorId           struct {
		EffectiveFamily               string `json:"EffectiveFamily"`
		EffectiveModel                string `json:"EffectiveModel"`
		IdentificationRegisters       string `json:"IdentificationRegisters"`
		MicrocodeInfo                 string `json:"MicrocodeInfo"`
		ProtectedIdentificationNumber string `json:"ProtectedIdentificationNumber"`
		Step                          string `json:"Step"`
		VendorID                      string `json:"VendorId"`
	} `json:"ProcessorId"`
	ProcessorType     string  `json:"ProcessorType"`
	Socket            string  `json:"Socket"`
	Status            Status  `json:"Status"`
	TDPWatts          float64 `json:"TDPWatts"`
	TotalCores        int     `json:"TotalCores"`
	TotalEnabledCores int     `json:"TotalEnabledCores"`
	TotalThreads      int     `json:"TotalThreads"`
	TurboState        string  `json:"TurboState"`
	Version           string  `json:"Version"`
	Oem               struct {
		Lenovo *struct {
			CurrentClockSpeedMHz int `json:"CurrentClockSpeedMHz"`
		} `json:"Lenovo"`
		Hpe *struct {
			VoltageVoltsX10 int `json:"VoltageVoltsX10"`
		} `json:"Hpe"`
		Dell *struct {
			DellProcessor struct {
				Volts string `json:"Volts"`
			} `json:"DellProcessor"`
		} `json:"Dell"`
	} `json:"Oem"`
}

type GPU struct {
	Id                    string  `json:"Id"`
	Name                  string  `json:"Name"`
	Description           string  `json:"Description"`
	Manufacturer          string  `json:"Manufacturer"`
	Model                 string  `json:"Model"`
	PartNumber            string  `json:"PartNumber"`
	Metrics               Odata  `json:"Metrics"`
	MemorySummary         struct {
        Metrics           Odata `json:"Metrics"`
	} `json:"MemorySummary"`
	ProcessorType     string  `json:"ProcessorType"`
	Status            Status  `json:"Status"`
}

type DellVideoMember struct {
	Id		     string  `json:"Id"`
	GPUGUID	     string  `json:"GPUGUID"`
	GPUHealth 	 string  `json:"GPUHealth"`
	GPUState 	 string  `json:"GPUState"`
	SerialNumber string  `json:"SerialNumber"`
}

type DellVideo struct {
	Members []DellVideoMember `json:"Members"`
}

type DellGPUSensorMember struct {
	Id		     string  `json:"Id"`
	BoardPowerSupplyStatus	 string  `json:"BoardPowerSupplyStatus"`
	MemoryTemperatureCelsius float64 `json:"MemoryTemperatureCelsius"`
	PowerBrakeStatus 	 string  `json:"PowerBrakeStatus"`
	PrimaryGPUTemperatureCelsius float64 `json:"PrimaryGPUTemperatureCelsius"`
	ThermalAlertStatus	 string  `json:"ThermalAlertStatus"`
}

type DellGPUSensors struct {
	Members []DellGPUSensorMember `json:"Members"`
}

type GPUMetrics struct {
	Id                    string  `json:"Id"`
    TemperatureCelsius	  float64 `json:"TemperatureCelsius"`
    ConsumedPowerWatt 	  float64 `json:"ConsumedPowerWatt"`
    OperatingSpeedMHz	  *float64 `json:"OperatingSpeedMHz"`
    BandwidthPercent      *float64 `json:"BandwidthPercent"`
	Oem				   *struct {
		Nvidia *struct {
			ThrottleReasons			   []string `json:"ThrottleReasons"`
			SMUtilizationPercent	   int      `json:"SMUtilizationPercent"`
			SMActivityPercent		   float64  `json:"SMActivityPercent"`
			SMOccupancyPercent		   float64  `json:"SMOccupancyPercent"`
			TensorCoreActivityPercent   float64  `json:"TensorCoreActivityPercent"`
			HMMAUtilizationPercent	   float64  `json:"HMMAUtilizationPercent"`
			PCIeRawTxBandwidthGbps	   float64  `json:"PCIeRawTxBandwidthGbps"`
			PCIeRawRxBandwidthGbps	   float64  `json:"PCIeRawRxBandwidthGbps"`
		} `json:"Nvidia"`
		Dell *struct {
			CurrentPCIeLinkSpeed     int     `json:"CurrentPCIeLinkSpeed"`
			MaxSupportedPCIeLinkSpeed int     `json:"MaxSupportedPCIeLinkSpeed"`
			DRAMUtilizationPercent    float64 `json:"DRAMUtilizationPercent"`
		} `json:"Dell"`
	} `json:"Oem"`
	PCIeErrors *struct {
		CorrectableErrorCount int `json:"CorrectableErrorCount"`
	} `json:"PCIeErrors"`
}
 

type GPUMemoryMetrics struct {
    BandwidthPercent	  float64 `json:"BandwidthPercent"`
    OperatingSpeedMHz 	  float64 `json:"OperatingSpeedMHz"`
}

type SystemResponse struct {
	IndicatorLED            string `json:"IndicatorLED"`
	LocationIndicatorActive *bool  `json:"LocationIndicatorActive"`
	Manufacturer            string `json:"Manufacturer"`
	AssetTag                string `json:"AssetTag"`
	PartNumber              string `json:"PartNumber"`
	Description             string `json:"Description"`
	HostName                string `json:"HostName"`
	PowerState              string `json:"PowerState"`
	Bios                    Odata  `json:"Bios"`
	BiosVersion             string `json:"BiosVersion"`
	Boot                    *struct {
		BootOptions                                    Odata    `json:"BootOptions"`
		Certificates                                   Odata    `json:"Certificates"`
		BootOrder                                      []string `json:"BootOrder"`
		BootSourceOverrideEnabled                      string   `json:"BootSourceOverrideEnabled"`
		BootSourceOverrideMode                         string   `json:"BootSourceOverrideMode"`
		BootSourceOverrideTarget                       string   `json:"BootSourceOverrideTarget"`
		UefiTargetBootSourceOverride                   any      `json:"UefiTargetBootSourceOverride"`
		BootSourceOverrideTargetRedfishAllowableValues []string `json:"BootSourceOverrideTarget@Redfish.AllowableValues"`
	} `json:"Boot"`
	EthernetInterfaces Odata `json:"EthernetInterfaces"`
	HostWatchdogTimer  *struct {
		FunctionEnabled bool   `json:"FunctionEnabled"`
		Status          Status `json:"Status"`
		TimeoutAction   string `json:"TimeoutAction"`
	} `json:"HostWatchdogTimer"`
	HostingRoles  []any `json:"HostingRoles"`
	Memory        Odata `json:"Memory"`
	MemorySummary *struct {
		MemoryMirroring      string  `json:"MemoryMirroring"`
		Status               Status  `json:"Status"`
		TotalSystemMemoryGiB float64 `json:"TotalSystemMemoryGiB"`
	} `json:"MemorySummary"`
	Model             string     `json:"Model"`
	Name              string     `json:"Name"`
	NetworkInterfaces Odata      `json:"NetworkInterfaces"`
	PCIeDevices       OdataSlice `json:"PCIeDevices"`
	PCIeFunctions     OdataSlice `json:"PCIeFunctions"`
	ProcessorSummary  *struct {
		Count                 int    `json:"Count"`
		LogicalProcessorCount int    `json:"LogicalProcessorCount"`
		Model                 string `json:"Model"`
		Status                Status `json:"Status"`
	} `json:"ProcessorSummary"`
	Processors     Odata  `json:"Processors"`
	SKU            string `json:"SKU"`
	SecureBoot     Odata  `json:"SecureBoot"`
	SerialNumber   string `json:"SerialNumber"`
	SimpleStorage  Odata  `json:"SimpleStorage"`
	Status         Status `json:"Status"`
	Storage        Odata  `json:"Storage"`
	SystemType     string `json:"SystemType"`
	TrustedModules []struct {
		FirmwareVersion string `json:"FirmwareVersion"`
		InterfaceType   string `json:"InterfaceType"`
		Status          Status `json:"Status"`
	} `json:"TrustedModules"`
	Oem struct {
		Hpe struct {
			IndicatorLED string `json:"IndicatorLED"`
		} `json:"Hpe"`
	} `json:"Oem"`
}
