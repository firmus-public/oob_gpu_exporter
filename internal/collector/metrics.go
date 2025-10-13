package collector

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

func gpuHealth2value(gpuHealth string) (bool, int) {
	switch gpuHealth {
	case "Critical":
		return true, 0
	case "Degraded":
		return true, 1
	case "OK":
		return true, 2
	case "Unknown":
		return true, 3
	default:
		return false, 0
	}
}
                
func gpuState2value(gpuState string) (bool, int) {
	switch gpuState {
	case "Available":
		return true, 0
	case "NotApplicable":
		return true, 1
	case "Unavailable":
		return true, 2
	default:
		return false, 0
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
		strings.TrimSpace(m.UUID),
	)
}

func (mc *Collector) NewGPUState(ch chan<- prometheus.Metric, m *DellVideoMember) {
	if ok, value := gpuState2value(m.GPUState); ok {
		ch <- prometheus.MustNewConstMetric(
			mc.GPUState,
			prometheus.GaugeValue,
			float64(value),
			m.Id,
			m.GPUState,
		)
	}
}

func (mc *Collector) NewGPUHealth(ch chan<- prometheus.Metric, m *DellVideoMember) {
	if ok, value := gpuHealth2value(m.GPUHealth); ok {
		ch <- prometheus.MustNewConstMetric(
			mc.GPUHealth,
			prometheus.GaugeValue,
			float64(value),
			m.Id,
			m.GPUHealth,
		)
	}
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

func (mc *Collector) NewGPUThrottleReasons(ch chan<- prometheus.Metric, v []string , id string) {
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

func (mc *Collector) NewGPUSMUtilizationPercent(ch chan<- prometheus.Metric, v int , id string) {
	ch <- prometheus.MustNewConstMetric(
		mc.GPUSMUtilizationPercent,
		prometheus.GaugeValue,
		float64(v),
		id,
	)
}

func (mc *Collector) NewGPUSMActivityPercent(ch chan<- prometheus.Metric, v float64 , id string) {
	ch <- prometheus.MustNewConstMetric(
		mc.GPUSMActivityPercent,
		prometheus.GaugeValue,
		v,
		id,
	)
}

func (mc *Collector) NewGPUSMOccupancyPercent(ch chan<- prometheus.Metric, v float64 , id string) {
	ch <- prometheus.MustNewConstMetric(
		mc.GPUSMOccupancyPercent,
		prometheus.GaugeValue,
		v,
		id,
	)
}

func (mc *Collector) NewGPUTensorCoreActivityPercent(ch chan<- prometheus.Metric, v float64 , id string) {
	ch <- prometheus.MustNewConstMetric(
		mc.GPUTensorCoreActivityPercent,
		prometheus.GaugeValue,
		v,
		id,
	)
}

func (mc *Collector) NewGPUHMMAUtilizationPercent(ch chan<- prometheus.Metric, v float64 , id string) {
	ch <- prometheus.MustNewConstMetric(
		mc.GPUHMMAUtilizationPercent,
		prometheus.GaugeValue,
		v,
		id,
	)
}

func (mc *Collector) NewGPUPCIeRawTxBandwidthGbps(ch chan<- prometheus.Metric, v float64 , id string) {
	ch <- prometheus.MustNewConstMetric(
		mc.GPUPCIeRawTxBandwidthGbps,
		prometheus.GaugeValue,
		v,
		id,
	)
}

func (mc *Collector) NewGPUPCIeRawRxBandwidthGbps(ch chan<- prometheus.Metric, v float64 , id string) {
	ch <- prometheus.MustNewConstMetric(
		mc.GPUPCIeRawRxBandwidthGbps,
		prometheus.GaugeValue,
		v,
		id,
	)
}

func (mc *Collector) NewGPUCurrentPCIeLinkSpeed(ch chan<- prometheus.Metric, v int , id string) {
	ch <- prometheus.MustNewConstMetric(
		mc.GPUCurrentPCIeLinkSpeed,
		prometheus.GaugeValue,
		float64(v),
		id,
	)
}

func (mc *Collector) NewGPUMaxSupportedPCIeLinkSpeed(ch chan<- prometheus.Metric, v int , id string) {
	ch <- prometheus.MustNewConstMetric(
		mc.GPUMaxSupportedPCIeLinkSpeed,
		prometheus.GaugeValue,
		float64(v),
		id,
	)
}

func (mc *Collector) NewGPUDRAMUtilizationPercent(ch chan<- prometheus.Metric, v float64 , id string) {
	ch <- prometheus.MustNewConstMetric(
		mc.GPUDRAMUtilizationPercent,
		prometheus.GaugeValue,
		v,
		id,
	)
}

func (mc *Collector) NewGPUPCIeCorrectableErrorCount(ch chan<- prometheus.Metric, v int , id string) {
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
