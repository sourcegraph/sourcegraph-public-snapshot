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
