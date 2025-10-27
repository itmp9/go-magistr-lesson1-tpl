package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	serverURL        = "http://srv.msk01.gigacorp.local/_stats"
	pollInterval     = 3 * time.Second
	loadAvgThreshold = 30
	memoryThreshold  = 0.80
	diskThreshold    = 0.90
	networkThreshold = 0.90
	maxErrors        = 3
)

type ServerStats struct {
	LoadAvg           float64
	TotalRAM          int64
	UsedRAM           int64
	TotalDisk         int64
	UsedDisk          int64
	TotalNetwork      int64
	CurrentNetwork    int64
}

func main() {
	consecutiveErrors := 0
	
	for {
		stats, err := fetchStats()
		
		if err != nil {
			consecutiveErrors++
			if consecutiveErrors == maxErrors {
				fmt.Println("Unable to fetch server statistic")
				consecutiveErrors = 0
			}
		} else {
			consecutiveErrors = 0
			checkThresholds(stats)
		}
		
		time.Sleep(pollInterval)
	}
}

func fetchStats() (*ServerStats, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	
	resp, err := client.Get(serverURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	stats, err := parseStats(string(body))
	if err != nil {
		return nil, err
	}
	
	return stats, nil
}

func parseStats(data string) (*ServerStats, error) {
	parts := strings.Split(strings.TrimSpace(data), ",")
	if len(parts) != 7 {
		return nil, fmt.Errorf("invalid data format: expected 7 values, got %d", len(parts))
	}
	
	var stats ServerStats
	var err error
	
	if stats.LoadAvg, err = strconv.ParseFloat(parts[0], 64); err != nil {
		return nil, err
	}
	
	if stats.TotalRAM, err = strconv.ParseInt(parts[1], 10, 64); err != nil {
		return nil, err
	}
	
	if stats.UsedRAM, err = strconv.ParseInt(parts[2], 10, 64); err != nil {
		return nil, err
	}
	
	if stats.TotalDisk, err = strconv.ParseInt(parts[3], 10, 64); err != nil {
		return nil, err
	}
	
	if stats.UsedDisk, err = strconv.ParseInt(parts[4], 10, 64); err != nil {
		return nil, err
	}
	
	if stats.TotalNetwork, err = strconv.ParseInt(parts[5], 10, 64); err != nil {
		return nil, err
	}
	
	if stats.CurrentNetwork, err = strconv.ParseInt(parts[6], 10, 64); err != nil {
		return nil, err
	}
	
	return &stats, nil
}

func checkThresholds(stats *ServerStats) {

	if stats.LoadAvg > loadAvgThreshold {
		fmt.Printf("Load Average is too high: %.0f\n", stats.LoadAvg)
	}

	if stats.TotalRAM > 0 {
		memoryUsage := float64(stats.UsedRAM) / float64(stats.TotalRAM)
		if memoryUsage > memoryThreshold {
			fmt.Printf("Memory usage too high: %d%%\n", int(memoryUsage*100))
		}
	}
	
	if stats.TotalDisk > 0 {
		diskUsage := float64(stats.UsedDisk) / float64(stats.TotalDisk)
		if diskUsage > diskThreshold {
			freeBytes := stats.TotalDisk - stats.UsedDisk
			freeMB := float64(freeBytes) / (1024 * 1024)
			fmt.Printf("Free disk space is too low: %d Mb left\n", int(freeMB))
		}
	}
	

	if stats.TotalNetwork > 0 {
		networkUsage := float64(stats.CurrentNetwork) / float64(stats.TotalNetwork)
		if networkUsage > networkThreshold {
			availableBytes := stats.TotalNetwork - stats.CurrentNetwork
			availableMbits := float64(availableBytes) / 1000000
			fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", int(availableMbits))
		}
	}
}
