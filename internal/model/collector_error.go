package model

// CollectErrors aggregates partial collection errors by subsystem.
type CollectErrors struct {
	CPU  []error
	Mem  []error
	Disk []error
	Net  []error
}

func (e *CollectErrors) HasError() bool {
	if e == nil {
		return false
	}
	return len(e.CPU)+len(e.Mem)+len(e.Disk)+len(e.Net) > 0
}
