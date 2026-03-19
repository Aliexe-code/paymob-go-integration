package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatAmount(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{100, "100"},
		{1000, "1,000"},
		{10000, "10,000"},
		{100000, "100,000"},
		{1000000, "1,000,000"},
		{1234567, "1,234,567"},
		{999, "999"},
		{1001, "1,001"},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.input)), func(t *testing.T) {
			result := FormatAmount(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStatusClass(t *testing.T) {
	tests := []struct {
		status   string
		expected string
	}{
		{"success", "bg-green-500/20 text-green-400"},
		{"failed", "bg-red-500/20 text-red-400"},
		{"pending", "bg-yellow-500/20 text-yellow-400"},
		{"cancelled", "bg-yellow-500/20 text-yellow-400"},
		{"unknown", "bg-yellow-500/20 text-yellow-400"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := StatusClass(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStatusText(t *testing.T) {
	tests := []struct {
		status   string
		expected string
	}{
		{"success", "Success"},
		{"failed", "Failed"},
		{"pending", "Pending"},
		{"cancelled", "Cancelled"},
		{"unknown", "Pending"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := StatusText(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}
