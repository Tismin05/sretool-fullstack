package collector

import (
	"context"
	"sretool-fullstack/internal/model"
)

// Collector 采集器接口
type Collector interface {
	Collect(ctx context.Context) (*model.Metrics, *model.CollectErrors)
}
