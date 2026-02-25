package main

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseCurrencyString(amountStr string) (float64, error) {
	amountStr = strings.ReplaceAll(amountStr, "Rp", "")
	amountStr = strings.ReplaceAll(amountStr, ".", "")
	amountStr = strings.ReplaceAll(amountStr, ",", "")
	amountStr = strings.TrimSpace(amountStr)
	return strconv.ParseFloat(amountStr, 64)
}

func main() {
	tests := []string{"Rp 0", "Rp 50.000", "Rp 150000", "0", "150000"}
	for _, t := range tests {
		val, err := ParseCurrencyString(t)
		fmt.Printf("Input: %q, Value: %f, Error: %v\n", t, val, err)
	}
}
