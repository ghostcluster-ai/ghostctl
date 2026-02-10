package utils

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ParseDuration parses a duration string like "1h", "30m", "10s"
func ParseDuration(durationStr string) (time.Duration, error) {
	// Handle common patterns
	patterns := map[string]string{
		`^(\d+)d$`: "${1}h", // days to hours
		`^(\d+)h$`: "${1}h", // hours
		`^(\d+)m$`: "${1}m", // minutes
		`^(\d+)s$`: "${1}s", // seconds
	}

	for pattern, replacement := range patterns {
		if matched, _ := regexp.MatchString(pattern, durationStr); matched {
			// Use time.ParseDuration for standard formats
			return time.ParseDuration(durationStr)
		}
	}

	return 0, fmt.Errorf("invalid duration format: %s", durationStr)
}

// FormatBytes formats bytes as human-readable string
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"B", "KB", "MB", "GB", "TB"}
	if exp >= len(units) {
		exp = len(units) - 1
	}

	return fmt.Sprintf("%.2f %s", float64(bytes)/float64(div), units[exp])
}

// ParseMemory parses memory string like "4Gi", "512Mi", "1024"
func ParseMemory(memStr string) (int64, error) {
	memStr = strings.TrimSpace(memStr)

	suffixes := map[string]int64{
		"Ki": 1024,
		"Mi": 1024 * 1024,
		"Gi": 1024 * 1024 * 1024,
		"Ti": 1024 * 1024 * 1024 * 1024,
		"K":  1000,
		"M":  1000 * 1000,
		"G":  1000 * 1000 * 1000,
		"T":  1000 * 1000 * 1000 * 1000,
	}

	// Check for suffix
	for suffix, multiplier := range suffixes {
		if strings.HasSuffix(memStr, suffix) {
			baseStr := strings.TrimSuffix(memStr, suffix)
			var value float64
			if _, err := fmt.Sscanf(baseStr, "%f", &value); err != nil {
				return 0, fmt.Errorf("invalid memory value: %s", memStr)
			}
			return int64(value * float64(multiplier)), nil
		}
	}

	// Try parsing as plain integer (bytes)
	var value int64
	if _, err := fmt.Sscanf(memStr, "%d", &value); err != nil {
		return 0, fmt.Errorf("invalid memory format: %s", memStr)
	}

	return value, nil
}

// ValidateClusterName validates a cluster name
func ValidateClusterName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("cluster name cannot be empty")
	}

	if len(name) > 63 {
		return fmt.Errorf("cluster name cannot exceed 63 characters")
	}

	// Should match [a-z0-9]([-a-z0-9]*[a-z0-9])?
	matched, err := regexp.MatchString(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`, name)
	if err != nil || !matched {
		return fmt.Errorf("cluster name must be lowercase alphanumeric with hyphens, starting and ending with alphanumeric")
	}

	return nil
}

// ValidateGPUType validates GPU type string
func ValidateGPUType(gpuType string) error {
	validTypes := map[string]bool{
		"nvidia-t4":   true,
		"nvidia-a100": true,
		"nvidia-v100": true,
		"nvidia-a40":  true,
		"amd-mi100":   true,
		"tpu-v4":      true,
	}

	if !validTypes[gpuType] {
		return fmt.Errorf("invalid GPU type: %s", gpuType)
	}

	return nil
}

// StringSliceContains checks if a string slice contains a value
func StringSliceContains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// StringSliceRemove removes a value from a string slice
func StringSliceRemove(slice []string, value string) []string {
	var result []string
	for _, item := range slice {
		if item != value {
			result = append(result, item)
		}
	}
	return result
}

// MergeStringMaps merges two string maps
func MergeStringMaps(m1, m2 map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range m1 {
		result[k] = v
	}
	for k, v := range m2 {
		result[k] = v
	}
	return result
}

// FilterMap filters a map based on a predicate function
func FilterMap(m map[string]string, predicate func(key, value string) bool) map[string]string {
	result := make(map[string]string)
	for k, v := range m {
		if predicate(k, v) {
			result[k] = v
		}
	}
	return result
}
