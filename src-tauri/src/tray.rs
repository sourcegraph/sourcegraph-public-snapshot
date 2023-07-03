use crate::common::{prompt_to_clear_all_data, show_logs, show_window};
use tauri::api::shell;
use tauri::{
    AppHandle, CustomMenuItem, Manager, SystemTray, SystemTrayEvent, SystemTrayMenu,
    SystemTrayMenuItem, SystemTraySubmenu,
};

pub fn create_system_tray() -> SystemTray {
    SystemTray::new().with_menu(create_system_tray_menu())
}

fn create_system_tray_menu() -> SystemTrayMenu {
    let view_logs_item = CustomMenuItem::new("view_logs".to_string(), "View Logs");
    let clear_all_data_item = CustomMenuItem::new("clear_all_data".to_string(), "Clear All Data");
    let troubleshooting_menu = SystemTraySubmenu::new(
        "Troubleshooting",
        SystemTrayMenu::new()
            .add_item(view_logs_item)
            .add_item(clear_all_data_item),
    );

    SystemTrayMenu::new()
        .add_item(CustomMenuItem::new("open".to_string(), "Open Cody"))
        .add_native_item(SystemTrayMenuItem::Separator)
        .add_item(
            CustomMenuItem::new("settings".to_string(), "Settings").accelerator("CmdOrCtrl+,"),
        )
        .add_item(CustomMenuItem::new(
            "update".to_string(),
            "Check for updates",
        ))
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
            "open" => show_window(app, "main"),
            "settings" => {
                let window = app.get_window("main").unwrap();
                window
                    .eval("window.location.href = '/user/app-settings'")
                    .unwrap();
                show_window(app, "main");
            }
            "view_logs" => show_logs(app),
            "clear_all_data" => prompt_to_clear_all_data(app),

            "about" => {
                shell::open(&app.shell_scope(), "https://about.sourcegraph.com", None)
                    .unwrap_or_else(|e| eprintln!("Failed to open URL: {:?}", e));
            }
            "restart" => app.restart(),
            "update" => app.trigger_global("tauri://update", None),
            "quit" => app.exit(0),
            _ => {}
        }
    }
}
