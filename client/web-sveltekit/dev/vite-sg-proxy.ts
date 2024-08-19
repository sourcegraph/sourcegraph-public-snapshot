import type { IncomingMessage } from 'http'
import { Transform } from 'stream'

import { createLogger, type Plugin, type ProxyOptions } from 'vite'

import { svelteKitRoutes, type SvelteKitRoute } from '../src/lib/routes'

interface Options {
    target: string
}

/**
 * This plugin proxies certain requests to a real Sourcegraph instance. These include
 * - API and auth requests
 * - asset requests for the React app
 * - requests for pages that are not known by the SvelteKit app (e.g. code insights)
 *
 * It does this by first fetching the sign-in page from the real Sourcegraph instance
 * and extracting the JS context object from it. This object contains a list of
 * routes known by the server, some of which will be handled by the SvelteKit app.
 * (the other data in JS context is ignored)
 * Those that are not will be proxied to the real Sourcegraph instance.
 *
 * Additionally, the plugin injects the JS context provided by the origin server into
 * locally generated HTML pages.
 *
 * This plugin is only enabled in 'serve' mode.
 */
export function sgProxy(options: Options): Plugin {
    const name = 'sg:proxy'
    const logger = createLogger(undefined, { prefix: `[${name}]` })
    // Needs to be kept in sync with app.html
    const contextPlaceholder = '// ---window.context---'

    // Additional endpoints that should be proxied to the real Sourcegraph instance.
    const additionalEndpoints = [
        // These are not part of the known routes list, but are required for the SvelteKit app to work
        // in development mode.
        '^/.api/',
        '^/.assets/',
        '^/.auth/',
        // Repo sub pages are also not part of the known routes list. They are listed here so that we
        // proxy them to the real Sourcegraph instance, for consistency with the production setup.
        '/-/raw/',
        '/-/batch-changes/?',
        '/-/settings/?',
        '/-/code-graph/?',
        '/-/own/?',
    ]

    // Routes known by the server that need to (potentially) be proxied to the real Sourcegraph instance.
    let knownServerRoutes: string[] = []

    function extractContextRaw(body: string): string | null {
        const match = body.match(/window\.context\s*=\s*{.*}/)
        return match?.[0] ?? null
    }

    function extractContext(body: string): Window['context'] | null {
        const context = extractContextRaw(body)
        if (!context) {
            return null
        }
        return new Function(`return ${context.match(/\{.*\}/)?.[0] ?? ''}`)()
    }

    /**
     * Returns true if the request should be handled by the SvelteKit app. This uses similar
     * logic to the `isRouteEnabled` function in the SvelteKit app.
     * If the request
     */
    function isHandledBySvelteKit(req: IncomingMessage, knownRoutes: string[]) {
        const url = new URL(req.url ?? '', `http://${req.headers.host}`)
        let foundRoute: SvelteKitRoute | undefined

        for (const route of svelteKitRoutes) {
            if (route.pattern.test(url.pathname)) {
                foundRoute = route
                if (!route.isRepoRoot) {
                    break
                }
            }
        }

        if (foundRoute) {
            return foundRoute.isRepoRoot ? !knownRoutes.some(route => new RegExp(route).test(url.pathname)) : true
        }
        return false
    }

    return {
        name,
        apply: 'serve',
        async config() {
            if (!options.target) {
                logger.info('No target specified, not proxying requests', { timestamp: true })
                return
            }

            let context: Window['context'] | null

            // At startup we fetch the sign-in page from the real Sourcegraph instance to extract the `knownRoutes` array
            // from the JS context object. This is used to determine which requests should be proxied to the real Sourcegraph
            // instance.
            // We keep trying to connect to the origin server in case it is not yet available (e.g. when just starting up a
            // local Sourcegraph instance).
            let backoff = 1
            while (true) {
                try {
                    logger.info(`Fetching JS context from ${options.target}`, { timestamp: true })
                    // The /sign-in endpoint is always available on dotcom and enterprise instances.
                    context = await fetch(`${options.target}/sign-in`)
                        .then(response => response.text())
                        .then(extractContext)
                    break
                } catch (error) {
                    logger.error(`Failed to fetch JS context: ${(error as Error).message}`, { timestamp: true })
                    logger.info(`Retrying in ${backoff} second(s)...`, { timestamp: true })
                    await new Promise(resolve => setTimeout(resolve, backoff * 1000))
                    backoff = Math.min(backoff * 2, 10)
                }
            }

            if (!context) {
                logger.error('Failed to extract JS context from origin', { timestamp: true })
                return
            }

            knownServerRoutes = context.svelteKit?.knownRoutes ?? []
            if (!knownServerRoutes.length) {
                logger.error('Failed to extract known routes from JS context', { timestamp: true })
                return
            }

            logger.info(`Known routes from origin JS context\n  - ${knownServerRoutes.join('\n  - ')}\n`, {
                timestamp: true,
            })

            const baseOptions: ProxyOptions = {
                target: options.target,
                changeOrigin: true,
                secure: false,
                headers: context.xhrHeaders,
            }

            const proxyConfig: Record<string, ProxyOptions> = {
                // Proxy requests to specific endpoints to a real Sourcegraph instance.
                [`${additionalEndpoints.join('|')}`]: baseOptions,
            }

            const dynamicOptions: ProxyOptions = {
                bypass(req) {
                    if (!req.url) {
                        return null
                    }
                    // If the request is for a SvelteKit route, we want to serve the SvelteKit app.
                    return isHandledBySvelteKit(req, knownServerRoutes) ? req.url : null
                },
                ...baseOptions,
            }

            for (const route of knownServerRoutes) {
                // vite's proxy server matches full URL, including query parameters.
                // That means a route regex like `^/search[/]?$` (which the server provides)
                // would not match `/search?q=foo`. We extend every route regex to allow
                // for any query parameters
                proxyConfig[route.replace(/\$$/, '(\\?.*)?$')] = dynamicOptions
            }

            return {
                server: {
                    proxy: proxyConfig,
                },
            }
        },
        configureServer(server) {
            if (!options.target) {
                return
            }

            server.middlewares.use(function proxyHTML(req, res, next) {
                // When a request is made for an HTML page that is handled by the SvelteKit
                // we want to inject the same JS context object that we would have fetched
                // from the origin server.
                // The implementation is quite hacky but apparently but it seems there is no
                // better way to do this. It was inspired by the express compression middleware:
                // https://github.com/expressjs/compression/blob/f3e6f389cb87e090438e13c04d67cec9e22f8098/index.js
                if (req.headers.accept?.includes('html') && isHandledBySvelteKit(req, knownServerRoutes)) {
                    const setHeader = res.setHeader
                    const write = res.write
                    const on = res.on
                    const end = res.end

                    const context = fetch(`${options.target}${req.url}`, {
                        headers: req.headers.cookie ? { cookie: req.headers.cookie } : {},
                    })
                        .then(response => response.text())
                        .then(body => {
                            const context = extractContextRaw(body)
                            if (!context) {
                                throw new Error('window.context not found in response from origin')
                            }
                            return context
                        })

                    const transform = new Transform({
                        transform(chunk, encoding, callback) {
                            context
                                .then(context => {
                                    let body = Buffer.from(chunk).toString()
                                    if (body.includes(contextPlaceholder)) {
                                        body = body.replace(contextPlaceholder, context)
                                        logger.info(`${req.url} - injected JS context`, { timestamp: true })
                                    }
                                    callback(null, body)
                                })
                                .catch(error => {
                                    logger.error(`Error fetching JS context: ${error.message}`, { timestamp: true })
                                    // We explicitly pass null to not cause the proxy to terminate
                                    callback(null, chunk)
                                })
                        },
                    })
                    transform
                        .on('data', chunk => {
                            // @ts-expect-error - the overload signature of write seems to prevent TS from recognizing the correct arguments
                            if (write.call(res, chunk) === false) {
                                transform.pause()
                            }
                        })
                        .on('end', () => {
                            // @ts-expect-error - the overload signature of end seems to prevent TS from recognizing the correct arguments
                            end.call(res)
                        })

                    res.on('drain', () => transform.resume())

                    let ended = false

                    res.setHeader = (name, value) => {
                        // content-length is set and sent before we have a chance to modify the response
                        // we need to ignore it, otherwise the browser will not render the page
                        // properly
                        return name === 'content-length' ? res : setHeader.call(res, name, value)
                    }

                    // @ts-expect-error - the overload signature of write seems to prevent TS from recognizing the correct arguments
                    res.write = (chunk, encoding, cb) => {
                        if (ended) {
                            return false
                        }
                        return transform.write(chunk, encoding, cb)
                    }

                    // @ts-expect-error - the overload signature of write seems to prevent TS from recognizing the correct arguments
                    res.end = (chunk, encoding, cb) => {
                        if (ended) {
                            return false
                        }
                        ended = true
                        return transform.end(chunk, encoding, cb)
                    }

                    // @ts-expect-error - the overload signature of write seems to prevent TS from recognizing the correct arguments
                    res.on = (type, listener) => {
                        if (type === 'drain') {
                            return transform.on(type, listener)
                        }
                        return on.call(res, type, listener)
                    }
                }
                next()
            })
        },
    }
}
