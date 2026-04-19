package parser

import (
	"os"
	"testing"
)

func TestParseCheckpoint_NonExistentDir(t *testing.T) {
	_, err := ParseCheckpoint("/nonexistent/dir", false)
	if err == nil {
		t.Error("ParseCheckpoint should error for non-existent directory")
	}
}

func TestParseCheckpoint_EmptyDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "criu-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	data, err := ParseCheckpoint(tmpDir, false)
	if err != nil {
		t.Errorf("ParseCheckpoint should succeed for empty directory, got: %v", err)
	}

	if data == nil {
		t.Error("ParseCheckpoint should return non-nil data")
	}
}

func TestParsePstree_EmptyJSON(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "pstree-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	// Write empty JSON array
	tmpFile.WriteString("[]")
	tmpFile.Close()

	entries, err := ParsePstree(tmpFile.Name())
	if err != nil {
		t.Errorf("ParsePstree should succeed for empty JSON array, got: %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("ParsePstree should return 0 entries, got %d", len(entries))
	}
}

func TestParsePstree_WithJSON(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "pstree-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	// Write binary header + JSON
	tmpFile.WriteString("BINARY_HEADER...")
	tmpFile.WriteString(`[{"pid":1,"ppid":0,"pgid":1,"sid":1,"comm":"init","nthreads":1}]`)
	tmpFile.Close()

	entries, err := ParsePstree(tmpFile.Name())
	if err != nil {
		t.Errorf("ParsePstree should succeed, got: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("ParsePstree should return 1 entry, got %d", len(entries))
	}

	if entries[0].PID != 1 {
		t.Errorf("Expected PID 1, got %d", entries[0].PID)
	}

	if entries[0].Comm != "init" {
		t.Errorf("Expected Comm 'init', got %s", entries[0].Comm)
	}
}

func TestParseInventory_ValidXML(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "inventory-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	xmlData := `<?xml version="1.0"?><inventory><magic>CRIU</magic><arch>x86_64</arch><pages_id>123</pages_id><gen_id>456</gen_id><entries>5</entries></inventory>`
	tmpFile.WriteString(xmlData)
	tmpFile.Close()

	inv, err := ParseInventory(tmpFile.Name())
	if err != nil {
		t.Errorf("ParseInventory should succeed for valid XML, got: %v", err)
	}

	if inv == nil {
		t.Fatal("ParseInventory should return non-nil inventory")
	}

	if inv.Magic != "CRIU" {
		t.Errorf("Expected Magic 'CRIU', got %s", inv.Magic)
	}

	if inv.Arch != "x86_64" {
		t.Errorf("Expected Arch 'x86_64', got %s", inv.Arch)
	}

	if inv.Entries != 5 {
		t.Errorf("Expected Entries 5, got %d", inv.Entries)
	}
}

func TestParseFds_ValidXML(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "fds-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	xmlData := `<?xml version="1.0"?><fds><fd id="1" fd="0" type="pipe" flags="02000000" /></fds>`
	tmpFile.WriteString(xmlData)
	tmpFile.Close()

	fds, err := ParseFds(tmpFile.Name())
	if err != nil {
		t.Errorf("ParseFds should succeed for valid XML, got: %v", err)
	}

	if len(fds) != 1 {
		t.Fatalf("ParseFds should return 1 FD, got %d", len(fds))
	}

	if fds[0].FD != 0 {
		t.Errorf("Expected FD 0, got %d", fds[0].FD)
	}

	if fds[0].Type != "pipe" {
		t.Errorf("Expected Type 'pipe', got %s", fds[0].Type)
	}
}
