//go:build !windows

package ui

func SetDebugEnabled(_ bool) {}

func InitOverlay() {}

func UpdateText(_ string) {}

func CloseOverlay() {}
