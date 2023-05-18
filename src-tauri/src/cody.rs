use tauri::{WindowBuilder, WindowUrl};
use tauri_plugin_positioner::{Position, WindowExt};

pub fn init_cody_window(handle: &tauri::AppHandle) {
    let app = handle.clone();
    tauri::async_runtime::spawn(async move {
        let cody_win = WindowBuilder::new(&app, "cody", WindowUrl::App("/cody-standalone".into()))
            .title("Ask Cody (Beta)")
            .fullscreen(false)
            .inner_size(480.0, 720.0)
            .always_on_top(false);

        cody_win
            .build()
            .unwrap()
            .move_window(Position::TopRight)
            .unwrap();
    });
}
