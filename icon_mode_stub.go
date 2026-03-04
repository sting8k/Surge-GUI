//go:build !darwin

package main

import "fmt"

func applyIconMode(mode string) error {
	if mode == iconModeDock {
		return nil
	}
	if mode == iconModeMenuBar {
		return fmt.Errorf("menu bar mode is only supported on macOS")
	}
	return fmt.Errorf("unsupported icon mode: %s", mode)
}
