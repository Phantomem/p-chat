// lib/monitor.go
package lib

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-ping/ping"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
)

type ServerMetrics struct {
	mu          sync.RWMutex
	CPUUsage    float64
	MemoryUsage float64
	DiskUsage   float64
	HTTPLatency map[string]time.Duration
	PingLatency map[string]time.Duration
	WorkerPool  struct {
		Active    int64
		Pending   int64
		Processed int64
	}
	WebSocketConnections int
	HTTPRequests         map[string]uint64
	StartTime            time.Time
}

var (
	metrics     ServerMetrics
	metricsOnce sync.Once
)

func InitMonitor() {
	metricsOnce.Do(func() {
		metrics = ServerMetrics{
			HTTPLatency:  make(map[string]time.Duration),
			PingLatency:  make(map[string]time.Duration),
			HTTPRequests: make(map[string]uint64),
			StartTime:    time.Now(),
		}

		// Start background metric collectors
		go collectSystemMetrics()
		go measureExternalResources()
		go printMetricsToConsole()
	})
}

func collectSystemMetrics() {
	for {
		if percent, err := cpu.Percent(time.Second, false); err == nil {
			metrics.mu.Lock()
			metrics.CPUUsage = percent[0]
			metrics.mu.Unlock()
		}

		if memStat, err := mem.VirtualMemory(); err == nil {
			metrics.mu.Lock()
			metrics.MemoryUsage = memStat.UsedPercent
			metrics.mu.Unlock()
		}

		if diskStat, err := disk.Usage("/"); err == nil {
			metrics.mu.Lock()
			metrics.DiskUsage = diskStat.UsedPercent
			metrics.mu.Unlock()
		}

		time.Sleep(5 * time.Second)
	}
}

func measureExternalResources() {
	websites := []string{"http://localhost:8080/status"}
	hosts := []string{"http://localhost:8080/status"}

	for {
		// Measure website latencies
		for _, url := range websites {
			start := time.Now()
			if resp, err := http.Get(url); err == nil {
				resp.Body.Close()
				metrics.mu.Lock()
				metrics.HTTPLatency[url] = time.Since(start)
				metrics.mu.Unlock()
			}
		}

		// Measure ping latencies
		for _, host := range hosts {
			pinger, err := ping.NewPinger(host)
			if err == nil {
				pinger.Count = 3
				pinger.Timeout = 5 * time.Second
				if err := pinger.Run(); err == nil {
					metrics.mu.Lock()
					metrics.PingLatency[host] = pinger.Statistics().AvgRtt
					metrics.mu.Unlock()
				}
			}
		}

		time.Sleep(30 * time.Second)
	}
}

func printMetricsToConsole() {
	for {
		metrics.mu.RLock()
		fmt.Printf("\n=== Server Metrics ===\n")
		fmt.Printf("Uptime: %s\n", time.Since(metrics.StartTime).Round(time.Second))
		fmt.Printf("CPU Usage: %.2f%%\n", metrics.CPUUsage)
		fmt.Printf("Memory Usage: %.2f%%\n", metrics.MemoryUsage)
		fmt.Printf("Disk Usage: %.2f%%\n", metrics.DiskUsage)
		fmt.Printf("WebSocket Connections: %d\n", metrics.WebSocketConnections)
		fmt.Printf("Worker Pool - Active: %d, Pending: %d, Processed: %d\n",
			metrics.WorkerPool.Active,
			metrics.WorkerPool.Pending,
			metrics.WorkerPool.Processed)
		metrics.mu.RUnlock()
		time.Sleep(10 * time.Second)
	}
}

// Exported functions to update metrics
func RecordWebSocketConnection(connected bool) {
	metrics.mu.Lock()
	defer metrics.mu.Unlock()
	if connected {
		metrics.WebSocketConnections++
	} else {
		metrics.WebSocketConnections--
	}
}

func RecordHTTPRequest(path string, duration time.Duration) {
	metrics.mu.Lock()
	defer metrics.mu.Unlock()
	metrics.HTTPRequests[path]++
}

func GetMetrics() ServerMetrics {
	metrics.mu.RLock()
	defer metrics.mu.RUnlock()
	return metrics
}

func CollectWorkerPoolMetrics(wp *WorkerPoolImpl) {
	for {
		active, pending, processed := wp.GetMetrics()

		metrics.mu.Lock()
		metrics.WorkerPool.Active = active
		metrics.WorkerPool.Pending = pending
		metrics.WorkerPool.Processed = processed
		metrics.mu.Unlock()

		time.Sleep(1 * time.Second)
	}
}
