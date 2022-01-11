import { storiesOf } from '@storybook/react'
import { parseISO } from 'date-fns'
import { createMemoryHistory } from 'history'
import React from 'react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    mockFetchAutoDefinedSearchContexts,
    mockFetchSearchContexts,
    mockGetUserSearchContextNamespaces,
} from '@sourcegraph/shared/src/testing/searchContexts/testHelpers'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { extensionsController } from '@sourcegraph/shared/src/util/searchTestHelpers'

import { WebStory } from '../../components/WebStory'
import { useExperimentalFeatures } from '../../stores'
import { ThemePreference } from '../../stores/themeState'
import { _fetchRecentFileViews, _fetchRecentSearches, _fetchSavedSearches, authUser } from '../panels/utils'

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
    parsedSearchQuery: 'r:golang/oauth2 test f:travis',
    platformContext: {} as any,
    keyboardShortcuts: [],
    searchContextsEnabled: true,
    selectedSearchContextSpec: '',
    setSelectedSearchContextSpec: () => {},
    defaultSearchContextSpec: '',
    isLightTheme: props.isLightTheme,
    fetchSavedSearches: _fetchSavedSearches,
    fetchRecentSearches: _fetchRecentSearches,
    fetchRecentFileViews: _fetchRecentFileViews,
    now: () => parseISO('2020-09-16T23:15:01Z'),
    fetchAutoDefinedSearchContexts: mockFetchAutoDefinedSearchContexts(),
    fetchSearchContexts: mockFetchSearchContexts,
    hasUserAddedRepositories: false,
    hasUserAddedExternalServices: false,
    getUserSearchContextNamespaces: mockGetUserSearchContextNamespaces,
    featureFlags: new Map(),
    extensionViews: () => null,
})

const { add } = storiesOf('web/search/home/SearchPage', module)
    .addParameters({
        design: {
            type: 'figma',
            url: 'https://www.figma.com/file/sPRyyv3nt5h0284nqEuAXE/12192-Sourcegraph-server-page-v1?node-id=255%3A3',
        },
        chromatic: { viewports: [544, 577, 769, 993, 1200] },
    })
    .addDecorator(Story => {
        useExperimentalFeatures.setState({ showSearchContext: false, showEnterpriseHomePanels: false })
        return <Story />
    })

add('Cloud with panels', () => (
    <WebStory>
        {webProps => {
            useExperimentalFeatures.setState({ showEnterpriseHomePanels: true })
            return <SearchPage {...defaultProps(webProps)} isSourcegraphDotCom={true} />
        }}
    </WebStory>
))

add('Cloud with community search contexts', () => (
    <WebStory>
        {webProps => <SearchPage {...defaultProps(webProps)} isSourcegraphDotCom={true} authenticatedUser={null} />}
    </WebStory>
))

add('Cloud with notebook onboarding', () => (
    <WebStory>
        {webProps => (
            <SearchPage
                {...defaultProps(webProps)}
                isSourcegraphDotCom={true}
                authenticatedUser={null}
                featureFlags={new Map([['search-notebook-onboarding', true]])}
            />
        )}
    </WebStory>
))

add('Server without panels', () => <WebStory>{webProps => <SearchPage {...defaultProps(webProps)} />}</WebStory>)

add('Server with panels', () => (
    <WebStory>
        {webProps => {
            useExperimentalFeatures.setState({ showEnterpriseHomePanels: true })
            return <SearchPage {...defaultProps(webProps)} />
        }}
    </WebStory>
))
