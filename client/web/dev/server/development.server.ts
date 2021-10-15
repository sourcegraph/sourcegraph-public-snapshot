import chalk from 'chalk'
import compression from 'compression'
import { RequestHandler } from 'express'
import { createProxyMiddleware, Options as HTTPProxyMiddlewareOptions } from 'http-proxy-middleware'
import { once } from 'lodash'
import signale from 'signale'
import createWebpackCompiler, { Configuration } from 'webpack'
import WebpackDevServer, { ProxyConfigArrayItem } from 'webpack-dev-server'

import { getManifest } from '../esbuild/manifestPlugin'
import { esbuildDevelopmentServer } from '../esbuild/server'
import {
    getCSRFTokenCookieMiddleware,
    environmentConfig,
    getAPIProxySettings,
    getCSRFTokenAndCookie,
    shouldCompressResponse,
    STATIC_ASSETS_PATH,
    STATIC_ASSETS_URL,
    HTTPS_WEB_SERVER_URL,
    HTTP_WEB_SERVER_URL,
    PROXY_ROUTES,
} from '../utils'
import { getHTMLPage } from '../webpack/get-html-webpack-plugins'

// TODO: migrate webpack.config.js to TS to use `import` in this file.
// eslint-disable-next-line @typescript-eslint/no-var-requires, @typescript-eslint/no-require-imports
const webpackConfig = require('../../webpack.config') as Configuration
const {
    SOURCEGRAPH_API_URL,
    SOURCEGRAPH_HTTPS_PORT,
    CLIENT_PROXY_DEVELOPMENT_PORT,
    IS_HOT_RELOAD_ENABLED,
} = environmentConfig

interface DevelopmentServerInit {
    proxyRoutes: string[]
    proxyMiddlewareOptions: HTTPProxyMiddlewareOptions
    csrfTokenCookieMiddleware: RequestHandler
}

async function startDevelopmentServer(): Promise<void> {
    signale.start(
        `Starting ${environmentConfig.DEV_WEB_BUILDER} dev server with environment config:\n`,
        environmentConfig
    )

    if (!SOURCEGRAPH_API_URL) {
        throw new Error('development.server.ts only supports *web-standalone* usage')
    }

    // Get CSRF token value from the `SOURCEGRAPH_API_URL`.
    const { csrfContextValue, csrfCookieValue } = await getCSRFTokenAndCookie(SOURCEGRAPH_API_URL)

    const init: DevelopmentServerInit = {
        proxyRoutes: PROXY_ROUTES,
        proxyMiddlewareOptions: getAPIProxySettings({
            csrfContextValue,
            apiURL: SOURCEGRAPH_API_URL,
        }),
        csrfTokenCookieMiddleware: getCSRFTokenCookieMiddleware(csrfCookieValue),
    }

    switch (environmentConfig.DEV_WEB_BUILDER) {
        case 'webpack':
            await startWebpackDevelopmentServer(init)
            break

        case 'esbuild':
            await startEsbuildDevelopmentServer(init)
            break
    }
}

async function startWebpackDevelopmentServer({
    proxyRoutes,
    proxyMiddlewareOptions,
    csrfTokenCookieMiddleware,
}: DevelopmentServerInit): Promise<void> {
    const proxyConfig: ProxyConfigArrayItem = {
        context: proxyRoutes,
        ...proxyMiddlewareOptions,
    }

    const developmentServerConfig: WebpackDevServer.Configuration = {
        // react-refresh plugin triggers page reload if needed.
        liveReload: false,
        allowedHosts: 'all',
        hot: IS_HOT_RELOAD_ENABLED,
        historyApiFallback: {
            disableDotRule: true,
        },
        port: CLIENT_PROXY_DEVELOPMENT_PORT,
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
        proxy: [proxyConfig],
        // Disable default DevServer compression. We need more fine grained compression to support streaming search.
        compress: false,
        onBeforeSetupMiddleware: developmentServer => {
            developmentServer.app.use(csrfTokenCookieMiddleware)
            // Re-enable gzip compression using our own `compression` filter.
            developmentServer.app.use(compression({ filter: shouldCompressResponse }))
        },
    }

    const compiler = createWebpackCompiler(webpackConfig)
    const server = new WebpackDevServer(developmentServerConfig, compiler)

    compiler.hooks.afterEmit.tap(
        'development-server-logger',
        once(() => {
            signale.success('Webpack build is ready!')
        })
    )

    await server.start()
    signale.info(`Development HTTP server is ready at ${chalk.blue.bold(HTTP_WEB_SERVER_URL)}`)
    signale.success(`Development HTTPS server is ready at ${chalk.blue.bold(HTTPS_WEB_SERVER_URL)}`)
    signale.await('Waiting for Webpack to compile assets')
}

async function startEsbuildDevelopmentServer({
    proxyRoutes,
    proxyMiddlewareOptions,
    csrfTokenCookieMiddleware,
}: DevelopmentServerInit): Promise<void> {
    const manifest = getManifest()
    const htmlPage = getHTMLPage({
        head: `<link rel="stylesheet" href="${manifest['app.css']}">`,
        bodyEnd: `<script src="${manifest['app.js']}" type="module"></script>`,
    })

    await esbuildDevelopmentServer({ host: '0.0.0.0', port: SOURCEGRAPH_HTTPS_PORT }, app => {
        app.use(csrfTokenCookieMiddleware)
        app.use(createProxyMiddleware(proxyRoutes, proxyMiddlewareOptions))
        app.get(/.*/, (_request, response) => {
            response.send(htmlPage)
        })
    })
}

startDevelopmentServer().catch(error => signale.error(error))
