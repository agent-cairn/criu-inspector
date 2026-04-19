use crate::error::{InspectorError, Result};
use crate::models::{CheckpointData, FileDescriptor, Inventory, ProcessEntry};
use std::fs;
use std::path::Path;

/// Parser for CRIU checkpoint files
pub struct CriuParser;

impl CriuParser {
    /// Parse a CRIU checkpoint directory
    pub fn parse_checkpoint_dir(dir: &Path, verbose: bool) -> Result<CheckpointData> {
        let mut data = CheckpointData::default();

        // Parse pstree.img (binary format with embedded JSON)
        let pstree_path = dir.join("pstree.img");
        if pstree_path.exists() {
            data.processes = Self::parse_pstree(&pstree_path)?;
        }

        // Parse inventory.img (XML format) - only in verbose mode
        if verbose {
            let inventory_path = dir.join("inventory.img");
            if inventory_path.exists() {
                data.inventory = Some(Self::parse_inventory(&inventory_path)?);
            }
        }

        // Parse fds.img (XML format) - only in verbose mode
        if verbose {
            let fds_path = dir.join("fds.img");
            if fds_path.exists() {
                data.file_descriptors = Self::parse_fds(&fds_path)?;
            }
        }

        Ok(data)
    }

    /// Parse pstree.img - extracts JSON from binary CRIU format
    fn parse_pstree(path: &Path) -> Result<Vec<ProcessEntry>> {
        let content = fs::read(path)?;

        // CRIU pstree.img format: binary header + JSON data
        // We'll search for the first '[' character to find JSON start
        let json_start = content
            .iter()
            .position(|&b| b == b'[')
            .ok_or_else(|| InspectorError::Parse("No JSON found in pstree.img".to_string()))?;

        let json_str = std::str::from_utf8(&content[json_start..])
            .map_err(|e| InspectorError::Parse(format!("Invalid UTF-8: {}", e)))?;

        let entries: Vec<ProcessEntry> = serde_json::from_str(json_str)
            .map_err(|e| InspectorError::Parse(format!("Failed to parse JSON: {}", e)))?;

        Ok(entries)
    }

    /// Parse inventory.img (XML format)
    fn parse_inventory(path: &Path) -> Result<Inventory> {
        let content = fs::read_to_string(path)?;

        let inventory: Inventory = quick_xml::de::from_str(&content)
            .map_err(|e| InspectorError::Parse(format!("Failed to parse inventory XML: {}", e)))?;

        Ok(inventory)
    }

    /// Parse fds.img (XML format)
    fn parse_fds(path: &Path) -> Result<Vec<FileDescriptor>> {
        let content = fs::read_to_string(path)?;

        // FDS XML format: <fd id="..." fd="..." type="..." ... />
        // We'll parse it manually for more flexibility
        let mut fds = Vec::new();
        let reader = &mut content.as_bytes();

        let mut parser = quick_xml::Reader::from_reader(reader);
        parser.config_mut().trim_text(true);

        let mut buf = Vec::new();

        loop {
            match parser.read_event_into(&mut buf) {
                Ok(quick_xml::events::Event::Start(e)) => {
                    if e.name().as_ref() == b"fd" {
                        let mut fd_entry = FileDescriptor {
                            fd: 0,
                            fd_type: String::new(),
                            path: None,
                            flags: None,
                            id: None,
                            mnt_id: None,
                        };

                        for attr in e.attributes() {
                            if let Ok(attr) = attr {
                                let key = std::str::from_utf8(attr.key.as_ref())
                                    .unwrap_or("");
                                let value = std::str::from_utf8(&attr.value)
                                    .unwrap_or("");

                                match key {
                                    "fd" => fd_entry.fd = value.parse().unwrap_or(0),
                                    "type" => fd_entry.fd_type = value.to_string(),
                                    "path" => fd_entry.path = Some(value.to_string()),
                                    "flags" => fd_entry.flags = Some(value.to_string()),
                                    "id" => fd_entry.id = Some(value.to_string()),
                                    "mnt_id" => fd_entry.mnt_id = value.parse().ok(),
                                    _ => {}
                                }
                            }
                        }

                        fds.push(fd_entry);
                    }
                }
                Ok(quick_xml::events::Event::Eof) => break,
                Err(e) => return Err(InspectorError::Parse(format!("XML parse error: {}", e))),
                _ => {}
            }
            buf.clear();
        }

        Ok(fds)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::io::Write;
    use tempfile::TempDir;

    #[test]
    fn test_parse_pstree_json() {
        let json_data = r#"[{
            "pid": 1,
            "ppid": 0,
            "pgid": 1,
            "sid": 1,
            "comm": "init",
            "nthreads": 1
        }, {
            "pid": 2,
            "ppid": 1,
            "pgid": 1,
            "sid": 1,
            "comm": "sh",
            "nthreads": 1
        }]"#;

        let mut binary_data = vec![0u8; 100]; // Mock binary header
        binary_data.extend_from_slice(json_data.as_bytes());

        let tmp = TempDir::new().unwrap();
        let pstree_path = tmp.path().join("pstree.img");
        fs::write(&pstree_path, binary_data).unwrap();

        let result = CriuParser::parse_pstree(&pstree_path).unwrap();

        assert_eq!(result.len(), 2);
        assert_eq!(result[0].pid, 1);
        assert_eq!(result[0].comm, "init");
        assert_eq!(result[1].pid, 2);
        assert_eq!(result[1].ppid, 1);
    }

    #[test]
    fn test_parse_inventory_xml() {
        let xml_data = r#"<?xml version="1.0" encoding="UTF-8"?>
<inventory>
    <magic>CRIUIMG</magic>
    <arch>x86_64</arch>
    <pages_id>12345</pages_id>
    <gen_id>67890</gen_id>
    <entries>42</entries>
    <image_pages>1024</image_pages>
</inventory>"#;

        let tmp = TempDir::new().unwrap();
        let inventory_path = tmp.path().join("inventory.img");
        fs::write(&inventory_path, xml_data).unwrap();

        let result = CriuParser::parse_inventory(&inventory_path).unwrap();

        assert_eq!(result.magic, "CRIUIMG");
        assert_eq!(result.arch, "x86_64");
        assert_eq!(result.pages_id, "12345");
        assert_eq!(result.entries, 42);
        assert_eq!(result.image_pages, Some(1024));
    }

    #[test]
    fn test_parse_fds_xml() {
        let xml_data = r#"<?xml version="1.0" encoding="UTF-8"?>
<fds>
    <fd id="1" fd="0" type="pipe" flags="02000000" />
    <fd id="2" fd="1" type="file" path="/etc/hosts" flags="0100000" />
    <fd id="3" fd="2" type="socket" flags="02000002" />
</fds>"#;

        let tmp = TempDir::new().unwrap();
        let fds_path = tmp.path().join("fds.img");
        fs::write(&fds_path, xml_data).unwrap();

        let result = CriuParser::parse_fds(&fds_path).unwrap();

        assert_eq!(result.len(), 3);
        assert_eq!(result[0].fd, 0);
        assert_eq!(result[0].fd_type, "pipe");
        assert_eq!(result[1].fd, 1);
        assert_eq!(result[1].path, Some("/etc/hosts".to_string()));
        assert_eq!(result[2].fd_type, "socket");
    }
}
