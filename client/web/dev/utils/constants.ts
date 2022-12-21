import path from 'path'

import { ROOT_PATH } from '@sourcegraph/build-config'

export const STATIC_ASSETS_URL = '/.assets/'
export const DEV_SERVER_LISTEN_ADDR = { host: '127.0.0.1', port: 3080 } as const
export const DEV_SERVER_PROXY_TARGET_ADDR = { host: '127.0.0.1', port: 3081 } as const
export const DEFAULT_SITE_CONFIG_PATH = path.resolve(ROOT_PATH, '../dev-private/enterprise/dev/site-config.json')

export const WEBPACK_STATS_OPTIONS = {
    all: false,
    timings: true,
    errors: true,
    warnings: true,
    colors: true,
} as const
