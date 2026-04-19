use std::fs::{self, File};
use std::io::Write;
use std::path::PathBuf;
use tempfile::TempDir;

/// Create a mock CRIU checkpoint directory with sample data
fn create_mock_checkpoint() -> TempDir {
    let tmp_dir = TempDir::new().unwrap();
    let dir = tmp_dir.path();

    // Create pstree.img (binary header + JSON)
    let pstree_json = r#"[
        {
            "pid": 1,
            "ppid": 0,
            "pgid": 1,
            "sid": 1,
            "comm": "systemd",
            "nthreads": 1
        },
        {
            "pid": 100,
            "ppid": 1,
            "pgid": 100,
            "sid": 1,
            "comm": "ssh",
            "nthreads": 2
        },
        {
            "pid": 200,
            "ppid": 1,
            "pgid": 100,
            "sid": 1,
            "comm": "nginx",
            "nthreads": 4
        }
    ]"#;

    let mut pstree_data = vec![0u8; 256]; // Simulate binary header
    pstree_data.extend_from_slice(pstree_json.as_bytes());
    fs::write(dir.join("pstree.img"), pstree_data).unwrap();

    // Create inventory.img (XML)
    let inventory_xml = r#"<?xml version="1.0" encoding="UTF-8"?>
<inventory>
    <magic>CRIUIMG</magic>
    <arch>x86_64</arch>
    <pages_id>ABC123</pages_id>
    <gen_id>XYZ789</gen_id>
    <entries>3</entries>
    <image_pages>4096</image_pages>
</inventory>"#;
    fs::write(dir.join("inventory.img"), inventory_xml).unwrap();

    // Create fds.img (XML)
    let fds_xml = r#"<?xml version="1.0" encoding="UTF-8"?>
<fds>
    <fd id="1" fd="0" type="pipe" flags="02000000" />
    <fd id="2" fd="1" type="pipe" flags="02000000" />
    <fd id="3" fd="2" type="file" path="/etc/hosts" flags="0100000" mnt_id="42" />
    <fd id="4" fd="3" type="file" path="/var/log/nginx/access.log" flags="0100002" mnt_id="42" />
    <fd id="5" fd="4" type="socket" flags="02000002" />
    <fd id="6" fd="5" type="socket" flags="02000002" />
</fds>"#;
    fs::write(dir.join("fds.img"), fds_xml).unwrap();

    tmp_dir
}

#[test]
fn test_full_checkpoint_parse_basic() {
    let tmp_dir = create_mock_checkpoint();
    let data = criu_inspector::CriuParser::parse_checkpoint_dir(tmp_dir.path(), false).unwrap();

    // Basic mode: only processes parsed
    assert_eq!(data.processes.len(), 3);
    assert!(data.inventory.is_none());
    assert!(data.file_descriptors.is_empty());
}

#[test]
fn test_full_checkpoint_parse_verbose() {
    let tmp_dir = create_mock_checkpoint();
    let data = criu_inspector::CriuParser::parse_checkpoint_dir(tmp_dir.path(), true).unwrap();

    // Verbose mode: all sections parsed
    assert_eq!(data.processes.len(), 3);

    // Check inventory
    assert!(data.inventory.is_some());
    let inventory = data.inventory.as_ref().unwrap();
    assert_eq!(inventory.magic, "CRIUIMG");
    assert_eq!(inventory.arch, "x86_64");
    assert_eq!(inventory.entries, 3);
    assert_eq!(inventory.image_pages, Some(4096));

    // Check file descriptors
    assert_eq!(data.file_descriptors.len(), 6);
    assert_eq!(data.file_descriptors[0].fd, 0);
    assert_eq!(data.file_descriptors[0].fd_type, "pipe");
    assert_eq!(data.file_descriptors[2].fd, 2);
    assert_eq!(data.file_descriptors[2].path, Some("/etc/hosts".to_string()));
    assert_eq!(data.file_descriptors[2].mnt_id, Some(42));
}

#[test]
fn test_partial_checkpoint_missing_optional_files() {
    let tmp_dir = TempDir::new().unwrap();
    let dir = tmp_dir.path();

    // Only create pstree.img
    let pstree_json = r#"[{
        "pid": 1,
        "ppid": 0,
        "pgid": 1,
        "sid": 1,
        "comm": "bash",
        "nthreads": 1
    }]"#;

    let mut pstree_data = vec![0u8; 100];
    pstree_data.extend_from_slice(pstree_json.as_bytes());
    fs::write(dir.join("pstree.img"), pstree_data).unwrap();

    // Parse in verbose mode - should handle missing files gracefully
    let data = criu_inspector::CriuParser::parse_checkpoint_dir(tmp_dir.path(), true).unwrap();

    assert_eq!(data.processes.len(), 1);
    assert!(data.inventory.is_none());
    assert!(data.file_descriptors.is_empty());
}

#[test]
fn test_checkpoint_serialization() {
    let tmp_dir = create_mock_checkpoint();
    let data = criu_inspector::CriuParser::parse_checkpoint_dir(tmp_dir.path(), true).unwrap();

    // Test JSON serialization
    let json = serde_json::to_string(&data).unwrap();
    let deserialized: criu_inspector::CheckpointData = serde_json::from_str(&json).unwrap();

    assert_eq!(data.processes.len(), deserialized.processes.len());
    assert_eq!(data.inventory, deserialized.inventory);
    assert_eq!(data.file_descriptors.len(), deserialized.file_descriptors.len());
}

#[test]
fn test_empty_directory() {
    let tmp_dir = TempDir::new().unwrap();
    let dir = tmp_dir.path();

    let data = criu_inspector::CriuParser::parse_checkpoint_dir(dir, true).unwrap();

    assert!(data.processes.is_empty());
    assert!(data.inventory.is_none());
    assert!(data.file_descriptors.is_empty());
}

#[test]
fn test_complex_fds_with_all_attributes() {
    let tmp_dir = TempDir::new().unwrap();
    let dir = tmp_dir.path();

    // Create FDS with all attributes
    let fds_xml = r#"<?xml version="1.0" encoding="UTF-8"?>
<fds>
    <fd id="123" fd="3" type="file" path="/tmp/test" flags="0100002" mnt_id="77" />
    <fd id="124" fd="4" type="eventpoll" flags="02000002" />
</fds>"#;

    fs::write(dir.join("fds.img"), fds_xml).unwrap();

    let data = criu_inspector::CriuParser::parse_checkpoint_dir(dir, true).unwrap();

    assert_eq!(data.file_descriptors.len(), 2);

    let fd1 = &data.file_descriptors[0];
    assert_eq!(fd1.fd, 3);
    assert_eq!(fd1.fd_type, "file");
    assert_eq!(fd1.path, Some("/tmp/test".to_string()));
    assert_eq!(fd1.flags, Some("0100002".to_string()));
    assert_eq!(fd1.id, Some("123".to_string()));
    assert_eq!(fd1.mnt_id, Some(77));

    let fd2 = &data.file_descriptors[1];
    assert_eq!(fd2.fd, 4);
    assert_eq!(fd2.fd_type, "eventpoll");
    assert_eq!(fd2.path, None);
    assert_eq!(fd2.mnt_id, None);
}
