use tauri::AppHandle;
use tauri::Manager;

pub fn show_window(app: &AppHandle) {
    let window = app.get_window("main").unwrap();
    window.show().unwrap();
    window.set_focus().unwrap();
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
pub fn is_scheme_url(url: &str, scheme: &str) -> bool {
    url.starts_with(scheme) && url[scheme.len()..].starts_with("://")
}
