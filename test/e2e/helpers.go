// Package e2e provides helper functions for end-to-end tests
package e2e

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// SmokeTestConfig provides configuration for smoke tests
type SmokeTestConfig struct {
	BaseURL       string        `json:"base_url"`
	Timeout       time.Duration `json:"timeout"`
	MaxRetries    int           `json:"max_retries"`
	RetryInterval time.Duration `json:"retry_interval"`
}

// DefaultSmokeTestConfig returns default configuration for smoke tests
func DefaultSmokeTestConfig() *SmokeTestConfig {
	return &SmokeTestConfig{
		BaseURL:       "http://localhost:8080",
		Timeout:       30 * time.Second,
		MaxRetries:    3,
		RetryInterval: 1 * time.Second,
	}
}

// SmokeTestResult represents the result of a smoke test
type SmokeTestResult struct {
	TestName     string        `json:"test_name"`
	Passed       bool          `json:"passed"`
	Duration     time.Duration `json:"duration"`
	Error        string        `json:"error,omitempty"`
	StatusCode   int           `json:"status_code,omitempty"`
	ResponseTime time.Duration `json:"response_time,omitempty"`
}

// SmokeTestRunner provides utilities for running smoke tests
type SmokeTestRunner struct {
	config *SmokeTestConfig
	client *http.Client
}

// NewSmokeTestRunner creates a new smoke test runner
func NewSmokeTestRunner(config *SmokeTestConfig) *SmokeTestRunner {
	if config == nil {
		config = DefaultSmokeTestConfig()
	}

	return &SmokeTestRunner{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// RunHealthCheck runs a health check smoke test
func (r *SmokeTestRunner) RunHealthCheck(ctx context.Context) *SmokeTestResult {
	result := &SmokeTestResult{
		TestName: "health_check",
	}
	start := time.Now()

	url := fmt.Sprintf("%s/health", r.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to create request: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	resp, err := r.client.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("Request failed: %v", err)
		result.Duration = time.Since(start)
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	result.ResponseTime = time.Since(start)
	result.Duration = result.ResponseTime
	result.Passed = resp.StatusCode == http.StatusOK

	if !result.Passed {
		result.Error = fmt.Sprintf("Health check failed with status %d", resp.StatusCode)
	}

	return result
}

// RunMetricsCheck runs a metrics endpoint smoke test
func (r *SmokeTestRunner) RunMetricsCheck(ctx context.Context) *SmokeTestResult {
	result := &SmokeTestResult{
		TestName: "metrics_check",
	}
	start := time.Now()

	url := fmt.Sprintf("%s/metrics", r.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to create request: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	resp, err := r.client.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("Request failed: %v", err)
		result.Duration = time.Since(start)
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	result.ResponseTime = time.Since(start)
	result.Duration = result.ResponseTime
	result.Passed = resp.StatusCode == http.StatusOK

	if !result.Passed {
		result.Error = fmt.Sprintf("Metrics check failed with status %d", resp.StatusCode)
	}

	return result
}

// RunCriticalPathCheck runs critical path smoke tests
func (r *SmokeTestRunner) RunCriticalPathCheck(ctx context.Context) []*SmokeTestResult {
	var results []*SmokeTestResult

	// Test health endpoint
	healthResult := r.RunHealthCheck(ctx)
	results = append(results, healthResult)

	// Test metrics endpoint
	metricsResult := r.RunMetricsCheck(ctx)
	results = append(results, metricsResult)

	return results
}

// GenerateSmokeTestReport generates a report from smoke test results
func GenerateSmokeTestReport(results []*SmokeTestResult) map[string]any {
	passed := 0
	failed := 0
	totalDuration := time.Duration(0)

	for _, result := range results {
		if result.Passed {
			passed++
		} else {
			failed++
		}
		totalDuration += result.Duration
	}

	return map[string]any{
		"total_tests":      len(results),
		"passed":           passed,
		"failed":           failed,
		"success_rate":     float64(passed) / float64(len(results)),
		"total_duration":   totalDuration.String(),
		"average_duration": (totalDuration / time.Duration(len(results))).String(),
		"results":          results,
	}
}
