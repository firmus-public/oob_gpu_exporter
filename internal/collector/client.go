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
)

type Client struct {
	redfish     *Redfish
	vendor      int
	systemPath  string
	procPath    string
	chassisPath string
	devicesPath string
	thermalPath string
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
	client := &Client{
		redfish: NewRedfish(
			h.Scheme,
			h.Hostname,
			h.Username,
			h.Password,
		),
	}

	client.redfish.CreateSession()
	ok := client.findAllEndpoints()
	if !ok {
		client.redfish.DeleteSession()
		return nil
	}

	return client
}

func (client *Client) findAllEndpoints() bool {
	var root V1Response
	var group GroupResponse
	var system SystemResponse
	var chassis ChassisResponse
	var ok bool

	// Root
	ok = client.redfish.Get(redfishRootPath, &root)
	if !ok {
		return false
	}

	// Chassis
	ok = client.redfish.Get(root.Chassis.OdataId, &group)
	if !ok {
		return false
	}

	client.chassisPath = group.Members[0].OdataId

	ok = client.redfish.Get(client.chassisPath, &chassis)
	if !ok {
		return false
	}

	// System
	ok = client.redfish.Get(root.Systems.OdataId, &group)
	if !ok {
		return false
	}

	client.systemPath = group.Members[0].OdataId

	ok = client.redfish.Get(client.systemPath, &system)
	if !ok {
		return false
	}

	client.procPath = system.Processors.OdataId
	client.devicesPath = chassis.PCIeDevices.OdataId
	client.thermalPath = chassis.Thermal.OdataId

	// Vendor
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
		client.vendor = SUPERMICRO
	}

	return true
}

func (client *Client) RefreshGPUs(mc *Collector, ch chan<- prometheus.Metric) bool {
	switch client.vendor {
	case DELL:
		return client.refreshDellGPUs(mc, ch)
	case SUPERMICRO:
		return client.refreshSupermicroGPUs(mc, ch)
	default:
		return false
	}
}

func (client *Client) refreshDellGPUs(mc *Collector, ch chan<- prometheus.Metric) bool {
	group := GroupResponse{}
	ok := client.redfish.Get(client.procPath, &group)
	if !ok {
		return false
	}

	// Get inventory information for Dell GPUs

	dellVideo := DellVideo{}

	// Get dell video inventory

	dellVideoPath := fmt.Sprintf("%s/Oem/Dell/DellVideo", client.systemPath)
	client.redfish.Get(dellVideoPath, &dellVideo)

    // GPU count
    var count int = len(dellVideo.Members)
    mc.NewGPUCount(ch, count)

	// Get dell GPU sensor metrics

    dellGPUSensorPath := fmt.Sprintf("%s/Oem/Dell/DellGPUSensors", client.systemPath)
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
	group := GroupResponse{}
	ok := client.redfish.Get(client.devicesPath, &group)
	if !ok {
		return false
	}

	// Get GPU metrics

	for _, c := range group.Members.GetLinks() {
		if !strings.Contains(c, "GPU") {
			continue
		}

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

	thermalResp := ThermalResponse{}
	ok = client.redfish.Get(client.thermalPath, &thermalResp)

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

	return true
}
