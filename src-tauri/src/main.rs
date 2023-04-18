#![cfg_attr(
    all(not(debug_assertions), target_os = "windows"),
    windows_subsystem = "windows"
)]

// use std::collections::HashMap;

// use tauri::api::process::Command;
// use tauri::api::process::CommandEvent;
use tauri::Manager;

fn main() {
    tauri::Builder::default()
        .setup(|app| {
            let window = app.get_window("main").unwrap();
            // let vars: HashMap<String, String> = HashMap::from([
            //     ("DEPLOY_TYPE".to_string(), "app".to_string()),
            //     ("USE_EMBEDDED_POSTGRESQL".to_string(), "0".to_string()),
            // ]);
            // let (mut rx, mut child) = Command::new_sidecar("backend")
            //     .expect("failed to create `backend` binary command")
            //     .envs(vars)
            //     .spawn()
            //     .expect("Failed to spawn backend sidecar");

            window.emit("backend-message", "hello world");

            // let _win_clone = window.clone();
            // tauri::async_runtime::spawn(async move {
            //     // read events such as stdout
            //     println!("INSIDE SPAWN");
            //     while let Some(event) = rx.recv().await {
            //         if let CommandEvent::Stdout(line) = event {
            //             println!("backend-message: {}", line);
            //             window
            //                 .emit("backend-message", Some(format!("'{}'", line)))
            //                 .expect("failed to emit event");
            //         }
            //     }
            // });

            // tauri::async_runtime::spawn(async move {
            //     win_clone.open_devtools();
            // });

            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
