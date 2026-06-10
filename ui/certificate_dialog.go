package ui

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	tea "charm.land/bubbletea/v2"
)

var errNativeSaveDialogUnavailable = errors.New("native save dialog is unavailable")
var errNativeSaveDialogCancelled = errors.New("native save dialog was cancelled")

func chooseCertificatePathCmd(defaultPath string) tea.Cmd {
	return func() tea.Msg {
		path, err := chooseCertificatePath(defaultPath)
		return certificatePathChoiceMsg{path: path, err: err}
	}
}

func chooseCertificatePath(defaultPath string) (string, error) {
	var command *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		script := `set chosenFile to choose file name with prompt "Save SSL certificate" default location (POSIX file ` +
			appleScriptQuote(directoryForPath(defaultPath)) + `) default name ` + appleScriptQuote(filenameForPath(defaultPath)) + `
POSIX path of chosenFile`
		command = exec.Command("osascript", "-e", script)
	case "linux":
		switch {
		case commandExists("zenity"):
			command = exec.Command("zenity", "--file-selection", "--save", "--confirm-overwrite", "--title=Save SSL certificate", "--filename="+defaultPath)
		case commandExists("kdialog"):
			command = exec.Command("kdialog", "--getsavefilename", defaultPath, "*.crt *.pem|SSL certificates")
		default:
			return "", errNativeSaveDialogUnavailable
		}
	case "windows":
		script := fmt.Sprintf(
			`Add-Type -AssemblyName System.Windows.Forms; $dialog = New-Object System.Windows.Forms.SaveFileDialog; $dialog.Title = 'Save SSL certificate'; $dialog.FileName = '%s'; $dialog.Filter = 'SSL certificates (*.crt;*.pem)|*.crt;*.pem|All files (*.*)|*.*'; if ($dialog.ShowDialog() -eq 'OK') { $dialog.FileName }`,
			strings.ReplaceAll(filenameForPath(defaultPath), "'", "''"),
		)
		command = exec.Command("powershell", "-NoProfile", "-STA", "-Command", script)
	default:
		return "", errNativeSaveDialogUnavailable
	}

	output, err := command.Output()
	path := strings.TrimSpace(string(output))
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			detail := strings.ToLower(string(bytes.TrimSpace(exitErr.Stderr)))
			if strings.Contains(detail, "cancel") || exitErr.ExitCode() == 1 {
				return "", errNativeSaveDialogCancelled
			}
		}
		return "", fmt.Errorf("open native save dialog: %w", err)
	}
	if path == "" {
		return "", errNativeSaveDialogCancelled
	}
	return path, nil
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func appleScriptQuote(value string) string {
	return `"` + strings.ReplaceAll(strings.ReplaceAll(value, `\`, `\\`), `"`, `\"`) + `"`
}

func directoryForPath(path string) string {
	directory := filepath.Dir(path)
	if directory == "." || directory == "" {
		return "/"
	}
	return directory
}

func filenameForPath(path string) string {
	filename := filepath.Base(path)
	if filename == "." || filename == string(filepath.Separator) || filename == "" {
		return "database-ca.crt"
	}
	return filename
}
