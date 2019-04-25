import { IGraphQLResponseRoot } from '../../../../shared/src/graphql/schema'
import { GraphQLRequestArgs } from '../shared/backend/graphql'

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
     * Allow error reporting.
     *
     * @duration permanent
     * @todo Since this is not really a feature flag, just unnest it into settings (and potentially get rid of the feature flags abstraction completely)
     */
    allowErrorReporting: boolean

    /**
     * Support link previews from extensions in content views (such as GitHub issues).
     */
    experimentalLinkPreviews: boolean

    /**
     * Support completion in text fields (such as on GitHub issues).
     */
    experimentalTextFieldCompletion: boolean
}

export const featureFlagDefaults: FeatureFlags = {
    allowErrorReporting: false,
    experimentalLinkPreviews: false,
    experimentalTextFieldCompletion: false,
}

export interface StorageItems {
    sourcegraphURL: string

    identity: string
    enterpriseUrls: string[]
    repoLocations: RepoLocations
    phabricatorMappings: PhabricatorMapping[]
    sourcegraphAnonymousUid: string
    disableExtension: boolean
    /**
     * Storage for feature flags.
     */
    featureFlags: Partial<FeatureFlags>
    clientConfiguration: ClientConfigurationDetails
    /**
     * Overrides settings from Sourcegraph.
     */
    clientSettings: string
    sideloadedExtensionURL: string | null
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
    sideloadedExtensionURL: null,
}

export interface BackgroundMessageHandlers {
    setIdentity({ identity }: { identity: string }): Promise<void>
    getIdentity(): Promise<string | undefined>

    setEnterpriseUrl(url: string): Promise<void>

    setSourcegraphUrl(url: string): Promise<void>

    removeEnterpriseUrl(url: string): Promise<void>

    insertCSS(details: { file: string; origin: string }): Promise<void>
    setBadgeText(text: string): void

    openOptionsPage(): Promise<void>

    createBlobURL(bundleUrl: string): Promise<string>

    requestGraphQL(params: GraphQLRequestArgs): Promise<IGraphQLResponseRoot>
}
