package ui

import (
	"fmt"
	"os/exec"
	"runtime"
)

// OpenFolderInFileManager membuka path direktori di aplikasi pengelola berkas bawaan OS (Explorer/Nautilus/Finder).
func OpenFolderInFileManager(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default: // Linux / BSD
		cmd = exec.Command("xdg-open", path)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to open folder in file manager: %w", err)
	}

	return nil
}
