package model

import "time"

// DiagnosticResult 诊断结果的标准化数据

// 枚举诊断结果的状态
type DiagnosticStatus string

const (
	StatusPass DiagnosticStatus = "PASS"
	StatusWarn DiagnosticStatus = "WARN"
	StatusFail DiagnosticStatus = "FAIL"
	StatusInfo DiagnosticStatus = "INFO"
)

// 数据结构定义
type DiagnosticResult struct {
	Title     string           `json:"title"`
	Timestamp time.Time        `json:"timestamp"` // 诊断执行时间
	Success   bool             `json:"success"`   // 操作成功与否
	Item      []DiagnosticItem `json:"item"`      // 具体的诊断结果
}

// 具体诊断项的结果结构体定义
type DiagnosticItem struct {
	Name        string           `json:"name"`        // 诊断项目的名称，如Check CPU info
	Description string           `json:"description"` // 注释
	Status      DiagnosticStatus `json:"status"`      // 诊断的结果FAIL/PASS/WARN
	Message     string           `json:"message"`     // 详细的输出信息是什么
	Errors      []string         `json:"errors"`      // 记录具体错误
	Suggestions []string         `json:"suggestions"` // 记录可能的操作建议
}

// 快速构建一个初始化的诊断结果对象
func NewDiagnosticResult(title string) *DiagnosticResult {
	return &DiagnosticResult{
		Title:     title,
		Timestamp: time.Now(),
		Success:   true,
		Item:      []DiagnosticItem{},
	}
}

func (r *DiagnosticResult) AddItem(name string, status DiagnosticStatus, message string, suggestions []string) {
	item := DiagnosticItem{
		Name:        name,
		Status:      status,
		Message:     message,
		Suggestions: suggestions,
	}
	r.Item = append(r.Item, item)

	if status == StatusFail {
		r.Success = false
	}
}
