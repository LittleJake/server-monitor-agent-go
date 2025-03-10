package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
)

func get_cpu() string {
	info, _ := cpu.Info()

	fmt.Printf("%vx %v\n", info[0].Cores, info[0].ModelName)
	return fmt.Sprintf("%vx %v", info[0].Cores, info[0].ModelName)
}

func get_uptime() string {
	bootTime, _ := host.BootTime()
	uptime := uint64(time.Now().Unix()) - bootTime

	delta := time.Duration(uptime) * time.Second

	days := int(delta.Hours() / 24)
	hours := int(delta.Hours()) % 24
	minutes := int(delta.Minutes()) % 60
	seconds := int(delta.Seconds()) % 60

	fmt.Printf("%d Days %02d:%02d:%02d\n", days, hours, minutes, seconds)
	return fmt.Sprintf("%d Days %02d:%02d:%02d", days, hours, minutes, seconds)
}

func get_connections() string {
	tcp, _ := net.Connections("tcp")
	udp, _ := net.Connections("udp")

	fmt.Printf("TCP: %v, UDP: %v\n", len(tcp), len(udp))
	return fmt.Sprintf("TCP: %v, UDP: %v", len(tcp), len(udp))
}

func get_sys_version() string {
	info, _ := host.Info()

	fmt.Printf("%v %v\n", info.Platform, info.PlatformVersion)
	return fmt.Sprintf("%v %v", info.Platform, info.PlatformVersion)
}

func disk_io_counters() {

}

func get_load_avg() string {
	// Get Load avg
	loadAvg, _ := load.Avg()

	fmt.Printf("%.2f, %.2f, %.2f\n", loadAvg.Load1, loadAvg.Load5, loadAvg.Load15)
	return fmt.Sprintf("%.2f, %.2f, %.2f", loadAvg.Load1, loadAvg.Load5, loadAvg.Load15)
}

func get_load() string {
	// Get CPU usage
	cpuTimes, _ := cpu.Times(false)

	if len(cpuTimes) == 0 {
		return "{}"
	}

	total := cpuTimes[0].User + cpuTimes[0].System + cpuTimes[0].Idle + cpuTimes[0].Nice +
		cpuTimes[0].Iowait + cpuTimes[0].Irq + cpuTimes[0].Softirq + cpuTimes[0].Steal +
		cpuTimes[0].Guest + cpuTimes[0].GuestNice
	percentages := map[string]float64{
		"user":      (cpuTimes[0].User / total) * 100,
		"system":    (cpuTimes[0].System / total) * 100,
		"idle":      (cpuTimes[0].Idle / total) * 100,
		"nice":      (cpuTimes[0].Nice / total) * 100,
		"iowait":    (cpuTimes[0].Iowait / total) * 100,
		"irq":       (cpuTimes[0].Irq / total) * 100,
		"softirq":   (cpuTimes[0].Softirq / total) * 100,
		"steal":     (cpuTimes[0].Steal / total) * 100,
		"guest":     (cpuTimes[0].Guest / total) * 100,
		"guestNice": (cpuTimes[0].GuestNice / total) * 100,
	}

	data, _ := json.Marshal(percentages)

	fmt.Print(string(data) + "\n")
	return string(data)
}

func get_mem_info() string {

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

	fmt.Print(string(data))
	return string(data)
}
