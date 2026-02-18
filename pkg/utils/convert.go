package utils

import "math"

type UnitsConversation struct{}

func ByteToKB(unitByte uint64) float64 {
	return math.Round(float64(unitByte)/1024*100) / 100
}

func (c UnitsConversation) ByteToMB(unitByte uint64) float64 {
	return math.Round(float64(unitByte)/1024/1024*100) / 100
}

func (c UnitsConversation) ByteToGB(unitByte uint64) float64 {
	return math.Round(float64(unitByte)/1024/1024/1024*100) / 100
}

func (c UnitsConversation) ByteToTB(unitByte uint64) float64 {
	return math.Round(float64(unitByte)/1024/1024/1024/1024*100) / 100
}
func KBtoByte(unitKB uint64) uint64 {
	return unitKB * 1024
}

func Pct(used, total uint64) float64 {
	if total == 0 {
		return 0
	}
	return float64(used) / float64(total) * 100
}
