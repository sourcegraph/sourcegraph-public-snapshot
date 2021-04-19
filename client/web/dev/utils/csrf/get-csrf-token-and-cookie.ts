import fetch from 'node-fetch'

const CSRF_CONTEXT_KEY = 'csrfToken'
const CSRF_CONTEXT_VALUE_REGEXP = new RegExp(`${CSRF_CONTEXT_KEY}":"(.*?)"`)

const CSRF_COOKIE_NAME = 'sg_csrf_token'
const CSRF_COOKIE_VALUE_REGEXP = new RegExp(`${CSRF_COOKIE_NAME}=(.*?);`)

interface CSFRTokenAndCookie {
    csrfContextValue: string
    csrfCookieValue: string
}

export async function getCSRFTokenAndCookie(proxyUrl: string): Promise<CSFRTokenAndCookie> {
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
