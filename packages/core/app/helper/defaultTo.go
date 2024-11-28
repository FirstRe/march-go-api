package helper

import (
	"fmt"
	"march-inventory/cmd/app/graph/types"
)

func DefaultTo[T comparable](value *T, defaultValue T) T {
	if value != nil {
		return *value
	}
	return defaultValue
}

func GetOptionalSize(s *types.SizeInventory) *string {
	weight := DefaultTo(s.Weight, 0)
	width := DefaultTo(s.Width, 0)
	length := DefaultTo(s.Length, 0)
	height := DefaultTo(s.Height, 0)

	response := fmt.Sprintf("%d|%d|%d|%d", weight, width, length, height)

	return &response
}
