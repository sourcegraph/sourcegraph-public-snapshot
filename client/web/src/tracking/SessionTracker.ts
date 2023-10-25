import cookies from 'js-cookie'

import { userCookieSettings, deviceSessionCookieSettings } from './cookieSettings'

const FIRST_SOURCE_URL_KEY = 'sourcegraphSourceUrl'
const LAST_SOURCE_URL_KEY = 'sourcegraphRecentSourceUrl'
const ORIGINAL_REFERRER_KEY = 'originalReferrer'
const MKTO_ORIGINAL_REFERRER_KEY = '_mkto_referrer'
const SESSION_REFERRER_KEY = 'sessionReferrer'
const SESSION_FIRST_URL_KEY = 'sessionFirstUrl'

/**
 * Configures and loads cookie properties for session tracking purposes.
 */
export class SessionTracker {
    /**
     * A lot of session-tracking is only done in Sourcegraph.com.
     */
    private isSourcegraphDotComMode = window.context?.sourcegraphDotComMode || false

    /**
     * We load initial values as the original code would check if we successfully
     * loaded a value, and if we didn't, try to load again - see getters on this
     * class.
     */
    private originalReferrer: string
    private sessionReferrer: string
    private sessionFirstURL: string
    private firstSourceURL: string
    private lastSourceURL: string

    constructor() {
        this.originalReferrer = this.getOriginalReferrer()
        this.sessionReferrer = this.getSessionReferrer()
        this.sessionFirstURL = this.getSessionFirstURL()
        this.firstSourceURL = this.getFirstSourceURL()
        this.lastSourceURL = this.getLastSourceURL()
    }

    public getOriginalReferrer(): string {
        if (!this.isSourcegraphDotComMode) {
            return ''
        }
        /**
         * Gets the original referrer from the cookie or, if it doesn't exist, the
         * mkto_referrer from the URL.
         */
        this.originalReferrer =
            this.originalReferrer ||
            cookies.get(ORIGINAL_REFERRER_KEY) ||
            cookies.get(MKTO_ORIGINAL_REFERRER_KEY) ||
            document.referrer

        cookies.set(ORIGINAL_REFERRER_KEY, this.originalReferrer, userCookieSettings)

        return this.originalReferrer
    }

    public getSessionReferrer(): string {
        // Gets the session referrer from the cookie
        if (!this.isSourcegraphDotComMode) {
            return ''
        }
        this.sessionReferrer = this.sessionReferrer || cookies.get(SESSION_REFERRER_KEY) || document.referrer

        cookies.set(SESSION_REFERRER_KEY, this.sessionReferrer, deviceSessionCookieSettings)
        return this.sessionReferrer
    }

    public getSessionFirstURL(): string {
        if (!this.isSourcegraphDotComMode) {
            return ''
        }
        this.sessionFirstURL = this.sessionFirstURL || cookies.get(SESSION_FIRST_URL_KEY) || location.href

        cookies.set(SESSION_FIRST_URL_KEY, this.sessionFirstURL, deviceSessionCookieSettings)
        return this.sessionFirstURL
    }

    public getFirstSourceURL(): string {
        if (!this.isSourcegraphDotComMode) {
            return ''
        }
        this.firstSourceURL = this.firstSourceURL || cookies.get(FIRST_SOURCE_URL_KEY) || location.href

        cookies.set(FIRST_SOURCE_URL_KEY, this.firstSourceURL, userCookieSettings)
        return this.firstSourceURL
    }

    public getLastSourceURL(): string {
        if (!this.isSourcegraphDotComMode) {
            return ''
        }

        /**
         * The cookie value gets overwritten each time a user visits a *.sourcegraph.com property.
         * This code lives in Google Tag Manager.
         */
        this.lastSourceURL = this.lastSourceURL || cookies.get(LAST_SOURCE_URL_KEY) || location.href

        cookies.set(LAST_SOURCE_URL_KEY, this.lastSourceURL, userCookieSettings)

        return this.lastSourceURL
    }

    public getReferrer(): string {
        if (this.isSourcegraphDotComMode) {
            return document.referrer
        }
        return ''
    }
}
