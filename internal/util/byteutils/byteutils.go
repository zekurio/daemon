package byteutils

import "fmt"

type ByteHelper uint64

// FormatToDecimal formats a given amount of bytes to a human readable string in decimal format. 1000 bytes = 1 kB
func (b ByteHelper) FormatToDecimal() string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := uint64(b) / uint64(unit); n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// FormatToBinary formats a given amount of bytes to a human readable string in binary format. 1024 bytes = 1 KiB
func (b ByteHelper) FormatToBinary() string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := uint64(b) / uint64(unit); n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
