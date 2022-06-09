import path from 'path'

export const ROOT_PATH = path.resolve(__dirname, '../../../../')
export const STATIC_ASSETS_PATH = path.resolve(ROOT_PATH, 'ui/assets')
export const STATIC_INDEX_PATH = path.resolve(STATIC_ASSETS_PATH, 'index.html')
export const STATIC_ASSETS_URL = '/.assets/'
export const DEV_SERVER_LISTEN_ADDR = { host: 'localhost', port: 3080 } as const
export const DEV_SERVER_PROXY_TARGET_ADDR = { host: 'localhost', port: 3081 } as const
export const DEFAULT_SITE_CONFIG_PATH = path.resolve(ROOT_PATH, '../dev-private/enterprise/dev/site-config.json')

export const WEBPACK_STATS_OPTIONS = {
    all: false,
    timings: true,
    errors: true,
    warnings: true,
    colors: true,
} as const
