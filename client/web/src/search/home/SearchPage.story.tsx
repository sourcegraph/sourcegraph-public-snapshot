import { DecoratorFn, Meta, Story } from '@storybook/react'
import { parseISO } from 'date-fns'
import { createMemoryHistory } from 'history'

import { getDocumentNode } from '@sourcegraph/http-client'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import {
    mockFetchAutoDefinedSearchContexts,
    mockFetchSearchContexts,
    mockGetUserSearchContextNamespaces,
} from '@sourcegraph/shared/src/testing/searchContexts/testHelpers'
import { extensionsController } from '@sourcegraph/shared/src/testing/searchTestHelpers'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { WebStory } from '../../components/WebStory'
import { MockedFeatureFlagsProvider } from '../../featureFlags/FeatureFlagsProvider'
import { useExperimentalFeatures } from '../../stores'
import { ThemePreference } from '../../theme'
import {
    HOME_PANELS_QUERY,
    RECENTLY_SEARCHED_REPOSITORIES_TO_LOAD,
    RECENT_FILES_TO_LOAD,
    RECENT_SEARCHES_TO_LOAD,
} from '../panels/HomePanels'
import {
    authUser,
    collaboratorsPayload,
    recentFilesPayload,
    recentSearchesPayload,
    savedSearchesPayload,
} from '../panels/utils'

import { SearchPage, SearchPageProps } from './SearchPage'

const history = createMemoryHistory()
const defaultProps = (props: ThemeProps): SearchPageProps => ({
    isSourcegraphDotCom: false,
    settingsCascade: {
        final: null,
        subjects: null,
    },
    location: history.location,
    history,
    extensionsController,
    telemetryService: NOOP_TELEMETRY_SERVICE,
    themePreference: ThemePreference.Light,
    onThemePreferenceChange: () => undefined,
    authenticatedUser: authUser,
    globbing: false,
    platformContext: {} as any,
    searchContextsEnabled: true,
    selectedSearchContextSpec: '',
    setSelectedSearchContextSpec: () => {},
    defaultSearchContextSpec: '',
    isLightTheme: props.isLightTheme,
    now: () => parseISO('2020-09-16T23:15:01Z'),
    fetchAutoDefinedSearchContexts: mockFetchAutoDefinedSearchContexts(),
    fetchSearchContexts: mockFetchSearchContexts,
    getUserSearchContextNamespaces: mockGetUserSearchContextNamespaces,
})

window.context.allowSignup = true

const decorator: DecoratorFn = Story => {
    useExperimentalFeatures.setState({ showSearchContext: false, showEnterpriseHomePanels: false })
    return <Story />
}

const config: Meta = {
    title: 'web/search/home/SearchPage',
    decorators: [decorator],
    parameters: {
        design: {
            type: 'figma',
            url: 'https://www.figma.com/file/sPRyyv3nt5h0284nqEuAXE/12192-Sourcegraph-server-page-v1?node-id=255%3A3',
        },
        chromatic: { viewports: [544, 577, 769, 993], disableSnapshot: false },
    },
}

export default config

function getMocks({
    enableSavedSearches,
    enableCollaborators,
}: {
    enableSavedSearches: boolean
    enableCollaborators: boolean
}) {
    return [
        {
            request: {
                query: getDocumentNode(HOME_PANELS_QUERY),
                variables: {
                    userId: '0',
                    firstRecentlySearchedRepositories: RECENTLY_SEARCHED_REPOSITORIES_TO_LOAD,
                    firstRecentSearches: RECENT_SEARCHES_TO_LOAD,
                    firstRecentFiles: RECENT_FILES_TO_LOAD,
                    enableSavedSearches,
                    enableCollaborators,
                },
            },
            result: {
                data: {
                    node: {
                        __typename: 'User',
                        recentlySearchedRepositoriesLogs: recentSearchesPayload(),
                        recentSearchesLogs: recentSearchesPayload(),
                        recentFilesLogs: recentFilesPayload(),
                        collaborators: enableCollaborators ? collaboratorsPayload() : undefined,
                    },
                    savedSearches: enableSavedSearches ? savedSearchesPayload() : undefined,
                },
            },
        },
    ]
}

export const CloudAuthedHome: Story = () => (
    <WebStory>
        {webProps => (
            <MockedTestProvider
                mocks={getMocks({
                    enableSavedSearches: false,
                    enableCollaborators: false,
                })}
            >
                <SearchPage {...defaultProps(webProps)} isSourcegraphDotCom={true} />
            </MockedTestProvider>
        )}
    </WebStory>
)

CloudAuthedHome.storyName = 'Cloud authenticated home'

export const CloudMarketingHome: Story = () => (
    <WebStory>
        {webProps => (
            <MockedFeatureFlagsProvider overrides={{}}>
                <SearchPage {...defaultProps(webProps)} isSourcegraphDotCom={true} authenticatedUser={null} />
            </MockedFeatureFlagsProvider>
        )}
    </WebStory>
)

CloudMarketingHome.storyName = 'Cloud marketing home'

export const ServerHome: Story = () => (
    <WebStory>
        {webProps => (
            <MockedTestProvider
                mocks={getMocks({
                    enableSavedSearches: true,
                    enableCollaborators: false,
                })}
            >
                <SearchPage {...defaultProps(webProps)} />
            </MockedTestProvider>
        )}
    </WebStory>
)

ServerHome.storyName = 'Server home'
