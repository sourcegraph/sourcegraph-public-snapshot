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
        if (window && window.telligent) {
            this.telligent = window.telligent
        } else {
            return
        }
        this.initialize(window.context.siteID, window.context.version === 'dev' ? this.DEV_ENV : this.PROD_ENV)
    }

    public isTelligentLoaded(): boolean {
        return Boolean(this.telligent)
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
        // for on-prem usage, we only want to collect high level event context
        // note user identification information is still captured through persistent `user_info`
        // metadata stored in a cookie
        if (!window.context.sourcegraphDotComMode) {
            const limitedEventProps = {
                event_action: eventProps.eventAction,
                event_category: eventProps.eventCategory,
                event_label: eventProps.eventLabel,
                page_title: eventProps.page_title,
                language: eventProps.language,
                code_search: eventProps.code_search
                    ? {
                          query_data: eventProps.code_search.query_data
                              ? {
                                    query: eventProps.code_search.query_data.query,
                                }
                              : undefined,
                          results: eventProps.code_search.results,
                          source: eventProps.code_search.source,
                          suggestion: eventProps.code_search.suggestion
                              ? {
                                    type: eventProps.code_search.suggestion.type,
                                }
                              : undefined,
                      }
                    : undefined,
                id: eventProps.id,
                platform: eventProps.platform,
                server: eventProps.server,
            }
            this.telligent('track', eventAction, limitedEventProps)
            return
        }
        this.telligent('track', eventAction, eventProps)
    }

    /**
     * Function to extract the Telligent user ID from the first-party cookie set by the Telligent JavaScript Tracker
     * @return string or bool The ID string if the cookie exists or null if the cookie has not been set yet
     */
    public getTelligentDuid(): string | null {
        const cookieProps = this.inspectTelligentCookie()
        return cookieProps ? cookieProps[0] : null
    }

    /**
     * Function to extract the Telligent session ID from the first-party cookie set by the Telligent JavaScript Tracker
     * @return string or bool The session ID string if the cookie exists or null if the cookie has not been set yet
     */
    public getTelligentSessionId(): string | null {
        const cookieProps = this.inspectTelligentCookie()
        return cookieProps ? cookieProps[5] : null
    }

    private initialize(siteID: string, env: string): void {
        if (!this.telligent) {
            return
        }
        const telligentUrl = window.context.sourcegraphDotComMode
            ? 'sourcegraph-logging.telligentdata.com'
            : `${window.location.host}/.api/telemetry`
        this.telligent('newTracker', 'sg', telligentUrl, {
            appId: siteID,
            platform: 'Web',
            encodeBase64: false,
            env,
            configUseCookies: true,
            useCookies: true,
            trackUrls: window.context.sourcegraphDotComMode,
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

        // If on-prem, record Sourcegraph Server version
        if (!window.context.sourcegraphDotComMode) {
            this.telligent('addStaticMetadata', 'sgVersion', window.context.version, 'header')
        }
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
