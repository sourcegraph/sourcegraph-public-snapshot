#![cfg_attr(
    all(not(debug_assertions), target_os = "windows"),
    windows_subsystem = "windows"
)]

#[cfg(not(dev))]
use {tauri::api::process::Command, tauri::api::process::CommandEvent};

use tauri::{
    api::shell, AppHandle, CustomMenuItem, Manager, SystemTray, SystemTrayEvent, SystemTrayMenu,
    SystemTrayMenuItem,
};

fn main() {
    match fix_path_env::fix() {
        Ok(_) => {}
        Err(e) => {
            println!("Error fixing path environment: {}", e);
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

                    let _ = window.eval(&format!(
                        "console.log(\":: {}\")",
                        line.replace("\"", "\\\"")
                    ));
                }
                CommandEvent::Stderr(line) => {
                    window
                        .emit("backend-stderr", Some(line.clone()))
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

fn create_tray_menu() -> SystemTrayMenu {
    SystemTrayMenu::new()
        .add_item(CustomMenuItem::new("open".to_string(), "Sourcegraph App"))
        .add_native_item(SystemTrayMenuItem::Separator)
        .add_item(
            CustomMenuItem::new("settings".to_string(), "Settings").accelerator("CmdOrCtrl+,"),
        )
        .add_item(CustomMenuItem::new(
            "about".to_string(),
            "About Sourcegraph",
        ))
        .add_native_item(SystemTrayMenuItem::Separator)
        .add_item(CustomMenuItem::new("toggle-status".to_string(), "Pause"))
        .add_item(CustomMenuItem::new("quit".to_string(), "Quit").accelerator("CmdOrCtrl+Q"))
}

fn open_window(app: &AppHandle) {
    let window = app.get_window("main").unwrap();
    match window.is_visible() {
        Ok(true) => {
            // noop
        }
        Ok(false) => {
            window.show();
        }
        Err(e) => {
            println!("Error getting window visibility: {}", e);
        }
    }
}

fn on_system_tray_event(app: &AppHandle, event: SystemTrayEvent) {
    if let SystemTrayEvent::MenuItemClick { id, .. } = event {
        let _item_handle = app.tray_handle().get_item(&id);
        match id.as_str() {
            "open" => open_window(app),
            "settings" => {
                let window = app.get_window("main").unwrap();
                window
                    .eval("window.location.href = '/setup-wizard'")
                    .unwrap();
                open_window(app);
            }
            "about" => {
                shell::open(&app.shell_scope(), "https://about.sourcegraph.com", None).unwrap()
            }
            "toggle-status" => {
                // Allow resuming/pausing depending on if the app is running
            }
            "quit" => app.exit(0),
            _ => {}
        }
    }
}
