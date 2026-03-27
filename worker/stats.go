package worker

import (
	"log"

	"github.com/c9s/goprocinfo/linux"
)

type Stats struct {
	MemStats       *linux.MemInfo
	DiskStats      *linux.Disk
	CpuStats       *linux.CPUStat
	LoadStats      *linux.LoadAvg
	TotalTaskCount int
}

func GetStats() *Stats {
	return &Stats{
		MemStats:  GetMemoryInfo(),
		DiskStats: GetDiskInfo(),
		CpuStats:  GetCpuInfo(),
		LoadStats: GetLoadAvg(),
	}
}

// MemTotal - mem total usage in kb
func (s *Stats) MemTotal() uint64 {
	return s.MemStats.MemTotal
}

func (s *Stats) MemAvalibale() uint64 {
	return s.MemStats.MemAvailable
}

func (s *Stats) MemUsed() uint64 {
	return s.MemTotal() - s.MemAvalibale()
}

func (s *Stats) MemPercent() uint64 {
	return s.MemAvalibale() / s.MemTotal()
}

// Disk gopprocinfos's
func (s *Stats) DiskTotal() uint64 {
	return s.DiskStats.All
}

func (s *Stats) DiskFree() uint64 {
	return s.DiskStats.Free
}

func (s *Stats) DiskUsed() uint64 {
	return s.DiskTotal() - s.DiskFree()
}

//Cpu gopprocinfos's

func (s *Stats) CpuUsage() float64 {
	idle := s.CpuStats.Idle + s.CpuStats.IOWait
	nonIdle := s.CpuStats.User + s.CpuStats.Nice + s.CpuStats.System + s.CpuStats.IRQ + s.CpuStats.SoftIRQ + s.CpuStats.Steal

	total := idle + nonIdle

	if total == 0 {
		return 0
	}

	return (float64(total) - float64(idle)) / float64(total)
}

func GetMemoryInfo() *linux.MemInfo {
	mem, err := linux.ReadMemInfo("proc/meminfo")
	if err != nil {
		log.Printf("Error  reading from /proc/meminfo")
		return &linux.MemInfo{}
	}

	return mem
}

func GetDiskInfo() *linux.Disk {
	disk, err := linux.ReadDisk("/")
	if err != nil {
		log.Printf("Error reading from /")
		return &linux.Disk{}
	}
	return disk

}

func GetCpuInfo() *linux.CPUStat {
	cpu, err := linux.ReadStat("/proc/stat")
	if err != nil {
		log.Printf("Error reading from /proc/stat")
		return &linux.CPUStat{}
	}
	return &cpu.CPUStatAll
}

func GetLoadAvg() *linux.LoadAvg {
	load, err := linux.ReadLoadAvg("/proc/loadavg")
	if err != nil {
		log.Printf("Error read from /proc/loadavg")
		return &linux.LoadAvg{}
	}
	return load
}
