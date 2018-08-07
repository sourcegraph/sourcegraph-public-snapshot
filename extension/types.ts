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
    newTooltips: boolean
}

export const featureFlagDefaults: FeatureFlags = {
    newTooltips: true,
}

export interface StorageItems {
    sourcegraphURL: string
    gitHubEnterpriseURL: string
    phabricatorURL: string
    inlineSymbolSearchEnabled: boolean
    renderMermaidGraphsEnabled: boolean
    repositoryFileTreeEnabled: boolean
    executeSearchEnabled: boolean
    sourcegraphRepoSearchToggled: boolean
    openEditorEnabled: boolean
    identity: string
    serverUrls: string[]
    enterpriseUrls: string[]
    serverUserId: string
    hasSeenServerModal: boolean
    repoLocations: RepoLocations
    phabricatorMappings: PhabricatorMapping[]
    openFileOnSourcegraph: boolean
    sourcegraphAnonymousUid: string
    disableExtension: boolean
    /**
     * Feature flag to use new LSP funcs (standardized with sourcegraph/sourcegraph). Soon this feature flag will
     * enable more CXP features.
     */
    useCXP: boolean
    /**
     * Storage for feature flags
     */
    featureFlags: FeatureFlags
}

export const defaultStorageItems: StorageItems = {
    sourcegraphURL: 'https://sourcegraph.com',
    serverUrls: ['https://sourcegraph.com'],
    gitHubEnterpriseURL: '',
    phabricatorURL: '',
    inlineSymbolSearchEnabled: true,
    renderMermaidGraphsEnabled: false,
    repositoryFileTreeEnabled: true,
    executeSearchEnabled: false,
    sourcegraphRepoSearchToggled: true,
    openEditorEnabled: false,
    identity: '',
    enterpriseUrls: [],
    serverUserId: '',
    hasSeenServerModal: false,
    repoLocations: {},
    phabricatorMappings: [],
    openFileOnSourcegraph: true,
    sourcegraphAnonymousUid: '',
    disableExtension: false,
    useCXP: false,
    featureFlags: featureFlagDefaults,
}

export type StorageChange = { [key in keyof StorageItems]: chrome.storage.StorageChange }
