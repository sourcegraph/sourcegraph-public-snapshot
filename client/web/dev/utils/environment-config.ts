/**
 * Unpack all `process.env.*` variables used during the build
 * time of the web application in this module to keep one source of truth.
 */
import { getEnvironmentBoolean } from '@sourcegraph/build-config'

import { DEFAULT_SITE_CONFIG_PATH } from './constants'

type WEB_BUILDER = 'esbuild' | 'webpack'

export const ENVIRONMENT_CONFIG = {
    /**
     * ----------------------------------------
     * Build configuration.
     * ----------------------------------------
     */
    NODE_ENV: process.env.NODE_ENV || 'development',
    // Determines if build is running on CI.
    CI: getEnvironmentBoolean('CI'),
    // Enables `embed` Webpack entry point.
    EMBED_DEVELOPMENT: getEnvironmentBoolean('EMBED_DEVELOPMENT'),

    // Should Webpack serve `index.html` with `HTMLWebpackPlugin`.
    WEBPACK_SERVE_INDEX: getEnvironmentBoolean('WEBPACK_SERVE_INDEX'),
    // Enables `StatoscopeWebpackPlugin` that allows to analyze application bundle.
    WEBPACK_BUNDLE_ANALYZER: getEnvironmentBoolean('WEBPACK_BUNDLE_ANALYZER'),

    // Allow overriding default Webpack naming behavior for debugging
    WEBPACK_USE_NAMED_CHUNKS: getEnvironmentBoolean('WEBPACK_USE_NAMED_CHUNKS'),

    //  Webpack is the default web build tool, and esbuild is an experimental option (see
    //  https://docs.sourcegraph.com/dev/background-information/web/build#esbuild).
    DEV_WEB_BUILDER: (process.env.DEV_WEB_BUILDER === 'esbuild' ? 'esbuild' : 'webpack') as WEB_BUILDER,

    /**
     * ----------------------------------------
     * Application features configuration.
     * ----------------------------------------
     */
    ENTERPRISE: getEnvironmentBoolean('ENTERPRISE'),
    SOURCEGRAPHDOTCOM_MODE: getEnvironmentBoolean('SOURCEGRAPHDOTCOM_MODE'),

    // Is reporting to Datadog/Sentry enabled.
    ENABLE_MONITORING: getEnvironmentBoolean('ENABLE_MONITORING'),

    /**
     * ----------------------------------------
     * Local environment configuration.
     * ----------------------------------------
     */
    SOURCEGRAPH_API_URL: process.env.SOURCEGRAPH_API_URL,
    SOURCEGRAPH_HTTPS_DOMAIN: process.env.SOURCEGRAPH_HTTPS_DOMAIN || 'sourcegraph.test',
    SOURCEGRAPH_HTTPS_PORT: Number(process.env.SOURCEGRAPH_HTTPS_PORT) || 3443,
    SOURCEGRAPH_HTTP_PORT: Number(process.env.SOURCEGRAPH_HTTP_PORT) || 3080,
    SITE_CONFIG_PATH: process.env.SITE_CONFIG_PATH || DEFAULT_SITE_CONFIG_PATH,
}

const { NODE_ENV, SOURCEGRAPH_HTTPS_DOMAIN, SOURCEGRAPH_HTTPS_PORT, SOURCEGRAPH_HTTP_PORT } = ENVIRONMENT_CONFIG

export const IS_DEVELOPMENT = NODE_ENV === 'development'
export const IS_PRODUCTION = NODE_ENV === 'production'

export const HTTPS_WEB_SERVER_URL = `https://${SOURCEGRAPH_HTTPS_DOMAIN}:${SOURCEGRAPH_HTTPS_PORT}`
export const HTTP_WEB_SERVER_URL = `http://localhost:${SOURCEGRAPH_HTTP_PORT}`
