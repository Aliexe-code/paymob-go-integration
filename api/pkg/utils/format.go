package utils

import "fmt"

// FormatAmount formats an integer with comma separators (e.g., 1000 -> "1,000")
func FormatAmount(n int) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var result []byte
	for i, c := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, c)
	}
	return string(result)
}

// StatusClass returns the CSS class for a payment status
func StatusClass(status string) string {
	switch status {
	case "success":
		return "bg-green-500/20 text-green-400"
	case "failed":
		return "bg-red-500/20 text-red-400"
	default:
		return "bg-yellow-500/20 text-yellow-400"
	}
}

// StatusText returns the display text for a payment status
func StatusText(status string) string {
	switch status {
	case "success":
		return "Success"
	case "failed":
		return "Failed"
	case "cancelled":
		return "Cancelled"
	default:
		return "Pending"
	}
}
