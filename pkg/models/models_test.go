package models

import "testing"

func TestProcessEntry_Fields(t *testing.T) {
	p := &ProcessEntry{
		PID:      1,
		PPID:     0,
		PGID:     1,
		SID:      1,
		Comm:     "init",
		NThreads: 1,
	}

	if p.PID != 1 {
		t.Errorf("Expected PID 1, got %d", p.PID)
	}

	if p.Comm != "init" {
		t.Errorf("Expected Comm 'init', got %s", p.Comm)
	}
}

func TestInventory_Fields(t *testing.T) {
	inv := &Inventory{
		Magic:   "CRIMG_V1",
		Arch:    "x86_64",
		PagesID: "test-pages",
		GenID:   "test-gen",
		Entries: 10,
	}

	if inv.Magic != "CRIMG_V1" {
		t.Errorf("Expected Magic 'CRIMG_V1', got %s", inv.Magic)
	}

	if inv.Entries != 10 {
		t.Errorf("Expected Entries 10, got %d", inv.Entries)
	}
}

func TestFileDescriptor_Fields(t *testing.T) {
	fd := &FileDescriptor{
		FD:    3,
		Type:  "file",
		Path:  "/tmp/test.txt",
		Flags: "O_RDWR",
	}

	if fd.FD != 3 {
		t.Errorf("Expected FD 3, got %d", fd.FD)
	}

	if fd.Type != "file" {
		t.Errorf("Expected Type 'file', got %s", fd.Type)
	}

	if fd.Path != "/tmp/test.txt" {
		t.Errorf("Expected Path '/tmp/test.txt', got %s", fd.Path)
	}
}

func TestCheckpointData_Aggregation(t *testing.T) {
	data := &CheckpointData{
		Processes: []*ProcessEntry{
			{PID: 1, Comm: "init"},
		},
		Inventory: &Inventory{
			Magic: "CRIMG_V1",
		},
		FileDescriptors: []*FileDescriptor{
			{FD: 0, Type: "pipe"},
		},
	}

	if len(data.Processes) != 1 {
		t.Errorf("Expected 1 process, got %d", len(data.Processes))
	}

	if data.Inventory == nil {
		t.Error("Expected non-nil inventory")
	}

	if len(data.FileDescriptors) != 1 {
		t.Errorf("Expected 1 file descriptor, got %d", len(data.FileDescriptors))
	}
}
