import path from 'path'

import { serve } from 'esbuild'
import express from 'express'
import { createProxyMiddleware } from 'http-proxy-middleware'
import signale from 'signale'

import { STATIC_ASSETS_PATH } from '../utils'

import { buildMonaco, BUILD_OPTIONS } from './build'
import { assetPathPrefix } from './manifestPlugin'

export const esbuildDevelopmentServer = async (
    listenAddress: { host: string; port: number },
    configureProxy: (app: express.Application) => void
): Promise<void> => {
    // One-time build (these files only change when the monaco-editor npm package is changed, which
    // is rare enough to ignore here).
    await buildMonaco()

    // Start esbuild's server on a random local port.
    const { host: esbuildHost, port: esbuildPort, wait: esbuildStopped } = await serve(
        { host: 'localhost', servedir: STATIC_ASSETS_PATH },
        BUILD_OPTIONS
    )

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
    return await new Promise<void>((resolve, reject) => {
        proxyServer.once('listening', () => {
            signale.success('esbuild server is ready')
            esbuildStopped.finally(() => proxyServer.close(error => (error ? reject(error) : resolve())))
        })
        proxyServer.once('error', error => reject(error))
    })
}
