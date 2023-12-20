import express from 'express'
import { createProxyMiddleware } from 'http-proxy-middleware'
import signale from 'signale'

import { ENVIRONMENT_CONFIG, HTTPS_WEB_SERVER_URL } from '../utils/environment-config'
import { getIndexHTML, getWebBuildManifest } from '../utils/get-index-html'
import { printSuccessBanner } from '../utils/success-banner'

import { getAPIProxySettings } from './apiProxySettings'

const { SOURCEGRAPH_API_URL, SOURCEGRAPH_HTTP_PORT } = ENVIRONMENT_CONFIG

interface DevelopmentServerInit {
    apiURL: string
    listenAddress?: { host: string; port: number }
}

async function startDevProxyServer({
    apiURL,
    listenAddress = { host: '127.0.0.1', port: SOURCEGRAPH_HTTP_PORT },
}: DevelopmentServerInit): Promise<void> {
    const { proxyRoutes, ...proxyMiddlewareOptions } = getAPIProxySettings({
        apiURL,
        getLocalIndexHTML(jsContextScript) {
            return getIndexHTML({ manifest: getWebBuildManifest(), jsContextScript })
        },
    })

    const proxyServer = express()
    proxyServer.use(createProxyMiddleware(proxyRoutes, proxyMiddlewareOptions))
    return new Promise<void>((_resolve, reject) => {
        proxyServer
            .listen(listenAddress)
            .once('listening', () => {
                printSuccessBanner(['✱ Sourcegraph is really ready now!', `Click here: ${HTTPS_WEB_SERVER_URL}`])
            })
            .once('error', error => reject(error))
    })
}

if (require.main === module) {
    signale.start('Starting dev server.', ENVIRONMENT_CONFIG)

    if (!SOURCEGRAPH_API_URL) {
        throw new Error(
            'development.server.ts only supports *web-standalone* usage (must set SOURCEGRAPH_API_URL env var)'
        )
    }

    startDevProxyServer({
        apiURL: SOURCEGRAPH_API_URL,
    }).catch(error => {
        signale.error(error)
        process.exit(1)
    })
}
