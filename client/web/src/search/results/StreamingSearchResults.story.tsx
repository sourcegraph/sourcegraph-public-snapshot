import { storiesOf } from '@storybook/react'
import { createBrowserHistory } from 'history'
import { EMPTY, NEVER, of } from 'rxjs'
import sinon from 'sinon'

import { SearchQueryStateStoreProvider } from '@sourcegraph/search'
import { AggregateStreamingSearchResults } from '@sourcegraph/shared/src/search/stream'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    extensionsController,
    HIGHLIGHTED_FILE_LINES_LONG,
    MULTIPLE_SEARCH_RESULT,
    REPO_MATCH_RESULTS_WITH_METADATA,
} from '@sourcegraph/shared/src/testing/searchTestHelpers'

import { AuthenticatedUser } from '../../auth'
import { WebStory } from '../../components/WebStory'
import { useExperimentalFeatures, useNavbarQueryState } from '../../stores'

import { StreamingSearchResults, StreamingSearchResultsProps } from './StreamingSearchResults'

const history = createBrowserHistory()
history.replace({ search: 'q=r:golang/oauth2+test+f:travis' })

const streamingSearchResult: AggregateStreamingSearchResults = {
    state: 'complete',
    results: [...MULTIPLE_SEARCH_RESULT.results, ...REPO_MATCH_RESULTS_WITH_METADATA],
    filters: MULTIPLE_SEARCH_RESULT.filters,
    progress: {
        durationMs: 500,
        matchCount: MULTIPLE_SEARCH_RESULT.progress.matchCount,
        skipped: [],
    },
}

const defaultProps: StreamingSearchResultsProps = {
    extensionsController,
    telemetryService: NOOP_TELEMETRY_SERVICE,

    history,
    location: history.location,
    authenticatedUser: {
        url: '/users/alice',
        displayName: 'Alice',
        username: 'alice',
        email: 'alice@email.test',
    } as AuthenticatedUser,
    isLightTheme: true,

    settingsCascade: {
        subjects: null,
        final: null,
    },
    platformContext: { forceUpdateTooltip: sinon.spy(), settings: NEVER, requestGraphQL: () => EMPTY },

    streamSearch: () => of(streamingSearchResult),

    fetchHighlightedFileLineRanges: () => of(HIGHLIGHTED_FILE_LINES_LONG),
    isSourcegraphDotCom: false,
    searchContextsEnabled: true,
}

const { add } = storiesOf('web/search/results/StreamingSearchResults', module)
    .addParameters({
        chromatic: { viewports: [577, 769, 993], disableSnapshot: false },
    })
    .addDecorator(Story => {
        useExperimentalFeatures.setState({ codeMonitoring: true, showSearchContext: true })
        useNavbarQueryState.setState({ searchQueryFromURL: 'r:golang/oauth2 test f:travis' })
        return <Story />
    })

add('standard render', () => (
    <WebStory>
        {() => (
            <SearchQueryStateStoreProvider useSearchQueryState={useNavbarQueryState}>
                <StreamingSearchResults {...defaultProps} />
            </SearchQueryStateStoreProvider>
        )}
    </WebStory>
))

add('unauthenticated user standard render', () => (
    <WebStory>
        {() => (
            <SearchQueryStateStoreProvider useSearchQueryState={useNavbarQueryState}>
                <StreamingSearchResults {...defaultProps} authenticatedUser={null} />
            </SearchQueryStateStoreProvider>
        )}
    </WebStory>
))

add('no results', () => {
    const result: AggregateStreamingSearchResults = {
        state: 'complete',
        results: [],
        filters: [],
        progress: {
            durationMs: 500,
            matchCount: 0,
            skipped: [],
        },
    }

    return (
        <WebStory>
            {() => (
                <SearchQueryStateStoreProvider useSearchQueryState={useNavbarQueryState}>
                    <StreamingSearchResults {...defaultProps} streamSearch={() => of(result)} />
                </SearchQueryStateStoreProvider>
            )}
        </WebStory>
    )
})

add('search with quotes', () => {
    useNavbarQueryState.setState({ searchQueryFromURL: 'r:golang/oauth2 test f:travis "test"' })
    return (
        <WebStory>
            {() => (
                <SearchQueryStateStoreProvider useSearchQueryState={useNavbarQueryState}>
                    <StreamingSearchResults {...defaultProps} />
                </SearchQueryStateStoreProvider>
            )}
        </WebStory>
    )
})

add('did you mean', () => {
    useNavbarQueryState.setState({ searchQueryFromURL: 'javascript test' })
    return (
        <WebStory>
            {() => (
                <SearchQueryStateStoreProvider useSearchQueryState={useNavbarQueryState}>
                    <StreamingSearchResults {...defaultProps} />
                </SearchQueryStateStoreProvider>
            )}
        </WebStory>
    )
})

add('progress with warnings', () => {
    const result: AggregateStreamingSearchResults = {
        state: 'complete',
        results: MULTIPLE_SEARCH_RESULT.results,
        filters: MULTIPLE_SEARCH_RESULT.filters,
        progress: {
            durationMs: 500,
            matchCount: MULTIPLE_SEARCH_RESULT.progress.matchCount,
            skipped: [
                {
                    reason: 'excluded-fork',
                    message: '10k forked repositories excluded',
                    severity: 'info',
                    title: '10k forked repositories excluded',
                    suggested: {
                        title: 'forked:yes',
                        queryExpression: 'forked:yes',
                    },
                },
                {
                    reason: 'excluded-archive',
                    message: '60k archived repositories excluded',
                    severity: 'info',
                    title: '60k archived repositories excluded',
                    suggested: {
                        title: 'archived:yes',
                        queryExpression: 'archived:yes',
                    },
                },
                {
                    reason: 'shard-timedout',
                    message: 'Search timed out',
                    severity: 'warn',
                    title: 'Search timed out',
                    suggested: {
                        title: 'timeout:2m',
                        queryExpression: 'timeout:2m',
                    },
                },
            ],
        },
    }

    return (
        <WebStory>
            {() => (
                <SearchQueryStateStoreProvider useSearchQueryState={useNavbarQueryState}>
                    <StreamingSearchResults {...defaultProps} streamSearch={() => of(result)} />
                </SearchQueryStateStoreProvider>
            )}
        </WebStory>
    )
})

add('loading with no results', () => (
    <WebStory>
        {() => (
            <SearchQueryStateStoreProvider useSearchQueryState={useNavbarQueryState}>
                <StreamingSearchResults {...defaultProps} streamSearch={() => NEVER} />
            </SearchQueryStateStoreProvider>
        )}
    </WebStory>
))

add('loading with some results', () => {
    const result: AggregateStreamingSearchResults = {
        state: 'loading',
        results: MULTIPLE_SEARCH_RESULT.results,
        filters: MULTIPLE_SEARCH_RESULT.filters,
        progress: {
            durationMs: 500,
            matchCount: MULTIPLE_SEARCH_RESULT.progress.matchCount,
            skipped: [],
        },
    }

    return (
        <WebStory>
            {() => (
                <SearchQueryStateStoreProvider useSearchQueryState={useNavbarQueryState}>
                    <StreamingSearchResults {...defaultProps} streamSearch={() => of(result)} />
                </SearchQueryStateStoreProvider>
            )}
        </WebStory>
    )
})

add('server-side alert', () => {
    const result: AggregateStreamingSearchResults = {
        state: 'complete',
        results: MULTIPLE_SEARCH_RESULT.results,
        filters: MULTIPLE_SEARCH_RESULT.filters,
        progress: {
            durationMs: 500,
            matchCount: MULTIPLE_SEARCH_RESULT.progress.matchCount,
            skipped: [],
        },
        alert: {
            proposedQueries: [{ query: 'test', description: 'new query' }],
            title: 'Test alert',
            description: 'This is an alert',
        },
    }

    return (
        <WebStory>
            {() => (
                <SearchQueryStateStoreProvider useSearchQueryState={useNavbarQueryState}>
                    <StreamingSearchResults {...defaultProps} streamSearch={() => of(result)} />
                </SearchQueryStateStoreProvider>
            )}
        </WebStory>
    )
})

add('server-side alert with no results', () => {
    const result: AggregateStreamingSearchResults = {
        state: 'complete',
        results: [],
        filters: [],
        progress: {
            durationMs: 500,
            matchCount: MULTIPLE_SEARCH_RESULT.progress.matchCount,
            skipped: [],
        },
        alert: {
            proposedQueries: [{ query: 'test', description: 'Test query' }],
            title: 'Test Alert',
            description: 'This is a test alert',
        },
    }

    return (
        <WebStory>
            {() => (
                <SearchQueryStateStoreProvider useSearchQueryState={useNavbarQueryState}>
                    <StreamingSearchResults {...defaultProps} streamSearch={() => of(result)} />
                </SearchQueryStateStoreProvider>
            )}
        </WebStory>
    )
})

add('error with no results', () => {
    const result: AggregateStreamingSearchResults = {
        state: 'error',
        results: [],
        filters: [],
        progress: {
            durationMs: 500,
            matchCount: MULTIPLE_SEARCH_RESULT.progress.matchCount,
            skipped: [],
        },
        error: new Error('test error'),
    }

    return (
        <WebStory>
            {() => (
                <SearchQueryStateStoreProvider useSearchQueryState={useNavbarQueryState}>
                    <StreamingSearchResults {...defaultProps} streamSearch={() => of(result)} />
                </SearchQueryStateStoreProvider>
            )}
        </WebStory>
    )
})

add('error with some results', () => {
    const result: AggregateStreamingSearchResults = {
        state: 'error',
        results: MULTIPLE_SEARCH_RESULT.results,
        filters: MULTIPLE_SEARCH_RESULT.filters,
        progress: {
            durationMs: 500,
            matchCount: MULTIPLE_SEARCH_RESULT.progress.matchCount,
            skipped: [],
        },
        error: new Error('test error'),
    }

    return (
        <WebStory>
            {() => (
                <SearchQueryStateStoreProvider useSearchQueryState={useNavbarQueryState}>
                    <StreamingSearchResults {...defaultProps} streamSearch={() => of(result)} />
                </SearchQueryStateStoreProvider>
            )}
        </WebStory>
    )
})

add('limit hit with some results', () => {
    const result: AggregateStreamingSearchResults = {
        state: 'complete',
        results: MULTIPLE_SEARCH_RESULT.results,
        filters: MULTIPLE_SEARCH_RESULT.filters,
        progress: {
            durationMs: 500,
            matchCount: MULTIPLE_SEARCH_RESULT.progress.matchCount,
            skipped: [
                {
                    reason: 'document-match-limit',
                    message: 'result limit hit',
                    severity: 'info',
                    title: 'result limit hit',
                },
            ],
        },
    }

    return (
        <WebStory>
            {() => (
                <SearchQueryStateStoreProvider useSearchQueryState={useNavbarQueryState}>
                    <StreamingSearchResults {...defaultProps} streamSearch={() => of(result)} />
                </SearchQueryStateStoreProvider>
            )}
        </WebStory>
    )
})
