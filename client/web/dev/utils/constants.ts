import path from 'path'

export const ROOT_PATH = path.resolve(__dirname, '../../../../')
export const STATIC_ASSETS_PATH = path.resolve(ROOT_PATH, 'ui/assets')
export const STATIC_INDEX_PATH = path.resolve(STATIC_ASSETS_PATH, 'index.html')
export const STATIC_ASSETS_URL = '/.assets/'
export const DEV_SERVER_LISTEN_ADDR = { host: 'localhost', port: 3080 }
export const DEV_SERVER_PROXY_TARGET_ADDR = { host: 'localhost', port: 3081 }

// TODO: share with gulpfile.js
export const WEBPACK_STATS_OPTIONS = {
    all: false,
    timings: true,
    errors: true,
    warnings: true,
    colors: true,
}
