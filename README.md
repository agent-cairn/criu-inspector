# criu-inspector

> A CLI tool for inspecting CRIU (Checkpoint/Restore In Userspace) checkpoint directories

**Demo-ready** output with color-coded tables and clear formatting — perfect for conference presentations and live demonstrations of container checkpoint/restore technology.

## What It Does

`criu-inspector` parses CRIU checkpoint directories and extracts structured information about:

- **Process Tree**: PID, PPID, process groups, sessions, and thread counts
- **Inventory Metadata**: Architecture, generation IDs, page counts
- **File Descriptors**: Open files, sockets, pipes with their flags and paths

Designed for live migration debugging, forensic analysis, and educational demos of CRIU internals.

## Install

### From Source

```bash
git clone https://github.com/agent-cairn/criu-inspector.git
cd criu-inspector
go build -o criu-inspector .
sudo mv criu-inspector /usr/local/bin/
```

### Requirements

- Go 1.22 or later
- A CRIU checkpoint directory (created with `criu dump`)

## Usage

### Basic Inspection

```bash
criu-inspector inspect /path/to/checkpoint
```

### Verbose Output (includes inventory and file descriptors)

```bash
criu-inspector inspect --verbose /path/to/checkpoint
```

### JSON Output (for automation)

```bash
criu-inspector inspect --json /path/to/checkpoint
```

### Pretty JSON

```bash
criu-inspector inspect --json --pretty /path/to/checkpoint
```

## Sample Output

```
═══════════════════════════════════════════════════════════════
  CRIU CHECKPOINT INSPECTION REPORT
═══════════════════════════════════════════════════════════════

📊 Process Tree
─────────────────────────────────────────────────────────────────
PID      PPID     PGID     SID      COMM                THREADS
─────────────────────────────────────────────────────────────────
1        0        1        1        systemd             1
1234     1        1234     1234     nginx               4
1235     1234     1234     1234     nginx: worker       4
1236     1234     1234     1234     nginx: worker       4
  Total processes: 4

─ Use --verbose for full details, --json for machine-readable output
```

With `--verbose`:

```
📦 Inventory Metadata
─────────────────────────────────────────────────────────────────
  Magic               : CRIMG_V1
  Architecture        : x86_64
  Pages ID            : 0x7f8e4a2b3c1d
  Generation ID       : 0x9d8e7f6a5b4c
  Entries             : 42
  Image Pages         : 15230

📂 File Descriptors
─────────────────────────────────────────────────────────────────
  FD  Type       Path                          Flags
─────────────────────────────────────────────────────────────────
  0   pipe       /tmp/nginx.sock               O_RDONLY
  1   file       /var/log/nginx/access.log     O_WRONLY|O_APPEND
  2   file       /var/log/nginx/error.log      O_WRONLY|O_APPEND
  3   socket     127.0.0.1:8080                O_RDWR
  Total FDs: 4
```

## Architecture

```
criu-inspector/
├── cmd/
│   ├── root.go          # Root command and CLI entry point
│   └── inspect.go       # Inspect subcommand with flags
├── pkg/
│   ├── models/
│   │   └── models.go    # Data structures for checkpoint data
│   ├── parser/
│   │   ├── parser.go    # Core parsing logic (JSON/binary/XML)
│   │   └── parser_test.go
│   └── output/
│       └── output.go    # Formatted, color-coded output
├── go.mod
└── README.md
```

### Parsing Strategy

CRIU checkpoint images use multiple formats:

1. **pstree.img**: Binary header with embedded JSON (primary format)
2. **inventory.img**: XML format (or fallback to manual parsing)
3. **fds.img**: XML format with manual extraction for robustness

The parser tries JSON first, then falls back to a heuristic binary parser for process tree data — sufficient for demos and debugging.

## CRIU Checkpoint Files

| File | Description | Parsed by Default |
|------|-------------|-------------------|
| `pstree.img` | Process hierarchy | ✅ Yes |
| `inventory.img` | Checkpoint metadata | `--verbose` only |
| `fds.img` | File descriptors | `--verbose` only |
| `core-*.img` | Process memory | Displayed in inventory |
| `mm-*.img` | Memory maps | Not yet implemented |

## Use Cases

- **Conference Demos**: Show audience what's inside a checkpoint with colorful, readable output
- **Live Migration Debugging**: Inspect checkpointed state before restore
- **Forensics**: Analyze process state from a frozen container
- **Education**: Teach CRIU internals by examining checkpoint files

## Creating Test Checkpoints

```bash
# Run a process
sleep 60 &
PID=$!

# Dump it with CRIU
sudo mkdir -p /tmp/checkpoint
sudo criu dump -t $PID -D /tmp/checkpoint --shell-job

# Inspect with criu-inspector
criu-inspector inspect --verbose /tmp/checkpoint
```

## Limitations

- Binary parsing is heuristic — not all CRIU versions may be fully supported
- Memory maps (`mm-*.img`) and network state not yet parsed
- Designed for inspection, not modification or restoration

## Contributing

This is a demo/educational tool for conference talks. Contributions welcome for:

- Additional checkpoint file formats (mm-*.img, net-*.img)
- Better binary format parsing
- Export to other formats (YAML, CSV)

## License

MIT

## Author

Built for conference demonstrations of CRIU and container checkpoint/restore technology.
