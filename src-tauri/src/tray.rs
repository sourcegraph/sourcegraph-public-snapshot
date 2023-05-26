use crate::cody::init_cody_window;
use crate::common::show_window;
use crate::BackendProcessId;

use std::process::Command as StdCommand;
use tauri::api::shell;
use tauri::{
    AppHandle, CustomMenuItem, Manager, SystemTray, SystemTrayEvent, SystemTrayMenu,
    SystemTrayMenuItem,
};

pub fn create_system_tray() -> SystemTray {
    SystemTray::new().with_menu(create_system_tray_menu())
}

fn create_system_tray_menu() -> SystemTrayMenu {
    SystemTrayMenu::new()
        .add_item(CustomMenuItem::new("open".to_string(), "Open Sourcegraph"))
        .add_item(CustomMenuItem::new("cody".to_string(), "Show Cody"))
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

pub fn on_system_tray_event(app: &AppHandle, event: SystemTrayEvent) {
    if let SystemTrayEvent::MenuItemClick { id, .. } = event {
        match id.as_str() {
            "open" => show_window(app, "main"),
            "cody" => {
                let win = app.get_window("cody");
                if win.is_none() {
                    init_cody_window(app);
                } else {
                    show_window(app, "cody")
                }
            }
            "settings" => {
                let window = app.get_window("main").unwrap();
                window.eval("window.location.href = '/settings'").unwrap();
                show_window(app, "main");
            }
            "troubleshoot" => {
                let log_dir_path = app.path_resolver().app_log_dir().unwrap();
                if let Some(log_path_str) = log_dir_path.to_str() {
                    let name = &app.package_info().name;
                    let combined_path = format!("{}/{}.log", log_path_str, name);
                    shell::open(&app.shell_scope(), &combined_path, None).unwrap()
                }
            }
            "about" => {
                shell::open(&app.shell_scope(), "https://about.sourcegraph.com", None).unwrap()
            }
            "restart" => {
                let backend_pid = &app.state::<BackendProcessId>().0;
                stop_process(app.clone(), *backend_pid);
                app.restart()
            }
            "quit" => {
                let backend_pid = &app.state::<BackendProcessId>().0;
                stop_process(app.clone(), *backend_pid);
                app.exit(0)
            }
            _ => {}
        }
    }
}

fn stop_process(app: AppHandle, process_id: u32) {
    // hide the windows to make it look like it's closing faster
    for (_name, window) in app.windows() {
        let _r = window.hide();
    }
    log::info!("stopping process pid: {}", process_id);
    StdCommand::new("kill")
        .args(["-s", "TERM", &process_id.to_string()])
        .spawn()
        .expect("failed to shutdown process")
        .wait()
        .expect("failed to wait for shutdown");

    // Sleep some to let the background process close properly before it gets killed by app closing
    // The above wait only waits for the kill command to finish, not the process itself
    // The alternative to sleeping is to continuously check if the process is still running
    // this requires a crate that makes the apple not accept the app to the app store
    // my testing showed the backend process was always closed in under 200 ms
    let half_second = std::time::Duration::from_millis(500);
    std::thread::sleep(half_second);
}
