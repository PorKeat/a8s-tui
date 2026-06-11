package ui

import (
	"bytes"
	"errors"
	"fmt"
	"os"
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

func chooseEnvironmentFileCmd(defaultDirectory string) tea.Cmd {
	return func() tea.Msg {
		path, err := chooseEnvironmentFile(defaultDirectory)
		return environmentFileChoiceMsg{path: path, err: err}
	}
}

func chooseEnvironmentFile(defaultDirectory string) (string, error) {
	defaultDirectory = environmentInitialDirectory(defaultDirectory)
	var command *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		script := `set chosenFile to choose file with prompt "Choose environment file" default location (POSIX file ` +
			appleScriptQuote(defaultDirectory) + `)
POSIX path of chosenFile`
		command = exec.Command("osascript", "-e", script)
	case "linux":
		switch {
		case commandExists("zenity"):
			command = exec.Command("zenity", "--file-selection", "--title=Choose environment file", "--filename="+defaultDirectory+"/", "--file-filter=Environment files | *.env", "--file-filter=All files | *")
		case commandExists("kdialog"):
			command = exec.Command("kdialog", "--getopenfilename", defaultDirectory, "*.env|Environment files")
		default:
			return "", errNativeSaveDialogUnavailable
		}
	case "windows":
		script := fmt.Sprintf(
			`Add-Type -AssemblyName System.Windows.Forms; $dialog = New-Object System.Windows.Forms.OpenFileDialog; $dialog.Title = 'Choose environment file'; $dialog.InitialDirectory = '%s'; $dialog.Filter = 'Environment files (*.env)|*.env|All files (*.*)|*.*'; if ($dialog.ShowDialog() -eq 'OK') { $dialog.FileName }`,
			strings.ReplaceAll(defaultDirectory, "'", "''"),
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
		return "", fmt.Errorf("open native file dialog: %w", err)
	}
	if path == "" {
		return "", errNativeSaveDialogCancelled
	}
	return path, nil
}

func environmentInitialDirectory(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		path = "."
	}
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		if absolute, absErr := filepath.Abs(path); absErr == nil {
			return absolute
		}
		return path
	}
	directory := filepath.Dir(path)
	if absolute, err := filepath.Abs(directory); err == nil {
		return absolute
	}
	return directory
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
