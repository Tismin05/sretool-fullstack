package model

// Alert 告警结构体
type Alert struct {
	Level     string  `json:"level"`
	Category  string  `json:"category"`
	Metric    string  `json:"metric"`
	Message   string  `json:"message"`
	Value     float64 `json:"value"`
	Threshold float64 `json:"threshold"`
	Unit      string  `json:"unit"`
	Host      string  `json:"host"`
}
