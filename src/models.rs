use serde::{Deserialize, Serialize};

/// Process tree entry from pstree.img
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct ProcessEntry {
    pub pid: u32,
    pub ppid: u32,
    pub pgid: u32,
    pub sid: u32,
    pub comm: String,
    pub nthreads: u32,
}

/// Inventory metadata from inventory.img (in XML format)
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct Inventory {
    #[serde(rename = "magic")]
    pub magic: String,

    #[serde(rename = "arch")]
    pub arch: String,

    #[serde(rename = "pages_id")]
    pub pages_id: String,

    #[serde(rename = "gen_id")]
    pub gen_id: String,

    #[serde(rename = "entries")]
    pub entries: u32,

    #[serde(rename = "image_pages")]
    pub image_pages: Option<u32>,
}

/// File descriptor entry from fds.img (in XML format)
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct FileDescriptor {
    #[serde(rename = "fd")]
    pub fd: i32,

    #[serde(rename = "type")]
    pub fd_type: String,

    #[serde(rename = "path")]
    pub path: Option<String>,

    #[serde(rename = "flags")]
    pub flags: Option<String>,

    #[serde(rename = "id")]
    pub id: Option<String>,

    #[serde(rename = "mnt_id")]
    pub mnt_id: Option<u32>,
}

/// Container for all parsed checkpoint data
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct CheckpointData {
    pub processes: Vec<ProcessEntry>,
    pub inventory: Option<Inventory>,
    pub file_descriptors: Vec<FileDescriptor>,
}

impl Default for CheckpointData {
    fn default() -> Self {
        Self {
            processes: Vec::new(),
            inventory: None,
            file_descriptors: Vec::new(),
        }
    }
}
