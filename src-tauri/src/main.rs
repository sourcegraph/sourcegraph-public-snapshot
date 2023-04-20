#![cfg_attr(
    all(not(debug_assertions), target_os = "windows"),
    windows_subsystem = "windows"
)]

use tauri::api::process::Command;
use tauri::api::process::CommandEvent;
use tauri::Manager;

fn main() {
    match fix_path_env::fix() {
        Ok(_) => {}
        Err(e) => {
            println!("Error fixing path environment: {}", e);
        }
    }
    tauri::Builder::default()
        .setup(|app| {
            let window = app.get_window("main").unwrap();
            let (mut rx, mut child) = Command::new_sidecar("backend")
                .expect("failed to create `backend` binary command")
                .spawn()
                .expect("Failed to spawn backend sidecar");

            tauri::async_runtime::spawn(async move {
                // read events such as stdout
                while let Some(event) = rx.recv().await {
                    if let CommandEvent::Stdout(line) = event {
                        window
                            .emit("message", Some(format!("'{}'", line)))
                            .expect("failed to emit event");
                        // write to stdin
                        child.write("message from Rust\n".as_bytes()).unwrap();
                    }
                }
            });

            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
