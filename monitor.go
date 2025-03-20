package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

var NET_FORMER []net.IOCountersStat = nil
var IO_FORMER map[string]interface{} = nil
var CPU_FORMER []cpu.TimesStat = nil

func getThroughput() string {
	counters, _ := net.IOCounters(false)
	rx := float32(counters[0].BytesRecv) / 1024 / 1024 / 1024
	tx := float32(counters[0].BytesSent) / 1024 / 1024 / 1024

	throughput := ""
	if rx > 1024 {
		rx = rx / 1024
		throughput += fmt.Sprintf("↓%.2f TB / ", rx)
	} else {
		throughput += fmt.Sprintf("↓%.2f GB / ", rx)
	}
	if tx > 1024 {
		tx = tx / 1024
		throughput += fmt.Sprintf("↑%.2f TB", tx)
	} else {
		throughput += fmt.Sprintf("↑%.2f GB", tx)
	}

	logMessage(DEBUG, throughput)
	return throughput
}

func getProcessNum() string {
	// Get process number
	processes, _ := process.Processes()
	logMessage(DEBUG, fmt.Sprintf("%v\n", len(processes)))
	return fmt.Sprintf("%v", len(processes))
}

func getDiskInfo(exclude ...string) string {
	// Set default value for exclude if not provided
	excludeOpts := ""
	if len(exclude) > 0 {
		excludeOpts = exclude[0]
	}
	excludeFS := ""
	if len(exclude) > 1 {
		excludeFS = exclude[1]
	}
	excludeDisk := ""
	if len(exclude) > 2 {
		excludeDisk = exclude[1]
	}
	// Get disk usage
	partitions, _ := disk.Partitions(false)

	disks := make(map[string]interface{})

	for _, partition := range partitions {
		usage, _ := disk.Usage(partition.Mountpoint)

		if excludeDisk != "" && strings.Contains(strings.ToLower(excludeDisk), strings.ToLower(partition.Device)) {
			continue
		}

		continueFlag := false
		for _, opts := range partition.Opts {
			if strings.Contains(strings.ToLower(excludeOpts), strings.ToLower(opts)) ||
				strings.Contains(strings.ToLower(excludeFS), strings.ToLower(partition.Fstype)) {
				continueFlag = true
				break
			}
		}
		if continueFlag {
			continue
		}

		disk := map[string]interface{}{
			"total":   fmt.Sprintf("%.2f", float32(usage.Total)/1024/1024),
			"used":    fmt.Sprintf("%.2f", float32(usage.Used)/1024/1024),
			"free":    fmt.Sprintf("%.2f", float32(usage.Free)/1024/1024),
			"percent": fmt.Sprintf("%.2f", usage.UsedPercent),
		}
		disks[partition.Mountpoint] = disk
	}

	data, _ := json.Marshal(disks)
	logMessage(DEBUG, string(data))
	return string(data)
}

func getIOSum() map[string]interface{} {
	counters, _ := disk.IOCounters()
	io := map[string]interface{}{
		"read": map[string]interface{}{
			"bytes": uint64(0),
			"count": uint64(0),
			"time":  uint64(0),
		},
		"write": map[string]interface{}{
			"bytes": uint64(0),
			"count": uint64(0),
			"time":  uint64(0),
		},
	}
	for _, counter := range counters {
		io["read"].(map[string]interface{})["bytes"] = io["read"].(map[string]interface{})["bytes"].(uint64) + counter.ReadBytes
		io["read"].(map[string]interface{})["count"] = io["read"].(map[string]interface{})["count"].(uint64) + counter.ReadCount
		io["read"].(map[string]interface{})["time"] = io["read"].(map[string]interface{})["time"].(uint64) + counter.ReadTime
		io["write"].(map[string]interface{})["bytes"] = io["write"].(map[string]interface{})["bytes"].(uint64) + counter.WriteBytes
		io["write"].(map[string]interface{})["count"] = io["write"].(map[string]interface{})["count"].(uint64) + counter.WriteCount
		io["write"].(map[string]interface{})["time"] = io["write"].(map[string]interface{})["time"].(uint64) + counter.WriteTime
	}

	return io
}

func getIO() string {
	// Get disk io counters
	counters := getIOSum()

	if IO_FORMER == nil {
		IO_FORMER = counters
	}

	io := map[string]interface{}{
		"read": map[string]interface{}{
			"bytes": func() uint64 {
				if counters["read"].(map[string]interface{})["bytes"].(uint64) > IO_FORMER["read"].(map[string]interface{})["bytes"].(uint64) {
					return counters["read"].(map[string]interface{})["bytes"].(uint64) - IO_FORMER["read"].(map[string]interface{})["bytes"].(uint64)
				}
				return 0
			}(),
			"count": func() uint64 {
				if counters["read"].(map[string]interface{})["count"].(uint64) > IO_FORMER["read"].(map[string]interface{})["count"].(uint64) {
					return counters["read"].(map[string]interface{})["count"].(uint64) - IO_FORMER["read"].(map[string]interface{})["count"].(uint64)
				}
				return 0
			}(),
			"time": func() uint64 {
				if counters["read"].(map[string]interface{})["time"].(uint64) > IO_FORMER["read"].(map[string]interface{})["time"].(uint64) {
					return counters["read"].(map[string]interface{})["time"].(uint64) - IO_FORMER["read"].(map[string]interface{})["time"].(uint64)
				}
				return 0
			}(),
		},
		"write": map[string]interface{}{
			"bytes": func() uint64 {
				if counters["write"].(map[string]interface{})["bytes"].(uint64) > IO_FORMER["write"].(map[string]interface{})["bytes"].(uint64) {
					return counters["write"].(map[string]interface{})["bytes"].(uint64) - IO_FORMER["write"].(map[string]interface{})["bytes"].(uint64)
				}
				return 0
			}(),
			"count": func() uint64 {
				if counters["write"].(map[string]interface{})["count"].(uint64) > IO_FORMER["write"].(map[string]interface{})["count"].(uint64) {
					return counters["write"].(map[string]interface{})["count"].(uint64) - IO_FORMER["write"].(map[string]interface{})["count"].(uint64)
				}
				return 0
			}(),
			"time": func() uint64 {
				if counters["write"].(map[string]interface{})["time"].(uint64) > IO_FORMER["write"].(map[string]interface{})["time"].(uint64) {
					return counters["write"].(map[string]interface{})["time"].(uint64) - IO_FORMER["write"].(map[string]interface{})["time"].(uint64)
				}
				return 0
			}(),
		},
	}

	IO_FORMER = counters
	data, _ := json.Marshal(io)
	return string(data)
}

func getNetwork() string {
	// Get network io counters
	var network map[string]interface{}
	counters, _ := net.IOCounters(false)

	if NET_FORMER == nil {
		NET_FORMER = counters
	}
	network = map[string]interface{}{
		"RX": map[string]interface{}{
			"bytes": func() uint64 {
				if counters[0].BytesRecv > NET_FORMER[0].BytesRecv {
					return counters[0].BytesRecv - NET_FORMER[0].BytesRecv
				}
				return 0
			}(),
			"packets": func() uint64 {
				if counters[0].PacketsRecv > NET_FORMER[0].PacketsRecv {
					return counters[0].PacketsRecv - NET_FORMER[0].PacketsRecv
				}
				return 0
			}(),
		},
		"TX": map[string]interface{}{
			"bytes": func() uint64 {
				if counters[0].BytesSent > NET_FORMER[0].BytesSent {
					return counters[0].BytesSent - NET_FORMER[0].BytesSent
				}
				return 0
			}(),
			"packets": func() uint64 {
				if counters[0].PacketsSent > NET_FORMER[0].PacketsSent {
					return counters[0].PacketsSent - NET_FORMER[0].PacketsSent
				}
				return 0
			}(),
		},
	}

	NET_FORMER = counters
	data, _ := json.Marshal(network)
	logMessage(DEBUG, fmt.Sprint(string(data)))
	return string(data)
}

func getCPUInfo() string {
	info, _ := cpu.Info()

	logMessage(DEBUG, fmt.Sprintf("%vx %v", info[0].Cores, info[0].ModelName))
	return fmt.Sprintf("%vx %v", info[0].Cores, info[0].ModelName)
}

func getUptime() string {
	bootTime, _ := host.BootTime()
	uptime := uint64(time.Now().Unix()) - bootTime

	delta := time.Duration(uptime) * time.Second

	days := int(delta.Hours() / 24)
	hours := int(delta.Hours()) % 24
	minutes := int(delta.Minutes()) % 60
	seconds := int(delta.Seconds()) % 60

	logMessage(DEBUG, fmt.Sprintf("%d Days %02d:%02d:%02d", days, hours, minutes, seconds))
	return fmt.Sprintf("%d Days %02d:%02d:%02d", days, hours, minutes, seconds)
}

func getConnections() string {
	tcp, _ := net.Connections("tcp")
	udp, _ := net.Connections("udp")

	logMessage(DEBUG, fmt.Sprintf("TCP: %v, UDP: %v\n", len(tcp), len(udp)))
	return fmt.Sprintf("TCP: %v, UDP: %v", len(tcp), len(udp))
}

func getSysVersion() string {
	info, _ := host.Info()

	logMessage(DEBUG, fmt.Sprintf("%v %v %v\n", info.OS, info.Platform, info.PlatformVersion))
	return fmt.Sprintf("%v %v %v", info.OS, info.Platform, info.PlatformVersion)
}

func getLoadAvg() string {
	// Get Load avg
	loadAvg, _ := load.Avg()

	logMessage(DEBUG, fmt.Sprintf("%.2f, %.2f, %.2f\n", loadAvg.Load1, loadAvg.Load5, loadAvg.Load15))
	return fmt.Sprintf("%.2f, %.2f, %.2f", loadAvg.Load1, loadAvg.Load5, loadAvg.Load15)
}

func getLoad() string {
	// Get CPU usage
	cpuTimes, _ := cpu.Times(false)

	if len(cpuTimes) == 0 {
		return "{}"
	}

	if CPU_FORMER == nil {
		CPU_FORMER = cpuTimes
		return "{}"
	}

	total := (cpuTimes[0].User + cpuTimes[0].System + cpuTimes[0].Idle + cpuTimes[0].Nice +
		cpuTimes[0].Iowait + cpuTimes[0].Irq + cpuTimes[0].Softirq + cpuTimes[0].Steal +
		cpuTimes[0].Guest + cpuTimes[0].GuestNice) - (CPU_FORMER[0].User + CPU_FORMER[0].System + CPU_FORMER[0].Idle + CPU_FORMER[0].Nice +
		CPU_FORMER[0].Iowait + CPU_FORMER[0].Irq + CPU_FORMER[0].Softirq + CPU_FORMER[0].Steal +
		CPU_FORMER[0].Guest + CPU_FORMER[0].GuestNice)

	percentages := map[string]string{
		"user":       fmt.Sprintf("%.2f", ((cpuTimes[0].User-CPU_FORMER[0].User)/total)*100),
		"system":     fmt.Sprintf("%.2f", ((cpuTimes[0].System-CPU_FORMER[0].System)/total)*100),
		"idle":       fmt.Sprintf("%.2f", ((cpuTimes[0].Idle-CPU_FORMER[0].Idle)/total)*100),
		"nice":       fmt.Sprintf("%.2f", ((cpuTimes[0].Nice-CPU_FORMER[0].Nice)/total)*100),
		"iowait":     fmt.Sprintf("%.2f", ((cpuTimes[0].Iowait-CPU_FORMER[0].Iowait)/total)*100),
		"irq":        fmt.Sprintf("%.2f", ((cpuTimes[0].Irq-CPU_FORMER[0].Irq)/total)*100),
		"softirq":    fmt.Sprintf("%.2f", ((cpuTimes[0].Softirq-CPU_FORMER[0].Softirq)/total)*100),
		"steal":      fmt.Sprintf("%.2f", ((cpuTimes[0].Steal-CPU_FORMER[0].Steal)/total)*100),
		"guest":      fmt.Sprintf("%.2f", ((cpuTimes[0].Guest-CPU_FORMER[0].Guest)/total)*100),
		"guest_nice": fmt.Sprintf("%.2f", ((cpuTimes[0].GuestNice-CPU_FORMER[0].GuestNice)/total)*100),
	}

	data, _ := json.Marshal(percentages)
	CPU_FORMER = cpuTimes

	logMessage(DEBUG, string(data))
	return string(data)
}

func getMemInfo() string {

	// Get memory usage
	memory, _ := mem.VirtualMemory()

	// Get swap usage
	swapMemory, _ := mem.SwapMemory()

	info := map[string]interface{}{
		"Mem": map[string]interface{}{
			"total":   fmt.Sprintf("%.2f", float32(memory.Total)/1024/1024),
			"used":    fmt.Sprintf("%.2f", float32(memory.Used)/1024/1024),
			"free":    fmt.Sprintf("%.2f", float32(memory.Free)/1024/1024),
			"percent": fmt.Sprintf("%.2f", memory.UsedPercent),
		},
		"Swap": map[string]interface{}{
			"total":   fmt.Sprintf("%.2f", float32(swapMemory.Total)/1024/1024),
			"used":    fmt.Sprintf("%.2f", float32(swapMemory.Used)/1024/1024),
			"free":    fmt.Sprintf("%.2f", float32(swapMemory.Free)/1024/1024),
			"percent": fmt.Sprintf("%.2f", swapMemory.UsedPercent),
		},
	}

	data, _ := json.Marshal(info)

	logMessage(DEBUG, string(data))
	return string(data)
}
