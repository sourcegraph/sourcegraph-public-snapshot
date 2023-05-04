use tauri::{WindowBuilder, WindowUrl};

pub fn init_cody_window(handle: &tauri::AppHandle) {
    let app = handle.clone();
    tauri::async_runtime::spawn(async move {
        let cody_win = WindowBuilder::new(&app, "cody", WindowUrl::App("/cody-standalone".into()))
            .title("Cody")
            .resizable(false)
            .fullscreen(false)
            .decorations(false)
            .inner_size(480.0, 720.0)
            .always_on_top(false);

        cody_win.build().unwrap().hide().unwrap();
    });
}
