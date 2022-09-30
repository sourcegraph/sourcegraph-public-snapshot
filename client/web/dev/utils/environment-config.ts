/**
 * Unpack all `process.env.*` variables used during the build
 * time of the web application in this module to keep one source of truth.
 */
import { getEnvironmentBoolean } from '@sourcegraph/build-config'

import { DEFAULT_SITE_CONFIG_PATH } from './constants'

type WEB_BUILDER = 'esbuild' | 'webpack'

const NODE_ENV = process.env.NODE_ENV || 'development'

export const IS_DEVELOPMENT = NODE_ENV === 'development'
export const IS_PRODUCTION = NODE_ENV === 'production'

export const ENVIRONMENT_CONFIG = {
    /**
     * ----------------------------------------
     * Build configuration.
     * ----------------------------------------
     */
    NODE_ENV,
    // Determines if build is running on CI.
    CI: getEnvironmentBoolean('CI'),
    // Determines if the build will be used for integration tests.
    // Can be used to expose global variables to integration tests (e.g., CodeMirror API).
    // Enabled in the dev environment to allow debugging integration tests with the dev server.
    INTEGRATION_TESTS: getEnvironmentBoolean('INTEGRATION_TESTS') || IS_DEVELOPMENT,
    // Enables `embed` Webpack entry point.
    EMBED_DEVELOPMENT: getEnvironmentBoolean('EMBED_DEVELOPMENT'),

    // Should Webpack serve `index.html` with `HTMLWebpackPlugin`.
    WEBPACK_SERVE_INDEX: getEnvironmentBoolean('WEBPACK_SERVE_INDEX'),
    // Enables `StatoscopeWebpackPlugin` that allows to analyze application bundle.
    WEBPACK_BUNDLE_ANALYZER: getEnvironmentBoolean('WEBPACK_BUNDLE_ANALYZER'),
    // The name used to generate Statoscope JSON stats and HTML report in the `/ui/assets` folder.
    WEBPACK_STATS_NAME: process.env.WEBPACK_STATS_NAME,
    // Allow overriding default Webpack naming behavior for debugging
    WEBPACK_USE_NAMED_CHUNKS: getEnvironmentBoolean('WEBPACK_USE_NAMED_CHUNKS'),

    // The commit SHA the client bundle was built with.
    COMMIT_SHA: process.env.COMMIT_SHA,
    // The current Docker image version, use to associate builds with Sentry's source maps.
    VERSION: process.env.VERSION,
    // Should sourcemaps be uploaded to Sentry.
    SENTRY_UPLOAD_SOURCE_MAPS: getEnvironmentBoolean('SENTRY_UPLOAD_SOURCE_MAPS'),
    // Sentry's Dotcom project's authentication token
    SENTRY_DOT_COM_AUTH_TOKEN: process.env.SENTRY_DOT_COM_AUTH_TOKEN,
    // Sentry organization
    SENTRY_ORGANIZATION: process.env.SENTRY_ORGANIZATION,
    // Sentry project
    SENTRY_PROJECT: process.env.SENTRY_PROJECT,

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

    // Is reporting to Sentry enabled.
    ENABLE_SENTRY: getEnvironmentBoolean('ENABLE_SENTRY'),
    // Is OpenTelemetry instrumentation enabled.
    ENABLE_OPEN_TELEMETRY: getEnvironmentBoolean('ENABLE_OPEN_TELEMETRY'),

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
    CLIENT_OTEL_EXPORTER_OTLP_ENDPOINT: process.env.CLIENT_OTEL_EXPORTER_OTLP_ENDPOINT || '-/debug/otlp',
}

const { SOURCEGRAPH_HTTPS_DOMAIN, SOURCEGRAPH_HTTPS_PORT, SOURCEGRAPH_HTTP_PORT } = ENVIRONMENT_CONFIG

export const HTTPS_WEB_SERVER_URL = `https://${SOURCEGRAPH_HTTPS_DOMAIN}:${SOURCEGRAPH_HTTPS_PORT}`
export const HTTP_WEB_SERVER_URL = `http://localhost:${SOURCEGRAPH_HTTP_PORT}`
