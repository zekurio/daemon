package timeutils

import (
	"fmt"
	"strings"
	"time"
)

func ParseDuration(s string) (time.Duration, error) {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, err
	}
	return d, nil
}

func FormatDuration(d time.Duration) string {
	var parts []string

	unitNames := []string{"hr", "min", "sec", "ms", "ns"}
	divisors := []int64{24 * 60 * 60 * 1e9, 60 * 60 * 1e9, 60 * 1e9, 1e9, 1e6, 1e3}
	remainder := d.Nanoseconds()

	days := remainder / divisors[0]
	if days > 0 {
		dayUnit := "day"
		if days > 1 {
			dayUnit = "days"
		}
		parts = append(parts, fmt.Sprintf("%d %s", days, dayUnit))
	}
	remainder %= divisors[0]

	for i, divisor := range divisors[1:] {
		value := remainder / divisor
		if value > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", value, unitNames[i]))
		}
		remainder %= divisor
	}

	return strings.Join(parts, " ")
}
