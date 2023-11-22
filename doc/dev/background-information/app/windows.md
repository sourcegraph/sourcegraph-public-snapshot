# Building App on Windows

The Windows build is still very early in development. Consider it highly experimental and unstable. 

## System requirements

Tauri requires Windows 10 (version 1803 or later) or Windows 11. (Only Windows 11 has been tested.)

## Dependencies

`sg setup` and `asdf` are not supported on Windows so you'll have to install each dependency manually.

- *Git*: install latest version from https://gitforwindows.org/
- *Go*: install version used in [.tools-versions](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/.tool-versions) from https://go.dev/dl/
- *Node*: install version used in [.tools-versions](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/.tool-versions) using NVM https://github.com/coreybutler/nvm-windows
- *pnpm*: install using PowerShell command `iwr https://get.pnpm.io/install.ps1 -useb | iex`
- *Python*: install from the Microsoft Store https://www.microsoft.com/store/productId/9NRWMJP3717K
- *Visual C++ build tools*: install with Visual Studio https://visualstudio.microsoft.com/thank-you-downloading-visual-studio/?sku=BuildTools, select "Desktop Development with C++"
- *GCC*: install MSYS2 from https://www.msys2.org/, then start MSYS2 UCRT64 shell, run `pacman -S mingw-w64-ucrt-x86_64-gcc`, and add `C:\msys64\ucrt64\bin` to PATH
- *Rust*: install 64-bit version from https://www.rust-lang.org/learn/get-started

## Build

From the Git Bash shell, run `./dev/app/build-windows.sh` to build. Bundling will fail but the main EXE will be created.

## Run

Run Sourcegraph.exe, located in `src-tauri\target\release`.

## Logs

The `Troubleshoot` option in the system tray does not work. You can view the logs by opening `%APPDATA%\com.sourcegraph.cody\logs\Cody.log`.

## Resources

- [How add directory to PATH on Windows](https://windowsloop.com/how-to-add-to-windows-path/)
