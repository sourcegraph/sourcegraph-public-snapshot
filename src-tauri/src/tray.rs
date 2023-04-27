use std::path::PathBuf;
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

pub fn on_system_tray_event(app: &AppHandle, event: SystemTrayEvent) {
    if let SystemTrayEvent::MenuItemClick { id, .. } = event {
        match id.as_str() {
            "open" => show_window(app),
            "settings" => {
                let window = app.get_window("main").unwrap();
                window.eval("window.location.href = '/settings'").unwrap();
                show_window(app);
            }
            "troubleshoot" => {
                let log_dir_path: PathBuf = tauri::api::path::app_log_dir(&app.config()).unwrap();
                if let Some(log_path_str) = log_dir_path.to_str() {
                    let name = app.config().package.product_name.as_ref().unwrap().clone();
                    let combined_path = format!("{}/{}.log", log_path_str, name);
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
