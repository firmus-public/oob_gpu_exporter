package collector

import (
	"fmt"
	"log"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/firmus-public/oob_gpu_exporter/internal/config"
	"github.com/firmus-public/oob_gpu_exporter/internal/version"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
)

var mu sync.Mutex
var collectors = map[string]*Collector{}

type Collector struct {
	// Internal variables
	client     *Client
	registry   *prometheus.Registry
	collected  *sync.Cond
	collecting bool
	errors     atomic.Uint64
	builder    *strings.Builder

	// Exporter
	ExporterBuildInfo         *prometheus.Desc
	ExporterScrapeErrorsTotal *prometheus.Desc

	// GPUs
	GPUCount                        *prometheus.Desc
	GPUInfo                         *prometheus.Desc
	GPUState                        *prometheus.Desc
	GPUHealth                       *prometheus.Desc
	GPUBoardPowerSupplyStatus       *prometheus.Desc
	GPUMemoryTemperatureCelsius     *prometheus.Desc
	GPUPowerBrakeStatus             *prometheus.Desc
	GPUPrimaryGPUTemperatureCelsius *prometheus.Desc
	GPUThermalAlertStatus           *prometheus.Desc
	GPUBandwidthPercent             *prometheus.Desc
	GPUConsumedPowerWatt            *prometheus.Desc
	GPUOperatingSpeedMHz            *prometheus.Desc
	GPUMemoryBandwidthPercent       *prometheus.Desc
	GPUMemoryOperatingSpeedMHz      *prometheus.Desc
	GPUThrottleReason               *prometheus.Desc
	GPUSMUtilizationPercent         *prometheus.Desc
	GPUSMActivityPercent            *prometheus.Desc
	GPUSMOccupancyPercent           *prometheus.Desc
	GPUTensorCoreActivityPercent    *prometheus.Desc
	GPUHMMAUtilizationPercent       *prometheus.Desc
	GPUPCIeRawTxBandwidthGbps       *prometheus.Desc
	GPUPCIeRawRxBandwidthGbps       *prometheus.Desc
	GPUCurrentPCIeLinkSpeed         *prometheus.Desc
	GPUMaxSupportedPCIeLinkSpeed    *prometheus.Desc
	GPUDRAMUtilizationPercent       *prometheus.Desc
	GPUPCIeCorrectableErrorCount    *prometheus.Desc
}

func NewCollector() *Collector {
	prefix := config.Config.MetricsPrefix

	collector := &Collector{
		ExporterBuildInfo: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu_exporter", "build_info"),
			"Constant metric with build information for the exporter",
			nil, prometheus.Labels{
				"version":   version.Version,
				"revision":  version.Revision,
				"goversion": runtime.Version(),
			},
		),
		ExporterScrapeErrorsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu_exporter", "scrape_errors_total"),
			"Total number of errors encountered while scraping target",
			nil, nil,
		),
        GPUCount: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "count"),
			"The number of GPUs detected",
			nil, nil,
		),
		GPUInfo: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "info"),
			"Information about the GPU",
			[]string{"id", "manufacturer", "model", "part_number", "serial_number", "guid", "slot"}, nil,
		),
		GPUState: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "state"),
			"State of the GPU",
			[]string{"id", "state"}, nil,
		),
		GPUHealth: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "health"),
			"Health status of the GPU",
			[]string{"id", "status"}, nil,
		),
		GPUBoardPowerSupplyStatus: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "board_power_supply_status"),
			"Status of the GPU board power supply",
			[]string{"id", "status"}, nil,
		),
		GPUMemoryTemperatureCelsius: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "memory_temperature_celsius"),
			"Temperature of the GPU memory in celsius",
			[]string{"id"}, nil,
		),
		GPUPowerBrakeStatus: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "power_brake_status"),
			"Status of the GPU power brake",
			[]string{"id", "status"}, nil,
		),
		GPUPrimaryGPUTemperatureCelsius: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "primary_gpu_temperature_celsius"),
			"Primary temperature of the GPU in celsius",
			[]string{"id"}, nil,
		),
		GPUThermalAlertStatus: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "thermal_alert_status"),
			"Thermal alert status of the GPU",
			[]string{"id", "status"}, nil,
		),
		GPUBandwidthPercent: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "bandwidth_percent"),
			"Utilization of the GPU in percent",
			[]string{"id"}, nil,
		),
		GPUConsumedPowerWatt: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "consumed_power_watt"),
			"Power consumed by the GPU in watts",
			[]string{"id"}, nil,
		),
		GPUOperatingSpeedMHz: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "operating_speed_mhz"),
			"Operating speed of the GPU in Mhz",
			[]string{"id"}, nil,
		),
		GPUMemoryBandwidthPercent: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "memory_bandwidth_percent"),
			"Utilization of the GPU memory in percent",
			[]string{"id"}, nil,
		),
		GPUMemoryOperatingSpeedMHz: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "memory_operating_speed_mhz"),
			"Operating speed of the GPU memory in Mhz",
			[]string{"id"}, nil,
		),
		GPUThrottleReason: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "throttle_reason"),
			"Reason for GPU throttling",
			[]string{"id", "reason"}, nil,
		),
		GPUSMUtilizationPercent: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "sm_utilization_percent"),
			"Streaming Multiprocessor (SM) utilization of the GPU in percent",
			[]string{"id"}, nil,
		),
		GPUSMActivityPercent: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "sm_activity_percent"),
			"Streaming Multiprocessor (SM) activity of the GPU in percent",
			[]string{"id"}, nil,
		),
		GPUSMOccupancyPercent: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "sm_occupancy_percent"),
			"Streaming Multiprocessor (SM) occupancy of the GPU in percent",
			[]string{"id"}, nil,
		),
		GPUTensorCoreActivityPercent: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "tensor_core_activity_percent"),
			"Tensor Core activity of the GPU in percent",
			[]string{"id"}, nil,
		),
		GPUHMMAUtilizationPercent: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "hmma_utilization_percent"),
			"HMMA (Hybrid Matrix Multiply-Accumulate) utilization of the GPU in percent",
			[]string{"id"}, nil,
		),
		GPUPCIeRawTxBandwidthGbps: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "pcie_raw_tx_bandwidth_gbps"),
			"PCIe raw transmit bandwidth of the GPU in Gbps",
			[]string{"id"}, nil,
		),
		GPUPCIeRawRxBandwidthGbps: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "pcie_raw_rx_bandwidth_gbps"),
			"PCIe raw receive bandwidth of the GPU in Gbps",
			[]string{"id"}, nil,
		),
		GPUCurrentPCIeLinkSpeed: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "current_pcie_link_speed"),
			"Current PCIe link speed of the GPU",
			[]string{"id"}, nil,
		),
		GPUMaxSupportedPCIeLinkSpeed: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "max_supported_pcie_link_speed"),
			"Maximum supported PCIe link speed of the GPU",
			[]string{"id"}, nil,
		),
		GPUDRAMUtilizationPercent: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "dram_utilization_percent"),
			"DRAM utilization of the GPU in percent",
			[]string{"id"}, nil,
		),
		GPUPCIeCorrectableErrorCount: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "pcie_correctable_error_count"),
			"Number of correctable PCIe errors of the GPU",
			[]string{"id"}, nil,
		),
	}

	collector.builder = new(strings.Builder)
	collector.collected = sync.NewCond(new(sync.Mutex))
	collector.registry = prometheus.NewRegistry()
	collector.registry.MustRegister(collector)

	return collector
}

func (collector *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.ExporterBuildInfo
	ch <- collector.ExporterScrapeErrorsTotal
	ch <- collector.GPUCount
	ch <- collector.GPUInfo
	ch <- collector.GPUHealth
	ch <- collector.GPUState
	ch <- collector.GPUBoardPowerSupplyStatus
	ch <- collector.GPUMemoryTemperatureCelsius
	ch <- collector.GPUPowerBrakeStatus
	ch <- collector.GPUPrimaryGPUTemperatureCelsius
	ch <- collector.GPUThermalAlertStatus
	ch <- collector.GPUBandwidthPercent
	ch <- collector.GPUConsumedPowerWatt
	ch <- collector.GPUOperatingSpeedMHz
	ch <- collector.GPUMemoryBandwidthPercent
	ch <- collector.GPUMemoryOperatingSpeedMHz
	ch <- collector.GPUThrottleReason
	ch <- collector.GPUSMUtilizationPercent
	ch <- collector.GPUSMActivityPercent
	ch <- collector.GPUSMOccupancyPercent
	ch <- collector.GPUTensorCoreActivityPercent
	ch <- collector.GPUHMMAUtilizationPercent
	ch <- collector.GPUPCIeRawTxBandwidthGbps
	ch <- collector.GPUPCIeRawRxBandwidthGbps
	ch <- collector.GPUCurrentPCIeLinkSpeed
	ch <- collector.GPUMaxSupportedPCIeLinkSpeed
	ch <- collector.GPUDRAMUtilizationPercent
	ch <- collector.GPUPCIeCorrectableErrorCount
}

func (collector *Collector) Collect(ch chan<- prometheus.Metric) {
	collector.client.redfish.RefreshSession()

	ok := collector.client.RefreshGPUs(collector, ch)
	if !ok {
		collector.errors.Add(1)
	}

	ch <- prometheus.MustNewConstMetric(collector.ExporterBuildInfo, prometheus.UntypedValue, 1)
	ch <- prometheus.MustNewConstMetric(collector.ExporterScrapeErrorsTotal, prometheus.CounterValue, float64(collector.errors.Load()))
}

func (collector *Collector) Gather() (string, error) {
	collector.collected.L.Lock()

	// If a collection is already in progress wait for it to complete and return the cached data
	if collector.collecting {
		collector.collected.Wait()
		metrics := collector.builder.String()
		collector.collected.L.Unlock()
		return metrics, nil
	}

	// Set collecting to true and let other goroutines enter in critical section
	collector.collecting = true
	collector.collected.L.Unlock()

	// Defer set collecting to false and wake waiting goroutines
	defer func() {
		collector.collected.L.Lock()
		collector.collected.Broadcast()
		collector.collecting = false
		collector.collected.L.Unlock()
	}()

	// Collect metrics
	collector.builder.Reset()

	m, err := collector.registry.Gather()
	if err != nil {
		return "", err
	}

	for i := range m {
		_, err := expfmt.MetricFamilyToText(collector.builder, m[i])
		if err != nil {
			log.Printf("Error converting metric to text: %v", err)
		}
	}

	return collector.builder.String(), nil
}

// Resets an existing collector of the given target
func Reset(target string) {
	mu.Lock()
	_, ok := collectors[target]
	if ok {
		delete(collectors, target)
	}
	mu.Unlock()
}

func GetCollector(target string) (*Collector, error) {
	mu.Lock()
	collector, ok := collectors[target]
	if !ok {
		collector = NewCollector()
		collectors[target] = collector
	}
	mu.Unlock()

	// Do not act concurrently on the same host
	collector.collected.L.Lock()
	defer collector.collected.L.Unlock()

	// Try to instantiate a new Redfish host
	if collector.client == nil {
		host := config.GetHostConfig(target)
		if host == nil {
			return nil, fmt.Errorf("failed to get host information")
		}
		c := NewClient(host)
		if c == nil {
			return nil, fmt.Errorf("failed to instantiate new client")
		} else {
			collector.client = c
		}
	}

	return collector, nil
}
