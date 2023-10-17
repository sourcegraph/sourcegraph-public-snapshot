import compression from 'compression'
import { createProxyMiddleware } from 'http-proxy-middleware'
import { once } from 'lodash'
import signale from 'signale'
import createWebpackCompiler, { type Configuration } from 'webpack'
import WebpackDevServer from 'webpack-dev-server'

import { STATIC_ASSETS_PATH } from '@sourcegraph/build-config'

import { getManifest } from '../esbuild/manifestPlugin'
import { esbuildDevelopmentServer } from '../esbuild/server'
import {
    ENVIRONMENT_CONFIG,
    getAPIProxySettings,
    getIndexHTML,
    getWebBuildManifest,
    HTTP_WEB_SERVER_URL,
    HTTPS_WEB_SERVER_URL,
    printSuccessBanner,
    shouldCompressResponse,
    STATIC_ASSETS_URL,
} from '../utils'

// TODO: migrate webpack.config.js to TS to use `import` in this file.
// eslint-disable-next-line @typescript-eslint/no-var-requires, @typescript-eslint/no-require-imports
const webpackConfig = require('../../webpack.config') as Configuration

const { SOURCEGRAPH_API_URL, SOURCEGRAPH_HTTPS_PORT, SOURCEGRAPH_HTTP_PORT } = ENVIRONMENT_CONFIG

interface DevelopmentServerInit {
    apiURL: string
}

async function startDevelopmentServer(): Promise<void> {
    signale.start(`Starting ${ENVIRONMENT_CONFIG.DEV_WEB_BUILDER} dev server.`, ENVIRONMENT_CONFIG)

    if (!SOURCEGRAPH_API_URL) {
        throw new Error('development.server.ts only supports *web-standalone* usage')
    }

    const init: DevelopmentServerInit = {
        apiURL: SOURCEGRAPH_API_URL,
    }

    switch (ENVIRONMENT_CONFIG.DEV_WEB_BUILDER) {
        case 'webpack':
            await startWebpackDevelopmentServer(init)
            break

        case 'esbuild':
            await startEsbuildDevelopmentServer(init)
            break
    }
}

async function startWebpackDevelopmentServer({ apiURL }: DevelopmentServerInit): Promise<void> {
    const compiler = createWebpackCompiler(webpackConfig)

    const { proxyRoutes, ...proxyConfig } = getAPIProxySettings({
        apiURL,
        getLocalIndexHTML(jsContextScript) {
            const manifestFile = getWebBuildManifest()
            return getIndexHTML({ manifestFile, jsContextScript })
        },
    })

    const developmentServerConfig: WebpackDevServer.Configuration = {
        // react-refresh plugin triggers page reload if needed.
        liveReload: false,
        allowedHosts: 'all',
        hot: true,
        historyApiFallback: {
            disableDotRule: true,
        },
        port: SOURCEGRAPH_HTTP_PORT,
        client: {
            overlay: false,
            webSocketTransport: 'ws',
            logging: 'verbose',
            webSocketURL: {
                port: SOURCEGRAPH_HTTPS_PORT,
                protocol: 'wss',
            },
        },
        static: {
            directory: STATIC_ASSETS_PATH,
            publicPath: [STATIC_ASSETS_URL, '/'],
        },
        proxy: [
            {
                context: proxyRoutes,
                ...proxyConfig,
            },
        ],
        // Disable default DevServer compression. We need more fine grained compression to support streaming search.
        compress: false,
        setupMiddlewares: (middlewares, developmentServer) => {
            // Re-enable gzip compression using our own `compression` filter.
            developmentServer.app!.use(compression({ filter: shouldCompressResponse }))
            return middlewares
        },
    }

    const server = new WebpackDevServer(developmentServerConfig, compiler)

    compiler.hooks.done.tap(
        'development-server-logger',
        once(() => {
            printSuccessBanner(
                [
                    'Webpack build is ready!',
                    `Development HTTP server is ready at ${HTTP_WEB_SERVER_URL}`,
                    `Development HTTPS server is ready at ${HTTPS_WEB_SERVER_URL}`,
                ],
                signale.log.bind(signale)
            )
        })
    )

    await server.start()
}

async function startEsbuildDevelopmentServer({ apiURL }: DevelopmentServerInit): Promise<void> {
    const manifestFile = getManifest()
    const htmlPage = getIndexHTML({ manifestFile })

    const { proxyRoutes, ...proxyMiddlewareOptions } = getAPIProxySettings({
        apiURL,
        getLocalIndexHTML(jsContextScript) {
            return getIndexHTML({ manifestFile, jsContextScript })
        },
    })

    await esbuildDevelopmentServer({ host: '0.0.0.0', port: SOURCEGRAPH_HTTP_PORT }, app => {
        app.use(createProxyMiddleware(proxyRoutes, proxyMiddlewareOptions))
        app.get(/.*/, (_request, response) => {
            response.send(htmlPage)
        })
    })
}

startDevelopmentServer().catch(error => signale.error(error))
