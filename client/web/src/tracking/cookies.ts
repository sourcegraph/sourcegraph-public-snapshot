import cookies, { type CookieAttributes } from 'js-cookie'

/**
 * Cookies is a simple interface over real cookies from 'js-cookie'.
 */
export interface Cookies {
    /**
     * Read cookie
     */
    get(name: string): string | undefined
    /**
     * Create a cookie
     */
    set(name: string, value: string, options?: CookieAttributes): string | undefined
}

/**
 * Alias for 'js-cookie' default implementation, behind the Cookies interface.
 */
export function defaultCookies(): Cookies {
    return cookies
}

export const userCookieSettings: CookieAttributes = {
    // 365 days expiry, but renewed on activity.
    expires: 365,
    // Enforce HTTPS
    secure: true,
    // We only read the cookie with JS so we don't need to send it cross-site nor on initial page requests.
    // However, we do need it on page redirects when users sign up via OAuth, hence using the Lax policy.
    // https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie/SameSite
    sameSite: 'Lax',
    // Specify the Domain attribute to ensure subdomains (about.sourcegraph.com) can receive this cookie.
    // https://developer.mozilla.org/en-US/docs/Web/HTTP/Cookies#define_where_cookies_are_sent
    domain: location.hostname,
}

export const deviceSessionCookieSettings: CookieAttributes = {
    // ~30 minutes expiry, but renewed on activity.
    expires: 0.0208,
    // Enforce HTTPS
    secure: true,
    // We only read the cookie with JS so we don't need to send it cross-site nor on initial page requests.
    // However, we do need it on page redirects when users sign up via OAuth, hence using the Lax policy.
    // https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie/SameSite
    sameSite: 'Lax',
    // Specify the Domain attribute to ensure subdomains (about.sourcegraph.com) can receive this cookie.
    // https://developer.mozilla.org/en-US/docs/Web/HTTP/Cookies#define_where_cookies_are_sent
    domain: location.hostname,
}
