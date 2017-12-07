interface PageError {
    statusCode: number
    statusText: string
    error: string
    errorID: string
}

interface Window {
    pageError?: PageError
    context: SourcegraphContext
}

/**
 * Represents user properties that are guaranteed to both (1) be set if the user is signed in,
 * and (2) not change over a user session
 */
interface ImmutableUser {
    readonly UID: string
    /**
     * IsAdmin defines whether the current user has admin privileges for their org.
     */
    readonly IsAdmin: boolean
}

interface License {
    AppID: string
    Expiry: DateTime
}

/**
 * The GraphQL ID type.
 */
type GQLID = string

/**
 * Defined in cmd/frontend/internal/app/jscontext/jscontext.go JSContext struct
 */
interface SourcegraphContext {
    xhrHeaders: { [key: string]: string }
    sessionID: string
    csrfToken: string
    userAgentIsBot: boolean

    /** The currently logged in user or null if the user is not signed in/authenticated */
    readonly user: ImmutableUser | null
    /**
     * requireUserBackfill is a temporary flag which is true for legacy users that must
     * backfill data (username, and optionally display name) to be added to the users table.
     * While this flag is true, the client should force the currently logged in user to
     * provide the backfill data before taking other actions in the application.
     */
    requireUserBackfill: boolean
    sentryDSN: string
    intercomHash: string

    /** Base URL for app (e.g., https://sourcegraph.com or http://localhost:3080) */
    appURL: string

    /** URL path to image/font/etc. assets on server */
    assetsRoot: string

    version: string
    auth0Domain: string
    auth0ClientID: string

    /**
     * onPrem is defined as the env var DEPLOYMENT_ON_PREM. True if the server is a privately hosted instance, as opposed
     * to the public sourcegraph.com.
     */
    onPrem: boolean

    /**
     * trackingAppID, set as "" by default server side, is required for the telligent environment to be set to production.
     * For Sourcegraph.com, it is SourcegraphWeb. For the node.aws.sgdev.org deployment, it might be something like SgdevWeb.
     * It is stored in telligent as a field called appID.
     */
    trackingAppID: string | null

    /**
     * repoHomePageRegex filter is for on-premises deployments, to ensure that only organization repos appear on the home page.
     * For instance, on node.aws.sgdev.org, it is set to ^gitolite\.aws\.sgdev\.org.
     */
    repoHomeRegexFilter: string

    /**
     * githubEnterpriseURLs is a map of GitHub Enerprise hosts to their full URLs for outbound GitHub links.
     */
    githubEnterpriseURLs: { [key: string]: string }

    /**
     * Status of license
     */
    licenseStatus: string

    /**
     * Server License
     */
    license: License | null

    /**
     * Emails support enabled
     */
    emailEnabled: boolean
}
