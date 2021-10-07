import path from 'path'

import { ROOT_PATH } from './constants'

const DEFAULT_SITE_CONFIG_PATH = path.resolve(ROOT_PATH, '../dev-private/enterprise/dev/site-config.json')

export const environmentConfig = {
    NODE_ENV: process.env.NODE_ENV || 'development',
    SOURCEGRAPH_API_URL: process.env.SOURCEGRAPH_API_URL,
    SOURCEGRAPH_HTTPS_DOMAIN: process.env.SOURCEGRAPH_HTTPS_DOMAIN || 'sourcegraph.test',
    SOURCEGRAPH_HTTPS_PORT: Number(process.env.SOURCEGRAPH_HTTPS_PORT) || 3443,
    WEBPACK_SERVE_INDEX: process.env.WEBPACK_SERVE_INDEX === 'true',
    SITE_CONFIG_PATH: process.env.SITE_CONFIG_PATH || DEFAULT_SITE_CONFIG_PATH,
    ENTERPRISE: Boolean(process.env.ENTERPRISE),

    // Webpack is the default web build tool, and esbuild is an experimental option (see
    // https://docs.sourcegraph.com/dev/background-information/web/build#esbuild).
    // eslint-disable-next-line @typescript-eslint/no-unnecessary-type-assertion
    DEV_WEB_BUILDER: (process.env.DEV_WEB_BUILDER === 'esbuild' ? 'esbuild' : 'webpack') as 'esbuild' | 'webpack',

    // TODO: do we use process.env.NO_HOT anywhere?
    IS_HOT_RELOAD_ENABLED: process.env.NO_HOT !== 'true',
}

const { SOURCEGRAPH_HTTPS_PORT, SOURCEGRAPH_HTTPS_DOMAIN } = environmentConfig

export const WEB_SERVER_URL = `http://${SOURCEGRAPH_HTTPS_DOMAIN}:${SOURCEGRAPH_HTTPS_PORT}`
