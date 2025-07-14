use tauri::{Manager, Result, Window};

#[tauri::command]
async fn start_dragging(window: Window) -> Result<()> {
    window.start_dragging()?;
    Ok(())
}

#[tauri::command]
async fn window_minimize(window: Window) -> Result<()> {
    window.minimize()?;
    Ok(())
}

#[tauri::command]
async fn window_toggle_maximize(window: Window) -> Result<()> {
    if window.is_maximized()? {
        window.unmaximize()?;
    } else {
        window.maximize()?;
    }
    Ok(())
}

#[tauri::command]
async fn window_close(window: Window) -> Result<()> {
    window.close()?;
    Ok(())
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    let mut builder = tauri::Builder::default();

    // Enable the Tauri devtools plugin in development builds
    #[cfg(debug_assertions)]
    {
        let devtools = tauri_plugin_devtools::init();
        builder = builder.plugin(devtools);
    }

    builder
        .plugin(tauri_plugin_os::init())
        // .plugin( /* Add your Tauri plugin here */ )
        // Add your commands here that you will call from your JS code
        .invoke_handler(tauri::generate_handler![
            start_dragging,
            window_minimize,
            window_toggle_maximize,
            window_close
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
