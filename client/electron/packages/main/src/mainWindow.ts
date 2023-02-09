import {app, BrowserWindow, Menu, nativeTheme, Tray} from 'electron';
import * as path from 'node:path';
import {join} from 'node:path';
import {URL} from 'node:url';

async function createWindow() {
  const browserWindow = new BrowserWindow({
    type: 'panel',
    transparent: true,
    useContentSize: true, // The width and height would be used as web page's size.
    frame: false,
    show: false, // Use the 'ready-to-show' event to show the instantiated BrowserWindow.
    webPreferences: {
      nodeIntegration: false,
      contextIsolation: true,
      sandbox: false, // Sandbox disabled because the demo of preload script depend on the Node.js api
      webviewTag: false, // The webview tag is not recommended. Consider alternatives like an iframe or Electron's BrowserView. @see https://www.electronjs.org/docs/latest/api/webview-tag#warning
      preload: join(app.getAppPath(), 'packages/preload/dist/index.cjs'),
    },
  });

  /**
   * Add the tray icon and menu.
   */
  const tray = new Tray(path.join(getIconPath(), getIcon()));
  tray.setPressedImage(path.join(getIconPath(), getIcon()));
  tray.setToolTip('Sourcegraph App');
  tray.setContextMenu(buildMenu());

  tray.on('click', function () {
    restoreOrCreateWindow();
  });

  /**
   * If the 'show' property of the BrowserWindow's constructor is omitted from the initialization options,
   * it then defaults to 'true'. This can cause flickering as the window loads the html content,
   * and it also has show problematic behaviour with the closing of the window.
   * Use `show: false` and listen to the  `ready-to-show` event to show the window.
   *
   * @see https://github.com/electron/electron/issues/25012 for the afford mentioned issue.
   */
  browserWindow.on('ready-to-show', () => {
    browserWindow?.show();

    if (import.meta.env.DEV) {
      browserWindow?.webContents.openDevTools({mode: 'detach'});
    }
  });

  /**
   * URL for main window.
   * Vite dev server for development.
   * `file://../renderer/index.html` for production and test.
   */
  const pageUrl =
    import.meta.env.DEV && import.meta.env.VITE_DEV_SERVER_URL !== undefined
      ? import.meta.env.VITE_DEV_SERVER_URL
      : new URL('../renderer/dist/index.html', 'file://' + __dirname).toString();

  await browserWindow.loadURL(pageUrl);

  return browserWindow;
}

/**
 * Restore an existing BrowserWindow or Create a new BrowserWindow.
 */
export async function restoreOrCreateWindow() {
  let window = BrowserWindow.getAllWindows().find(w => !w.isDestroyed());

  if (window === undefined) {
    window = await createWindow();
  } else {
    if (window.isVisible()) {
      window.hide();
    } else {
      window.restore();
      window.show();
      window.focus();
    }
  }
}

/**
 * Tray icon helpers.
 */

const getIconPath = () => {
  return path.join(__dirname, '..', '..', '..', 'buildResources');
};

const getIcon = () => {
  if (process.platform === 'win32') return 'icon-light@2x.ico';
  if (nativeTheme.shouldUseDarkColors) return 'icon-light@2x.png';
  return 'icon-dark@2x.png';
};

const buildMenu = () => {
  const menu = Menu.buildFromTemplate([
    {
      label: 'Settings',
      click() {
        openSettings();
      },
    },
    {type: 'separator'},
    {
      label: 'Quit',
      click() {
        app.quit();
      },
      accelerator: 'CommandOrControl+Q',
    },
  ]);

  return menu;
};

const openSettings = () => {
  /* Open settings page on the browser. */
};
