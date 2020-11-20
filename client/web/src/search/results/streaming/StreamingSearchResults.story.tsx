import { storiesOf } from '@storybook/react'
import { createBrowserHistory } from 'history'
import React from 'react'
import { of } from 'rxjs'
import sinon from 'sinon'
import { SearchPatternType } from '../../../../../shared/src/graphql-operations'
import { NOOP_TELEMETRY_SERVICE } from '../../../../../shared/src/telemetry/telemetryService'
import { extensionsController, MULTIPLE_SEARCH_RESULT } from '../../../../../shared/src/util/searchTestHelpers'
import { WebStory } from '../../../components/WebStory'
import { AggregateStreamingSearchResults } from '../../stream'
import { StreamingSearchResults, StreamingSearchResultsProps } from './StreamingSearchResults'

const history = createBrowserHistory()
history.replace({ search: 'q=r:golang/oauth2+test+f:travis' })

const streamingSearchResult: AggregateStreamingSearchResults = {
    results: MULTIPLE_SEARCH_RESULT.results,
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

    extensionsController,
    telemetryService: NOOP_TELEMETRY_SERVICE,

    history,
    location: history.location,

    navbarSearchQueryState: { query: '', cursorPosition: 0 },

    settingsCascade: {
        subjects: null,
        final: null,
    },

    streamSearch: () => of(streamingSearchResult),
}

const { add } = storiesOf('web/search/results/streaming/StreamingSearchResults', module).addParameters({
    chromatic: { viewports: [769, 993] },
})

add('render', () => <WebStory>{() => <StreamingSearchResults {...defaultProps} />}</WebStory>)
