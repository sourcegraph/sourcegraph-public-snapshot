import { Options } from 'http-proxy-middleware'

// One of the API routes: "/-/sign-in".
export const PROXY_ROUTES = ['/.api', '/search/stream', '/-', '/.auth']

interface GetAPIProxySettingsOptions {
    csrfContextValue: string
    apiURL: string
}

export const getAPIProxySettings = ({ csrfContextValue, apiURL }: GetAPIProxySettingsOptions): Options => ({
    target: apiURL,
    secure: false,
    changeOrigin: true,
    headers: {
        'x-csrf-token': csrfContextValue,
    },
    cookieDomainRewrite: '',
    onProxyRes: proxyResponse => {
        if (proxyResponse.headers['set-cookie']) {
            const cookies = proxyResponse.headers['set-cookie'].map(cookie =>
                cookie.replace(/; secure/gi, '').replace(/; samesite=.+/gi, '')
            )

            proxyResponse.headers['set-cookie'] = cookies
        }
    },
    // TODO: share with `client/web/gulpfile.js`
    // Avoid crashing on "read ECONNRESET".
    onError: () => undefined,
    // Don't log proxy errors, these usually just contain
    // ECONNRESET errors caused by the browser cancelling
    // requests. This should not be needed to actually debug something.
    logLevel: 'silent',
    onProxyReqWs: (_proxyRequest, _request, socket) =>
        socket.on('error', error => console.error('WebSocket proxy error:', error)),
})
