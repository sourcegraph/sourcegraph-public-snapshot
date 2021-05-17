import { Options } from 'http-proxy-middleware'

// One of the API routes: "/-/sign-in".
export const PROXY_ROUTES = ['/.api', '/search/stream', '/-', '/.auth']

interface GetAPIProxySettingsOptions {
    csrfContextValue: string
    apiURL: string
}

export const getAPIProxySettings = ({ csrfContextValue, apiURL }: GetAPIProxySettingsOptions): Options => ({
    target: apiURL,
    // Do not SSL certificate.
    secure: false,
    // Change the origin of the host header to the target URL.
    changeOrigin: true,
    // Attach `x-csrf-token` header to every request to avoid "CSRF token is invalid" API error.
    headers: {
        'x-csrf-token': csrfContextValue,
    },
    // Rewrite domain of `set-cookie` headers for all cookies received.
    cookieDomainRewrite: '',
    onProxyRes: proxyResponse => {
        if (proxyResponse.headers['set-cookie']) {
            // Remove `Secure` and `SameSite` from `set-cookie` headers.
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
