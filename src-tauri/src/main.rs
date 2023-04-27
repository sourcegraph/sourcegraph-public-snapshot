#![cfg_attr(
    all(not(debug_assertions), target_os = "windows"),
    windows_subsystem = "windows"
)]

use std::path::PathBuf;

#[cfg(not(dev))]
use {tauri::api::process::Command, tauri::api::process::CommandEvent};

use {
    tauri::api::shell, tauri::AppHandle, tauri::CustomMenuItem, tauri::Manager, tauri::SystemTray,
    tauri::SystemTrayEvent, tauri::SystemTrayMenu, tauri::SystemTrayMenuItem,
    tauri_plugin_log::LogTarget,
};

fn main() {
    match fix_path_env::fix() {
        Ok(_) => {}
        Err(e) => {
            println!("error fixing path environment: {}", e);
        }
    }

    let tray_menu = create_tray_menu();
    let tray = SystemTray::new().with_menu(tray_menu);

    tauri::Builder::default()
        .system_tray(tray)
        .on_system_tray_event(on_system_tray_event)
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
                .targets([LogTarget::LogDir, LogTarget::Webview])
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
        // read events such as stdout
        while let Some(event) = rx.recv().await {
            match event {
                CommandEvent::Stdout(line) => log::info!("{}", line),
                CommandEvent::Stderr(line) => log::error!("{}", line),
                _ => continue,
            };
        }
    });
}

fn create_tray_menu() -> SystemTrayMenu {
    SystemTrayMenu::new()
        .add_item(CustomMenuItem::new("open".to_string(), "Sourcegraph App"))
        .add_native_item(SystemTrayMenuItem::Separator)
        .add_item(
            CustomMenuItem::new("settings".to_string(), "Settings").accelerator("CmdOrCtrl+,"),
        )
        .add_item(CustomMenuItem::new(
            "troubleshoot".to_string(),
            "Troubleshoot",
        ))
        .add_native_item(SystemTrayMenuItem::Separator)
        .add_item(CustomMenuItem::new(
            "about".to_string(),
            "About Sourcegraph",
        ))
        .add_native_item(SystemTrayMenuItem::Separator)
        .add_item(CustomMenuItem::new("restart".to_string(), "Restart"))
        .add_item(CustomMenuItem::new("quit".to_string(), "Quit").accelerator("CmdOrCtrl+Q"))
}

fn show_window(app: &AppHandle) {
    let window = app.get_window("main").unwrap();
    if !window.is_visible().unwrap() {
        window.show().expect("failed to open window");
    }
}

fn on_system_tray_event(app: &AppHandle, event: SystemTrayEvent) {
    if let SystemTrayEvent::MenuItemClick { id, .. } = event {
        match id.as_str() {
            "open" => show_window(app),
            "settings" => {
                let window = app.get_window("main").unwrap();
                window.eval("window.location.href = '/settings'").unwrap();
                show_window(app);
            }
            "troubleshoot" => {
                let log_path: PathBuf = tauri::api::path::app_log_dir(&app.config()).unwrap();

                if let Some(log_path_str) = log_path.to_str() {
                    let combined_path = format!("{}/Sourcegraph App.log", log_path_str);
                    shell::open(&app.shell_scope(), &combined_path, None).unwrap()
                }
            }
            "about" => {
                shell::open(&app.shell_scope(), "https://about.sourcegraph.com", None).unwrap()
            }
            "restart" => app.restart(),
            "quit" => app.exit(0),
            _ => {}
        }
    }
}
