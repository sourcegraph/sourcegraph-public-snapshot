#![cfg_attr(
    all(not(debug_assertions), target_os = "windows"),
    windows_subsystem = "windows"
)]

#[cfg(not(dev))]
use {tauri::api::process::Command, tauri::api::process::CommandEvent};

use tauri::Manager;

fn main() {
    match fix_path_env::fix() {
        Ok(_) => {}
        Err(e) => {
            println!("error fixing path environment: {}", e);
        }
    }

    tauri::Builder::default()
        .setup(|app| {
            let window = app.get_window("main").unwrap();
            start_embedded_services(window);
            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}

#[cfg(dev)]
fn start_embedded_services(_window: tauri::Window) {
    println!("embedded Sourcegraph services disabled for local development");
}

#[cfg(not(dev))]
fn start_embedded_services(window: tauri::Window) {
    let sidecar = "sourcegraph-backend";
    let (mut rx, _child) = Command::new_sidecar(sidecar)
        .expect(format!("failed to create `{sidecar}` binary command").as_str())
        .spawn()
        .expect(format!("failed to spawn {sidecar} sidecar").as_str());

    tauri::async_runtime::spawn(async move {
        // read events such as stdout
        while let Some(event) = rx.recv().await {
            match event {
                CommandEvent::Stdout(line) => {
                    window
                        .emit(format!("{sidecar}-stdout").as_str(), Some(line.clone()))
                        .expect("failed to emit event");

                    let _ = window.eval(&format!(
                        "console.log(\":: {}\")",
                        line.replace("\"", "\\\"")
                    ));
                }
                CommandEvent::Stderr(line) => {
                    window
                        .emit(format!("{sidecar}-stderr").as_str(), Some(line.clone()))
                        .expect("failed to emit event");

                    let _ = window.eval(&format!(
                        "console.log(\":: {}\")",
                        line.replace("\"", "\\\"")
                    ));
                }
                _ => continue,
            };
        }
    });
}
