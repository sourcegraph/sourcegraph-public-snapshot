use crate::common::show_window;
use std::fs;
use std::path::PathBuf;
use tauri::api::dialog::blocking::message;
use tauri::api::dialog::confirm;
use tauri::api::shell;
use tauri::{
    AppHandle, CustomMenuItem, Manager, SystemTray, SystemTrayEvent, SystemTrayMenu,
    SystemTrayMenuItem, SystemTraySubmenu,
};

pub fn create_system_tray() -> SystemTray {
    SystemTray::new().with_menu(create_system_tray_menu())
}

fn create_system_tray_menu() -> SystemTrayMenu {
    let logs_item = CustomMenuItem::new("viewlogs".to_string(), "View Logs");
    let clear_data_item = CustomMenuItem::new("cleardata".to_string(), "Clear All Data");
    let troubleshooting_menu = SystemTraySubmenu::new(
        "Troubleshooting",
        SystemTrayMenu::new()
            .add_item(logs_item)
            .add_item(clear_data_item),
    );

    SystemTrayMenu::new()
        .add_item(CustomMenuItem::new(
            "open".to_string(),
            "Open Sourcegraph App",
        ))
        .add_native_item(SystemTrayMenuItem::Separator)
        .add_item(
            CustomMenuItem::new("settings".to_string(), "Settings").accelerator("CmdOrCtrl+,"),
        )
        .add_submenu(troubleshooting_menu)
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
            "open" => show_window(app),
            "settings" => {
                let window = app.get_window("main").unwrap();
                window.eval("window.location.href = '/settings'").unwrap();
                show_window(app);
            }
            "viewlogs" => {
                let log_dir_path = app.path_resolver().app_log_dir().unwrap();
                if let Some(log_path_str) = log_dir_path.to_str() {
                    let name = &app.package_info().name;
                    let combined_path = format!("{}/{}.log", log_path_str, name);
                    shell::open(&app.shell_scope(), &combined_path, None).unwrap()
                }
            }
            "cleardata" => restart_app(app),
            "about" => {
                shell::open(&app.shell_scope(), "https://about.sourcegraph.com", None).unwrap()
            }
            "restart" => app.restart(),
            "quit" => app.exit(0),
            _ => {}
        }
    }
}

fn restart_app(app: &AppHandle) {
    let window = app.get_window("main").unwrap();
    let app_clone = app.clone();

    confirm(
        Some(&window),
        "Sourcegraph",
        "This will remove all data.\nAre you sure?",
        move |answer| {
            if answer {
                let data_dir =
                    join_path(app_clone.path_resolver().app_data_dir(), "postgresql/data");
                match data_dir {
                    Some(db_dir) => {
                        log::warn!("attempting to remove: {}", db_dir.to_string_lossy());
                        match fs::remove_dir_all(db_dir) {
                            Ok(_) => app_clone.restart(),
                            Err(err) => {
                                log::error!("{}", err);
                            }
                        }
                    }
                    None => {}
                }
            }
        },
    );
}

fn join_path(base_folder: Option<PathBuf>, subfolder: &str) -> Option<PathBuf> {
    if let Some(path) = base_folder {
        Some(path.join(subfolder))
    } else {
        None
    }
}
