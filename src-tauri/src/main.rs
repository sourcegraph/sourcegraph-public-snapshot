#![cfg_attr(
    all(not(debug_assertions), target_os = "windows"),
    windows_subsystem = "windows"
)]

#[cfg(not(dev))]
use {tauri::api::process::Command, tauri::api::process::CommandEvent};

mod common;
mod tray;
use common::{extract_path_from_scheme_url, show_window};
use std::collections::HashMap;
use std::sync::RwLock;
use tauri::Manager;
use tauri_utils::config::RemoteDomainAccessScope;

#[cfg(not(target_os = "macos"))]
use common::is_scheme_url;

// The URL to open the frontend on, if launched with a scheme url.
static LAUNCH_PATH: RwLock<String> = RwLock::new(String::new());

#[tauri::command]
fn get_launch_path(window: tauri::Window) -> String {
    if window.label() == "cody" {
        return "/cody-standalone".to_string();
    }
    LAUNCH_PATH.read().unwrap().clone()
}

#[tauri::command]
fn app_shell_loaded() -> Option<AppShellReadyPayload> {
    return APP_SHELL_READY_PAYLOAD.read().unwrap().clone();
}

#[tauri::command]
fn show_main_window(app_handle: tauri::AppHandle) {
    show_window(&app_handle, "main");
}

#[tauri::command]
fn reload_cody_window(app_handle: tauri::AppHandle) {
    let win = app_handle.get_window("cody");
    if win.is_some() {
        win.unwrap().eval("window.location.reload();").unwrap();
    }
}

#[tauri::command]
fn show_logs(app_handle: tauri::AppHandle) {
    common::show_logs(&app_handle);
}

#[tauri::command]
fn restart_app(app_handle: tauri::AppHandle) {
    app_handle.restart();
}

#[tauri::command]
fn clear_all_data(app_handle: tauri::AppHandle) {
    common::prompt_to_clear_all_data(&app_handle);
}

fn set_launch_path(url: String) {
    *LAUNCH_PATH.write().unwrap() = url;
}

// Url scheme for sourcegraph:// urls.
const SCHEME: &str = "sourcegraph";
const BUNDLE_IDENTIFIER: &str = "com.sourcegraph.cody";

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

    let scope = RemoteDomainAccessScope {
        scheme: Some("http".to_string()),
        domain: "localhost".to_string(),
        windows: vec!["main".to_string(), "cody".to_string()],
        plugins: vec![],
        enable_tauri_api: true,
    };
    let mut context = tauri::generate_context!();
    context
        .config_mut()
        .tauri
        .security
        .dangerous_remote_domain_ipc_access = vec![scope];

    tauri::Builder::default()
        .system_tray(tray)
        .on_system_tray_event(tray::on_system_tray_event)
        .on_system_tray_event(|app, event| {
            tauri_plugin_positioner::on_tray_event(app, &event);
        })
        .on_window_event(|event| match event.event() {
            tauri::WindowEvent::CloseRequested { api, .. } => {
                // Ensure the app stays open after the last window is closed.
                if event.window().label() == "main" {
                    // We use `tauri::AppHandle::hide` instead of `event.window().hide` because
                    // hiding the app allows clicking the dock icon to show the app again.
                    // This is a temporary solution that only works if the app has a single window.
                    // If we need to add more windows in the future, we need to wait until
                    // https://github.com/tauri-apps/tauri/issues/3084 is fixed.
                    #[allow(unused_unsafe)]
                    #[cfg(not(target_os = "macos"))]
                    {
                        event.window().hide().unwrap();
                    }

                    #[allow(unused_unsafe)]
                    #[cfg(target_os = "macos")]
                    unsafe {
                        tauri::AppHandle::hide(&event.window().app_handle()).unwrap();
                    }
                    api.prevent_close();
                }
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
        .plugin(tauri_plugin_positioner::init())
        .setup(|app| {
            let handle = app.handle();
            start_embedded_services(&handle);
            // Register handler for sourcegraph:// scheme urls.
            tauri_plugin_deep_link::register(SCHEME, move |request| {
                if let Some(path) = extract_path_from_scheme_url(&request, SCHEME) {
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
                    show_window(&handle, "main");
                }
            })
            .unwrap();

            // If launched with a scheme url, on non-mac the app receives the url as an argument.
            // On mac, this is handled by the same handler that receives the url when the app is
            // already running.
            #[cfg(not(target_os = "macos"))]
            if let Some(url) = std::env::args().nth(1) {
                if is_scheme_url(&url, SCHEME) {
                    let path = extract_path_from_scheme_url(&url, SCHEME);
                    set_launch_path(url)
                }
            }
            Ok(())
        })
        // Define a handler so that invoke("get_launch_scheme_url") can be
        // called on the frontend. (The Tauri invoke_handler function, despite
        // its name which may suggest that it invokes something, actually only
        // *defines* an invoke() handler and does not invoke anything during
        // setup here.)
        .invoke_handler(tauri::generate_handler![
            get_launch_path,
            app_shell_loaded,
            show_main_window,
            reload_cody_window,
            show_logs,
            restart_app,
            clear_all_data
        ])
        .run(context)
        .expect("error while running Cody app");
}

#[cfg(dev)]
fn start_embedded_services(app_handle: &tauri::AppHandle) {
    let args = get_sourcegraph_args(app_handle);
    println!("embedded Sourcegraph services disabled for local development");
    println!("Sourcegraph would start with args: {:?}", args);
}

#[derive(Clone, serde::Serialize)]
struct AppShellReadyPayload {
    sign_in_url: String,
}

// The URL to open the frontend on, if launched with a scheme url.
static APP_SHELL_READY_PAYLOAD: RwLock<Option<AppShellReadyPayload>> = RwLock::new(None);

#[cfg(not(dev))]
fn start_embedded_services(handle: &tauri::AppHandle) {
    let app = handle.clone();
    let sidecar = "sourcegraph-backend";
    let args = get_sourcegraph_args(&app);
    println!("Sourcegraph starting with args: {:?}", args);
    let (mut rx, _child) = Command::new_sidecar(sidecar)
        .expect(format!("failed to create `{sidecar}` binary command").as_str())
        .args(args)
        .envs(HashMap::from([
            (
                "SRC_REPOS_DESIRED_PERCENT_FREE".to_string(),
                "0".to_string(),
            ),
            ("SRC_PROF_HTTP".to_string(), "".to_string()),
        ]))
        .spawn()
        .expect(format!("failed to spawn {sidecar} sidecar").as_str());

    tauri::async_runtime::spawn(async move {
        while let Some(event) = rx.recv().await {
            match event {
                CommandEvent::Stdout(line) => log::info!("{}", line),
                CommandEvent::Stderr(line) => {
                    if line.contains("tauri:sign-in-url: ") {
                        let url = line.splitn(2, ' ').last().unwrap();
                        *APP_SHELL_READY_PAYLOAD.write().unwrap() = Some(AppShellReadyPayload {
                            sign_in_url: url.to_string(),
                        });
                        app.get_window("main")
                            .unwrap()
                            .emit(
                                "app-shell-ready",
                                APP_SHELL_READY_PAYLOAD.read().unwrap().clone(),
                            )
                            .unwrap();
                    }
                    log::error!("{}", line);
                }
                CommandEvent::Error(err) => {
                    show_error_screen(&app);
                    log::error!("Error running the Cody app backend: {:#?}", err)
                }
                CommandEvent::Terminated(payload) => {
                    show_error_screen(&app);

                    if let Some(code) = payload.code {
                        log::error!("Cody app backend terminated with exit code {}", code);
                    }

                    if let Some(signal) = payload.signal {
                        log::error!("Cody app backend terminated due to signal {}", signal);
                    }
                }
                _ => continue,
            };
        }
    });
}

/// Show the error page in the main window
fn show_error_screen(app_handle: &tauri::AppHandle) {
    app_handle
        .get_window("main")
        .unwrap()
        .eval("window.location.href = 'tauri://localhost/error.html';")
        .unwrap();
    show_window(app_handle, "main");
}

fn get_sourcegraph_args(app_handle: &tauri::AppHandle) -> Vec<String> {
    let data_dir = app_handle.path_resolver().app_data_dir();
    let cache_dir = app_handle.path_resolver().app_cache_dir();
    let mut args = Vec::new();

    // cache_dir is where the cache goes
    if let Some(cache_dir) = cache_dir {
        args.push("--cacheDir".to_string());
        args.push(cache_dir.to_string_lossy().to_string())
    }

    // configDir is where the database goes
    if let Some(data_dir) = data_dir {
        args.push("--configDir".to_string());
        args.push(data_dir.to_string_lossy().to_string())
    }
    return args;
}
