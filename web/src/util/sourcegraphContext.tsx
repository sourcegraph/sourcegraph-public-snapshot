/**
 * Represents user properties that are guaranteed to both (1) be set if the user is signed in,
 * and (2) not change over a user session
 */
export interface ImmutableUser {
    readonly UID: string
}

/**
 * SourcegraphContext is defined in cmd/frontend/internal/app/jscontext/jscontext.go JSContext struct
 */
export class SourcegraphContext {
    public xhrHeaders: { [key: string]: string }
    public sessionID: string
    public csrfToken: string
    public userAgentIsBot: boolean
    /**
     * user is an ImmutableUser object, which is only non-null if the user is signed in/authenticated
     */
    public readonly user: ImmutableUser | null
    public sentryDSN: string
    public intercomHash: string

    public appURL: string // base URL for app (e.g., https://sourcegraph.com or http://localhost:3080)
    public assetsRoot: string // URL path to image/font/etc. assets on server
    public version: string
    public auth0Domain: string
    public auth0ClientID: string
    /**
     * authEnabled, set as AUTH_ENABLED as an env var and enabled by default, causes Sourcegraph to require GitHub.com authentication.
     * With authEnabled set to false, no sign in is required or possible, and repositories are pulled from local disk. Used for on-prem.
     */
    public authEnabled: boolean
    /**
     * onPrem is defined as the env var DEPLOYMENT_ON_PREM. True if the server is a privately hosted instance, as opposed
     * to the public sourcegraph.com.
     */
    public onPrem: boolean
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
}

declare global {
    interface Window {
        context: SourcegraphContext
    }
}

export const sourcegraphContext = new SourcegraphContext(window.context)
