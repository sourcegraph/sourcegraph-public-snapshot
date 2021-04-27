import path from 'path'

import { environmentConfig } from './environment-config'

const { SOURCEGRAPH_HTTPS_PORT, SOURCEGRAPH_HTTPS_DOMAIN } = environmentConfig

export const ROOT_PATH = path.resolve(__dirname, '../../../../')
export const STATIC_ASSETS_PATH = path.resolve(ROOT_PATH, 'ui/assets')
export const STATIC_ASSETS_URL = '/.assets/'

export const WEB_SERVER_URL = `http://${SOURCEGRAPH_HTTPS_DOMAIN}:${SOURCEGRAPH_HTTPS_PORT}`

// TODO: share with gulpfile.js
export const WEBPACK_STATS_OPTIONS = {
    all: false,
    timings: true,
    errors: true,
    warnings: true,
    colors: true,
}
