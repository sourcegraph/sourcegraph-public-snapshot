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
