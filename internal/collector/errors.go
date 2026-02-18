package collector

// CollectErrors 采集错误聚合结构体
type CollectErrors struct {
	CPU  []error
	Mem  []error
	Disk []error
	Net  []error
}

// HasError 检查是否有错误
func (e *CollectErrors) HasError() bool {
	if e == nil {
		return false
	}
	return len(e.CPU)+len(e.Mem)+len(e.Disk)+len(e.Net) > 0
}
