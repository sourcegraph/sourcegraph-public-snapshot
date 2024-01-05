use std::{fs::File, io::BufReader, path::PathBuf};

use protobuf::{CodedInputStream, Message};

pub fn read_index_from_file(file: PathBuf) -> scip::types::Index {
    let mut candidate_idx = scip::types::Index::new();
    let candidate_f = File::open(file).unwrap();

    let mut reader = BufReader::new(candidate_f);
    let mut cis = CodedInputStream::from_buf_read(&mut reader);

    candidate_idx.merge_from(&mut cis).unwrap();

    candidate_idx
}
