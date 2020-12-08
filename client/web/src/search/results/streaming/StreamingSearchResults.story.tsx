import { storiesOf } from '@storybook/react'
import { createBrowserHistory } from 'history'
import React from 'react'
import { NEVER, of } from 'rxjs'
import sinon from 'sinon'
import { SearchPatternType } from '../../../../../shared/src/graphql-operations'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { NOOP_TELEMETRY_SERVICE } from '../../../../../shared/src/telemetry/telemetryService'
import { WebStory } from '../../../components/WebStory'
import { AggregateStreamingSearchResults } from '../../stream'
import { StreamingSearchResults, StreamingSearchResultsProps } from './StreamingSearchResults'
import {
    extensionsController,
    HIGHLIGHTED_FILE_LINES_LONG,
    MULTIPLE_SEARCH_RESULT,
    REPO_MATCH_RESULT,
} from '../../../../../shared/src/util/searchTestHelpers'

const history = createBrowserHistory()
history.replace({ search: 'q=r:golang/oauth2+test+f:travis' })

const streamingSearchResult: AggregateStreamingSearchResults = {
    results: [...MULTIPLE_SEARCH_RESULT.results, REPO_MATCH_RESULT] as GQL.SearchResult[],
    filters: MULTIPLE_SEARCH_RESULT.dynamicFilters,
    progress: {
        done: true,
        durationMs: 500,
        matchCount: MULTIPLE_SEARCH_RESULT.matchCount,
        skipped: [],
    },
}

const defaultProps: StreamingSearchResultsProps = {
    caseSensitive: false,
    setCaseSensitivity: sinon.spy(),
    patternType: SearchPatternType.literal,
    setPatternType: sinon.spy(),
    versionContext: undefined,
    setVersionContext: sinon.spy(),
    availableVersionContexts: [],
    previousVersionContext: null,

    extensionsController,
    telemetryService: NOOP_TELEMETRY_SERVICE,

    history,
    location: history.location,
    authenticatedUser: null,
    isLightTheme: true,

    navbarSearchQueryState: { query: '', cursorPosition: 0 },

    settingsCascade: {
        subjects: null,
        final: null,
    },
    platformContext: { forceUpdateTooltip: sinon.spy(), settings: NEVER },

    streamSearch: () => of(streamingSearchResult),

    fetchHighlightedFileLines: () => of(HIGHLIGHTED_FILE_LINES_LONG),
}

const { add } = storiesOf('web/search/results/streaming/StreamingSearchResults', module).addParameters({
    chromatic: { viewports: [769, 993] },
})

add('standard render', () => <WebStory>{() => <StreamingSearchResults {...defaultProps} />}</WebStory>)

add('no results', () => {
    const result: AggregateStreamingSearchResults = {
        results: [],
        filters: [],
        progress: {
            done: true,
            durationMs: 500,
            matchCount: 0,
            skipped: [],
        },
    }

    return <WebStory>{() => <StreamingSearchResults {...defaultProps} streamSearch={() => of(result)} />}</WebStory>
})

add('diffs tab selected', () => {
    const history = createBrowserHistory()
    history.replace({ search: 'q=r:golang/oauth2+test+f:travis+type:diff' })

    return (
        <WebStory>
            {() => <StreamingSearchResults {...defaultProps} history={history} location={history.location} />}
        </WebStory>
    )
})

add('search with quotes', () => {
    const history = createBrowserHistory()
    history.replace({ search: 'q=r:golang/oauth2+test+f:travis+"test"' })

    return (
        <WebStory>
            {() => <StreamingSearchResults {...defaultProps} history={history} location={history.location} />}
        </WebStory>
    )
})

add('progress with warnings', () => {
    const result: AggregateStreamingSearchResults = {
        results: MULTIPLE_SEARCH_RESULT.results,
        filters: MULTIPLE_SEARCH_RESULT.dynamicFilters,
        progress: {
            done: true,
            durationMs: 500,
            matchCount: MULTIPLE_SEARCH_RESULT.matchCount,
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

    return <WebStory>{() => <StreamingSearchResults {...defaultProps} streamSearch={() => of(result)} />}</WebStory>
})

add('show version context warning', () => {
    const history = createBrowserHistory()
    history.replace({ search: 'q=r:golang/oauth2+test+f:travis&c=test' })

    return (
        <WebStory>
            {() => (
                <StreamingSearchResults
                    {...defaultProps}
                    history={history}
                    location={history.location}
                    previousVersionContext={null}
                    availableVersionContexts={[
                        { name: 'test', revisions: [] },
                        { name: 'other', revisions: [] },
                    ]}
                />
            )}
        </WebStory>
    )
})

add('loading with no results', () => (
    <WebStory>{() => <StreamingSearchResults {...defaultProps} streamSearch={() => NEVER} />}</WebStory>
))

add('loading with some results', () => {
    const result: AggregateStreamingSearchResults = {
        results: MULTIPLE_SEARCH_RESULT.results,
        filters: MULTIPLE_SEARCH_RESULT.dynamicFilters,
        progress: {
            done: false,
            durationMs: 500,
            matchCount: MULTIPLE_SEARCH_RESULT.matchCount,
            skipped: [],
        },
    }

    return <WebStory>{() => <StreamingSearchResults {...defaultProps} streamSearch={() => of(result)} />}</WebStory>
})
