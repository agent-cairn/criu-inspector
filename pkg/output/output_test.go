package output

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/agent-cairn/criu-inspector/pkg/models"
)

// TestPrintHumanReadable_Basic tests basic process tree output
func TestPrintHumanReadable_Basic(t *testing.T) {
	data := &models.CheckpointData{
		Processes: []*models.ProcessEntry{
			{PID: 1, PPID: 0, PGID: 1, SID: 1, Comm: "systemd", NThreads: 1},
			{PID: 1234, PPID: 1, PGID: 1234, SID: 1234, Comm: "nginx", NThreads: 4},
		},
		Inventory: nil,
	}

	output := captureOutput(func() {
		PrintHumanReadable(data, false)
	})

	// Verify key elements are present
	if !strings.Contains(output, "CRIU CHECKPOINT INSPECTION REPORT") {
		t.Error("Missing report header")
	}
	if !strings.Contains(output, "Process Tree") {
		t.Error("Missing process tree section")
	}
	if !strings.Contains(output, "systemd") {
		t.Error("Missing process name")
	}
	if !strings.Contains(output, "nginx") {
		t.Error("Missing process name")
	}
	if !strings.Contains(output, "Total processes: 2") {
		t.Error("Missing process count")
	}
}

// TestPrintHumanReadable_Verbose tests verbose mode with inventory and FDs
func TestPrintHumanReadable_Verbose(t *testing.T) {
	data := &models.CheckpointData{
		Processes: []*models.ProcessEntry{
			{PID: 1, PPID: 0, PGID: 1, SID: 1, Comm: "init", NThreads: 1},
		},
		Inventory: &models.Inventory{
			Magic:     "CRIMG_V1",
			Arch:      "x86_64",
			PagesID:   "0x12345678",
			GenID:     "0x87654321",
			Entries:   42,
			ImagePages: 15230,
		},
		FileDescriptors: []*models.FileDescriptor{
			{FD: 0, Type: "file", Path: "/var/log/app.log", Flags: "O_RDONLY", ID: "fd-1"},
			{FD: 1, Type: "pipe", Path: "/tmp/pipe.sock", Flags: "O_RDWR", ID: "fd-2"},
			{FD: 3, Type: "socket", Path: "127.0.0.1:8080", Flags: "O_RDWR", ID: "fd-3"},
		},
	}

	output := captureOutput(func() {
		PrintHumanReadable(data, true)
	})

	// Verify verbose sections are present
	if !strings.Contains(output, "Inventory Metadata") {
		t.Error("Missing inventory section in verbose mode")
	}
	if !strings.Contains(output, "File Descriptors") {
		t.Error("Missing FD section in verbose mode")
	}
	if !strings.Contains(output, "CRIMG_V1") {
		t.Error("Missing inventory data")
	}
	if !strings.Contains(output, "x86_64") {
		t.Error("Missing architecture")
	}
	if !strings.Contains(output, "Total file descriptors: 3") {
		t.Error("Missing FD count")
	}
}

// TestPrintProcessTree_Empty tests empty process tree
func TestPrintProcessTree_Empty(t *testing.T) {
	data := &models.CheckpointData{
		Processes: []*models.ProcessEntry{},
	}

	output := captureOutput(func() {
		PrintHumanReadable(data, false)
	})

	if !strings.Contains(output, "No process data found") {
		t.Error("Missing empty process tree message")
	}
}

// TestPrintProcessTree_LongComm tests command name truncation
func TestPrintProcessTree_LongComm(t *testing.T) {
	longComm := strings.Repeat("a", 30)
	data := &models.CheckpointData{
		Processes: []*models.ProcessEntry{
			{PID: 1, PPID: 0, PGID: 1, SID: 1, Comm: longComm, NThreads: 1},
		},
	}

	output := captureOutput(func() {
		PrintHumanReadable(data, false)
	})

	// Should be truncated with ellipsis
	if !strings.Contains(output, "…") {
		t.Error("Long command name should be truncated")
	}
}

// TestPrintInventory_NoImagePages tests inventory without image pages
func TestPrintInventory_NoImagePages(t *testing.T) {
	data := &models.CheckpointData{
		Processes: []*models.ProcessEntry{},
		Inventory: &models.Inventory{
			Magic:   "CRIMG_V1",
			Arch:    "x86_64",
			PagesID: "0x12345678",
			GenID:   "0x87654321",
			Entries: 42,
			// ImagePages is 0, should not display
		},
	}

	output := captureOutput(func() {
		PrintHumanReadable(data, true)
	})

	if strings.Contains(output, "Image Pages") {
		t.Error("Image pages should not display when 0")
	}
}

// TestPrintFileDescriptors_Empty tests empty FD list
func TestPrintFileDescriptors_Empty(t *testing.T) {
	data := &models.CheckpointData{
		Processes:        []*models.ProcessEntry{},
		Inventory:        nil,
		FileDescriptors: []*models.FileDescriptor{},
	}

	output := captureOutput(func() {
		PrintHumanReadable(data, true)
	})

	if !strings.Contains(output, "No file descriptors found") {
		t.Error("Missing empty FD message")
	}
}

// TestPrintFileDescriptors_LongPath tests path truncation
func TestPrintFileDescriptors_LongPath(t *testing.T) {
	longPath := "/very/long/path/that/should/be/truncated/" + strings.Repeat("x", 50)
	data := &models.CheckpointData{
		Processes: []*models.ProcessEntry{},
		FileDescriptors: []*models.FileDescriptor{
			{FD: 4, Type: "file", Path: longPath, Flags: "O_RDONLY", ID: "fd-4"},
		},
	}

	output := captureOutput(func() {
		PrintHumanReadable(data, true)
	})

	if !strings.Contains(output, "…") {
		t.Error("Long path should be truncated")
	}
}

// TestPrintFileDescriptors_NoPathOrFlags tests FD with missing path/flags
func TestPrintFileDescriptors_NoPathOrFlags(t *testing.T) {
	data := &models.CheckpointData{
		Processes: []*models.ProcessEntry{},
		FileDescriptors: []*models.FileDescriptor{
			{FD: 5, Type: "eventfd", Path: "", Flags: "", ID: ""},
		},
	}

	output := captureOutput(func() {
		PrintHumanReadable(data, true)
	})

	// Should show (none) for missing values
	if !strings.Contains(output, "(none)") {
		t.Error("Missing path/flags should show (none)")
	}
	// Should show - for missing ID
	if !strings.Contains(output, "-") {
		t.Error("Missing ID should show -")
	}
}

// TestPrintFileDescriptors_Stdio highlights stdin/stdout/stderr
func TestPrintFileDescriptors_Stdio(t *testing.T) {
	data := &models.CheckpointData{
		Processes: []*models.ProcessEntry{},
		FileDescriptors: []*models.FileDescriptor{
			{FD: 0, Type: "file", Path: "/dev/stdin", Flags: "O_RDONLY", ID: "fd-0"},
			{FD: 1, Type: "file", Path: "/dev/stdout", Flags: "O_WRONLY", ID: "fd-1"},
			{FD: 2, Type: "file", Path: "/dev/stderr", Flags: "O_WRONLY", ID: "fd-2"},
			{FD: 3, Type: "file", Path: "/tmp/file.txt", Flags: "O_RDWR", ID: "fd-3"},
		},
	}

	output := captureOutput(func() {
		PrintHumanReadable(data, true)
	})

	// All FDs should be present
	if !strings.Contains(output, "Total file descriptors: 4") {
		t.Error("Wrong FD count")
	}
}

// TestInitTerminal tests terminal initialization
func TestInitTerminal(t *testing.T) {
	// This function sets a color package global
	// Just ensure it doesn't panic
	InitTerminal()
}

// TestColorType tests FD type coloring logic
func TestColorType(t *testing.T) {
	tests := []struct {
		fdType        string
		wantNonEmpty bool
	}{
		{"file", true},
		{"pipe", true},
		{"socket", true},
		{"eventfd", true},
		{"epoll", true},
		{"unknown", true},
	}

	for _, tt := range tests {
		result := colorType(tt.fdType)
		if tt.wantNonEmpty && result == "" {
			t.Errorf("colorType(%q) returned empty string", tt.fdType)
		}
	}
}

// Helper function to capture stdout
func captureOutput(f func()) string {
	// Save original stdout
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	// Create pipe
	r, w, _ := os.Pipe()

	// Redirect stdout and stderr to pipe
	os.Stdout = w
	os.Stderr = w

	// Run function
	f()

	// Close writer and restore stdout
	w.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// Read all output
	var buf bytes.Buffer
	io.Copy(&buf, r)

	return buf.String()
}
