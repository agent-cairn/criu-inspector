use clap::Parser;
use std::path::PathBuf;

mod error;
mod models;
mod parser;

use error::Result;
use parser::CriuParser;

/// CRIU checkpoint inspector tool
#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    /// Path to CRIU checkpoint directory
    #[arg(short, long)]
    dir: PathBuf,

    /// Verbose output (includes inventory and file descriptors)
    #[arg(short, long)]
    verbose: bool,

    /// Output as pretty-printed JSON (default: compact)
    #[arg(long)]
    pretty: bool,
}

fn main() -> Result<()> {
    let args = Args::parse();

    if !args.dir.exists() {
        return Err(error::InspectorError::FileNotFound(format!(
            "Directory not found: {}",
            args.dir.display()
        )));
    }

    if !args.dir.is_dir() {
        return Err(error::InspectorError::Parse(format!(
            "Not a directory: {}",
            args.dir.display()
        )));
    }

    let data = CriuParser::parse_checkpoint_dir(&args.dir, args.verbose)?;

    let output = if args.pretty {
        serde_json::to_string_pretty(&data)?
    } else {
        serde_json::to_string(&data)?
    };

    println!("{}", output);

    Ok(())
}
