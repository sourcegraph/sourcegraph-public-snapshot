interface PageError {
    statusCode: number
    statusText: string
    error: string
    errorID: string
}

interface Window {
    pageError?: PageError
    context: SourcegraphContext
    MonacoEnvironment: {
        getWorkerUrl(moduleId: string, label: string): string
    }
}

/**
 * Represents user properties that are guaranteed to both (1) be set if the user is signed in,
 * and (2) not change over a user session
 */
interface ImmutableUser {
    readonly UID: number
}

/**
 * Defined in cmd/frontend/internal/app/jscontext/jscontext.go JSContext struct
 */
interface SourcegraphContext {
    xhrHeaders: { [key: string]: string }
    csrfToken: string
    userAgentIsBot: boolean

    /**
     * The currently logged in user or null if the user is not signed
     * in/authenticated.
     *
     * @deprecated use currentUser in ./auth.ts instead
     */
    readonly user: ImmutableUser | null

    sentryDSN: string

    /** Base URL for app (e.g., https://sourcegraph.com or http://localhost:3080) */
    appURL: string

    /** URL path to image/font/etc. assets on server */
    assetsRoot: string

    version: string

    /**
     * Debug is whether debug mode is enabled.
     */
    debug: boolean

    sourcegraphDotComMode: boolean

    /**
     * siteID is the identifier of the Sourcegraph site. It is also the Telligent app ID.
     */
    siteID: string

    /**
     * Status of onboarding
     */
    showOnboarding: boolean

    /**
     * Emails support enabled
     */
    emailEnabled: boolean

    /**
     * A subset of the site configuration. Not all fields are set.
     */
    site: {
        'auth.public': boolean
    }

    /** Whether access tokens are enabled. */
    accessTokensEnabled: boolean

    /** Whether the reset-password flow is enabled. */
    resetPasswordEnabled: boolean

    /**
     * Likely running within a Docker container under a Mac host OS.
     */
    likelyDockerOnMac: boolean

    /**
     * Whether or not the server needs to restart in order to apply a pending
     * configuration change.
     */
    needServerRestart: boolean

    /**
     * Whether or not the server is running via datacenter deployment.
     */
    isRunningDataCenter: boolean

    /**
     * Whether the hostSurveysLocally feature flag is enabled.
     */
    hostSurveysLocallyEnabled: boolean

    /** Whether signup is allowed on the site. */
    allowSignup: boolean

    /** Authentication provider instances in site config. */
    authProviders?: {
        displayName: string
        isBuiltin: boolean
        authenticationURL?: string
    }[]
}

declare module '*.json' {
    const value: any
    export default value
}
