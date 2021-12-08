import chalk from 'chalk'
import historyApiFallback from 'connect-history-api-fallback'
import express, { RequestHandler } from 'express'
import expressStaticGzip from 'express-static-gzip'
import { createProxyMiddleware } from 'http-proxy-middleware'
import signale from 'signale'

import {
    PROXY_ROUTES,
    getAPIProxySettings,
    environmentConfig,
    STATIC_ASSETS_PATH,
    STATIC_INDEX_PATH,
    HTTP_WEB_SERVER_URL,
    HTTPS_WEB_SERVER_URL,
} from '../utils'

const { SOURCEGRAPH_API_URL, CLIENT_PROXY_DEVELOPMENT_PORT } = environmentConfig

function startProductionServer(): void {
    if (!SOURCEGRAPH_API_URL) {
        throw new Error('production.server.ts only supports *web-standalone* usage')
    }

    signale.await('Production server', { ...environmentConfig })

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

    // Proxy API requests to the `process.env.SOURCEGRAPH_API_URL`.
    app.use(
        PROXY_ROUTES,
        createProxyMiddleware(
            getAPIProxySettings({
                apiURL: SOURCEGRAPH_API_URL,
            })
        )
    )

    // Redirect remaining routes to index.html
    app.get('/*', (_request, response) => response.sendFile(STATIC_INDEX_PATH))

    app.listen(CLIENT_PROXY_DEVELOPMENT_PORT, () => {
        signale.info(`Production HTTP server is ready at ${chalk.blue.bold(HTTP_WEB_SERVER_URL)}`)
        signale.success(`Production HTTPS server is ready at ${chalk.blue.bold(HTTPS_WEB_SERVER_URL)}`)
    })
}

startProductionServer()
