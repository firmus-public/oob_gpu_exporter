package collector

import (
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

func gpuHealth2value(gpuHealth string) int {
	switch strings.ToLower(gpuHealth) {
	case "critical":
		return 0
	case "degraded", "warning":
		return 1
	case "ok":
		return 2
	case "unknown":
		return 3
	default:
		return -1
	}
}

func gpuState2value(gpuState string) int {
	switch strings.ToLower(gpuState) {
	case "available", "enabled":
		return 0
	case "notapplicable":
		return 1
	case "unavailable", "disabled":
		return 2
	default:
		return -1
	}
}

func boardPowerSupplyStatus2value(boardPowerSupplyStatus string) (bool, int) {
	switch boardPowerSupplyStatus {
	case "NotApplicable":
		return true, 0
	case "SufficientPower":
		return true, 1
	case "UnderPowered":
		return true, 2
	default:
		return false, 0
	}
}

func powerBrakeStatus2value(powerBrakeStatus string) (bool, int) {
	switch powerBrakeStatus {
	case "NotApplicable":
		return true, 0
	case "Released":
		return true, 1
	case "Set":
		return true, 2
	default:
		return false, 0
	}
}

func thermalAlertStatus2value(thermalAlertStatus string) (bool, int) {
	switch thermalAlertStatus {
	case "NotApplicable":
		return true, 0
	case "NotPending":
		return true, 1
	case "Pending":
		return true, 2
	default:
		return false, 0
	}
}

func (mc *Collector) NewGPUCount(ch chan<- prometheus.Metric, count int) {
	ch <- prometheus.MustNewConstMetric(
		mc.GPUCount,
		prometheus.CounterValue,
		float64(count),
	)
}

func (mc *Collector) NewGPUInfo(ch chan<- prometheus.Metric, m *GPUInfo) {
	ch <- prometheus.MustNewConstMetric(
		mc.GPUInfo,
		prometheus.UntypedValue,
		1.0,
		m.Id,
		strings.TrimSpace(m.Manufacturer),
		strings.TrimSpace(m.Model),
		strings.TrimSpace(m.PartNumber),
		strings.TrimSpace(m.SerialNumber),
		strings.TrimSpace(m.GPUGUID),
		strconv.Itoa(m.Slot),
	)
}

func (mc *Collector) NewDellGPUState(ch chan<- prometheus.Metric, m *DellVideoMember) {
	value := gpuState2value(m.GPUState)
	ch <- prometheus.MustNewConstMetric(
		mc.GPUState,
		prometheus.GaugeValue,
		float64(value),
		m.Id,
		m.GPUState,
	)
}

func (mc *Collector) NewDellGPUHealth(ch chan<- prometheus.Metric, m *DellVideoMember) {
	value := gpuHealth2value(m.GPUHealth)
	ch <- prometheus.MustNewConstMetric(
		mc.GPUHealth,
		prometheus.GaugeValue,
		float64(value),
		m.Id,
		m.GPUHealth,
	)
}

func (mc *Collector) NewSupermicroGPUHealth(ch chan<- prometheus.Metric, m *PCIeDeviceResponse) {
	value := gpuHealth2value(m.Status.Health)
	ch <- prometheus.MustNewConstMetric(
		mc.GPUHealth,
		prometheus.GaugeValue,
		float64(value),
		m.ID,
		m.Status.Health,
	)
}

func (mc *Collector) NewSupermicroGPUState(ch chan<- prometheus.Metric, m *PCIeDeviceResponse) {
	value := gpuState2value(m.Status.State)
	ch <- prometheus.MustNewConstMetric(
		mc.GPUState,
		prometheus.GaugeValue,
		float64(value),
		m.ID,
		m.Status.State,
	)
}

func (mc *Collector) NewGBNVLGPUHealth(ch chan<- prometheus.Metric, id string, status string) {
	value := gpuHealth2value(status)
	ch <- prometheus.MustNewConstMetric(
		mc.GPUHealth,
		prometheus.GaugeValue,
		float64(value),
		id,
		status,
	)
}

func (mc *Collector) NewGBNVLGPUState(ch chan<- prometheus.Metric, id string, state string) {
	value := gpuState2value(state)
	ch <- prometheus.MustNewConstMetric(
		mc.GPUState,
		prometheus.GaugeValue,
		float64(value),
		id,
		state,
	)
}

func (mc *Collector) NewSmcGPUTemp(ch chan<- prometheus.Metric, id string, value float64) {
	ch <- prometheus.MustNewConstMetric(
		mc.GPUPrimaryGPUTemperatureCelsius,
		prometheus.GaugeValue,
		value,
		id,
	)
}

func (mc *Collector) NewSmcGPUMemoryTemp(ch chan<- prometheus.Metric, id string, value float64) {
	ch <- prometheus.MustNewConstMetric(
		mc.GPUMemoryTemperatureCelsius,
		prometheus.GaugeValue,
		value,
		id,
	)
}

func (mc *Collector) NewBoardPowerSupplyStatus(ch chan<- prometheus.Metric, m *DellGPUSensorMember) {
	if ok, value := boardPowerSupplyStatus2value(m.BoardPowerSupplyStatus); ok {
		ch <- prometheus.MustNewConstMetric(
			mc.GPUBoardPowerSupplyStatus,
			prometheus.GaugeValue,
			float64(value),
			m.Id,
			m.BoardPowerSupplyStatus,
		)
	}
}

func (mc *Collector) NewMemoryTemperatureCelsius(ch chan<- prometheus.Metric, m *DellGPUSensorMember) {
	ch <- prometheus.MustNewConstMetric(
		mc.GPUMemoryTemperatureCelsius,
		prometheus.GaugeValue,
		m.MemoryTemperatureCelsius,
		m.Id,
	)
}

func (mc *Collector) NewPowerBrakeStatus(ch chan<- prometheus.Metric, m *DellGPUSensorMember) {
	if ok, value := powerBrakeStatus2value(m.PowerBrakeStatus); ok {
		ch <- prometheus.MustNewConstMetric(
			mc.GPUPowerBrakeStatus,
			prometheus.GaugeValue,
			float64(value),
			m.Id,
			m.PowerBrakeStatus,
		)
	}
}

func (mc *Collector) NewPrimaryGPUTemperatureCelsius(ch chan<- prometheus.Metric, m *DellGPUSensorMember) {
	ch <- prometheus.MustNewConstMetric(
		mc.GPUPrimaryGPUTemperatureCelsius,
		prometheus.GaugeValue,
		m.PrimaryGPUTemperatureCelsius,
		m.Id,
	)
}

func (mc *Collector) NewThermalAlertStatus(ch chan<- prometheus.Metric, m *DellGPUSensorMember) {
	if ok, value := thermalAlertStatus2value(m.ThermalAlertStatus); ok {
		ch <- prometheus.MustNewConstMetric(
			mc.GPUThermalAlertStatus,
			prometheus.GaugeValue,
			float64(value),
			m.Id,
			m.ThermalAlertStatus,
		)
	}
}

func (mc *Collector) NewGPUOperatingSpeedMHz(ch chan<- prometheus.Metric, m *GPUMetrics) {
	if m.OperatingSpeedMHz == nil {
		return
	}
	ch <- prometheus.MustNewConstMetric(
		mc.GPUOperatingSpeedMHz,
		prometheus.GaugeValue,
		*m.OperatingSpeedMHz,
		m.Id,
	)
}

func (mc *Collector) NewGPUThrottleReasons(ch chan<- prometheus.Metric, v []string, id string) {
	for _, reason := range v {
		// TODO: default all possible reason metrics to zero when known
		ch <- prometheus.MustNewConstMetric(
			mc.GPUThrottleReason,
			prometheus.GaugeValue,
			1.0,
			id,
			reason,
		)
	}
}

func (mc *Collector) NewGPUSMUtilizationPercent(ch chan<- prometheus.Metric, v float64, id string) {
	ch <- prometheus.MustNewConstMetric(
		mc.GPUSMUtilizationPercent,
		prometheus.GaugeValue,
		v,
		id,
	)
}

func (mc *Collector) NewGPUSMActivityPercent(ch chan<- prometheus.Metric, v float64, id string) {
	ch <- prometheus.MustNewConstMetric(
		mc.GPUSMActivityPercent,
		prometheus.GaugeValue,
		v,
		id,
	)
}

func (mc *Collector) NewGPUSMOccupancyPercent(ch chan<- prometheus.Metric, v float64, id string) {
	ch <- prometheus.MustNewConstMetric(
		mc.GPUSMOccupancyPercent,
		prometheus.GaugeValue,
		v,
		id,
	)
}

func (mc *Collector) NewGPUTensorCoreActivityPercent(ch chan<- prometheus.Metric, v float64, id string) {
	ch <- prometheus.MustNewConstMetric(
		mc.GPUTensorCoreActivityPercent,
		prometheus.GaugeValue,
		v,
		id,
	)
}

func (mc *Collector) NewGPUHMMAUtilizationPercent(ch chan<- prometheus.Metric, v float64, id string) {
	ch <- prometheus.MustNewConstMetric(
		mc.GPUHMMAUtilizationPercent,
		prometheus.GaugeValue,
		v,
		id,
	)
}

func (mc *Collector) NewGPUPCIeRawTxBandwidthGbps(ch chan<- prometheus.Metric, v float64, id string) {
	ch <- prometheus.MustNewConstMetric(
		mc.GPUPCIeRawTxBandwidthGbps,
		prometheus.GaugeValue,
		v,
		id,
	)
}

func (mc *Collector) NewGPUPCIeRawRxBandwidthGbps(ch chan<- prometheus.Metric, v float64, id string) {
	ch <- prometheus.MustNewConstMetric(
		mc.GPUPCIeRawRxBandwidthGbps,
		prometheus.GaugeValue,
		v,
		id,
	)
}

func (mc *Collector) NewGPUCurrentPCIeLinkSpeed(ch chan<- prometheus.Metric, v int, id string) {
	ch <- prometheus.MustNewConstMetric(
		mc.GPUCurrentPCIeLinkSpeed,
		prometheus.GaugeValue,
		float64(v),
		id,
	)
}

func (mc *Collector) NewGPUMaxSupportedPCIeLinkSpeed(ch chan<- prometheus.Metric, v int, id string) {
	ch <- prometheus.MustNewConstMetric(
		mc.GPUMaxSupportedPCIeLinkSpeed,
		prometheus.GaugeValue,
		float64(v),
		id,
	)
}

func (mc *Collector) NewGPUDRAMUtilizationPercent(ch chan<- prometheus.Metric, v float64, id string) {
	ch <- prometheus.MustNewConstMetric(
		mc.GPUDRAMUtilizationPercent,
		prometheus.GaugeValue,
		v,
		id,
	)
}

func (mc *Collector) NewGPUPCIeCorrectableErrorCount(ch chan<- prometheus.Metric, v int, id string) {
	ch <- prometheus.MustNewConstMetric(
		mc.GPUPCIeCorrectableErrorCount,
		prometheus.CounterValue,
		float64(v),
		id,
	)
}

func (mc *Collector) NewGPUBandwidthPercent(ch chan<- prometheus.Metric, m *GPUMetrics) {
	if m.BandwidthPercent == nil {
		return
	}
	ch <- prometheus.MustNewConstMetric(
		mc.GPUBandwidthPercent,
		prometheus.GaugeValue,
		*m.BandwidthPercent,
		m.Id,
	)
}

func (mc *Collector) NewGPUMemoryOperatingSpeedMHz(ch chan<- prometheus.Metric, id string, m *GPUMemoryMetrics) {
	ch <- prometheus.MustNewConstMetric(
		mc.GPUMemoryOperatingSpeedMHz,
		prometheus.GaugeValue,
		m.OperatingSpeedMHz,
		id,
	)
}

func (mc *Collector) NewGPUMemoryBandwidthPercent(ch chan<- prometheus.Metric, id string, m *GPUMemoryMetrics) {
	ch <- prometheus.MustNewConstMetric(
		mc.GPUMemoryBandwidthPercent,
		prometheus.GaugeValue,
		m.BandwidthPercent,
		id,
	)
}

func (mc *Collector) NewGPUConsumedPowerWatt(ch chan<- prometheus.Metric, m *GPUMetrics) {
	ch <- prometheus.MustNewConstMetric(
		mc.GPUConsumedPowerWatt,
		prometheus.GaugeValue,
		m.ConsumedPowerWatt,
		m.Id,
	)
}

func (mc *Collector) NewGPUPowerLimitWatts(ch chan<- prometheus.Metric, id string, value float64) {
	ch <- prometheus.MustNewConstMetric(
		mc.GPUPowerLimitWatts,
		prometheus.GaugeValue,
		value,
		id,
	)
}

func (mc *Collector) NewGPUCoreVoltageVolts(ch chan<- prometheus.Metric, id string, value float64) {
	ch <- prometheus.MustNewConstMetric(
		mc.GPUCoreVoltageVolts,
		prometheus.GaugeValue,
		value,
		id,
	)
}

func (mc *Collector) NewGPUMemoryECCErrors(ch chan<- prometheus.Metric, id string, correctable int, uncorrectable int) {
	ch <- prometheus.MustNewConstMetric(
		mc.GPUMemoryCorrectableECCErrorCount,
		prometheus.CounterValue,
		float64(correctable),
		id,
	)
	ch <- prometheus.MustNewConstMetric(
		mc.GPUMemoryUncorrectableECCErrorCount,
		prometheus.CounterValue,
		float64(uncorrectable),
		id,
	)
}

func (mc *Collector) NewGPUPowerSmoothingMetrics(ch chan<- prometheus.Metric, id string, m *PowerSmoothing) {
	ch <- prometheus.MustNewConstMetric(mc.GPUPowerSmoothingSupported, prometheus.GaugeValue, boolToFloat(m.PowerSmoothingSupported), id)
	ch <- prometheus.MustNewConstMetric(mc.GPUPowerSmoothingEnabled, prometheus.GaugeValue, boolToFloat(m.Enabled), id)
	ch <- prometheus.MustNewConstMetric(mc.GPUPowerSmoothingImmediateRampDown, prometheus.GaugeValue, boolToFloat(m.ImmediateRampDown), id)

	ch <- prometheus.MustNewConstMetric(mc.GPUPowerSmoothingMaxAllowedTMPFloorPercent, prometheus.GaugeValue, m.MaxAllowedTMPFloorPercent, id)
	ch <- prometheus.MustNewConstMetric(mc.GPUPowerSmoothingMinAllowedTMPFloorPercent, prometheus.GaugeValue, m.MinAllowedTMPFloorPercent, id)
	ch <- prometheus.MustNewConstMetric(mc.GPUPowerSmoothingRampDownHysteresisSeconds, prometheus.GaugeValue, m.RampDownHysteresisSeconds, id)
	ch <- prometheus.MustNewConstMetric(mc.GPUPowerSmoothingRampDownWattsPerSecond, prometheus.GaugeValue, m.RampDownWattsPerSecond, id)
	ch <- prometheus.MustNewConstMetric(mc.GPUPowerSmoothingRampUpWattsPerSecond, prometheus.GaugeValue, m.RampUpWattsPerSecond, id)
	ch <- prometheus.MustNewConstMetric(mc.GPUPowerSmoothingRemainingLifetimeCircuitryPercent, prometheus.GaugeValue, m.RemainingLifetimeCircuitryPercent, id)
	ch <- prometheus.MustNewConstMetric(mc.GPUPowerSmoothingTMPFloorPercent, prometheus.GaugeValue, m.TMPFloorPercent, id)
	ch <- prometheus.MustNewConstMetric(mc.GPUPowerSmoothingTMPFloorWatts, prometheus.GaugeValue, m.TMPFloorWatts, id)
	ch <- prometheus.MustNewConstMetric(mc.GPUPowerSmoothingTMPWatts, prometheus.GaugeValue, m.TMPWatts, id)
}

func (mc *Collector) NewGPUNVLinkPortMetrics(ch chan<- prometheus.Metric, gpuID string, portID string, portResp *PortResponse, metricsResp *PortMetricsResponse) {
	status := 0.0
	if portResp.Status.State == "Enabled" && portResp.LinkStatus == "LinkUp" {
		status = 1.0
	}
	ch <- prometheus.MustNewConstMetric(mc.GPUNVLinkStatus, prometheus.GaugeValue, status, gpuID, portID)
	ch <- prometheus.MustNewConstMetric(mc.GPUNVLinkCurrentSpeedGbps, prometheus.GaugeValue, portResp.CurrentSpeedGbps, gpuID, portID)

	ch <- prometheus.MustNewConstMetric(mc.GPUNVLinkRXBytes, prometheus.CounterValue, float64(metricsResp.RXBytes), gpuID, portID)
	ch <- prometheus.MustNewConstMetric(mc.GPUNVLinkTXBytes, prometheus.CounterValue, float64(metricsResp.TXBytes), gpuID, portID)

	nvidia := metricsResp.Oem.Nvidia
	ch <- prometheus.MustNewConstMetric(mc.GPUNVLinkBitErrorRate, prometheus.GaugeValue, nvidia.BitErrorRate, gpuID, portID)
	ch <- prometheus.MustNewConstMetric(mc.GPUNVLinkLinkDownedCount, prometheus.CounterValue, float64(nvidia.LinkDownedCount), gpuID, portID)
	ch <- prometheus.MustNewConstMetric(mc.GPUNVLinkSymbolErrors, prometheus.CounterValue, float64(nvidia.SymbolErrors), gpuID, portID)
	ch <- prometheus.MustNewConstMetric(mc.GPUNVLinkRecoveryCount, prometheus.CounterValue, float64(nvidia.LinkErrorRecoveryCount), gpuID, portID)

	// NVLink error flags
	if nvidia.NVLinkErrors != nil {
		ch <- prometheus.MustNewConstMetric(mc.GPUNVLinkRuntimeError, prometheus.GaugeValue, boolToFloat(nvidia.NVLinkErrors.RuntimeError), gpuID, portID)
		ch <- prometheus.MustNewConstMetric(mc.GPUNVLinkTrainingError, prometheus.GaugeValue, boolToFloat(nvidia.NVLinkErrors.TrainingError), gpuID, portID)
	}
}

func boolToFloat(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}

func (mc *Collector) NewGPUGaugeMetric(ch chan<- prometheus.Metric, desc *prometheus.Desc, id string, value float64) {
	ch <- prometheus.MustNewConstMetric(
		desc,
		prometheus.GaugeValue,
		value,
		id,
	)
}

func (mc *Collector) NewGPUCounterMetric(ch chan<- prometheus.Metric, desc *prometheus.Desc, id string, value int) {
	ch <- prometheus.MustNewConstMetric(
		desc,
		prometheus.CounterValue,
		float64(value),
		id,
	)
}

func (mc *Collector) NewGPUTotalNVLinks(ch chan<- prometheus.Metric, id string, count int) {
	ch <- prometheus.MustNewConstMetric(
		mc.GPUTotalNVLinks,
		prometheus.GaugeValue,
		float64(count),
		id,
	)
}

func (mc *Collector) NewGPUEnergyJoules(ch chan<- prometheus.Metric, id string, value float64) {
	ch <- prometheus.MustNewConstMetric(
		mc.GPUEnergyJoules,
		prometheus.CounterValue,
		value,
		id,
	)
}

func (mc *Collector) NewGPUCacheECCErrors(ch chan<- prometheus.Metric, id string, correctable int, uncorrectable int) {
	ch <- prometheus.MustNewConstMetric(
		mc.GPUCacheCorrectableECCErrorCount,
		prometheus.CounterValue,
		float64(correctable),
		id,
	)
	ch <- prometheus.MustNewConstMetric(
		mc.GPUCacheUncorrectableECCErrorCount,
		prometheus.CounterValue,
		float64(uncorrectable),
		id,
	)
}

func (mc *Collector) NewGPUResetMetrics(ch chan<- prometheus.Metric, id string, m *ProcessorResetMetrics) {
	ch <- prometheus.MustNewConstMetric(mc.GPUConventionalResetEntryCount, prometheus.CounterValue, float64(m.ConventionalResetEntryCount), id)
	ch <- prometheus.MustNewConstMetric(mc.GPUConventionalResetExitCount, prometheus.CounterValue, float64(m.ConventionalResetExitCount), id)
	ch <- prometheus.MustNewConstMetric(mc.GPUFundamentalResetEntryCount, prometheus.CounterValue, float64(m.FundamentalResetEntryCount), id)
	ch <- prometheus.MustNewConstMetric(mc.GPUFundamentalResetExitCount, prometheus.CounterValue, float64(m.FundamentalResetExitCount), id)
	ch <- prometheus.MustNewConstMetric(mc.GPUPFFLRResetEntryCount, prometheus.CounterValue, float64(m.PF_FLR_ResetEntryCount), id)
	ch <- prometheus.MustNewConstMetric(mc.GPUPFFLRResetExitCount, prometheus.CounterValue, float64(m.PF_FLR_ResetExitCount), id)
}
