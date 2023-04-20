#![cfg_attr(
    all(not(debug_assertions), target_os = "windows"),
    windows_subsystem = "windows"
)]

fn main() {
    match fix_path_env::fix() {
        Ok(_) => {}
        Err(e) => {
            println!("Error fixing path environment: {}", e);
        }
    }
    tauri::Builder::default()
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
