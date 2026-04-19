use thiserror::Error;

#[derive(Debug, Error)]
pub enum InspectorError {
    #[error("IO error: {0}")]
    Io(#[from] std::io::Error),

    #[error("Parse error: {0}")]
    Parse(String),

    #[error("XML error: {0}")]
    Xml(#[from] quick_xml::DeError),

    #[error("File not found: {0}")]
    FileNotFound(String),
}

pub type Result<T> = std::result::Result<T, InspectorError>;
