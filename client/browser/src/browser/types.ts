interface RepoLocations {
    [key: string]: string
}

interface PhabricatorMapping {
    callsign: string
    path: string
}

/**
 * The feature flags available.
 */
export interface FeatureFlags {
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
    inlineSymbolSearchEnabled: true,
    allowErrorReporting: false,
}

// TODO(chris) Switch to Partial<StorageItems> to eliminate bugs caused by
// missing items.
export interface StorageItems {
    sourcegraphURL: string

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
    sideloadedExtensionURL: string | null
    NeedsServerConfigurationAlertDismissed?: boolean
    NeedsRepoConfigurationAlertDismissed?: {
        [repoName: string]: boolean
    }
}

interface ClientConfigurationDetails {
    contentScriptUrls: string[]
    parentSourcegraph: {
        url: string
    }
}

export const defaultStorageItems: StorageItems = {
    sourcegraphURL: 'https://sourcegraph.com',

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
    sideloadedExtensionURL: '',
}

export type StorageChange = { [key in keyof StorageItems]: chrome.storage.StorageChange }
