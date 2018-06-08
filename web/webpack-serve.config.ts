import httpProxyMiddleware from 'http-proxy-middleware'
const convert = require('koa-connect') // no TypeScript typings available
import { Options } from 'webpack-serve'

// We use webpack-serve to enable auto-page-reloading after changes to web code. (It does NOT do
// "hot" reloading, which is updating the code without a page reload. But it's almost as good, and
// it's more reliable.)
//
// We used to use webpack-dev-server for this, but it is deprecated. webpack-serve is its successor.

export default {
    clipboard: false,
    content: '../ui/assets',
    port: 3088,
    hot: false,
    dev: {
        publicPath: '/.assets/',
    },
    add: (app, middleware, options) => {
        // Since we're manipulating the order of middleware added, we need to handle
        // adding these two internal middleware functions.
        middleware.webpack()
        middleware.content()

        // Proxy *must* be the last middleware added.
        app.use(
            convert(
                // Proxy all requests (that are not for webpack-built assets) to the Sourcegraph
                // frontend server, and we make the Sourcegraph appURL equal to the URL of
                // webpack-serve. This is how webpack-serve needs to work (because it does a bit
                // more magic in injecting scripts that use WebSockets into proxied requests).
                httpProxyMiddleware({
                    target: 'http://localhost:3080',
                    logLevel: 'debug',
                    // ... see: https://github.com/chimurai/http-proxy-middleware#options
                })
            )
        )
    },
} as Options
