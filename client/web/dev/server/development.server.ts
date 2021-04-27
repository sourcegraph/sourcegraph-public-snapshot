import 'dotenv/config'

import chalk from 'chalk'
import signale from 'signale'
import createWebpackCompiler, { Configuration } from 'webpack'
import WebpackDevServer from 'webpack-dev-server'

import {
    getCSRFTokenCookieMiddleware,
    PROXY_ROUTES,
    environmentConfig,
    getAPIProxySettings,
    getCSRFTokenAndCookie,
    STATIC_ASSETS_PATH,
    STATIC_ASSETS_URL,
    WEBPACK_STATS_OPTIONS,
    WEB_SERVER_URL,
} from '../utils'

// TODO: migrate webpack.config.js to TS to use `import` in this file.
// eslint-disable-next-line @typescript-eslint/no-var-requires, @typescript-eslint/no-require-imports
const webpackConfig = require('../../webpack.config') as Configuration
const { SOURCEGRAPH_API_URL, SOURCEGRAPH_HTTPS_PORT, IS_HOT_RELOAD_ENABLED } = environmentConfig

export async function startDevelopmentServer(): Promise<void> {
    // Get CSRF token value from the `SOURCEGRAPH_API_URL`.
    const { csrfContextValue, csrfCookieValue } = await getCSRFTokenAndCookie(SOURCEGRAPH_API_URL)
    signale.await('Development server', { ...environmentConfig, csrfContextValue, csrfCookieValue })

    const options: WebpackDevServer.Configuration = {
        hot: IS_HOT_RELOAD_ENABLED,
        // TODO: resolve https://github.com/webpack/webpack-dev-server/issues/2313 and enable HTTPS.
        https: false,
        historyApiFallback: true,
        port: SOURCEGRAPH_HTTPS_PORT,
        publicPath: STATIC_ASSETS_URL,
        contentBase: STATIC_ASSETS_PATH,
        contentBasePublicPath: [STATIC_ASSETS_URL, '/'],
        stats: WEBPACK_STATS_OPTIONS,
        noInfo: false,
        disableHostCheck: true,
        proxy: [
            {
                context: PROXY_ROUTES,
                ...getAPIProxySettings({
                    csrfContextValue,
                    apiURL: SOURCEGRAPH_API_URL,
                }),
            },
        ],
        before(app) {
            app.use(getCSRFTokenCookieMiddleware(csrfCookieValue))
        },
    }

    WebpackDevServer.addDevServerEntrypoints(webpackConfig, options)

    const server = new WebpackDevServer(createWebpackCompiler(webpackConfig), options)

    server.listen(SOURCEGRAPH_HTTPS_PORT, '0.0.0.0', () => {
        signale.success(`Development server is ready at ${chalk.blue.bold(WEB_SERVER_URL)}`)
    })
}

startDevelopmentServer().catch(error => signale.error(error))
