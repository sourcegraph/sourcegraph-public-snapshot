import { EmailAddrList, ExternalToken, User } from 'sourcegraph/util/types'

/**
 * exported because webworkers need to be able to use it, and methods not transferred with context object
 */
export function isOnPremInstance(authEnabled: boolean): boolean {
    return !authEnabled
}

/**
 * SourcegraphContext is defined in cmd/frontend/internal/app/jscontext/jscontext.go JSContext struct
 */
export class SourcegraphContext {
    public xhrHeaders: { [key: string]: string }
    public csrfToken: string
    public userAgentIsBot: boolean
    public user: User | null
    public emails: EmailAddrList | null
    public gitHubToken: ExternalToken | null
    public sentryDSN: string
    public intercomHash: string

    public appURL: string // base URL for app (e.g., https://sourcegraph.com or http://localhost:3080)
    public assetsRoot: string // URL path to image/font/etc. assets on server
    public version: string
    /**
     * authEnabled, set as AUTH_ENABLED as an env var and enabled by default, causes Sourcegraph to require GitHub.com authentication.
     * With authEnabled set to false, no sign in is required or possible, and repositories are pulled from local disk. Used for on-prem.
     */
    public authEnabled: boolean
    /**
     * trackingAppID, set as "" by default server side, is required for the telligent environment to be set to production.
     * For Sourcegraph.com, it is SourcegraphWeb. For the node.aws.sgdev.org deployment, it might be something like SgdevWeb.
     * It is stored in telligent as a field called appID.
     */
    public trackingAppID: string | null
    /**
     * repoHomePageRegex filter is for on-premises deployments, to ensure that only organization repos appear on the home page.
     * For instance, on node.aws.sgdev.org, it is set to ^gitolite\.aws\.sgdev\.org.
     */
    public repoHomeRegexFilter: string

    constructor(ctx: any) {
        Object.assign(this, ctx)
    }

    /**
     * the browser extension is detected when it creates a div with id `sourcegraph-app-background` on page.
     * for on-premise or testing instances of Sourcegraph, the browser extension never runs, so this will return false.
     * proceed with caution.
     */
    public hasBrowserExtensionInstalled(): boolean {
        return document.getElementById('sourcegraph-app-background') !== null
    }

    public primaryEmail(): string | null {
        if (this.emails && this.emails.EmailAddrs) {
            return (this.emails.EmailAddrs.filter(e => e.Primary).map(e => e.Email)[0]) || null
        }
        return null
    }
}

declare global {
    interface Window {
        context: SourcegraphContext
    }
}

export const sourcegraphContext = new SourcegraphContext(window.context)
