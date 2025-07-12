package utils

import (
	"database/sql"
	"fmt"
	"os"
	"runtime"
	"time"
)

type SystemMonitor struct {
	db *sql.DB
}

type SystemMetrics struct {
	// System metrics
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsage   float64 `json:"disk_usage"`
	Uptime      int64   `json:"uptime"`
	
	// Database metrics
	DatabaseConnections int     `json:"database_connections"`
	DatabaseSize        int64   `json:"database_size"`
	QueryPerformance    float64 `json:"query_performance"`
	
	// Application metrics
	ActiveUsers         int     `json:"active_users"`
	TotalUsers          int     `json:"total_users"`
	PaymentVolume       int64   `json:"payment_volume"`
	ErrorRate           float64 `json:"error_rate"`
	
	// Bot metrics
	BotUptime          int64 `json:"bot_uptime"`
	MessagesProcessed  int   `json:"messages_processed"`
	CommandsExecuted   int   `json:"commands_executed"`
	WebhooksReceived   int   `json:"webhooks_received"`
	
	Timestamp time.Time `json:"timestamp"`
}

func NewSystemMonitor(db *sql.DB) *SystemMonitor {
	return &SystemMonitor{db: db}
}

func (sm *SystemMonitor) GetMetrics() (*SystemMetrics, error) {
	metrics := &SystemMetrics{
		Timestamp: time.Now(),
	}
	
	// Get system metrics
	if err := sm.getSystemMetrics(metrics); err != nil {
		return nil, fmt.Errorf("failed to get system metrics: %w", err)
	}
	
	// Get database metrics
	if err := sm.getDatabaseMetrics(metrics); err != nil {
		return nil, fmt.Errorf("failed to get database metrics: %w", err)
	}
	
	// Get application metrics
	if err := sm.getApplicationMetrics(metrics); err != nil {
		return nil, fmt.Errorf("failed to get application metrics: %w", err)
	}
	
	return metrics, nil
}

func (sm *SystemMonitor) getSystemMetrics(metrics *SystemMetrics) error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	// Memory usage in MB
	metrics.MemoryUsage = float64(m.Alloc) / 1024 / 1024
	
	// CPU usage (simplified - in production use proper CPU monitoring)
	metrics.CPUUsage = 0.0 // TODO: Implement proper CPU monitoring
	
	// Disk usage (simplified)
	metrics.DiskUsage = 0.0 // TODO: Implement proper disk monitoring
	
	// Uptime (simplified)
	metrics.Uptime = time.Now().Unix() - startTime
	
	return nil
}

func (sm *SystemMonitor) getDatabaseMetrics(metrics *SystemMetrics) error {
	if sm.db == nil {
		return nil
	}
	
	// Database connections
	stats := sm.db.Stats()
	metrics.DatabaseConnections = stats.OpenConnections
	
	// Database size
	var dbSize int64
	err := sm.db.QueryRow(`
		SELECT pg_database_size(current_database())
	`).Scan(&dbSize)
	if err != nil {
		dbSize = 0
	}
	metrics.DatabaseSize = dbSize
	
	// Query performance (average query time)
	var avgQueryTime float64
	err = sm.db.QueryRow(`
		SELECT COALESCE(AVG(mean_exec_time), 0) 
		FROM pg_stat_statements 
		WHERE calls > 0
	`).Scan(&avgQueryTime)
	if err != nil {
		avgQueryTime = 0
	}
	metrics.QueryPerformance = avgQueryTime
	
	return nil
}

func (sm *SystemMonitor) getApplicationMetrics(metrics *SystemMetrics) error {
	if sm.db == nil {
		return nil
	}
	
	// Total users
	err := sm.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&metrics.TotalUsers)
	if err != nil {
		metrics.TotalUsers = 0
	}
	
	// Active users (last 24 hours)
	err = sm.db.QueryRow(`
		SELECT COUNT(*) 
		FROM users 
		WHERE last_active > NOW() - INTERVAL '24 hours'
	`).Scan(&metrics.ActiveUsers)
	if err != nil {
		metrics.ActiveUsers = 0
	}
	
	// Payment volume (last 24 hours)
	err = sm.db.QueryRow(`
		SELECT COALESCE(SUM(amount), 0) 
		FROM payments 
		WHERE status = 'completed' 
		AND created_at > NOW() - INTERVAL '24 hours'
	`).Scan(&metrics.PaymentVolume)
	if err != nil {
		metrics.PaymentVolume = 0
	}
	
	// Error rate (simplified)
	var totalRequests, errorRequests int
	err = sm.db.QueryRow(`
		SELECT COUNT(*) 
		FROM system_logs 
		WHERE created_at > NOW() - INTERVAL '1 hour'
	`).Scan(&totalRequests)
	if err != nil {
		totalRequests = 0
	}
	
	err = sm.db.QueryRow(`
		SELECT COUNT(*) 
		FROM system_logs 
		WHERE level = 'ERROR' 
		AND created_at > NOW() - INTERVAL '1 hour'
	`).Scan(&errorRequests)
	if err != nil {
		errorRequests = 0
	}
	
	if totalRequests > 0 {
		metrics.ErrorRate = float64(errorRequests) / float64(totalRequests) * 100
	}
	
	return nil
}

func (sm *SystemMonitor) LogMetrics(metrics *SystemMetrics) error {
	if sm.db == nil {
		return nil
	}
	
	_, err := sm.db.Exec(`
		INSERT INTO system_metrics (
			cpu_usage, memory_usage, disk_usage, uptime,
			database_connections, database_size, query_performance,
			active_users, total_users, payment_volume, error_rate,
			bot_uptime, messages_processed, commands_executed, webhooks_received,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`, 
		metrics.CPUUsage, metrics.MemoryUsage, metrics.DiskUsage, metrics.Uptime,
		metrics.DatabaseConnections, metrics.DatabaseSize, metrics.QueryPerformance,
		metrics.ActiveUsers, metrics.TotalUsers, metrics.PaymentVolume, metrics.ErrorRate,
		metrics.BotUptime, metrics.MessagesProcessed, metrics.CommandsExecuted, metrics.WebhooksReceived,
		metrics.Timestamp,
	)
	
	return err
}

func (sm *SystemMonitor) GetHealthCheck() map[string]interface{} {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"services":  make(map[string]interface{}),
	}
	
	// Check database connection
	if sm.db != nil {
		err := sm.db.Ping()
		if err != nil {
			health["services"].(map[string]interface{})["database"] = map[string]interface{}{
				"status": "unhealthy",
				"error":  err.Error(),
			}
			health["status"] = "unhealthy"
		} else {
			health["services"].(map[string]interface{})["database"] = map[string]interface{}{
				"status": "healthy",
			}
		}
	}
	
	// Check disk space
	diskUsage := sm.getDiskUsage()
	if diskUsage > 90 {
		health["services"].(map[string]interface{})["disk"] = map[string]interface{}{
			"status": "warning",
			"usage":  diskUsage,
		}
		if health["status"] == "healthy" {
			health["status"] = "warning"
		}
	} else {
		health["services"].(map[string]interface{})["disk"] = map[string]interface{}{
			"status": "healthy",
			"usage":  diskUsage,
		}
	}
	
	// Check memory usage
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memUsage := float64(m.Alloc) / 1024 / 1024
	
	health["services"].(map[string]interface{})["memory"] = map[string]interface{}{
		"status": "healthy",
		"usage":  memUsage,
	}
	
	return health
}

func (sm *SystemMonitor) getDiskUsage() float64 {
	// This is a simplified implementation
	// In production, use proper disk usage monitoring
	return 0.0
}

func (sm *SystemMonitor) StartMonitoring() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				metrics, err := sm.GetMetrics()
				if err != nil {
					fmt.Printf("Error getting metrics: %v\n", err)
					continue
				}
				
				// Log metrics to database
				if err := sm.LogMetrics(metrics); err != nil {
					fmt.Printf("Error logging metrics: %v\n", err)
				}
				
				// Check for alerts
				sm.checkAlerts(metrics)
			}
		}
	}()
}

func (sm *SystemMonitor) checkAlerts(metrics *SystemMetrics) {
	// Memory usage alert
	if metrics.MemoryUsage > 500 { // 500MB
		sm.sendAlert("HIGH_MEMORY_USAGE", fmt.Sprintf("Memory usage: %.2f MB", metrics.MemoryUsage))
	}
	
	// Error rate alert
	if metrics.ErrorRate > 5 { // 5%
		sm.sendAlert("HIGH_ERROR_RATE", fmt.Sprintf("Error rate: %.2f%%", metrics.ErrorRate))
	}
	
	// Database connections alert
	if metrics.DatabaseConnections > 50 {
		sm.sendAlert("HIGH_DB_CONNECTIONS", fmt.Sprintf("Database connections: %d", metrics.DatabaseConnections))
	}
}

func (sm *SystemMonitor) sendAlert(alertType, message string) {
	// Log alert
	fmt.Printf("ALERT [%s]: %s\n", alertType, message)
	
	// In production, send to monitoring service or admin notifications
	if sm.db != nil {
		sm.db.Exec(`
			INSERT INTO system_alerts (alert_type, message, created_at)
			VALUES ($1, $2, $3)
		`, alertType, message, time.Now())
	}
}

// Global start time for uptime calculation
var startTime = time.Now().Unix()

func GetSystemInfo() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return map[string]interface{}{
		"version":     "1.0.0",
		"go_version":  runtime.Version(),
		"os":          runtime.GOOS,
		"arch":        runtime.GOARCH,
		"num_cpu":     runtime.NumCPU(),
		"num_goroutine": runtime.NumGoroutine(),
		"memory_mb":   float64(m.Alloc) / 1024 / 1024,
		"uptime":      time.Now().Unix() - startTime,
		"pid":         os.Getpid(),
	}
}