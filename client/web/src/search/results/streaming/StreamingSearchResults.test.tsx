import { mount } from 'enzyme'
import { createBrowserHistory } from 'history'
import React from 'react'
import { act } from 'react-dom/test-utils'
import { BrowserRouter } from 'react-router-dom'
import { NEVER, of } from 'rxjs'
import sinon from 'sinon'
import { SearchPatternType } from '../../../../../shared/src/graphql-operations'
import { NOOP_TELEMETRY_SERVICE } from '../../../../../shared/src/telemetry/telemetryService'
import { extensionsController, MULTIPLE_SEARCH_RESULT } from '../../../../../shared/src/util/searchTestHelpers'
import { AggregateStreamingSearchResults } from '../../stream'
import { SearchResultsInfoBar } from '../SearchResultsInfoBar'
import { StreamingProgress } from './progress/StreamingProgress'
import { StreamingSearchResults, StreamingSearchResultsProps } from './StreamingSearchResults'

describe('StreamingSearchResults', () => {
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
        authenticatedUser: null,

        navbarSearchQueryState: { query: '', cursorPosition: 0 },

        settingsCascade: {
            subjects: null,
            final: null,
        },
        platformContext: { forceUpdateTooltip: sinon.spy(), settings: NEVER },

        streamSearch: () => of(streamingSearchResult),
    }

    it('should call streaming search API with the right parameters from URL', () => {
        const history = createBrowserHistory()
        history.replace({ search: 'q=r:golang/oauth2+test+f:travis+case:yes&patternType=regexp&c=test' })

        const searchSpy = sinon.spy(defaultProps.streamSearch)

        const element = mount(
            <BrowserRouter>
                <StreamingSearchResults
                    {...defaultProps}
                    history={history}
                    location={history.location}
                    streamSearch={searchSpy}
                />
            </BrowserRouter>
        )

        sinon.assert.calledOnce(searchSpy)
        sinon.assert.calledWith(
            searchSpy,
            'r:golang/oauth2 test f:travis  case:yes',
            'V2',
            SearchPatternType.regexp,
            'test'
        )

        element.unmount()
    })

    it('should update patternType if different between URL and context', () => {
        const history = createBrowserHistory()
        history.replace({ search: 'q=r:golang/oauth2+test+f:travis&patternType=regexp' })

        const setPatternTypeSpy = sinon.spy()

        const element = mount(
            <BrowserRouter>
                <StreamingSearchResults
                    {...defaultProps}
                    history={history}
                    location={history.location}
                    patternType={SearchPatternType.literal}
                    setPatternType={setPatternTypeSpy}
                />
            </BrowserRouter>
        )

        sinon.assert.calledOnce(setPatternTypeSpy)
        sinon.assert.calledWith(setPatternTypeSpy, SearchPatternType.regexp)

        element.unmount()
    })

    it('should not update patternType if URL and context are the same', () => {
        const history = createBrowserHistory()
        history.replace({ search: 'q=r:golang/oauth2+test+f:travis&patternType=regexp' })

        const setPatternTypeSpy = sinon.spy()

        const element = mount(
            <BrowserRouter>
                <StreamingSearchResults
                    {...defaultProps}
                    history={history}
                    location={history.location}
                    patternType={SearchPatternType.regexp}
                    setPatternType={setPatternTypeSpy}
                />
            </BrowserRouter>
        )

        sinon.assert.notCalled(setPatternTypeSpy)

        element.unmount()
    })

    it('should update caseSensitive if different between URL and context', () => {
        const history = createBrowserHistory()
        history.replace({ search: 'q=r:golang/oauth2+test+f:travis case:yes' })

        const setCaseSensitivitySpy = sinon.spy()

        const element = mount(
            <BrowserRouter>
                <StreamingSearchResults
                    {...defaultProps}
                    history={history}
                    location={history.location}
                    caseSensitive={false}
                    setCaseSensitivity={setCaseSensitivitySpy}
                />
            </BrowserRouter>
        )

        sinon.assert.calledOnce(setCaseSensitivitySpy)
        sinon.assert.calledWith(setCaseSensitivitySpy, true)

        element.unmount()
    })

    it('should not update caseSensitive if URL and context are the same', () => {
        const history = createBrowserHistory()
        history.replace({ search: 'q=r:golang/oauth2+test+f:travis+case:yes' })

        const setCaseSensitivitySpy = sinon.spy()

        const element = mount(
            <BrowserRouter>
                <StreamingSearchResults
                    {...defaultProps}
                    history={history}
                    location={history.location}
                    caseSensitive={true}
                    setCaseSensitivity={setCaseSensitivitySpy}
                />
            </BrowserRouter>
        )

        sinon.assert.notCalled(setCaseSensitivitySpy)

        element.unmount()
    })

    it('should render progress with data from API', () => {
        const element = mount(
            <BrowserRouter>
                <StreamingSearchResults {...defaultProps} />
            </BrowserRouter>
        )

        const progress = element.find(StreamingProgress)
        expect(progress.prop('progress')).toEqual(streamingSearchResult.progress)

        element.unmount()
    })

    it('should expand and collapse results when event from infobar is triggered', () => {
        const element = mount(
            <BrowserRouter>
                <StreamingSearchResults {...defaultProps} />
            </BrowserRouter>
        )

        let infobar = element.find(SearchResultsInfoBar)
        expect(infobar.prop('allExpanded')).toBe(false)

        act(() => infobar.prop('onExpandAllResultsToggle')())
        element.update()

        infobar = element.find(SearchResultsInfoBar)
        expect(infobar.prop('allExpanded')).toBe(true)

        act(() => infobar.prop('onExpandAllResultsToggle')())
        element.update()

        infobar = element.find(SearchResultsInfoBar)
        expect(infobar.prop('allExpanded')).toBe(false)
    })
})
