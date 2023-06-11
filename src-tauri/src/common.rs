use std::fs;
use tauri::api::dialog::confirm;
use tauri::api::shell;
use tauri::AppHandle;
use tauri::Manager;

pub fn show_window(app: &AppHandle, label: &str) {
    let window = app.get_window(label).unwrap();
    if !window.is_visible().unwrap() {
        window.show().unwrap()
    }
    window.set_focus().unwrap()
}

/// Extracts the path from a URL that starts with the scheme followed by `://`.
///
/// # Examples
///
/// ```
/// let url = "sourcegraph://settings";
/// assert_eq!(extract_path_from_url(url), "/settings");
///
/// let url = "sourcegraph://user/admin";
/// assert_eq!(extract_path_from_url(url), "/user/admin");
/// ```
pub fn extract_path_from_scheme_url<'a>(url: &'a str, scheme: &str) -> &'a str {
    &url[(scheme.len() + 2)..]
}

/// Checks if a URL starts with the scheme (sourcegraph://)
#[cfg(not(target_os = "macos"))]
pub fn is_scheme_url(url: &str, scheme: &str) -> bool {
    url.starts_with(scheme) && url[scheme.len()..].starts_with("://")
}

pub fn show_logs(app: &AppHandle) {
    let log_dir_path = app.path_resolver().app_log_dir().unwrap();
    if let Some(log_path_str) = log_dir_path.to_str() {
        let name = &app.package_info().name;
        let combined_path = format!("{}/{}.log", log_path_str, name);
        shell::open(&app.shell_scope(), &combined_path, None).unwrap()
    }
}

pub fn prompt_to_clear_all_data(app: &AppHandle) {
    let window = app.get_window("main").unwrap();
    let path_resolver = app.path_resolver();
    let app_clone = app.clone(); // Clone the app for use in the closure

    confirm(
        Some(&window),
        "Sourcegraph",
        "This will remove all data.\nAre you sure?",
        move |answer| {
            if answer {
                clear_all_data_and_restart(&app_clone)
            }
        },
    );
}

fn clear_all_data_and_restart(app: &AppHandle) {
    let path_resolver = app.path_resolver();

    // Delete app data dir
    if let Some(app_data_dir_path) = path_resolver.app_data_dir() {
        if let Err(err) = fs::remove_dir_all(&app_data_dir_path) {
            log::error!("{}", err);
        }
    }

    // Delete app config dir
    if let Some(app_config_dir_path) = path_resolver.app_config_dir() {
        if let Err(err) = fs::remove_dir_all(&app_config_dir_path) {
            log::error!("{}", err);
        }
    }

    clear_local_storage(app);

    app.restart();
}

fn clear_local_storage(app: &AppHandle) {
    let window = app.get_window("main").unwrap();
    // Note that this will clear localStorage only for the current origin, which
    // is fine assuming the webview is still on localhost:3080.
    window.eval("localStorage.clear();").unwrap();
}
