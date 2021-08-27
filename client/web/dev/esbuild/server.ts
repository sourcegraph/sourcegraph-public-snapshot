import http from 'http'

import { serve } from 'esbuild'
import signale from 'signale'

import { BUILD_OPTIONS } from './build'

export const esbuildDevelopmentServer = async (): Promise<void> => {
    // Start esbuild's server on a random local port.
    const { host: esbuildHost, port: esbuildPort } = await serve({ host: 'localhost' }, BUILD_OPTIONS)
    const upstreamHost = 'localhost'
    const upstreamPort = 3081

    // Start a proxy at :3080 to serve esbuild assets (if found) or otherwise forward to the upstream.
    const proxyServer = http
        .createServer((request, response) => {
            const commonRequestOptions: http.RequestOptions = {
                path: request.url,
                method: request.method,
                headers: request.headers,
            }

            // First, try serving each request from esbuild.
            const esbuildRequest = http.request(
                { ...commonRequestOptions, hostname: esbuildHost, port: esbuildPort },
                proxyResponse => {
                    // If esbuild returns "not found" for the request, it's probably not for an esbuild
                    // asset. Forward to the upstream.
                    if (proxyResponse.statusCode === 404) {
                        const upstreamRequest = http.request(
                            { ...commonRequestOptions, hostname: upstreamHost, port: upstreamPort },
                            upstreamResponse => {
                                response.writeHead(upstreamResponse.statusCode!, upstreamResponse.headers)
                                upstreamResponse.pipe(response, { end: true })
                            }
                        )
                        request.pipe(upstreamRequest, { end: true }) // forward the request body to the upstream
                        return
                    }

                    // Otherwise, return the asset response from esbuild.
                    response.writeHead(proxyResponse.statusCode!, proxyResponse.headers)
                    proxyResponse.pipe(response, { end: true })
                }
            )
            request.pipe(esbuildRequest, { end: true })
        })
        .listen({ host: 'localhost', port: 3080 })
    await new Promise<void>((resolve, reject) => {
        proxyServer.once('listening', () => {
            signale.success('esbuild server is ready')
            resolve()
        })
        proxyServer.once('error', error => reject(error))
    })
}
