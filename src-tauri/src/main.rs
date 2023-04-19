#![cfg_attr(
    all(not(debug_assertions), target_os = "windows"),
    windows_subsystem = "windows"
)]

fn main() {
    fix_path_env::fix();
    tauri::Builder::default()
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
