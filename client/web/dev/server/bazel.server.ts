import compression from 'compression'
import type WebpackDevServer from 'webpack-dev-server'

import { STATIC_ASSETS_PATH } from '@sourcegraph/build-config'

import {
    ENVIRONMENT_CONFIG,
    getAPIProxySettings,
    getIndexHTML,
    getWebBuildManifest,
    shouldCompressResponse,
    STATIC_ASSETS_URL,
} from '../utils'

const { SOURCEGRAPH_API_URL, SOURCEGRAPH_HTTPS_PORT, SOURCEGRAPH_HTTP_PORT } = ENVIRONMENT_CONFIG

export function createDevelopmentServerConfig(): WebpackDevServer.Configuration {
    console.log(ENVIRONMENT_CONFIG)
    // if (!SOURCEGRAPH_API_URL) {
    //     throw new Error('bazel.server.ts only supports *web-standalone* usage')
    // }

    const apiURL = SOURCEGRAPH_API_URL!

    const { proxyRoutes, ...proxyConfig } = getAPIProxySettings({
        apiURL,
        getLocalIndexHTML(jsContextScript) {
            const manifestFile = getWebBuildManifest()
            return getIndexHTML({ manifestFile, jsContextScript })
        },
    })

    return {
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
}
