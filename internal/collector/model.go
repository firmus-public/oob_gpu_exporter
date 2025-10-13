package collector

import "strconv"

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

// Redundancy is a common structure used in any entity with redundancy
type Redundancy struct {
	Name              string  `json:"Name"`
	MaxNumSupported   int     `json:"MaxNumSupported"`
	MinNumNeeded      int     `json:"MinNumNeeded"`
	Mode              xstring `json:"Mode"`
	RedundancyEnabled bool    `json:"RedundancyEnabled"`
	RedundancySet     []any   `json:"RedundancySet"`
	Status            Status  `json:"Status"`
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
	Id            string `json:"Id"`
	Name          string `json:"Name"`
	Description   string `json:"Description"`
	Manufacturer  string `json:"Manufacturer"`
	Model         string `json:"Model"`
	PartNumber    string `json:"PartNumber"`
	Metrics       Odata  `json:"Metrics"`
	MemorySummary struct {
		Metrics Odata `json:"Metrics"`
	} `json:"MemorySummary"`
	ProcessorType string `json:"ProcessorType"`
	Status        Status `json:"Status"`
}

type DellVideoMember struct {
	Id           string `json:"Id"`
	GPUGUID      string `json:"GPUGUID"`
	GPUHealth    string `json:"GPUHealth"`
	GPUState     string `json:"GPUState"`
	SerialNumber string `json:"SerialNumber"`
}

type DellVideo struct {
	Members []DellVideoMember `json:"Members"`
}

type DellGPUSensorMember struct {
	Id                           string  `json:"Id"`
	BoardPowerSupplyStatus       string  `json:"BoardPowerSupplyStatus"`
	MemoryTemperatureCelsius     float64 `json:"MemoryTemperatureCelsius"`
	PowerBrakeStatus             string  `json:"PowerBrakeStatus"`
	PrimaryGPUTemperatureCelsius float64 `json:"PrimaryGPUTemperatureCelsius"`
	ThermalAlertStatus           string  `json:"ThermalAlertStatus"`
}

type DellGPUSensors struct {
	Members []DellGPUSensorMember `json:"Members"`
}

type GPUMetrics struct {
	Id                 string   `json:"Id"`
	TemperatureCelsius float64  `json:"TemperatureCelsius"`
	ConsumedPowerWatt  float64  `json:"ConsumedPowerWatt"`
	OperatingSpeedMHz  *float64 `json:"OperatingSpeedMHz"`
	BandwidthPercent   *float64 `json:"BandwidthPercent"`
	Oem                *struct {
		Nvidia *struct {
			ThrottleReasons           []string `json:"ThrottleReasons"`
			SMUtilizationPercent      int      `json:"SMUtilizationPercent"`
			SMActivityPercent         float64  `json:"SMActivityPercent"`
			SMOccupancyPercent        float64  `json:"SMOccupancyPercent"`
			TensorCoreActivityPercent float64  `json:"TensorCoreActivityPercent"`
			HMMAUtilizationPercent    float64  `json:"HMMAUtilizationPercent"`
			PCIeRawTxBandwidthGbps    float64  `json:"PCIeRawTxBandwidthGbps"`
			PCIeRawRxBandwidthGbps    float64  `json:"PCIeRawRxBandwidthGbps"`
		} `json:"Nvidia"`
		Dell *struct {
			CurrentPCIeLinkSpeed      int     `json:"CurrentPCIeLinkSpeed"`
			MaxSupportedPCIeLinkSpeed int     `json:"MaxSupportedPCIeLinkSpeed"`
			DRAMUtilizationPercent    float64 `json:"DRAMUtilizationPercent"`
		} `json:"Dell"`
	} `json:"Oem"`
	PCIeErrors *struct {
		CorrectableErrorCount int `json:"CorrectableErrorCount"`
	} `json:"PCIeErrors"`
}
 

type GPUMemoryMetrics struct {
	BandwidthPercent  float64 `json:"BandwidthPercent"`
	OperatingSpeedMHz float64 `json:"OperatingSpeedMHz"`
}

type ChassisResponse struct {
	Name                    string `json:"Name"`
	AssetTag                string `json:"AssetTag"`
	SerialNumber            string `json:"SerialNumber"`
	PartNumber              string `json:"PartNumber"`
	Model                   string `json:"Model"`
	ChassisType             string `json:"ChassisType"`
	Manufacturer            string `json:"Manufacturer"`
	Description             string `json:"Description"`
	SKU                     string `json:"SKU"`
	PowerState              string `json:"PowerState"`
	EnvironmentalClass      string `json:"EnvironmentalClass"`
	IndicatorLED            string `json:"IndicatorLED"`
	LocationIndicatorActive *bool  `json:"LocationIndicatorActive"`
	Assembly                Odata  `json:"Assembly"`
	Location                *struct {
		Info       string `json:"Info"`
		InfoFormat string `json:"InfoFormat"`
		Placement  struct {
			Rack string `json:"Rack"`
			Row  string `json:"Row"`
		} `json:"Placement"`
		PostalAddress struct {
			Building string `json:"Building"`
			Room     string `json:"Room"`
		} `json:"PostalAddress"`
	} `json:"Location"`
	Memory           Odata  `json:"Memory"`
	NetworkAdapters  Odata  `json:"NetworkAdapters"`
	PCIeDevices      Odata  `json:"PCIeDevices"`
	PCIeSlots        Odata  `json:"PCIeSlots"`
	Power            Odata  `json:"Power"`
	Sensors          Odata  `json:"Sensors"`
	Status           Status `json:"Status"`
	Thermal          Odata  `json:"Thermal"`
	PhysicalSecurity *struct {
		IntrusionSensor       string `json:"IntrusionSensor"`
		IntrusionSensorNumber int    `json:"IntrusionSensorNumber"`
		IntrusionSensorReArm  string `json:"IntrusionSensorReArm"`
	} `json:"PhysicalSecurity"`
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

type PCIeDeviceResponse struct {
	ID              string `json:"Id"`
	Name            string `json:"Name"`
	Description     string `json:"Description"`
	Model           string `json:"Model"`
	SerialNumber    string `json:"SerialNumber"`
	PartNumber      string `json:"PartNumber"`
	FirmwareVersion string `json:"FirmwareVersion"`
	DeviceType      string `json:"DeviceType"`
	Status          Status `json:"Status"`
	PCIeInterface struct {
		PCIeType    string `json:"PCIeType"`
		MaxPCIeType string `json:"MaxPCIeType"`
		LanesInUse  int    `json:"LanesInUse"`
		MaxLanes    int    `json:"MaxLanes"`
	} `json:"PCIeInterface"`
	PCIeFunctions Odata `json:"PCIeFunctions"`
	Oem struct {
		Supermicro struct {
			OdataType        string `json:"@odata.type"`
			GPUSlot          int    `json:"GPUSlot"`
			BoardPartNumber  string `json:"BoardPartNumber"`
			Driver           string `json:"Driver"`
			MemoryVendor     string `json:"MemoryVendor"`
			MemoryPartNumber string `json:"MemoryPartNumber"`
			GPUGUID          string `json:"GPUGuid"`
			InfoROMVersion   string `json:"InfoROMVersion"`
			GPUVendor        string `json:"GPUVendor"`
		} `json:"Supermicro"`
	} `json:"Oem"`
}

type ThermalResponse struct {
	Name         string        `json:"Name"`
	Description  string        `json:"Description"`
	Fans         []Fan         `json:"Fans"`
	Temperatures []Temperature `json:"Temperatures"`
	Redundancy   []Redundancy  `json:"Redundancy"`
}

type Fan struct {
	Name                      string       `json:"Name"`
	FanName                   string       `json:"FanName"`
	MemberId                  string       `json:"MemberId"`
	Assembly                  Odata        `json:"Assembly"`
	HotPluggable              bool         `json:"HotPluggable"`
	MaxReadingRange           any          `json:"MaxReadingRange"`
	MinReadingRange           any          `json:"MinReadingRange"`
	PhysicalContext           string       `json:"PhysicalContext"`
	Reading                   float64      `json:"Reading"`
	CurrentReading            float64      `json:"CurrentReading"`
	Units                     string       `json:"Units"`
	ReadingUnits              string       `json:"ReadingUnits"`
	Redundancy                []Redundancy `json:"Redundancy"`
	Status                    Status       `json:"Status"`
	LowerThresholdCritical    any          `json:"LowerThresholdCritical"`
	LowerThresholdFatal       any          `json:"LowerThresholdFatal"`
	LowerThresholdNonCritical any          `json:"LowerThresholdNonCritical"`
	UpperThresholdCritical    any          `json:"UpperThresholdCritical"`
	UpperThresholdFatal       any          `json:"UpperThresholdFatal"`
	UpperThresholdNonCritical any          `json:"UpperThresholdNonCritical"`
}

func (f *Fan) GetName() string {
	if f.FanName != "" {
		return f.FanName
	}
	return f.Name
}

func (f *Fan) GetReading() float64 {
	if f.Reading > 0 {
		return f.Reading
	}
	return f.CurrentReading
}

func (f *Fan) GetUnits() string {
	if f.ReadingUnits != "" {
		return f.ReadingUnits
	}
	return f.Units
}

func (f *Fan) GetId(fallback int) string {
	if len(f.MemberId) > 0 {
		return f.MemberId
	}
	return strconv.Itoa(fallback)
}

type Temperature struct {
	Name                      string  `json:"Name"`
	Number                    int     `json:"Number"`
	MemberId                  string  `json:"MemberId"`
	ReadingCelsius            float64 `json:"ReadingCelsius"`
	MaxReadingRangeTemp       float64 `json:"MaxReadingRangeTemp"`
	MinReadingRangeTemp       float64 `json:"MinReadingRangeTemp"`
	PhysicalContext           string  `json:"PhysicalContext"`
	LowerThresholdCritical    float64 `json:"LowerThresholdCritical"`
	LowerThresholdFatal       float64 `json:"LowerThresholdFatal"`
	LowerThresholdNonCritical float64 `json:"LowerThresholdNonCritical"`
	UpperThresholdCritical    float64 `json:"UpperThresholdCritical"`
	UpperThresholdFatal       float64 `json:"UpperThresholdFatal"`
	UpperThresholdNonCritical float64 `json:"UpperThresholdNonCritical"`
	Status                    Status  `json:"Status"`
	RelatedItem               []Odata `json:"RelatedItem"`
}

func (t *Temperature) GetId(fallback int) string {
	if len(t.MemberId) > 0 {
		return t.MemberId
	}
	if t.Number > 0 {
		return strconv.Itoa(t.Number)
	}
	return strconv.Itoa(fallback)
}
