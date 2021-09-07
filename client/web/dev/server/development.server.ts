import chalk from 'chalk'
import { once } from 'lodash'
import signale from 'signale'
import createWebpackCompiler, { Configuration } from 'webpack'
import WebpackDevServer, { ProxyConfigArrayItem } from 'webpack-dev-server'

import {
    getCSRFTokenCookieMiddleware,
    PROXY_ROUTES,
    environmentConfig,
    getAPIProxySettings,
    getCSRFTokenAndCookie,
    STATIC_ASSETS_PATH,
    STATIC_ASSETS_URL,
    WEB_SERVER_URL,
} from '../utils'

// TODO: migrate webpack.config.js to TS to use `import` in this file.
// eslint-disable-next-line @typescript-eslint/no-var-requires, @typescript-eslint/no-require-imports
const webpackConfig = require('../../webpack.config') as Configuration
const { SOURCEGRAPH_API_URL, SOURCEGRAPH_HTTPS_PORT, IS_HOT_RELOAD_ENABLED } = environmentConfig

export async function startDevelopmentServer(): Promise<void> {
    // Get CSRF token value from the `SOURCEGRAPH_API_URL`.
    const { csrfContextValue, csrfCookieValue } = await getCSRFTokenAndCookie(SOURCEGRAPH_API_URL)
    signale.start('Starting webpack-dev-server with environment config:\n', environmentConfig)

    const proxyConfig: ProxyConfigArrayItem = {
        context: PROXY_ROUTES,
        ...getAPIProxySettings({
            csrfContextValue,
            apiURL: SOURCEGRAPH_API_URL,
        }),
    }

    const developmentServerConfig: WebpackDevServer.Configuration = {
        // react-refresh plugin triggers page reload if needed.
        liveReload: false,
        allowedHosts: 'all',
        hot: IS_HOT_RELOAD_ENABLED,
        // TODO: resolve https://github.com/webpack/webpack-dev-server/issues/2313 and enable HTTPS.
        https: false,
        historyApiFallback: {
            disableDotRule: true,
        },
        port: SOURCEGRAPH_HTTPS_PORT,
        client: {
            overlay: false,
        },
        static: {
            directory: STATIC_ASSETS_PATH,
            publicPath: [STATIC_ASSETS_URL, '/'],
        },
        proxy: [proxyConfig],
        onBeforeSetupMiddleware: developmentServer => {
            developmentServer.app.use(getCSRFTokenCookieMiddleware(csrfCookieValue))
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
    signale.success(`Development server is ready at ${chalk.blue.bold(WEB_SERVER_URL)}`)
    signale.await('Waiting for Webpack to compile assets')
}

startDevelopmentServer().catch(error => signale.error(error))
