package output

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/agent-cairn/criu-inspector/pkg/models"
)

// ANSI color codes for demo-friendly output
var (
	boldGreen  = color.New(color.FgGreen, color.Bold).SprintFunc()
	boldBlue   = color.New(color.FgBlue, color.Bold).SprintFunc()
	boldYellow = color.New(color.FgYellow, color.Bold).SprintFunc()
	boldCyan   = color.New(color.FgCyan, color.Bold).SprintFunc()
	boldWhite  = color.New(color.FgWhite, color.Bold).SprintFunc()
	gray       = color.New(color.FgHiBlack).SprintFunc()
)

// PrintHumanReadable prints checkpoint data in a formatted, color-coded table
func PrintHumanReadable(data *models.CheckpointData, verbose bool) {
	fmt.Println()
	fmt.Println(boldBlue("═══════════════════════════════════════════════════════════════"))
	fmt.Println(boldBlue("  CRIU CHECKPOINT INSPECTION REPORT"))
	fmt.Println(boldBlue("═══════════════════════════════════════════════════════════════"))
	fmt.Println()

	// Process Tree
	printProcessTree(data.Processes)

	// Inventory (verbose only)
	if verbose && data.Inventory != nil {
		printInventory(data.Inventory)
	}

	// File Descriptors (verbose only)
	if verbose && len(data.FileDescriptors) > 0 {
		printFileDescriptors(data.FileDescriptors)
	}

	fmt.Println()
	fmt.Println(gray("─ Use --verbose for full details, --json for machine-readable output"))
	fmt.Println()
}

// printProcessTree prints the process hierarchy
func printProcessTree(processes []*models.ProcessEntry) {
	fmt.Println(boldGreen("📊 Process Tree"))
	fmt.Println(gray("─────────────────────────────────────────────────────────────────"))

	if len(processes) == 0 {
		fmt.Println(gray("  No process data found"))
		fmt.Println()
		return
	}

	// Table header
	fmt.Printf("%-8s %-8s %-8s %-8s %-20s %-8s\n",
		boldYellow("PID"),
		boldYellow("PPID"),
		boldYellow("PGID"),
		boldYellow("SID"),
		boldYellow("COMM"),
		boldYellow("THREADS"))
	fmt.Println(gray("─────────────────────────────────────────────────────────────────"))

	// Process entries
	for _, p := range processes {
		comm := p.Comm
		if len(comm) > 18 {
			comm = comm[:18] + "…"
		}

		// Highlight init process (PID 1)
		pidStr := fmt.Sprintf("%d", p.PID)
		if p.PID == 1 {
			pidStr = boldGreen(pidStr)
		}

		fmt.Printf("%-8s %-8d %-8d %-8d %-20s %-8d\n",
			pidStr, p.PPID, p.PGID, p.SID, comm, p.NThreads)
	}

	fmt.Printf(gray("  Total processes: %d\n\n"), len(processes))
}

// printInventory prints inventory metadata
func printInventory(inventory *models.Inventory) {
	fmt.Println(boldGreen("📦 Inventory Metadata"))
	fmt.Println(gray("─────────────────────────────────────────────────────────────────"))

	fmt.Printf("  %-18s : %s\n", boldYellow("Magic"), inventory.Magic)
	fmt.Printf("  %-18s : %s\n", boldYellow("Architecture"), boldCyan(inventory.Arch))
	fmt.Printf("  %-18s : %s\n", boldYellow("Pages ID"), inventory.PagesID)
	fmt.Printf("  %-18s : %s\n", boldYellow("Generation ID"), inventory.GenID)
	fmt.Printf("  %-18s : %d\n", boldYellow("Entries"), inventory.Entries)

	if inventory.ImagePages > 0 {
		fmt.Printf("  %-18s : %d pages\n", boldYellow("Image Pages"), inventory.ImagePages)
	}

	fmt.Println()
}

// printFileDescriptors prints file descriptor information
func printFileDescriptors(fds []*models.FileDescriptor) {
	fmt.Println(boldGreen("📂 File Descriptors"))
	fmt.Println(gray("─────────────────────────────────────────────────────────────────"))

	if len(fds) == 0 {
		fmt.Println(gray("  No file descriptors found"))
		fmt.Println()
		return
	}

	// Table header
	fmt.Printf("%-6s %-12s %-40s %-10s %-10s\n",
		boldYellow("FD"),
		boldYellow("TYPE"),
		boldYellow("PATH"),
		boldYellow("FLAGS"),
		boldYellow("ID"))
	fmt.Println(gray("─────────────────────────────────────────────────────────────────"))

	// FD entries
	for _, fd := range fds {
		fdType := colorType(fd.Type)
		fdNum := fmt.Sprintf("%d", fd.FD)
		if fd.FD == 0 || fd.FD == 1 || fd.FD == 2 {
			fdNum = boldYellow(fdNum) // Highlight stdin/stdout/stderr
		}

		path := fd.Path
		if path == "" {
			path = gray("(none)")
		} else if len(path) > 38 {
			path = path[:38] + "…"
		}

		flags := fd.Flags
		if flags == "" {
			flags = gray("(none)")
		}

		id := fd.ID
		if id == "" {
			id = gray("-")
		}

		fmt.Printf("%-6s %-12s %-40s %-10s %-10s\n",
			fdNum, fdType, path, flags, id)
	}

	fmt.Printf(gray("  Total file descriptors: %d\n\n"), len(fds))
}

// colorType returns a color-coded file descriptor type
func colorType(fdType string) string {
	switch fdType {
	case "file":
		return boldCyan(fdType)
	case "pipe":
		return boldGreen(fdType)
	case "socket":
		return boldYellow(fdType)
	case "eventfd":
		return color.New(color.FgMagenta, color.Bold).SprintFunc()(fdType)
	case "epoll":
		return color.New(color.FgHiMagenta, color.Bold).SprintFunc()(fdType)
	default:
		return fdType
	}
}

// InitTerminal ensures terminal output is properly initialized
func InitTerminal() {
	// Force enable colors on all terminals
	color.NoColor = false
}
