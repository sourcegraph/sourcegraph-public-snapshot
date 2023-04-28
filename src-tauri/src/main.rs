#![cfg_attr(
    all(not(debug_assertions), target_os = "windows"),
    windows_subsystem = "windows"
)]

#[cfg(not(dev))]
use {tauri::api::process::Command, tauri::api::process::CommandEvent};

mod tray;

fn main() {
    match fix_path_env::fix() {
        Ok(_) => {}
        Err(e) => {
            println!("error fixing path environment: {}", e);
        }
    }

    let tray = tray::create_system_tray();

    tauri::Builder::default()
        .system_tray(tray)
        .on_system_tray_event(tray::on_system_tray_event)
        .on_window_event(|event| match event.event() {
            tauri::WindowEvent::CloseRequested { api, .. } => {
                // Ensure the app stays open after the last window is closed.
                event.window().hide().unwrap();
                api.prevent_close();
            }
            _ => {}
        })
        .plugin(
            tauri_plugin_log::Builder::default()
                .targets([
                    tauri_plugin_log::LogTarget::LogDir,
                    tauri_plugin_log::LogTarget::Webview,
                ])
                .level(log::LevelFilter::Info)
                .build(),
        )
        .setup(|_app| {
            start_embedded_services();
            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}

#[cfg(dev)]
fn start_embedded_services() {
    println!("embedded Sourcegraph services disabled for local development");
}

#[cfg(not(dev))]
fn start_embedded_services() {
    let sidecar = "sourcegraph-backend";
    let (mut rx, _child) = Command::new_sidecar(sidecar)
        .expect(format!("failed to create `{sidecar}` binary command").as_str())
        .spawn()
        .expect(format!("failed to spawn {sidecar} sidecar").as_str());

    tauri::async_runtime::spawn(async move {
        while let Some(event) = rx.recv().await {
            match event {
                CommandEvent::Stdout(line) => log::info!("{}", line),
                CommandEvent::Stderr(line) => log::error!("{}", line),
                _ => continue,
            };
        }
    });
}
