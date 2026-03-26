package collector

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/firmus-public/oob_gpu_exporter/internal/config"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	UNKNOWN = iota
	DELL
	HPE
	LENOVO
	INSPUR
	H3C
	INVENTEC
	FUJITSU
	SUPERMICRO
	SUPERMICRO_GB_NVL
)

// ChassisData holds paths related to a single chassis member
type ChassisData struct {
	chassisID   string
	chassisPath string
	devicesPath string
	thermalPath string
}

// extractChassisID extracts the chassis ID from a chassis path
// e.g., "/redfish/v1/Chassis/HGX_GPU_0" -> "HGX_GPU_0"
func extractChassisID(chassisPath string) string {
	parts := strings.Split(chassisPath, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return chassisPath
}

// SystemData holds paths related to a single system member
type SystemData struct {
	systemID   string
	systemPath string
	procPath   string
}

type Client struct {
	redfish      *Redfish
	vendor       int
	systemItems  []SystemData
	chassisItems []ChassisData
	product      string
}

// getDefaultChassisID returns the first chassis ID or empty string if none available
func (client *Client) getDefaultChassisID() string {
	if len(client.chassisItems) > 0 {
		return client.chassisItems[0].chassisID
	}
	return ""
}

// getDefaultSystemPath returns the first system path or empty string if none available
func (client *Client) getDefaultSystemPath() string {
	if len(client.systemItems) > 0 {
		return client.systemItems[0].systemPath
	}
	return ""
}

// getDefaultProcPath returns the first processors path or empty string if none available
func (client *Client) getDefaultProcPath() string {
	if len(client.systemItems) > 0 {
		return client.systemItems[0].procPath
	}
	return ""
}

type GPUInfo struct {
	Id           string
	Manufacturer string
	Model        string
	PartNumber   string
	SerialNumber string
	GPUGUID      string
	Slot         int
}

func NewClient(h *config.HostConfig) *Client {
	fmt.Println("Creating new client for host", h.Hostname)
	client := &Client{
		redfish: NewRedfish(
			h.Scheme,
			h.Hostname,
			h.Username,
			h.Password,
		),
	}
	fmt.Println("Created new client for host", h.Hostname)
	client.redfish.CreateSession()
	fmt.Println("Created session for host", h.Hostname)
	ok := client.findAllEndpoints()
	if !ok {
		fmt.Println("Failed to find endpoints for host", h.Hostname)
		client.redfish.DeleteSession()
		return nil
	}
	fmt.Println("Found endpoints for host", h.Hostname)
	return client
}

func (client *Client) findAllEndpoints() bool {
	var root V1Response
	var group GroupResponse
	var system SystemResponse
	var ok bool

	// Root
	ok = client.redfish.Get(redfishRootPath, &root)
	if !ok {
		return false
	}
	client.product = root.Product

	// Chassis - iterate over all members
	ok = client.redfish.Get(root.Chassis.OdataId, &group)
	if !ok {
		return false
	}

	client.chassisItems = make([]ChassisData, 0, len(group.Members))
	for _, member := range group.Members {
		var chassis ChassisResponse
		ok = client.redfish.Get(member.OdataId, &chassis)
		if !ok {
			continue
		}

		chassisData := ChassisData{
			chassisID:   extractChassisID(member.OdataId),
			chassisPath: member.OdataId,
			devicesPath: chassis.PCIeDevices.OdataId,
			thermalPath: chassis.Thermal.OdataId,
		}
		client.chassisItems = append(client.chassisItems, chassisData)
	}

	// Note: We don't require chassis items to be present, as some vendors
	// (like GB-NVL) use system-level processors for GPU data, not chassis devices.

	// Systems - iterate over all members
	ok = client.redfish.Get(root.Systems.OdataId, &group)
	if !ok {
		return false
	}

	client.systemItems = make([]SystemData, 0, len(group.Members))
	for _, member := range group.Members {
		ok = client.redfish.Get(member.OdataId, &system)
		if !ok {
			continue
		}

		systemData := SystemData{
			systemID:   extractChassisID(member.OdataId), // reuse the same ID extraction logic
			systemPath: member.OdataId,
			procPath:   system.Processors.OdataId,
		}
		client.systemItems = append(client.systemItems, systemData)

		// Detect vendor from the first valid system
		if client.vendor == UNKNOWN {
			m := strings.ToLower(system.Manufacturer)
			if strings.Contains(m, "dell") || strings.Contains(m, "sustainable") {
				client.vendor = DELL
			} else if strings.Contains(m, "hpe") {
				client.vendor = HPE
			} else if strings.Contains(m, "lenovo") {
				client.vendor = LENOVO
			} else if strings.Contains(m, "inspur") {
				client.vendor = INSPUR
			} else if strings.Contains(m, "h3c") {
				client.vendor = H3C
			} else if strings.Contains(m, "inventec") {
				client.vendor = INVENTEC
			} else if strings.Contains(m, "fujitsu") {
				client.vendor = FUJITSU
			} else if strings.Contains(m, "supermicro") {
				if client.product == "GB NVL" {
					client.vendor = SUPERMICRO_GB_NVL
				} else {
					client.vendor = SUPERMICRO
				}
			}
		}
	}

	if len(client.systemItems) == 0 {
		return false
	}

	return true
}

func (client *Client) RefreshGPUs(mc *Collector, ch chan<- prometheus.Metric) bool {
	switch client.vendor {
	case DELL:
		return client.refreshDellGPUs(mc, ch)
	case SUPERMICRO:
		return client.refreshSupermicroGPUs(mc, ch)
	case SUPERMICRO_GB_NVL:
		return client.refreshSupermicroGBNVL(mc, ch)
	default:
		return false
	}
}

func (client *Client) refreshDellGPUs(mc *Collector, ch chan<- prometheus.Metric) bool {
	systemPath := client.getDefaultSystemPath()
	procPath := client.getDefaultProcPath()

	group := GroupResponse{}
	ok := client.redfish.Get(procPath, &group)
	if !ok {
		return false
	}

	// Get inventory information for Dell GPUs

	dellVideo := DellVideo{}

	// Get dell video inventory

	dellVideoPath := fmt.Sprintf("%s/Oem/Dell/DellVideo", systemPath)
	client.redfish.Get(dellVideoPath, &dellVideo)

	// GPU count
	var count = len(dellVideo.Members)
	mc.NewGPUCount(ch, count)

	// Get dell GPU sensor metrics

	dellGPUSensorPath := fmt.Sprintf("%s/Oem/Dell/DellGPUSensors", systemPath)
	dellGPUSensors := DellGPUSensors{}
	if ok := client.redfish.Get(dellGPUSensorPath, &dellGPUSensors); ok {
		for _, v := range dellGPUSensors.Members {
			mc.NewBoardPowerSupplyStatus(ch, &v)
			mc.NewMemoryTemperatureCelsius(ch, &v)
			mc.NewPowerBrakeStatus(ch, &v)
			mc.NewPrimaryGPUTemperatureCelsius(ch, &v)
			mc.NewThermalAlertStatus(ch, &v)
		}
	}

	// Get GPU metrics

	for _, c := range group.Members.GetLinks() {
		resp := GPU{}
		ok = client.redfish.Get(c, &resp)
		if !ok {
			continue
		}

		if resp.ProcessorType != "GPU" {
			continue
		}

		if resp.Status.State != StateEnabled {
			continue
		}

		gpuInfo := GPUInfo{}
		gpuInfo.Id = resp.Id
		gpuInfo.Manufacturer = resp.Manufacturer
		gpuInfo.Model = resp.Model
		gpuInfo.PartNumber = resp.PartNumber

		for _, v := range dellVideo.Members {
			if v.Id == resp.Id {
				gpuInfo.GPUGUID = v.GPUGUID
				gpuInfo.SerialNumber = v.SerialNumber
				mc.NewDellGPUState(ch, &v)
				mc.NewDellGPUHealth(ch, &v)
				break
			}
		}

		mc.NewGPUInfo(ch, &gpuInfo)

		if resp.Metrics.OdataId != "" {
			gpuMetrics := GPUMetrics{}
			ok = client.redfish.Get(resp.Metrics.OdataId, &gpuMetrics)
			if !ok {
				break
			}

			mc.NewGPUBandwidthPercent(ch, &gpuMetrics)
			mc.NewGPUConsumedPowerWatt(ch, &gpuMetrics)
			mc.NewGPUOperatingSpeedMHz(ch, &gpuMetrics)

			if gpuMetrics.Oem != nil {
				nvidia := gpuMetrics.Oem.Nvidia
				if nvidia != nil {
					mc.NewGPUThrottleReasons(ch, nvidia.ThrottleReasons, gpuMetrics.Id)
					mc.NewGPUSMUtilizationPercent(ch, nvidia.SMUtilizationPercent, gpuMetrics.Id)
					mc.NewGPUSMActivityPercent(ch, nvidia.SMActivityPercent, gpuMetrics.Id)
					mc.NewGPUSMOccupancyPercent(ch, nvidia.SMOccupancyPercent, gpuMetrics.Id)
					mc.NewGPUTensorCoreActivityPercent(ch, nvidia.TensorCoreActivityPercent, gpuMetrics.Id)
					mc.NewGPUHMMAUtilizationPercent(ch, nvidia.HMMAUtilizationPercent, gpuMetrics.Id)
					mc.NewGPUPCIeRawTxBandwidthGbps(ch, nvidia.PCIeRawTxBandwidthGbps, gpuMetrics.Id)
					mc.NewGPUPCIeRawRxBandwidthGbps(ch, nvidia.PCIeRawRxBandwidthGbps, gpuMetrics.Id)
				}
				dell := gpuMetrics.Oem.Dell
				if dell != nil {
					mc.NewGPUCurrentPCIeLinkSpeed(ch, dell.CurrentPCIeLinkSpeed, gpuMetrics.Id)
					mc.NewGPUMaxSupportedPCIeLinkSpeed(ch, dell.MaxSupportedPCIeLinkSpeed, gpuMetrics.Id)
					mc.NewGPUDRAMUtilizationPercent(ch, dell.DRAMUtilizationPercent, gpuMetrics.Id)
				}
			}

			if gpuMetrics.PCIeErrors != nil {
				mc.NewGPUPCIeCorrectableErrorCount(ch, gpuMetrics.PCIeErrors.CorrectableErrorCount, gpuMetrics.Id)
			}
		}

		if resp.MemorySummary.Metrics.OdataId != "" {
			gpuMemoryMetrics := GPUMemoryMetrics{}
			ok = client.redfish.Get(resp.MemorySummary.Metrics.OdataId, &gpuMemoryMetrics)
			if !ok {
				break
			}

			mc.NewGPUMemoryBandwidthPercent(ch, resp.Id, &gpuMemoryMetrics)
			mc.NewGPUMemoryOperatingSpeedMHz(ch, resp.Id, &gpuMemoryMetrics)
		}
	}

	return true
}

var GPU_REGEXP = regexp.MustCompile(`GPU (.*) Temp`)
var HBM_REGEXP = regexp.MustCompile(`HBM (.*) Temp`)

func (client *Client) refreshSupermicroGPUs(mc *Collector, ch chan<- prometheus.Metric) bool {
	var totalCount = 0

	// Iterate over all chassis members
	for _, chassisData := range client.chassisItems {
		if chassisData.devicesPath == "" {
			continue
		}

		group := GroupResponse{}
		ok := client.redfish.Get(chassisData.devicesPath, &group)
		if !ok {
			continue
		}

		// Get GPU metrics
		for _, c := range group.Members.GetLinks() {
			if !strings.Contains(c, "GPU") {
				continue
			}
			totalCount++

			resp := PCIeDeviceResponse{}

			ok = client.redfish.Get(c, &resp)
			if !ok {
				continue
			}

			gpuInfo := GPUInfo{}
			gpuInfo.Id = resp.ID
			gpuInfo.Model = resp.Model
			gpuInfo.PartNumber = resp.PartNumber
			gpuInfo.SerialNumber = resp.SerialNumber

			if resp.Oem != nil && resp.Oem.Supermicro != nil {
				gpuInfo.Manufacturer = resp.Oem.Supermicro.GPUVendor
				if resp.Oem.Supermicro.GPUGUID1 != "" {
					gpuInfo.GPUGUID = resp.Oem.Supermicro.GPUGUID1
				} else {
					gpuInfo.GPUGUID = resp.Oem.Supermicro.GPUGUID2
				}
				gpuInfo.Slot = resp.Oem.Supermicro.GPUSlot
			}

			mc.NewGPUInfo(ch, &gpuInfo)
			mc.NewSupermicroGPUHealth(ch, &resp)
			mc.NewSupermicroGPUState(ch, &resp)
		}

		// Get thermal data for this chassis
		if chassisData.thermalPath == "" {
			continue
		}

		thermalResp := ThermalResponse{}
		ok = client.redfish.Get(chassisData.thermalPath, &thermalResp)

		if ok {
			for _, t := range thermalResp.Temperatures {
				if t.Name == "GPU Temp" && t.Oem != nil && t.Oem.Supermicro != nil {
					for name, value := range t.Oem.Supermicro.Details {
						matches := GPU_REGEXP.FindStringSubmatch(name)
						if matches != nil {
							temp, err := strconv.ParseFloat(value, 64)
							if err == nil {
								id := "GPU" + matches[1]
								mc.NewSmcGPUTemp(ch, id, temp)
							}
						}
					}
					continue
				}

				if t.Name == "HBM Temp" && t.Oem != nil && t.Oem.Supermicro != nil {
					for name, value := range t.Oem.Supermicro.Details {
						matches := HBM_REGEXP.FindStringSubmatch(name)
						if matches != nil {
							temp, err := strconv.ParseFloat(value, 64)
							if err == nil {
								id := "GPU" + matches[1]
								mc.NewSmcGPUMemoryTemp(ch, id, temp)
							}
						}
					}
					continue
				}

				re := regexp.MustCompile(`(GPU\d+) Temp`)
				matches := re.FindStringSubmatch(t.Name)
				if matches != nil {
					id := matches[1]
					mc.NewSmcGPUTemp(ch, id, t.ReadingCelsius)
					continue
				}
			}
		}
	}

	mc.NewGPUCount(ch, totalCount)

	return totalCount > 0 || len(client.chassisItems) > 0
}

func (client *Client) refreshSupermicroGBNVL(mc *Collector, ch chan<- prometheus.Metric) bool {
	gpuCount := 0

	// Iterate over all systems to find GPUs
	for _, systemData := range client.systemItems {
		if systemData.procPath == "" {
			continue
		}

		group := GroupResponse{}
		ok := client.redfish.Get(systemData.procPath, &group)
		if !ok {
			continue
		}

		for _, c := range group.Members.GetLinks() {
			resp := GPU{}
			ok = client.redfish.Get(c, &resp)
			if !ok {
				continue
			}

			if resp.ProcessorType != "GPU" {
				continue
			}

			gpuCount++

			gpuInfo := GPUInfo{}
			// Basic Info
			gpuInfo.Id = resp.Id
			gpuInfo.Manufacturer = resp.Manufacturer
			gpuInfo.Model = resp.Model
			gpuInfo.PartNumber = resp.PartNumber
			gpuInfo.SerialNumber = resp.SerialNumber

			// Status
			if resp.Status.Health != "" {
				mc.NewGBNVLGPUHealth(ch, resp.Id, resp.Status.Health)
			}
			if resp.Status.State != "" {
				mc.NewGBNVLGPUState(ch, resp.Id, resp.Status.State)
			}

			// Parse EnvironmentMetrics for Power, Temp, and Energy
			if resp.EnvironmentMetrics != nil && resp.EnvironmentMetrics.OdataId != "" {
				envMetrics := EnvironmentMetricsResponse{}
				ok = client.redfish.Get(resp.EnvironmentMetrics.OdataId, &envMetrics)
				if ok {
					mc.NewGPUConsumedPowerWatt(ch, &GPUMetrics{Id: resp.Id, ConsumedPowerWatt: envMetrics.PowerWatts.Reading})
					mc.NewGPUPowerLimitWatts(ch, resp.Id, envMetrics.PowerLimitWatts.Reading)

					// Using NewSmcGPUTemp as it maps to GPUPrimaryGPUTemperatureCelsius which is what we want
					mc.NewSmcGPUTemp(ch, resp.Id, envMetrics.TemperatureCelsius.Reading)

					// Energy consumption
					mc.NewGPUEnergyJoules(ch, resp.Id, envMetrics.EnergyJoules.Reading)
				}
			}

			// TotalNumberNVLinks from GPU OEM data
			if resp.Oem.Nvidia.TotalNumberNVLinks > 0 {
				mc.NewGPUTotalNVLinks(ch, resp.Id, resp.Oem.Nvidia.TotalNumberNVLinks)
			}

			mc.NewGPUInfo(ch, &gpuInfo)

			// Processor Metrics
			if resp.Metrics.OdataId != "" {
				gpuMetrics := GPUMetrics{}
				ok = client.redfish.Get(resp.Metrics.OdataId, &gpuMetrics)
				if !ok {
					continue
				}

				// Override with the GPU ID since ProcessorMetrics always has Id="ProcessorMetrics"
				gpuMetrics.Id = resp.Id

				mc.NewGPUBandwidthPercent(ch, &gpuMetrics)
				mc.NewGPUOperatingSpeedMHz(ch, &gpuMetrics)
				mc.NewGPUCoreVoltageVolts(ch, resp.Id, gpuMetrics.CoreVoltage.Reading)

				if gpuMetrics.Oem != nil && gpuMetrics.Oem.Nvidia != nil {
					nvidia := gpuMetrics.Oem.Nvidia
					mc.NewGPUThrottleReasons(ch, nvidia.ThrottleReasons, gpuMetrics.Id)
					mc.NewGPUSMUtilizationPercent(ch, nvidia.SMUtilizationPercent, gpuMetrics.Id)
					mc.NewGPUSMActivityPercent(ch, nvidia.SMActivityPercent, gpuMetrics.Id)
					mc.NewGPUSMOccupancyPercent(ch, nvidia.SMOccupancyPercent, gpuMetrics.Id)
					mc.NewGPUTensorCoreActivityPercent(ch, nvidia.TensorCoreActivityPercent, gpuMetrics.Id)
					mc.NewGPUHMMAUtilizationPercent(ch, nvidia.HMMAUtilizationPercent, gpuMetrics.Id)
					mc.NewGPUPCIeRawTxBandwidthGbps(ch, nvidia.PCIeRawTxBandwidthGbps, gpuMetrics.Id)
					mc.NewGPUPCIeRawRxBandwidthGbps(ch, nvidia.PCIeRawRxBandwidthGbps, gpuMetrics.Id)

					mc.NewGPUGaugeMetric(ch, mc.GPUFP16ActivityPercent, gpuMetrics.Id, nvidia.FP16ActivityPercent)
					mc.NewGPUGaugeMetric(ch, mc.GPUFP32ActivityPercent, gpuMetrics.Id, nvidia.FP32ActivityPercent)
					mc.NewGPUGaugeMetric(ch, mc.GPUFP64ActivityPercent, gpuMetrics.Id, nvidia.FP64ActivityPercent)
					mc.NewGPUGaugeMetric(ch, mc.GPUIntegerActivityUtilizationPercent, gpuMetrics.Id, nvidia.IntegerActivityUtilizationPercent)
					mc.NewGPUGaugeMetric(ch, mc.GPUNVLinkDataRxBandwidthGbps, gpuMetrics.Id, nvidia.NVLinkDataRxBandwidthGbps)
					mc.NewGPUGaugeMetric(ch, mc.GPUNVLinkDataTxBandwidthGbps, gpuMetrics.Id, nvidia.NVLinkDataTxBandwidthGbps)

					// New NVDEC/NVJPG utilization metrics
					mc.NewGPUGaugeMetric(ch, mc.GPUNVDecUtilizationPercent, gpuMetrics.Id, nvidia.NVDecUtilizationPercent)
					mc.NewGPUGaugeMetric(ch, mc.GPUNVJpgUtilizationPercent, gpuMetrics.Id, nvidia.NVJpgUtilizationPercent)

					// NVLink raw bandwidth (GPU-level aggregate)
					mc.NewGPUGaugeMetric(ch, mc.GPUNVLinkGpuRawRxBandwidthGbps, gpuMetrics.Id, nvidia.NVLinkRawRxBandwidthGbps)
					mc.NewGPUGaugeMetric(ch, mc.GPUNVLinkGpuRawTxBandwidthGbps, gpuMetrics.Id, nvidia.NVLinkRawTxBandwidthGbps)
				}

				// Cache ECC errors
				if gpuMetrics.CacheMetricsTotal != nil {
					mc.NewGPUCacheECCErrors(ch, gpuMetrics.Id,
						gpuMetrics.CacheMetricsTotal.LifeTime.CorrectableECCErrorCount,
						gpuMetrics.CacheMetricsTotal.LifeTime.UncorrectableECCErrorCount)
				}

				if gpuMetrics.PCIeErrors != nil {
					mc.NewGPUPCIeCorrectableErrorCount(ch, gpuMetrics.PCIeErrors.CorrectableErrorCount, gpuMetrics.Id)
				}
			}

			// Memory Metrics
			if resp.MemorySummary.Metrics.OdataId != "" {
				gpuMemoryMetrics := GPUMemoryMetrics{}
				ok = client.redfish.Get(resp.MemorySummary.Metrics.OdataId, &gpuMemoryMetrics)
				if ok {
					mc.NewGPUMemoryBandwidthPercent(ch, resp.Id, &gpuMemoryMetrics)
					mc.NewGPUMemoryOperatingSpeedMHz(ch, resp.Id, &gpuMemoryMetrics)
					mc.NewGPUMemoryECCErrors(ch, resp.Id, gpuMemoryMetrics.LifeTime.CorrectableECCErrorCount, gpuMemoryMetrics.LifeTime.UncorrectableECCErrorCount)
				}
			}

			// Power Smoothing
			if resp.Oem.Nvidia.PowerSmoothing.OdataId != "" {
				powerSmoothing := PowerSmoothing{}
				ok = client.redfish.Get(resp.Oem.Nvidia.PowerSmoothing.OdataId, &powerSmoothing)
				if ok {
					mc.NewGPUPowerSmoothingMetrics(ch, resp.Id, &powerSmoothing)
				}
			}

			// Processor Reset Metrics
			if resp.Oem.Nvidia.ProcessorResetMetrics != nil && resp.Oem.Nvidia.ProcessorResetMetrics.OdataId != "" {
				resetMetrics := ProcessorResetMetrics{}
				ok = client.redfish.Get(resp.Oem.Nvidia.ProcessorResetMetrics.OdataId, &resetMetrics)
				if ok {
					mc.NewGPUResetMetrics(ch, resp.Id, &resetMetrics)
				}
			}

			// NVLink Ports
			if resp.Ports != nil && resp.Ports.OdataId != "" {
				portsResp := PortCollection{}
				ok = client.redfish.Get(resp.Ports.OdataId, &portsResp)
				if ok {
					for _, link := range portsResp.Members {
						// We only care about NVLink ports
						// The ID/Link usually contains the ID, but let's fetch checking the ID
						// Optimization: We could check if the URL contains "NVLink" before fetching,
						// but let's fetch to be safe and check ID.
						// Actually, checking URL is safer to avoid fetching PCIe ports if they are in the same collection
						if !strings.Contains(link.OdataId, "NVLink") {
							continue
						}

						portResp := PortResponse{}
						ok = client.redfish.Get(link.OdataId, &portResp)
						if !ok {
							continue
						}

						if !strings.HasPrefix(portResp.Id, "NVLink") {
							continue
						}

						if portResp.Metrics != nil && portResp.Metrics.OdataId != "" {
							portMetricsResp := PortMetricsResponse{}
							ok = client.redfish.Get(portResp.Metrics.OdataId, &portMetricsResp)
							if ok {
								mc.NewGPUNVLinkPortMetrics(ch, resp.Id, portResp.Id, &portResp, &portMetricsResp)
							}
						}
					}
				}
			}
		}
	}

	mc.NewGPUCount(ch, gpuCount)

	return true
}
