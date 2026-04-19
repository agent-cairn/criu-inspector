# criu-inspector

CRIU (Checkpoint/Restore In Userspace) checkpoint inspector tool. Parses CRIU checkpoint directories and extracts structured information about processes, inventory metadata, and file descriptors.

## Features

- **Process Tree Parsing**: Extracts process hierarchy from `pstree.img`
- **Inventory Metadata**: Parses `inventory.img` for dump architecture, pages info, and entry counts
- **File Descriptors**: Parses `fds.img` for open file descriptors, sockets, and pipes
- **Verbose Mode**: Optional detailed output including inventory and file descriptors
- **JSON Output**: Structured JSON output (compact or pretty-printed)

## Installation

```bash
cargo install --git https://github.com/agent-cairn/criu-inspector
```

## Usage

### Basic usage (process tree only)
```bash
criu-inspector --dir /path/to/checkpoint
```

### Verbose output (includes inventory and file descriptors)
```bash
criu-inspector --dir /path/to/checkpoint --verbose
```

### Pretty-printed JSON
```bash
criu-inspector --dir /path/to/checkpoint --verbose --pretty
```

## Output Format

### Basic Output
```json
{
  "processes": [
    {
      "pid": 1,
      "ppid": 0,
      "pgid": 1,
      "sid": 1,
      "comm": "systemd",
      "nthreads": 1
    }
  ],
  "inventory": null,
  "file_descriptors": []
}
```

### Verbose Output
```json
{
  "processes": [...],
  "inventory": {
    "magic": "CRIUIMG",
    "arch": "x86_64",
    "pages_id": "ABC123",
    "gen_id": "XYZ789",
    "entries": 3,
    "image_pages": 4096
  },
  "file_descriptors": [
    {
      "fd": 0,
      "type": "pipe",
      "path": null,
      "flags": "02000000",
      "id": "1",
      "mnt_id": null
    },
    {
      "fd": 1,
      "type": "file",
      "path": "/etc/hosts",
      "flags": "0100000",
      "id": "2",
      "mnt_id": 42
    }
  ]
}
```

## Development

### Run tests
```bash
cargo test
```

### Run with clippy
```bash
cargo clippy -- -D warnings
```

### Format code
```bash
cargo fmt
```

### Build
```bash
cargo build --release
```

## License

MIT
