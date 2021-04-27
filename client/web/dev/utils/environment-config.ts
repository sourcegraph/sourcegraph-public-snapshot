export const environmentConfig = {
    NODE_ENV: process.env.NODE_ENV || 'development',
    SOURCEGRAPH_API_URL: process.env.SOURCEGRAPH_API_URL || 'https://k8s.sgdev.org',
    SOURCEGRAPH_HTTPS_DOMAIN: process.env.SOURCEGRAPH_HTTPS_DOMAIN || 'sourcegraph.test',
    SOURCEGRAPH_HTTPS_PORT: Number(process.env.SOURCEGRAPH_HTTPS_PORT) || 3443,
    WEBPACK_SERVE_INDEX: process.env.WEBPACK_SERVE_INDEX === 'true',

    // TODO: do we use process.env.NO_HOT anywhere?
    IS_HOT_RELOAD_ENABLED: process.env.NO_HOT !== 'true',
}
