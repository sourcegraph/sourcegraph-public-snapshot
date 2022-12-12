import path from 'path'

import { serve } from 'esbuild'
import express from 'express'
import { createProxyMiddleware } from 'http-proxy-middleware'
import signale from 'signale'

import { STATIC_ASSETS_PATH, buildMonaco } from '@sourcegraph/build-config'

import { HTTPS_WEB_SERVER_URL, printSuccessBanner } from '../utils'

import { BUILD_OPTIONS } from './build'
import { assetPathPrefix } from './manifestPlugin'

export const esbuildDevelopmentServer = async (
    listenAddress: { host: string; port: number },
    configureProxy: (app: express.Application) => void
): Promise<void> => {
    const start = performance.now()

    // One-time build (these files only change when the monaco-editor npm package is changed, which
    // is rare enough to ignore here).
    await buildMonaco(STATIC_ASSETS_PATH)

    // Start esbuild's server on a random local port.
    const {
        host: esbuildHost,
        port: esbuildPort,
        wait: esbuildStopped,
    } = await serve({ host: 'localhost', servedir: STATIC_ASSETS_PATH }, BUILD_OPTIONS)

    // Start a proxy at :3080. Asset requests (underneath /.assets/) go to esbuild; all other
    // requests go to the upstream.
    const proxyApp = express()
    proxyApp.use(
        assetPathPrefix,
        createProxyMiddleware({
            target: { protocol: 'http:', host: esbuildHost, port: esbuildPort },
            pathRewrite: { [`^${assetPathPrefix}`]: '' },
            onProxyRes: (proxyResponse, request) => {
                // Cache chunks because their filename includes a hash of the content.
                const isCacheableChunk = path.basename(request.url).startsWith('chunk-')
                proxyResponse.headers['Cache-Control'] = isCacheableChunk ? 'max-age=3600' : 'no-cache'
            },
            logLevel: 'error',
        })
    )
    configureProxy(proxyApp)

    const proxyServer = proxyApp.listen(listenAddress)
    // eslint-disable-next-line @typescript-eslint/return-await
    return await new Promise<void>((resolve, reject) => {
        proxyServer.once('listening', () => {
            signale.success(`esbuild server is ready after ${Math.round(performance.now() - start)}ms`)
            printSuccessBanner(['✱ Sourcegraph is really ready now!', `Click here: ${HTTPS_WEB_SERVER_URL}`])
            esbuildStopped.finally(() => proxyServer.close(error => (error ? reject(error) : resolve())))
        })
        proxyServer.once('error', error => reject(error))
    })
}
