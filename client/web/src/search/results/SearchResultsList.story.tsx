import { storiesOf } from '@storybook/react'
import { createBrowserHistory } from 'history'
import * as React from 'react'
import _VisibilitySensor from 'react-visibility-sensor'
import { NEVER, of } from 'rxjs'
import sinon from 'sinon'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { ISearchResults } from '@sourcegraph/shared/src/graphql/schema'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    extensionsController,
    HIGHLIGHTED_FILE_LINES_REQUEST,
    MULTIPLE_SEARCH_REQUEST,
    REPO_MATCH_RESULT,
    SEARCH_REQUEST,
} from '@sourcegraph/shared/src/util/searchTestHelpers'

import { WebStory } from '../../components/WebStory'

import { SearchResultsList, SearchResultsListProps } from './SearchResultsList'

const history = createBrowserHistory()
history.replace({ search: 'q=r:golang/oauth2+test+f:travis' })

window.context.deployType = 'dev'

const defaultProps: SearchResultsListProps = {
    location: history.location,
    history,
    authenticatedUser: null,
    isSourcegraphDotCom: false,
    deployType: 'dev',

    resultsOrError: SEARCH_REQUEST(),
    onShowMoreResultsClick: sinon.spy(),

    allExpanded: true,
    onExpandAllResultsToggle: sinon.spy(),

    showSavedQueryModal: false,
    onSavedQueryModalClose: sinon.spy(),
    onSaveQueryClick: sinon.spy(),

    fetchHighlightedFileLineRanges: HIGHLIGHTED_FILE_LINES_REQUEST,

    isLightTheme: true,
    settingsCascade: {
        subjects: null,
        final: null,
    },
    extensionsController: { ...extensionsController, executeCommand: sinon.spy() },
    platformContext: { forceUpdateTooltip: sinon.spy(), settings: NEVER },
    telemetryService: NOOP_TELEMETRY_SERVICE,
    parsedSearchQuery: 'r:golang/oauth2 test f:travis',
    patternType: SearchPatternType.regexp,
    setPatternType: sinon.spy(),
    caseSensitive: false,
    setCaseSensitivity: sinon.spy(),

    versionContext: undefined,
    selectedSearchContextSpec: 'global',

    navbarSearchQueryState: { query: '' },

    shouldDisplayPerformanceWarning: () => of(false),
    enableCodeMonitoring: false,
}

const { add } = storiesOf('web/search/results/SearchResultsList', module).addParameters({
    chromatic: { viewports: [769, 993] },
})

add('loading', () => <WebStory>{() => <SearchResultsList {...defaultProps} resultsOrError={undefined} />}</WebStory>)

add('single result', () => <WebStory>{() => <SearchResultsList {...defaultProps} />}</WebStory>)

add('multiple results', () => (
    <WebStory>{() => <SearchResultsList {...defaultProps} resultsOrError={MULTIPLE_SEARCH_REQUEST()} />}</WebStory>
))

add('no results with quote tip in infobar', () => {
    const resultsOrError: ISearchResults = {
        ...(defaultProps.resultsOrError as ISearchResults),
        results: [],
        matchCount: 0,
        approximateResultCount: '0',
    }

    const location = {
        ...history.location,
        search: 'q="test"',
    }

    return (
        <WebStory>
            {() => (
                <SearchResultsList
                    {...defaultProps}
                    resultsOrError={resultsOrError}
                    patternType={SearchPatternType.literal}
                    location={location}
                />
            )}
        </WebStory>
    )
})

add('error', () => (
    <WebStory>
        {() => <SearchResultsList {...defaultProps} resultsOrError={{ message: 'test error', name: 'TestError' }} />}
    </WebStory>
))

add('show performance warning', () => {
    const shouldDisplayPerformanceWarning = () => of(true)

    return (
        <WebStory>
            {() => (
                <SearchResultsList
                    {...defaultProps}
                    shouldDisplayPerformanceWarning={shouldDisplayPerformanceWarning}
                />
            )}
        </WebStory>
    )
})

add('show server side alert', () => {
    const shouldDisplayPerformanceWarning = () => of(true)
    const resultsOrError: ISearchResults = {
        ...(defaultProps.resultsOrError as ISearchResults),
        alert: {
            __typename: 'SearchAlert',
            description: 'This is a test alert',
            proposedQueries: [{ __typename: 'SearchQueryDescription', description: 'Test query', query: 'test' }],
            title: 'Test Alert',
        },
    }

    return (
        <WebStory>
            {() => (
                <SearchResultsList
                    {...defaultProps}
                    resultsOrError={resultsOrError}
                    shouldDisplayPerformanceWarning={shouldDisplayPerformanceWarning}
                />
            )}
        </WebStory>
    )
})

add('show server side alert with timeout warning', () => {
    const shouldDisplayPerformanceWarning = () => of(true)
    const resultsOrError: ISearchResults = {
        ...(defaultProps.resultsOrError as ISearchResults),
        alert: {
            __typename: 'SearchAlert',
            description: 'This is a test alert',
            proposedQueries: [{ __typename: 'SearchQueryDescription', description: 'Test query', query: 'test' }],
            title: 'Test Alert',
        },
        timedout: [REPO_MATCH_RESULT],
        repositoriesCount: 1,
    }

    return (
        <WebStory>
            {() => (
                <SearchResultsList
                    {...defaultProps}
                    resultsOrError={resultsOrError}
                    shouldDisplayPerformanceWarning={shouldDisplayPerformanceWarning}
                />
            )}
        </WebStory>
    )
})
