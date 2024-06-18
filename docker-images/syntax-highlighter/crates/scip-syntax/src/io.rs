use std::{fs::File, io::BufReader};

use anyhow::{Context, Result};
use camino::Utf8Path;
use protobuf::{CodedInputStream, Message};

pub fn read_index_from_file(file: &Utf8Path) -> Result<scip::types::Index> {
    let mut candidate_idx = scip::types::Index::new();
    let candidate_f =
        File::open(file).context(format!("when trying to read an index from {file}"))?;

    let mut reader = BufReader::new(candidate_f);
    let mut cis = CodedInputStream::from_buf_read(&mut reader);

    candidate_idx.merge_from(&mut cis)?;

    Ok(candidate_idx)
}
