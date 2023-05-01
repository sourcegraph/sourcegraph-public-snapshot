#![cfg_attr(
    all(not(debug_assertions), target_os = "windows"),
    windows_subsystem = "windows"
)]

#[cfg(not(dev))]
use {tauri::api::process::Command, tauri::api::process::CommandEvent};

mod tray;
use std::sync::RwLock;
use tauri::Manager;

// The URL to open the frontend on, if launched with a scheme url.
static LAUNCH_PATH: RwLock<String> = RwLock::new(String::new());

#[tauri::command]
fn get_launch_path() -> String {
    LAUNCH_PATH.read().unwrap().clone()
}

fn set_launch_path(url: String) {
    *LAUNCH_PATH.write().unwrap() = url;
}

// Url scheme for sourcegraph:// urls.
const SCHEME: &str = "sourcegraph";
const BUNDLE_IDENTIFIER: &str = "com.sourcegraph.app";

/// Extracts the path from a URL that starts with the `SCHEME` followed by `://`.
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
fn extract_path_from_scheme_url(url: &str) -> &str {
    &url[(SCHEME.len() + 2)..]
}

fn main() {
    // Prepare handler for sourcegraph:// scheme urls.
    tauri_plugin_deep_link::prepare(BUNDLE_IDENTIFIER);

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
        .setup(|app| {
            start_embedded_services();

            // Register handler for sourcegraph:// scheme urls.
            let handle = app.handle();
            tauri_plugin_deep_link::register(SCHEME, move |request| {
                let path: &str = extract_path_from_scheme_url(&request);

                // Case 1: the app has been *launched* with the scheme
                // url. In the frontend, app-shell.tsx will read it with
                // getLaunchPath().
                set_launch_path(path.to_string());

                // Case 2: the app was *already running* when the scheme url was
                // opened. This currently doesn't collide with Case 1 because it
                // doesn't do anything while we're still launching, probably
                // because the webview isn't ready yet.
                // TODO(marek) add a guard to check whether we're still launching.
                handle
                    .get_window("main")
                    .unwrap()
                    .eval(&format!("window.location.href = '{}'", path))
                    .unwrap();
            })
            .unwrap();

            Ok(())
        })
        // Define a handler so that invoke("get_launch_scheme_url") can be
        // called on the frontend. (The Tauri invoke_handler function, despite
        // its name which may suggest that it invokes something, actually only
        // *defines* an invoke() handler and does not invoke anything during
        // setup here.)
        .invoke_handler(tauri::generate_handler![get_launch_path])
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
