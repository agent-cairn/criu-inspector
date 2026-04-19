package parser

import (
	"encoding/binary"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/agent-cairn/criu-inspector/pkg/models"
)

// ParseCheckpoint parses a CRIU checkpoint directory
func ParseCheckpoint(dir string, verbose bool) (*models.CheckpointData, error) {
	// Validate directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("checkpoint directory not found: %s", dir)
	}
	data := &models.CheckpointData{}

	// Parse pstree.img (always)
	pstreePath := filepath.Join(dir, "pstree.img")
	if _, err := os.Stat(pstreePath); err == nil {
		processes, err := ParsePstree(pstreePath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse pstree.img: %w", err)
		}
		data.Processes = processes
	}

	// Parse inventory.img (verbose only)
	if verbose {
		inventoryPath := filepath.Join(dir, "inventory.img")
		if _, err := os.Stat(inventoryPath); err == nil {
			inventory, err := ParseInventory(inventoryPath)
			if err != nil {
				return nil, fmt.Errorf("failed to parse inventory.img: %w", err)
			}
			data.Inventory = inventory
		}
	}

	// Parse fds.img (verbose only)
	if verbose {
		fdsPath := filepath.Join(dir, "fds.img")
		if _, err := os.Stat(fdsPath); err == nil {
			fds, err := ParseFds(fdsPath)
			if err != nil {
				return nil, fmt.Errorf("failed to parse fds.img: %w", err)
			}
			data.FileDescriptors = fds
		}
	}

	return data, nil
}

// ParsePstree parses the pstree.img file (binary format with embedded JSON)
func ParsePstree(path string) ([]*models.ProcessEntry, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read pstree.img: %w", err)
	}

	// CRIU pstree.img has a binary header, JSON starts after '['
	// Search for the first '[' character
	jsonStart := -1
	for i, b := range content {
		if b == '[' {
			jsonStart = i
			break
		}
	}

	if jsonStart == -1 {
		// No JSON found, try binary parsing
		entries, err := parsePstreeBinary(content)
		if err != nil {
			return nil, fmt.Errorf("failed to parse pstree: no JSON found and binary parse failed: %w", err)
		}
		return entries, nil
	}

	// Parse JSON portion
	jsonBytes := content[jsonStart:]
	var entries []*models.ProcessEntry
	if err := json.Unmarshal(jsonBytes, &entries); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return entries, nil
}

// parsePstreeBinary attempts to parse pstree.img as a binary format
// This is a fallback for when no JSON is embedded
func parsePstreeBinary(content []byte) ([]*models.ProcessEntry, error) {
	// This is a simplified parser for the binary format
	// Real CRIU uses a complex protobuf-like format
	// For demo purposes, we'll extract what we can

	entries := []*models.ProcessEntry{}

	// Skip header (first 100 bytes typically)
	if len(content) < 100 {
		return nil, fmt.Errorf("content too short to be valid pstree.img")
	}

	// Try to find patterns that look like PID/PPID pairs
	// This is heuristic and may not work for all checkpoint formats
	offset := 100
	for offset < len(content)-16 {
		// Look for reasonable PID values (1-65535)
		pid := binary.LittleEndian.Uint32(content[offset : offset+4])
		ppid := binary.LittleEndian.Uint32(content[offset+4 : offset+8])

		// Heuristic: valid process tree entries
		if pid > 0 && pid < 65536 && ppid < 65536 {
			entry := &models.ProcessEntry{
				PID:      pid,
				PPID:     ppid,
				PGID:     pid, // Assume process group = PID
				SID:      pid, // Assume session = PID
				Comm:     "process", // Unknown without full parsing
				NThreads: 1,
			}
			entries = append(entries, entry)
			offset += 16 // Skip to next potential entry
		} else {
			offset++
		}
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("no valid process entries found in binary format")
	}

	return entries, nil
}

// ParseInventory parses inventory.img (XML format)
func ParseInventory(path string) (*models.Inventory, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read inventory.img: %w", err)
	}

	// Try to parse as XML
	var inventory models.Inventory
	if err := xml.Unmarshal(content, &inventory); err != nil {
		// Fallback: extract key-value pairs manually
		return parseInventoryFallback(string(content))
	}

	return &inventory, nil
}

// parseInventoryFallback extracts inventory data manually if XML parsing fails
func parseInventoryFallback(content string) (*models.Inventory, error) {
	inventory := &models.Inventory{
		Magic:   "UNKNOWN",
		Arch:    "UNKNOWN",
		PagesID: "UNKNOWN",
		GenID:   "UNKNOWN",
	}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "<magic>") {
			start := strings.Index(line, "<magic>") + 7
			end := strings.Index(line, "</magic>")
			if end > start {
				inventory.Magic = line[start:end]
			}
		} else if strings.Contains(line, "<arch>") {
			start := strings.Index(line, "<arch>") + 6
			end := strings.Index(line, "</arch>")
			if end > start {
				inventory.Arch = line[start:end]
			}
		} else if strings.Contains(line, "<pages_id>") {
			start := strings.Index(line, "<pages_id>") + 10
			end := strings.Index(line, "</pages_id>")
			if end > start {
				inventory.PagesID = line[start:end]
			}
		} else if strings.Contains(line, "<gen_id>") {
			start := strings.Index(line, "<gen_id>") + 8
			end := strings.Index(line, "</gen_id>")
			if end > start {
				inventory.GenID = line[start:end]
			}
		} else if strings.Contains(line, "<entries>") {
			start := strings.Index(line, "<entries>") + 9
			end := strings.Index(line, "</entries>")
			if end > start {
				var entries uint32
				fmt.Sscanf(line[start:end], "%d", &entries)
				inventory.Entries = entries
			}
		} else if strings.Contains(line, "<image_pages>") {
			start := strings.Index(line, "<image_pages>") + 12
			end := strings.Index(line, "</image_pages>")
			if end > start {
				var pages uint32
				fmt.Sscanf(line[start:end], "%d", &pages)
				inventory.ImagePages = pages
			}
		}
	}

	return inventory, nil
}

// ParseFds parses fds.img (XML format)
func ParseFds(path string) ([]*models.FileDescriptor, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read fds.img: %w", err)
	}

	// Parse XML structure manually for robustness
	fds := []*models.FileDescriptor{}
	lines := strings.Split(string(content), "\n")

	var currentFD *models.FileDescriptor
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Start of FD element
		if strings.Contains(line, "<fd ") {
			currentFD = &models.FileDescriptor{
				Type: "unknown",
			}

			// Parse attributes
			if idx := strings.Index(line, `fd="`); idx >= 0 {
				start := idx + 4
				end := strings.Index(line[start:], `"`)
				if end > 0 {
					var fd int32
					fmt.Sscanf(line[start:start+end], "%d", &fd)
					currentFD.FD = fd
				}
			}

			if idx := strings.Index(line, `type="`); idx >= 0 {
				start := idx + 6
				end := strings.Index(line[start:], `"`)
				if end > 0 {
					currentFD.Type = line[start : start+end]
				}
			}

			if idx := strings.Index(line, `path="`); idx >= 0 {
				start := idx + 6
				end := strings.Index(line[start:], `"`)
				if end > 0 {
					currentFD.Path = line[start : start+end]
				}
			}

			if idx := strings.Index(line, `flags="`); idx >= 0 {
				start := idx + 7
				end := strings.Index(line[start:], `"`)
				if end > 0 {
					currentFD.Flags = line[start : start+end]
				}
			}

			if idx := strings.Index(line, `id="`); idx >= 0 {
				start := idx + 4
				end := strings.Index(line[start:], `"`)
				if end > 0 {
					currentFD.ID = line[start : start+end]
				}
			}

			if idx := strings.Index(line, `mnt_id="`); idx >= 0 {
				start := idx + 8
				end := strings.Index(line[start:], `"`)
				if end > 0 {
					var mntID uint32
					fmt.Sscanf(line[start:start+end], "%d", &mntID)
					currentFD.MntID = mntID
				}
			}

			// Self-closing tag
			if strings.Contains(line, "/>") {
				fds = append(fds, currentFD)
				currentFD = nil
			}
		} else if strings.Contains(line, "</fd>") && currentFD != nil {
			fds = append(fds, currentFD)
			currentFD = nil
		}
	}

	return fds, nil
}
