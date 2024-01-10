import chalk from 'chalk'
import historyApiFallback from 'connect-history-api-fallback'
import express, { type RequestHandler } from 'express'
import expressStaticGzip from 'express-static-gzip'
import { createProxyMiddleware } from 'http-proxy-middleware'
import signale from 'signale'

import {
    getAPIProxySettings,
    ENVIRONMENT_CONFIG,
    HTTP_WEB_SERVER_URL,
    HTTPS_WEB_SERVER_URL,
    getWebBuildManifest,
    STATIC_INDEX_PATH,
    getIndexHTML,
} from '../utils'

const { SOURCEGRAPH_API_URL, SOURCEGRAPH_HTTP_PORT, STATIC_ASSETS_PATH } = ENVIRONMENT_CONFIG

function startProductionServer(): void {
    if (!SOURCEGRAPH_API_URL) {
        throw new Error('production.server.ts only supports *web-standalone* usage')
    }

    signale.await('Starting production server', ENVIRONMENT_CONFIG)

    const app = express()

    // Serve index.html in place of any 404 responses.
    app.use(historyApiFallback() as RequestHandler)

    // Serve build artifacts.
    app.use(
        '/.assets',
        expressStaticGzip(STATIC_ASSETS_PATH, {
            enableBrotli: true,
            orderPreference: ['br', 'gz'],
            index: false,
        })
    )

    const { proxyRoutes, ...proxyConfig } = getAPIProxySettings({
        apiURL: SOURCEGRAPH_API_URL,
        ...(ENVIRONMENT_CONFIG.WEB_BUILDER_SERVE_INDEX && {
            getLocalIndexHTML(jsContextScript) {
                const manifestFile = getWebBuildManifest()
                return getIndexHTML({ manifestFile, jsContextScript })
            },
        }),
    })

    // Proxy API requests to the `process.env.SOURCEGRAPH_API_URL`.
    app.use(proxyRoutes, createProxyMiddleware(proxyConfig))

    // Redirect remaining routes to index.html
    app.get('/*', (_request, response) => response.sendFile(STATIC_INDEX_PATH))

    app.listen(SOURCEGRAPH_HTTP_PORT, () => {
        signale.info(`Production HTTP server is ready at ${chalk.blue.bold(HTTP_WEB_SERVER_URL)}`)
        signale.success(`Production HTTPS server is ready at ${chalk.blue.bold(HTTPS_WEB_SERVER_URL)}`)
    })
}

startProductionServer()
