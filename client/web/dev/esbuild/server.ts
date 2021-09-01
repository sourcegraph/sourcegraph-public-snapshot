import http from 'http'
import path from 'path'

import { serve } from 'esbuild'
import signale from 'signale'

import { DEV_SERVER_LISTEN_ADDR, DEV_SERVER_PROXY_TARGET_ADDR, STATIC_ASSETS_PATH } from '../utils'

import { buildMonaco, BUILD_OPTIONS } from './build'
import { assetPathPrefix } from './manifestPlugin'

export const esbuildDevelopmentServer = async (): Promise<void> => {
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
    const proxyServer = http
        .createServer((request, response) => {
            const commonRequestOptions: http.RequestOptions = {
                method: request.method,
                headers: request.headers,
            }

            const isAssetRequest = request.url!.startsWith(assetPathPrefix)
            if (isAssetRequest) {
                // Forward to esbuild.
                const esbuildRequest = http.request(
                    {
                        ...commonRequestOptions,
                        hostname: esbuildHost,
                        port: esbuildPort,
                        path: request.url!.slice(assetPathPrefix.length - 1),
                    },
                    proxyResponse => {
                        const isCacheableChunk = path.basename(request.url!).startsWith('chunk-')

                        response.writeHead(proxyResponse.statusCode!, {
                            ...proxyResponse.headers,

                            // Cache chunks because their filename includes a hash of the content.
                            'Cache-Control': isCacheableChunk ? 'max-age=3600' : 'no-cache',
                        })
                        proxyResponse.pipe(response, { end: true })
                    }
                )
                request.pipe(esbuildRequest, { end: true })
            } else {
                // Forward to upstream.
                const upstreamRequest = http.request(
                    {
                        ...commonRequestOptions,
                        hostname: DEV_SERVER_PROXY_TARGET_ADDR.host,
                        port: DEV_SERVER_PROXY_TARGET_ADDR.port,
                        path: request.url!,
                    },
                    proxyResponse => {
                        response.writeHead(proxyResponse.statusCode!, proxyResponse.headers)
                        proxyResponse.pipe(response, { end: true })
                    }
                )
                request.pipe(upstreamRequest, { end: true })
            }
        })
        .listen(DEV_SERVER_LISTEN_ADDR)
    await new Promise<void>((resolve, reject) => {
        proxyServer.once('listening', () => {
            signale.success('esbuild server is ready')
            esbuildStopped.finally(() => proxyServer.close(error => (error ? reject(error) : resolve())))
        })
        proxyServer.once('error', error => reject(error))
    })
}
