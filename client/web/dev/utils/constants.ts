import path from 'path'

export const ROOT_PATH = path.resolve(__dirname, '../../../../')
export const STATIC_ASSETS_PATH = path.resolve(ROOT_PATH, 'ui/assets')
export const STATIC_ASSETS_URL = '/.assets/'

// TODO: share with gulpfile.js
export const WEBPACK_STATS_OPTIONS = {
    all: false,
    timings: true,
    errors: true,
    warnings: true,
    colors: true,
}
