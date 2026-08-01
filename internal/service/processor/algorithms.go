package processor

import "math"

const (
	VoltageMin         = 200.0
	VoltageMax         = 250.0
	AnomalyPowerFactor = 1.8 // power > avg*factor -> anomaly
)

type DetectionResult struct {
	AnomalyScore float64
	IsAnomaly    bool
	AlertType    string
	Message      string
}

func DetectAnomaly(voltage, power, avgPower float64, samples int64) DetectionResult {
	score := 0.0
	msg := ""
	alertType := ""

	if voltage < VoltageMin || voltage > VoltageMax {
		score = math.Max(score, 0.9)
		alertType = "voltage_out_of_range"
		msg = "voltage out of allowed range"
	}

	if samples >= 5 && avgPower > 0 && power > avgPower*AnomalyPowerFactor {
		spike := math.Min(1.0, (power/avgPower)/AnomalyPowerFactor)
		if spike > score {
			score = spike
			alertType = "power_spike"
			msg = "power spike vs sliding window"
		}
	}

	return DetectionResult{
		AnomalyScore: score,
		IsAnomaly:    score >= 0.7,
		AlertType:    alertType,
		Message:      msg,
	}
}
