use std::{
    io,
    io::{stdout, BufWriter, Read, Write},
};

use scip_syntax::ctags::{generate_tags, Reply, Request};

fn main() {
    println!(
        "{}\n",
        serde_json::to_string(&Reply::Program {
            name: "SCIP Ctags".to_string(),
            version: "5.9.0".to_string(),
        })
        .unwrap()
    );

    loop {
        let mut line = String::new();
        std::io::stdin()
            .read_line(&mut line)
            .expect("Could not read line");

        if line.is_empty() {
            break;
        }

        let mut buf_writer = BufWriter::new(stdout());

        let request = serde_json::from_str::<Request>(&line).unwrap();
        match request {
            Request::GenerateTags { filename, size } => {
                // eprintln!("{filename}");

                let mut file_data = vec![0; size];
                io::stdin()
                    .read_exact(&mut file_data)
                    .expect("Could not read file data");
                generate_tags(&mut buf_writer, filename, &file_data);
            }
        }

        Reply::Completed {
            command: "generate-tags".to_string(),
        }
        .write(&mut buf_writer);

        buf_writer.flush().unwrap();
    }
}
