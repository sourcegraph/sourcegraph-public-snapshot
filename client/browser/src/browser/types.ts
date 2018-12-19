export interface RepoLocations {
    [key: string]: string
}

export interface PhabricatorMapping {
    callsign: string
    path: string
}

/**
 * The feature flags available.
 */
export interface FeatureFlags {
    /**
     * Whether or not to render [Mermaid](https://mermaidjs.github.io/) graphs
     * in markdown files viewed on GitHub.
     *
     * @duration permanent
     */
    renderMermaidGraphsEnabled: boolean
    /**
     * Whether or not to use the new inject method for code intelligence.
     *
     * @duration temporary - to be removed November first.
     */
    newInject: boolean
    /**
     * Enable inline symbol search by typing `!symbolQueryText` inside of GitHub PR comments (requires reload after toggling).
     *
     * @duration temporary - needs feedback from users.
     */
    inlineSymbolSearchEnabled: boolean

    /**
     * Allow error reporting.
     *
     * @duration permanent
     */
    allowErrorReporting: boolean
}

export const featureFlagDefaults: FeatureFlags = {
    newInject: false,
    renderMermaidGraphsEnabled: true,
    inlineSymbolSearchEnabled: true,
    allowErrorReporting: false,
}

export interface AccessToken {
    id: string
    token: string
}

/** A map where the key is the server URL and the value is the token. */
export interface AccessTokens {
    [url: string]: AccessToken
}

// TODO(chris) Switch to Partial<StorageItems> to eliminate bugs caused by
// missing items.
export interface StorageItems {
    sourcegraphURL: string
    /**
     * The current users access tokens the different sourcegraphUrls they have
     * had configured.
     */
    accessTokens: AccessTokens

    identity: string
    enterpriseUrls: string[]
    hasSeenServerModal: boolean
    repoLocations: RepoLocations
    phabricatorMappings: PhabricatorMapping[]
    sourcegraphAnonymousUid: string
    disableExtension: boolean
    /**
     * Storage for feature flags.
     */
    featureFlags: FeatureFlags
    clientConfiguration: ClientConfigurationDetails
    /**
     * Overrides settings from Sourcegraph.
     */
    clientSettings: string
    unpackedExtensionURL: string
}

interface ClientConfigurationDetails {
    contentScriptUrls: string[]
    parentSourcegraph: {
        url: string
    }
}

export const defaultStorageItems: StorageItems = {
    sourcegraphURL: 'https://sourcegraph.com',
    accessTokens: {},

    identity: '',
    enterpriseUrls: [],
    hasSeenServerModal: false,
    repoLocations: {},
    phabricatorMappings: [],
    sourcegraphAnonymousUid: '',
    disableExtension: false,
    featureFlags: featureFlagDefaults,
    clientConfiguration: {
        contentScriptUrls: [],
        parentSourcegraph: {
            url: 'https://sourcegraph.com',
        },
    },
    clientSettings: '',
    unpackedExtensionURL: '',
}

export type StorageChange = { [key in keyof StorageItems]: chrome.storage.StorageChange }
