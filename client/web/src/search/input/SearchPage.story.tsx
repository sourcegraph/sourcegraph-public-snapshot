import { storiesOf } from '@storybook/react'
import { parseISO } from 'date-fns'
import { createMemoryHistory } from 'history'
import React from 'react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { extensionsController } from '@sourcegraph/shared/src/util/searchTestHelpers'

import { WebStory } from '../../components/WebStory'
import { SearchPatternType } from '../../graphql-operations'
import { mockFetchAutoDefinedSearchContexts, mockFetchSearchContexts } from '../../searchContexts/testHelpers'
import { ThemePreference } from '../../theme'
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
    setVersionContext: () => Promise.resolve(undefined),
    availableVersionContexts: [],
    globbing: false,
    enableSmartQuery: false,
    parsedSearchQuery: 'r:golang/oauth2 test f:travis',
    patternType: SearchPatternType.literal,
    setPatternType: () => undefined,
    caseSensitive: false,
    setCaseSensitivity: () => undefined,
    platformContext: {} as any,
    keyboardShortcuts: [],
    copyQueryButton: false,
    versionContext: undefined,
    showSearchContext: false,
    showSearchContextManagement: false,
    selectedSearchContextSpec: '',
    setSelectedSearchContextSpec: () => {},
    defaultSearchContextSpec: '',
    showRepogroupHomepage: false,
    showEnterpriseHomePanels: false,
    showOnboardingTour: false,
    showQueryBuilder: false,
    isLightTheme: props.isLightTheme,
    fetchSavedSearches: _fetchSavedSearches,
    fetchRecentSearches: _fetchRecentSearches,
    fetchRecentFileViews: _fetchRecentFileViews,
    now: () => parseISO('2020-09-16T23:15:01Z'),
    fetchAutoDefinedSearchContexts: mockFetchAutoDefinedSearchContexts(),
    fetchSearchContexts: mockFetchSearchContexts,
})

const { add } = storiesOf('web/search/input/SearchPage', module).addParameters({
    design: {
        type: 'figma',
        url: 'https://www.figma.com/file/sPRyyv3nt5h0284nqEuAXE/12192-Sourcegraph-server-page-v1?node-id=255%3A3',
    },
    chromatic: { viewports: [544, 577, 769, 993, 1200] },
})

add('Cloud with panels', () => (
    <WebStory>
        {webProps => (
            <SearchPage {...defaultProps(webProps)} isSourcegraphDotCom={true} showEnterpriseHomePanels={true} />
        )}
    </WebStory>
))

add('Cloud without repogroups or panels', () => (
    <WebStory>{webProps => <SearchPage {...defaultProps(webProps)} isSourcegraphDotCom={true} />}</WebStory>
))

add('Cloud with repogroups', () => (
    <WebStory>
        {webProps => <SearchPage {...defaultProps(webProps)} isSourcegraphDotCom={true} showRepogroupHomepage={true} />}
    </WebStory>
))

add('Server without panels', () => <WebStory>{webProps => <SearchPage {...defaultProps(webProps)} />}</WebStory>)

add('Server without panels, with query builder', () => (
    <WebStory>{webProps => <SearchPage {...defaultProps(webProps)} showQueryBuilder={true} />}</WebStory>
))

add('Server with panels', () => (
    <WebStory>{webProps => <SearchPage {...defaultProps(webProps)} showEnterpriseHomePanels={true} />}</WebStory>
))

add('Server with panels and query builder', () => (
    <WebStory>
        {webProps => <SearchPage {...defaultProps(webProps)} showEnterpriseHomePanels={true} showQueryBuilder={true} />}
    </WebStory>
))
