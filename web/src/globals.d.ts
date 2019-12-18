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

type DeployType = 'cluster' | 'docker-container' | 'dev'

/**
 * Defined in cmd/frontend/internal/app/jscontext/jscontext.go JSContext struct
 */
interface SourcegraphContext
    extends Pick<Required<import('./schema/site.schema').SiteConfiguration>, 'experimentalFeatures'> {
    xhrHeaders: { [key: string]: string }
    csrfToken: string
    userAgentIsBot: boolean

    /**
     * Whether the user is authenticated. Use authenticatedUser in ./auth.ts to obtain information about the user.
     */
    readonly isAuthenticatedUser: boolean

    readonly sentryDSN: string | null

    /** Externally accessible URL for Sourcegraph (e.g., https://sourcegraph.com or http://localhost:3080). */
    externalURL: string

    /** URL path to image/font/etc. assets on server */
    assetsRoot: string

    version: string

    /**
     * Debug is whether debug mode is enabled.
     */
    debug: boolean

    sourcegraphDotComMode: boolean

    /**
     * siteID is the identifier of the Sourcegraph site.
     */
    siteID: string

    /** The GraphQL ID of the Sourcegraph site. */
    siteGQLID: GQL.ID

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
        'update.channel': string
    }

    /** Whether access tokens are enabled. */
    accessTokensAllow: 'all-users-create' | 'site-admin-create' | 'none'

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
     * The kind of deployment.
     */
    deployType: DeployType

    /** Whether signup is allowed on the site. */
    allowSignup: boolean

    /** Authentication provider instances in site config. */
    authProviders?: {
        displayName: string
        isBuiltin: boolean
        authenticationURL?: string
    }[]

    /** Custom branding for the homepage and search icon. */
    branding?: {
        /** The URL of the favicon to be used for your instance */
        favicon?: string

        /** Override style for light themes */
        light?: BrandAssets
        /** Override style for dark themes */
        dark?: BrandAssets

        /** Prevents the icon in the top-left corner of the screen from spinning. */
        disableSymbolSpin?: boolean

        brandName: string
    }
}

interface BrandAssets {
    /** The URL to the logo used on the homepage */
    logo?: string
    /** The URL to the symbol used as the search icon */
    symbol?: string
}

/**
 * For Web Worker entrypoints using Webpack's worker-loader.
 *
 * See https://github.com/webpack-contrib/worker-loader#integrating-with-typescript.
 */
declare module 'worker-loader?*' {
    class WebpackWorker extends Worker {
        constructor()
    }
    export default WebpackWorker
}
