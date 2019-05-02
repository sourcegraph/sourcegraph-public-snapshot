import { IGraphQLResponseRoot } from '../../../../shared/src/graphql/schema'
import { GraphQLRequestArgs } from '../shared/backend/graphql'
import { DEFAULT_SOURCEGRAPH_URL } from '../shared/util/context';

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

    enterpriseUrls: string[]
    phabricatorMappings: PhabricatorMapping[]

    /**
     * Storage for feature flags.
     */
    featureFlags: Partial<FeatureFlags>

    /**
     * Configuration details for the browser extension, editor extensions, etc from GraphQL.
     * See GraphQL schema.
     */
    clientConfiguration: ClientConfigurationDetails

    /**
     * Overrides settings from Sourcegraph.
     */
    clientSettings: string

    sideloadedExtensionURL: string | null
}

/**
 * Configuration details for the browser extension, editor extensions, etc from GraphQL.
 * See GraphQL schema.
 */
interface ClientConfigurationDetails {
    /**
     * The list of phabricator/gitlab/bitbucket/etc instance URLs that specifies
     * which pages the content script will be injected into.
     */
    contentScriptUrls: string[]

    /**
     * Returns details about the parent Sourcegraph instance.
     */
    parentSourcegraph: {
        /**
         * Sourcegraph instance URL.
         */
        url: string
    }
}

export const defaultStorageItems: StorageItems = {
    sourcegraphURL: DEFAULT_SOURCEGRAPH_URL,
    enterpriseUrls: [],
    phabricatorMappings: [],
    featureFlags: featureFlagDefaults,
    clientConfiguration: {
        contentScriptUrls: [],
        parentSourcegraph: {
            url: DEFAULT_SOURCEGRAPH_URL,
        },
    },
    clientSettings: '',
    sideloadedExtensionURL: null,
}

export interface BackgroundMessageHandlers {
    openOptionsPage(): Promise<void>
    createBlobURL(bundleUrl: string): Promise<string>
    requestGraphQL(params: GraphQLRequestArgs): Promise<IGraphQLResponseRoot>
}
