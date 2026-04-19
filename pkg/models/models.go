package models

// ProcessEntry represents a process from pstree.img
type ProcessEntry struct {
	PID      uint32 `json:"pid"`
	PPID     uint32 `json:"ppid"`
	PGID     uint32 `json:"pgid"`
	SID      uint32 `json:"sid"`
	Comm     string `json:"comm"`
	NThreads uint32 `json:"nthreads"`
}

// Inventory represents metadata from inventory.img
type Inventory struct {
	Magic      string `json:"magic" xml:"magic"`
	Arch       string `json:"arch" xml:"arch"`
	PagesID    string `json:"pages_id" xml:"pages_id"`
	GenID      string `json:"gen_id" xml:"gen_id"`
	Entries    uint32 `json:"entries" xml:"entries"`
	ImagePages uint32 `json:"image_pages,omitempty" xml:"image_pages"`
}

// FileDescriptor represents an FD from fds.img
type FileDescriptor struct {
	FD     int32  `json:"fd"`
	Type   string `json:"type"`
	Path   string `json:"path,omitempty"`
	Flags  string `json:"flags,omitempty"`
	ID     string `json:"id,omitempty"`
	MntID  uint32 `json:"mnt_id,omitempty"`
}

// CheckpointData aggregates all parsed checkpoint information
type CheckpointData struct {
	Processes       []*ProcessEntry   `json:"processes"`
	Inventory       *Inventory        `json:"inventory,omitempty"`
	FileDescriptors []*FileDescriptor `json:"file_descriptors,omitempty"`
}
