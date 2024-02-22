use csv;
use serde::Deserialize;
use std::collections::{HashMap, HashSet};
use std::{error::Error, io};

#[derive(Debug, Deserialize)]
struct Record {
    id: u32,
    name_segment: String,
    prefix_id: Option<u32>,
}

fn main() -> Result<(), Box<dyn Error>> {
    let mut rdr = csv::Reader::from_reader(io::stdin());
    let mut records = HashMap::new();
    let mut deleted_ids = HashSet::new();
    for result in rdr.deserialize() {
        let record: Record = result?;
        records.insert(record.id, record);
    }

    let initial_size = records.len();

    loop {
        let mut seen: HashMap<(String, Option<u32>), u32> = HashMap::new();
        let mut remap = HashMap::new();

        for (id, record) in &records {
            let key = (record.name_segment.clone(), record.prefix_id);
            if let Some(prev) = seen.get(&key) {
                remap.insert(*id, *prev);
            } else {
                seen.insert(key, *id);
            }
        }

        // Fixpoint reached
        if remap.is_empty() {
            break;
        }

        let keys: Vec<u32> = records.keys().copied().collect();

        for id in keys {
            if remap.contains_key(&id) {
                deleted_ids.insert(id);
                records.remove(&id);
                continue;
            }

            records.entry(id).and_modify(|e| {
                if let Some(parent) = e.prefix_id {
                    e.prefix_id = Some(remap.get(&parent).copied().unwrap_or(parent))
                }
            });
        }
    }

    let compressed_size = records.len();
    let saved = (initial_size as f64 - compressed_size as f64) / initial_size as f64 * 100f64;

    println!("from {initial_size} rows\nto   {compressed_size} rows\nreduction by {saved:.2}%");

    Ok(())
}
