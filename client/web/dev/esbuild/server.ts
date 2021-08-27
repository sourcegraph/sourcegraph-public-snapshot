import http from 'http'

import { serve } from 'esbuild'
import signale from 'signale'

import { BUILD_OPTIONS } from './build'
import { assetPathPrefix } from './manifestPlugin'

export const esbuildDevelopmentServer = async (): Promise<void> => {
    // Start esbuild's server on a random local port.
    const { host: esbuildHost, port: esbuildPort, wait: esbuildStopped } = await serve(
        { host: 'localhost' },
        BUILD_OPTIONS
    )
    const upstreamHost = 'localhost'
    const upstreamPort = 3081

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
                        response.writeHead(proxyResponse.statusCode!, proxyResponse.headers)
                        proxyResponse.pipe(response, { end: true })
                    }
                )
                request.pipe(esbuildRequest, { end: true })
            } else {
                // Forward to upstream.
                const upstreamRequest = http.request(
                    {
                        ...commonRequestOptions,
                        hostname: upstreamHost,
                        port: upstreamPort,
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
        .listen({ host: 'localhost', port: 3080 })
    await new Promise<void>((resolve, reject) => {
        proxyServer.once('listening', () => {
            signale.success('esbuild server is ready')
            esbuildStopped.finally(() => proxyServer.close(error => (error ? reject(error) : resolve())))
        })
        proxyServer.once('error', error => reject(error))
    })
}
