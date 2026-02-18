package engine

import (
	"time"
	"tisminSRETool/internal/model"
)

func CalculateRate(prev, cur model.Metrics, interval time.Duration) model.Metrics {
	if interval.Seconds() <= 0 {
		return cur
	}
	seconds := interval.Seconds()
	res := cur
	var used uint64

	// CPU相关信息计算
	diffTotal := cur.CPU.TotalTicks - prev.CPU.TotalTicks
	diffIdle := cur.CPU.IdleTicks - prev.CPU.IdleTicks
	if diffTotal > 0 && diffIdle > 0 && diffIdle > diffTotal {
		used = diffTotal - diffIdle
	}
	res.CPU.UsagePercent = float64(used) / float64(diffTotal) * 100

	// 内存相关信息计算

	// 磁盘相关信息计算
	for i := range res.Disk {
		if i < len(prev.Disk) && res.Disk[i].Device == prev.Disk[i].Device {
			res.Disk[i].ReadSpeed = float64(res.Disk[i].Read-prev.Disk[i].Read) / seconds
			res.Disk[i].WriteSpeed = float64(res.Disk[i].Write-prev.Disk[i].Write) / seconds

			diffReadIOs := res.Disk[i].Read - prev.Disk[i].Read
			diffWriteIOs := res.Disk[i].Write - prev.Disk[i].Write
			diffIOTime := float64(res.Disk[i].IOQueueTime - prev.Disk[i].IOQueueTime)
			totalIOs := diffReadIOs + diffWriteIOs

			if totalIOs > 0 && diffIOTime > 0 {
				res.Disk[i].Await = diffIOTime / float64(diffReadIOs)
				totalMs := seconds * 1000
				res.Disk[i].Util = diffIOTime / totalMs * 100
			}
		}
	}

	// 网络相关信息计算
	for i := range res.Net {
		if i < len(prev.Net) && res.Net[i].Name == prev.Net[i].Name {
			res.Net[i].RxSpeed = float64(res.Net[i].RxBytes-prev.Net[i].RxBytes) / seconds
			res.Net[i].TxSpeed = float64(res.Net[i].TxBytes-prev.Net[i].TxBytes) / seconds
		}
	}
	return res
}
