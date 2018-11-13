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
     * Use the simpler options menu.
     *
     * @duration temporary - will be on by default after testing.
     */
    simpleOptionsMenu: boolean
    /**
     * Whether or not to render [Mermaid](https://mermaidjs.github.io/) graphs
     * in markdown files viewed on GitHub.
     *
     * @duration permanent
     */
    renderMermaidGraphsEnabled: boolean
    /**
     * Open files from the fuzzy file finder omnibar tool (src :f <file_name>)
     * on Sourcegraph or the codehost.
     *
     * @duration permanent
     */
    openFileOnSourcegraph: boolean
    /**
     * Whether or not to use the new inject method for code intelligence.
     *
     * @duration temporary - to be removed November first.
     */
    newInject: boolean
    /**
     * Enable the use of Sourcegraph extensions.
     *
     * @duration temporary - to be removed by @chris when extensions are stable and out of
     * beta.
     */
    useExtensions: boolean
    /**
     * Enable inline symbol search by typing `!symbolQueryText` inside of GitHub PR comments (requires reload after toggling).
     *
     * @duration temporary - needs feedback from users.
     */
    inlineSymbolSearchEnabled: boolean
    /**
     * Whether or not to execute a search on Sourcegraph when a search is
     * executed on the code host.
     *
     * @duration permanent
     */
    executeSearchEnabled: boolean
    /**
     * Display the Sourcegraph file tree in the code host when viewing a repository.
     *
     * @duration permanent
     */
    repositoryFileTreeEnabled: boolean
}

export const featureFlagDefaults: FeatureFlags = {
    simpleOptionsMenu: true,
    newInject: false,
    renderMermaidGraphsEnabled: false,
    useExtensions: false,
    openFileOnSourcegraph: true,
    inlineSymbolSearchEnabled: true,
    executeSearchEnabled: false,
    repositoryFileTreeEnabled: true,
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

    gitHubEnterpriseURL: string
    phabricatorURL: string
    identity: string
    serverUrls: string[]
    enterpriseUrls: string[]
    serverUserId: string
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

    serverUrls: ['https://sourcegraph.com'],
    gitHubEnterpriseURL: '',
    phabricatorURL: '',
    identity: '',
    enterpriseUrls: [],
    serverUserId: '',
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
}

export type StorageChange = { [key in keyof StorageItems]: chrome.storage.StorageChange }
