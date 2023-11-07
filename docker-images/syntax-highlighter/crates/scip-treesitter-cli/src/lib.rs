pub mod evaluate;
pub mod index;

pub mod progress {
    use indicatif::{ProgressBar, ProgressStyle};
    pub fn create_spinner() -> ProgressBar {
        let bar = ProgressBar::new_spinner();

        bar.set_style(
            ProgressStyle::with_template("{spinner:.blue} {msg}")
                .unwrap()
                .tick_strings(&[
                    "▹▹▹▹▹",
                    "▸▹▹▹▹",
                    "▹▸▹▹▹",
                    "▹▹▸▹▹",
                    "▹▹▹▸▹",
                    "▹▹▹▹▸",
                    "▪▪▪▪▪",
                ]),
        );

        bar
    }

    pub fn create_progress_bar(len: u64) -> ProgressBar {
        let bar = ProgressBar::new(len);

        bar.set_style(
            ProgressStyle::with_template(
                "[{elapsed_precise}] {bar:40.cyan/blue} {pos:>7}/{len:7}\n {msg}",
            )
            .unwrap()
            .progress_chars("##-"),
        );

        bar
    }
}

pub mod io {

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
}
