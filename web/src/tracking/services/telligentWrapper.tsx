declare global {
    interface Window {
        telligent(...args: any[]): void
    }
}

class TelligentWrapper {
    private telligent?: (...args: any[]) => void | null
    private DEV_ENV = 'development'
    private PROD_ENV = 'production'

    constructor() {
        // Never log anything in self-hosted Sourcegraph instances.
        if (!window.context || !window.context.sourcegraphDotComMode) {
            return
        }

        if (window && window.telligent) {
            this.telligent = window.telligent
        } else {
            return
        }
        this.initialize(window.context.siteID, window.context.version === 'dev' ? this.DEV_ENV : this.PROD_ENV)
    }

    public addStaticMetadataObject(metadata: any): void {
        if (!this.telligent) {
            return
        }
        this.telligent('addStaticMetadataObject', metadata)
    }

    public setUserProperty(property: string, value: any): void {
        if (!this.telligent) {
            return
        }
        this.telligent('addStaticMetadata', property, value, 'userInfo')
    }

    public track(eventAction: string, eventProps: any): void {
        if (!this.telligent) {
            return
        }
        this.telligent('track', eventAction, eventProps)
    }

    /**
     * Function to extract the Telligent user ID from the first-party cookie set by the Telligent JavaScript Tracker
     * @return string or boolean The ID string if the cookie exists or null if the cookie has not been set yet
     */
    public getTelligentDuid(): string | null {
        const cookieProps = this.inspectTelligentCookie()
        return cookieProps ? cookieProps[0] : null
    }

    /**
     * Function to extract the Telligent session ID from the first-party cookie set by the Telligent JavaScript Tracker
     * @return string or boolean The session ID string if the cookie exists or null if the cookie has not been set yet
     */
    public getTelligentSessionId(): string | null {
        const cookieProps = this.inspectTelligentCookie()
        return cookieProps ? cookieProps[5] : null
    }

    private initialize(siteID: string, env: string): void {
        if (!this.telligent) {
            return
        }
        const telligentUrl = 'sourcegraph-logging.telligentdata.com'
        this.telligent('newTracker', 'sg', telligentUrl, {
            appId: siteID,
            platform: 'Web',
            encodeBase64: false,
            env,
            configUseCookies: true,
            useCookies: true,
            trackUrls: true,
            /**
             * NOTE: do not use window.location.hostname (which includes subdomains) as the cookieDomain
             * on sourcegraph.com subdomains (such as about.sourcegraph.com). Subdomains should be removed
             * from the cookieDomain property to ensure analytics user profiles sync across all Sourcegraph sites.
             */
            cookieDomain: window.location.hostname,
            metadata: {
                gaCookies: true,
                performanceTiming: true,
                augurIdentityLite: true,
                webPage: true,
            },
        })
    }

    private inspectTelligentCookie(): string[] | null {
        const cookieName = '_te_'
        const matcher = new RegExp(cookieName + 'id\\.[a-f0-9]+=([^;]+);?')
        const match = window.document.cookie.match(matcher)
        if (match && match[1]) {
            return match[1].split('.')
        } else {
            return null
        }
    }
}

export const telligent = new TelligentWrapper()
