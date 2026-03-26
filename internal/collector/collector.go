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
	GPUCount                             *prometheus.Desc
	GPUInfo                              *prometheus.Desc
	GPUState                             *prometheus.Desc
	GPUHealth                            *prometheus.Desc
	GPUBoardPowerSupplyStatus            *prometheus.Desc
	GPUMemoryTemperatureCelsius          *prometheus.Desc
	GPUPowerBrakeStatus                  *prometheus.Desc
	GPUPrimaryGPUTemperatureCelsius      *prometheus.Desc
	GPUThermalAlertStatus                *prometheus.Desc
	GPUBandwidthPercent                  *prometheus.Desc
	GPUConsumedPowerWatt                 *prometheus.Desc
	GPUOperatingSpeedMHz                 *prometheus.Desc
	GPUMemoryBandwidthPercent            *prometheus.Desc
	GPUMemoryOperatingSpeedMHz           *prometheus.Desc
	GPUThrottleReason                    *prometheus.Desc
	GPUSMUtilizationPercent              *prometheus.Desc
	GPUSMActivityPercent                 *prometheus.Desc
	GPUSMOccupancyPercent                *prometheus.Desc
	GPUTensorCoreActivityPercent         *prometheus.Desc
	GPUHMMAUtilizationPercent            *prometheus.Desc
	GPUPCIeRawTxBandwidthGbps            *prometheus.Desc
	GPUPCIeRawRxBandwidthGbps            *prometheus.Desc
	GPUCurrentPCIeLinkSpeed              *prometheus.Desc
	GPUMaxSupportedPCIeLinkSpeed         *prometheus.Desc
	GPUDRAMUtilizationPercent            *prometheus.Desc
	GPUPCIeCorrectableErrorCount         *prometheus.Desc
	GPUPowerLimitWatts                   *prometheus.Desc
	GPUCoreVoltageVolts                  *prometheus.Desc
	GPUMemoryCorrectableECCErrorCount    *prometheus.Desc
	GPUMemoryUncorrectableECCErrorCount  *prometheus.Desc
	GPUFP16ActivityPercent               *prometheus.Desc
	GPUFP32ActivityPercent               *prometheus.Desc
	GPUFP64ActivityPercent               *prometheus.Desc
	GPUIntegerActivityUtilizationPercent *prometheus.Desc
	GPUNVLinkDataRxBandwidthGbps         *prometheus.Desc
	GPUNVLinkDataTxBandwidthGbps         *prometheus.Desc
	GPUTotalNVLinks                      *prometheus.Desc
	GPUEnergyJoules                      *prometheus.Desc
	GPUCacheCorrectableECCErrorCount     *prometheus.Desc
	GPUCacheUncorrectableECCErrorCount   *prometheus.Desc
	GPUNVDecUtilizationPercent           *prometheus.Desc
	GPUNVJpgUtilizationPercent           *prometheus.Desc
	GPUNVLinkGpuRawRxBandwidthGbps       *prometheus.Desc
	GPUNVLinkGpuRawTxBandwidthGbps       *prometheus.Desc

	// Reset Metrics
	GPUConventionalResetEntryCount *prometheus.Desc
	GPUConventionalResetExitCount  *prometheus.Desc
	GPUFundamentalResetEntryCount  *prometheus.Desc
	GPUFundamentalResetExitCount   *prometheus.Desc
	GPUPFFLRResetEntryCount        *prometheus.Desc
	GPUPFFLRResetExitCount         *prometheus.Desc

	// Power Smoothing
	GPUPowerSmoothingSupported                         *prometheus.Desc
	GPUPowerSmoothingEnabled                           *prometheus.Desc
	GPUPowerSmoothingImmediateRampDown                 *prometheus.Desc
	GPUPowerSmoothingMaxAllowedTMPFloorPercent         *prometheus.Desc
	GPUPowerSmoothingMinAllowedTMPFloorPercent         *prometheus.Desc
	GPUPowerSmoothingRampDownHysteresisSeconds         *prometheus.Desc
	GPUPowerSmoothingRampDownWattsPerSecond            *prometheus.Desc
	GPUPowerSmoothingRampUpWattsPerSecond              *prometheus.Desc
	GPUPowerSmoothingRemainingLifetimeCircuitryPercent *prometheus.Desc
	GPUPowerSmoothingTMPFloorPercent                   *prometheus.Desc
	GPUPowerSmoothingTMPFloorWatts                     *prometheus.Desc
	GPUPowerSmoothingTMPWatts                          *prometheus.Desc

	// NVLink Port Metrics
	GPUNVLinkStatus           *prometheus.Desc
	GPUNVLinkCurrentSpeedGbps *prometheus.Desc
	GPUNVLinkRXBytes          *prometheus.Desc
	GPUNVLinkTXBytes          *prometheus.Desc
	GPUNVLinkBitErrorRate     *prometheus.Desc
	GPUNVLinkLinkDownedCount  *prometheus.Desc
	GPUNVLinkSymbolErrors     *prometheus.Desc
	GPUNVLinkRecoveryCount    *prometheus.Desc
	GPUNVLinkRuntimeError     *prometheus.Desc
	GPUNVLinkTrainingError    *prometheus.Desc
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
			prometheus.BuildFQName(prefix, "gpu", "num_gpus"),
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
		GPUPowerLimitWatts: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "power_limit_watt"),
			"Power limit of the GPU in watts",
			[]string{"id"}, nil,
		),
		GPUCoreVoltageVolts: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "core_voltage_volts"),
			"Core voltage of the GPU in volts",
			[]string{"id"}, nil,
		),
		GPUMemoryCorrectableECCErrorCount: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "memory_correctable_ecc_error_count"),
			"Number of correctable memory ECC errors of the GPU",
			[]string{"id"}, nil,
		),
		GPUMemoryUncorrectableECCErrorCount: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "memory_uncorrectable_ecc_error_count"),
			"Number of uncorrectable memory ECC errors of the GPU",
			[]string{"id"}, nil,
		),
		GPUFP16ActivityPercent: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "fp16_activity_percent"),
			"FP16 activity of the GPU in percent",
			[]string{"id"}, nil,
		),
		GPUFP32ActivityPercent: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "fp32_activity_percent"),
			"FP32 activity of the GPU in percent",
			[]string{"id"}, nil,
		),
		GPUFP64ActivityPercent: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "fp64_activity_percent"),
			"FP64 activity of the GPU in percent",
			[]string{"id"}, nil,
		),
		GPUIntegerActivityUtilizationPercent: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "integer_activity_utilization_percent"),
			"Integer activity utilization of the GPU in percent",
			[]string{"id"}, nil,
		),
		GPUNVLinkDataRxBandwidthGbps: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "nvlink_data_rx_bandwidth_gbps"),
			"NVLink data receive bandwidth of the GPU in Gbps",
			[]string{"id"}, nil,
		),
		GPUNVLinkDataTxBandwidthGbps: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "nvlink_data_tx_bandwidth_gbps"),
			"NVLink data transmit bandwidth of the GPU in Gbps",
			[]string{"id"}, nil,
		),
		GPUTotalNVLinks: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "total_nvlinks"),
			"Total number of NVLinks on the GPU",
			[]string{"id"}, nil,
		),
		GPUEnergyJoules: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "energy_joules"),
			"Total energy consumed by the GPU in joules",
			[]string{"id"}, nil,
		),
		GPUCacheCorrectableECCErrorCount: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "cache_correctable_ecc_error_count"),
			"Number of correctable cache ECC errors of the GPU",
			[]string{"id"}, nil,
		),
		GPUCacheUncorrectableECCErrorCount: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "cache_uncorrectable_ecc_error_count"),
			"Number of uncorrectable cache ECC errors of the GPU",
			[]string{"id"}, nil,
		),
		GPUNVDecUtilizationPercent: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "nvdec_utilization_percent"),
			"NVDEC video decoder utilization of the GPU in percent",
			[]string{"id"}, nil,
		),
		GPUNVJpgUtilizationPercent: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "nvjpg_utilization_percent"),
			"NVJPG JPEG decoder utilization of the GPU in percent",
			[]string{"id"}, nil,
		),
		GPUNVLinkGpuRawRxBandwidthGbps: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "nvlink_gpu_raw_rx_bandwidth_gbps"),
			"NVLink raw receive bandwidth of the GPU in Gbps (aggregate)",
			[]string{"id"}, nil,
		),
		GPUNVLinkGpuRawTxBandwidthGbps: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "nvlink_gpu_raw_tx_bandwidth_gbps"),
			"NVLink raw transmit bandwidth of the GPU in Gbps (aggregate)",
			[]string{"id"}, nil,
		),

		// Reset Metrics
		GPUConventionalResetEntryCount: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "conventional_reset_entry_count"),
			"Number of conventional reset entries on the GPU",
			[]string{"id"}, nil,
		),
		GPUConventionalResetExitCount: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "conventional_reset_exit_count"),
			"Number of conventional reset exits on the GPU",
			[]string{"id"}, nil,
		),
		GPUFundamentalResetEntryCount: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "fundamental_reset_entry_count"),
			"Number of fundamental reset entries on the GPU",
			[]string{"id"}, nil,
		),
		GPUFundamentalResetExitCount: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "fundamental_reset_exit_count"),
			"Number of fundamental reset exits on the GPU",
			[]string{"id"}, nil,
		),
		GPUPFFLRResetEntryCount: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "pf_flr_reset_entry_count"),
			"Number of PF FLR reset entries on the GPU",
			[]string{"id"}, nil,
		),
		GPUPFFLRResetExitCount: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "pf_flr_reset_exit_count"),
			"Number of PF FLR reset exits on the GPU",
			[]string{"id"}, nil,
		),

		// Power Smoothing
		GPUPowerSmoothingSupported: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "power_smoothing_supported"),
			"Indicates if logic power smoothing is supported on the GPU",
			[]string{"id"}, nil,
		),
		GPUPowerSmoothingEnabled: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "power_smoothing_enabled"),
			"Indicates if power smoothing is enabled on the GPU",
			[]string{"id"}, nil,
		),
		GPUPowerSmoothingImmediateRampDown: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "power_smoothing_immediate_ramp_down"),
			"Indicates if immediate ramp down is enabled on the GPU",
			[]string{"id"}, nil,
		),
		GPUPowerSmoothingMaxAllowedTMPFloorPercent: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "power_smoothing_max_allowed_tmp_floor_percent"),
			"Maximum allowed TMP floor percentage for power smoothing",
			[]string{"id"}, nil,
		),
		GPUPowerSmoothingMinAllowedTMPFloorPercent: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "power_smoothing_min_allowed_tmp_floor_percent"),
			"Minimum allowed TMP floor percentage for power smoothing",
			[]string{"id"}, nil,
		),
		GPUPowerSmoothingRampDownHysteresisSeconds: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "power_smoothing_ramp_down_hysteresis_seconds"),
			"Ramp down hysteresis in seconds for power smoothing",
			[]string{"id"}, nil,
		),
		GPUPowerSmoothingRampDownWattsPerSecond: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "power_smoothing_ramp_down_watts_per_second"),
			"Ramp down rate in watts per second for power smoothing",
			[]string{"id"}, nil,
		),
		GPUPowerSmoothingRampUpWattsPerSecond: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "power_smoothing_ramp_up_watts_per_second"),
			"Ramp up rate in watts per second for power smoothing",
			[]string{"id"}, nil,
		),
		GPUPowerSmoothingRemainingLifetimeCircuitryPercent: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "power_smoothing_remaining_lifetime_circuitry_percent"),
			"Remaining lifetime of the power smoothing circuitry in percent",
			[]string{"id"}, nil,
		),
		GPUPowerSmoothingTMPFloorPercent: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "power_smoothing_tmp_floor_percent"),
			"Current TMP floor percentage for power smoothing",
			[]string{"id"}, nil,
		),
		GPUPowerSmoothingTMPFloorWatts: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "power_smoothing_tmp_floor_watts"),
			"Current TMP floor in watts for power smoothing",
			[]string{"id"}, nil,
		),
		GPUPowerSmoothingTMPWatts: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "power_smoothing_tmp_watts"),
			"Current TMP in watts for power smoothing",
			[]string{"id"}, nil,
		),

		// NVLink Port Metrics
		GPUNVLinkStatus: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "nvlink_status"),
			"Status of the NVLink port (1=Up, 0=Down)",
			[]string{"id", "port"}, nil,
		),
		GPUNVLinkCurrentSpeedGbps: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "nvlink_current_speed_gbps"),
			"Current speed of the NVLink port in Gbps",
			[]string{"id", "port"}, nil,
		),
		GPUNVLinkRXBytes: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "nvlink_rx_bytes"),
			"Bytes received on the NVLink port",
			[]string{"id", "port"}, nil,
		),
		GPUNVLinkTXBytes: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "nvlink_tx_bytes"),
			"Bytes transmitted on the NVLink port",
			[]string{"id", "port"}, nil,
		),
		GPUNVLinkBitErrorRate: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "nvlink_bit_error_rate"),
			"Bit error rate of the NVLink port",
			[]string{"id", "port"}, nil,
		),
		GPUNVLinkLinkDownedCount: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "nvlink_link_downed_count"),
			"Number of times the NVLink port has gone down",
			[]string{"id", "port"}, nil,
		),
		GPUNVLinkSymbolErrors: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "nvlink_symbol_errors"),
			"Symbol errors on the NVLink port",
			[]string{"id", "port"}, nil,
		),
		GPUNVLinkRecoveryCount: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "nvlink_recovery_count"),
			"Link error recovery count on the NVLink port",
			[]string{"id", "port"}, nil,
		),
		GPUNVLinkRuntimeError: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "nvlink_runtime_error"),
			"NVLink runtime error on the port (1=error, 0=no error)",
			[]string{"id", "port"}, nil,
		),
		GPUNVLinkTrainingError: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "gpu", "nvlink_training_error"),
			"NVLink training error on the port (1=error, 0=no error)",
			[]string{"id", "port"}, nil,
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
	ch <- collector.GPUPowerLimitWatts
	ch <- collector.GPUCoreVoltageVolts
	ch <- collector.GPUMemoryCorrectableECCErrorCount
	ch <- collector.GPUMemoryUncorrectableECCErrorCount
	ch <- collector.GPUFP16ActivityPercent
	ch <- collector.GPUFP32ActivityPercent
	ch <- collector.GPUFP64ActivityPercent
	ch <- collector.GPUIntegerActivityUtilizationPercent
	ch <- collector.GPUNVLinkDataRxBandwidthGbps
	ch <- collector.GPUNVLinkDataTxBandwidthGbps
	ch <- collector.GPUTotalNVLinks
	ch <- collector.GPUEnergyJoules
	ch <- collector.GPUCacheCorrectableECCErrorCount
	ch <- collector.GPUCacheUncorrectableECCErrorCount
	ch <- collector.GPUNVDecUtilizationPercent
	ch <- collector.GPUNVJpgUtilizationPercent
	ch <- collector.GPUNVLinkGpuRawRxBandwidthGbps
	ch <- collector.GPUNVLinkGpuRawTxBandwidthGbps
	ch <- collector.GPUConventionalResetEntryCount
	ch <- collector.GPUConventionalResetExitCount
	ch <- collector.GPUFundamentalResetEntryCount
	ch <- collector.GPUFundamentalResetExitCount
	ch <- collector.GPUPFFLRResetEntryCount
	ch <- collector.GPUPFFLRResetExitCount
	ch <- collector.GPUPowerSmoothingSupported
	ch <- collector.GPUPowerSmoothingEnabled
	ch <- collector.GPUPowerSmoothingImmediateRampDown
	ch <- collector.GPUPowerSmoothingMaxAllowedTMPFloorPercent
	ch <- collector.GPUPowerSmoothingMinAllowedTMPFloorPercent
	ch <- collector.GPUPowerSmoothingRampDownHysteresisSeconds
	ch <- collector.GPUPowerSmoothingRampDownWattsPerSecond
	ch <- collector.GPUPowerSmoothingRampUpWattsPerSecond
	ch <- collector.GPUPowerSmoothingRemainingLifetimeCircuitryPercent
	ch <- collector.GPUPowerSmoothingTMPFloorPercent
	ch <- collector.GPUPowerSmoothingTMPFloorWatts
	ch <- collector.GPUPowerSmoothingTMPWatts

	ch <- collector.GPUNVLinkStatus
	ch <- collector.GPUNVLinkCurrentSpeedGbps
	ch <- collector.GPUNVLinkRXBytes
	ch <- collector.GPUNVLinkTXBytes
	ch <- collector.GPUNVLinkBitErrorRate
	ch <- collector.GPUNVLinkLinkDownedCount
	ch <- collector.GPUNVLinkSymbolErrors
	ch <- collector.GPUNVLinkRecoveryCount
	ch <- collector.GPUNVLinkRuntimeError
	ch <- collector.GPUNVLinkTrainingError
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
