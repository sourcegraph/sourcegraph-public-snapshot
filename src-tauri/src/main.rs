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
            let (mut rx, _child) = Command::new_sidecar("backend")
                .expect("failed to create `backend` binary command")
                .spawn()
                .expect("Failed to spawn backend sidecar");

            tauri::async_runtime::spawn(async move {
                // read events such as stdout
                while let Some(event) = rx.recv().await {
                    match event {
                        CommandEvent::Stdout(line) => {
                            window
                                .emit("backend-stdout", Some(line.clone()))
                                .expect("failed to emit event");

                            window.eval(&format!(
                                "console.log(\":: {}\")",
                                line.replace("\"", "\\\"")
                            ));
                        }
                        CommandEvent::Stderr(line) => {
                            window
                                .emit("backend-stderr", Some(line.clone()))
                                .expect("failed to emit event");

                            window.eval(&format!(
                                "console.log(\":: {}\")",
                                line.replace("\"", "\\\"")
                            ));
                        }
                        _ => continue,
                    };
                }
            });

            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
