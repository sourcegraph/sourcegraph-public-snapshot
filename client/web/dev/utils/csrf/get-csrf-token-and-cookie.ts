import fetch from 'node-fetch'

export const CSRF_CONTEXT_KEY = 'csrfToken'
const CSRF_CONTEXT_VALUE_REGEXP = new RegExp(`${CSRF_CONTEXT_KEY}":"(.*?)"`)

export const CSRF_COOKIE_NAME = 'sg_csrf_token'
const CSRF_COOKIE_VALUE_REGEXP = new RegExp(`${CSRF_COOKIE_NAME}=(.*?);`)

interface CSRFTokenAndCookie {
    csrfContextValue: string
    csrfCookieValue: string
}

/**
 *
 * Fetch `${proxyUrl}/sign-in` and extract two values from the response:
 *
 * 1. `set-cookie` value for `CSRF_COOKIE_NAME`.
 * 2. value from JS context under `CSRF_CONTEXT_KEY` key.
 *
 */
export async function getCSRFTokenAndCookie(proxyUrl: string): Promise<CSRFTokenAndCookie> {
    const response = await fetch(`${proxyUrl}/sign-in`)

    const html = await response.text()
    const cookieHeader = response.headers.get('set-cookie')

    if (!cookieHeader) {
        throw new Error(`"set-cookie" header not found in "${proxyUrl}/sign-in" response`)
    }

    const csrfHeaderMatches = CSRF_CONTEXT_VALUE_REGEXP.exec(html)
    const csrfCookieMatches = CSRF_COOKIE_VALUE_REGEXP.exec(cookieHeader)

    if (!csrfHeaderMatches || !csrfCookieMatches) {
        throw new Error('CSRF value not found!')
    }

    return {
        csrfContextValue: csrfHeaderMatches[1],
        csrfCookieValue: csrfCookieMatches[1],
    }
}
