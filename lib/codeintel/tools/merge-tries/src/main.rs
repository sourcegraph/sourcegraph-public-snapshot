use csv;
use serde::Deserialize;
use std::collections::HashMap;
use std::{error::Error, io};

#[derive(Debug, Deserialize)]
struct Record {
    id: u32,
    name_segment: String,
    prefix_id: Option<u32>,
}

fn main() -> Result<(), Box<dyn Error>> {
    let mut rdr = csv::Reader::from_reader(io::stdin());
    let mut records = Vec::new();
    for result in rdr.deserialize() {
        let record: Record = result?;
        records.push(record);
    }
    records.sort_by_key(|r| r.id);

    let initial_size = records.len();

    // id -> id
    let mut remap: HashMap<u32, u32> = HashMap::new();
    let mut earliest_ids = HashMap::new();
    for record in records.iter() {
        let mut prefix_id = record.prefix_id;
        if let Some(p_id) = prefix_id {
            if let Some(old_id) = remap.get(&p_id) {
                prefix_id = Some(*old_id);
            }
        }

        let key = (prefix_id, &record.name_segment);
        if let Some(earliest_id) = earliest_ids.get(&key) {
            remap.insert(record.id, *earliest_id);
            continue;
        }

        earliest_ids.insert(key, record.id);
    }

    let compressed_size = earliest_ids.len();
    let saved = (initial_size as f64 - compressed_size as f64) / initial_size as f64 * 100f64;

    println!("from {initial_size} rows\nto   {compressed_size} rows\nreduction by {saved:.2}%");

    Ok(())
}
