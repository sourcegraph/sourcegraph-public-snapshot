# Building App on Windows

Git -> install from https://gitforwindows.org/
Go -> install version used in .toolsversion from https://go.dev/dl/
Node -> install version used in .toolsversion using NVM https://github.com/coreybutler/nvm-windows
pnpm -> install using PowerShell command `iwr https://get.pnpm.io/install.ps1 -useb | iex`
Python -> install from Microsoft Store https://www.microsoft.com/store/productId/9NRWMJP3717K
Visual C++ build tools -> install with Visual Studio https://visualstudio.microsoft.com/thank-you-downloading-visual-studio/?sku=BuildTools, select "Desktop Development with C++" (more info: https://github.com/nodejs/node-gyp#on-windows)
MSYS2 -> install from https://www.msys2.org/
GCC -> run `pacman -S mingw-w64-ucrt-x86_64-gcc` in MSYS2, then add `C:\msys64\ucrt64\bin` to PATH
Rust -> install 64-bit version from https://www.rust-lang.org/learn/get-started
Lua -> install from https://github.com/rjpcomputing/luaforwindows/releases
