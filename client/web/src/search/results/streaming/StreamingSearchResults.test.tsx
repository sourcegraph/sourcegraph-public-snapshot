import { mount } from 'enzyme'
import { createBrowserHistory } from 'history'
import React from 'react'
import { act } from 'react-dom/test-utils'
import { BrowserRouter } from 'react-router-dom'
import { NEVER, of } from 'rxjs'
import sinon from 'sinon'
import { FileMatch } from '../../../../../shared/src/components/FileMatch'
import { VirtualList } from '../../../../../shared/src/components/VirtualList'
import { SearchPatternType } from '../../../../../shared/src/graphql-operations'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { NOOP_TELEMETRY_SERVICE } from '../../../../../shared/src/telemetry/telemetryService'
import { SearchResult } from '../../../components/SearchResult'
import { SavedSearchModal } from '../../../savedSearches/SavedSearchModal'
import * as helpers from '../../helpers'
import { AggregateStreamingSearchResults } from '../../stream'
import { SearchResultsInfoBar } from '../SearchResultsInfoBar'
import { VersionContextWarning } from '../VersionContextWarning'
import { StreamingProgress } from './progress/StreamingProgress'
import { StreamingSearchResults, StreamingSearchResultsProps } from './StreamingSearchResults'
import {
    extensionsController,
    HIGHLIGHTED_FILE_LINES_REQUEST,
    MULTIPLE_SEARCH_RESULT,
    REPO_MATCH_RESULT,
    RESULT,
} from '../../../../../shared/src/util/searchTestHelpers'

describe('StreamingSearchResults', () => {
    const history = createBrowserHistory()

    const streamingSearchResult: AggregateStreamingSearchResults = {
        state: 'complete',
        results: MULTIPLE_SEARCH_RESULT.results,
        filters: MULTIPLE_SEARCH_RESULT.dynamicFilters,
        progress: {
            durationMs: 500,
            matchCount: MULTIPLE_SEARCH_RESULT.matchCount,
            skipped: [],
        },
    }

    const defaultProps: StreamingSearchResultsProps = {
        parsedSearchQuery: 'r:golang/oauth2 test f:travis',
        caseSensitive: false,
        patternType: SearchPatternType.literal,
        versionContext: undefined,
        selectedSearchContextSpec: 'global',
        availableVersionContexts: [],
        previousVersionContext: null,

        extensionsController,
        telemetryService: NOOP_TELEMETRY_SERVICE,

        history,
        location: history.location,
        authenticatedUser: null,

        navbarSearchQueryState: { query: '' },

        settingsCascade: {
            subjects: null,
            final: null,
        },
        platformContext: { forceUpdateTooltip: sinon.spy(), settings: NEVER },

        streamSearch: () => of(streamingSearchResult),

        fetchHighlightedFileLineRanges: HIGHLIGHTED_FILE_LINES_REQUEST,
        isLightTheme: true,
        enableCodeMonitoring: false,
    }

    it('should call streaming search API with the right parameters from URL', () => {
        const searchSpy = sinon.spy(defaultProps.streamSearch)

        const element = mount(
            <BrowserRouter>
                <StreamingSearchResults
                    {...defaultProps}
                    parsedSearchQuery="r:golang/oauth2 test f:travis"
                    patternType={SearchPatternType.regexp}
                    caseSensitive={true}
                    versionContext="test"
                    streamSearch={searchSpy}
                    availableVersionContexts={[{ name: 'test', revisions: [] }]}
                />
            </BrowserRouter>
        )

        sinon.assert.calledOnce(searchSpy)
        sinon.assert.calledWith(searchSpy, {
            query: 'r:golang/oauth2 test f:travis',
            version: 'V2',
            patternType: SearchPatternType.regexp,
            caseSensitive: true,
            versionContext: 'test',
            searchContextSpec: 'global',
            trace: undefined,
        })

        element.unmount()
    })

    it('should call streaming search API with no version context if parameter is invalid', () => {
        const history = createBrowserHistory()
        history.replace({ search: 'q=r:golang/oauth2+test+f:travis+case:yes&patternType=regexp&c=test' })

        const searchSpy = sinon.spy(defaultProps.streamSearch)

        const element = mount(
            <BrowserRouter>
                <StreamingSearchResults
                    {...defaultProps}
                    parsedSearchQuery="r:golang/oauth2 test f:travis"
                    patternType={SearchPatternType.regexp}
                    caseSensitive={false}
                    versionContext="test"
                    streamSearch={searchSpy}
                    availableVersionContexts={[{ name: 'something', revisions: [] }]}
                />
            </BrowserRouter>
        )

        sinon.assert.calledOnce(searchSpy)
        sinon.assert.calledWith(searchSpy, {
            query: 'r:golang/oauth2 test f:travis',
            version: 'V2',
            patternType: SearchPatternType.regexp,
            caseSensitive: false,
            versionContext: undefined,
            searchContextSpec: 'global',
            trace: undefined,
        })

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
        let results = element.find(FileMatch)
        expect(results.map(result => result.prop('allExpanded'))).not.toContain(true)

        act(() => infobar.prop('onExpandAllResultsToggle')())
        element.update()

        infobar = element.find(SearchResultsInfoBar)
        expect(infobar.prop('allExpanded')).toBe(true)
        results = element.find(FileMatch)
        expect(results.map(result => result.prop('allExpanded'))).not.toContain(false)

        act(() => infobar.prop('onExpandAllResultsToggle')())
        element.update()

        infobar = element.find(SearchResultsInfoBar)
        expect(infobar.prop('allExpanded')).toBe(false)
        results = element.find(FileMatch)
        expect(results.map(result => result.prop('allExpanded'))).not.toContain(true)

        element.unmount()
    })

    it('should show version context warning if version context has changed from URL', () => {
        const history = createBrowserHistory()
        history.replace({ search: 'q=r:golang/oauth2+test+f:travis&c=test' })

        const element = mount(
            <BrowserRouter>
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
            </BrowserRouter>
        )

        const warning = element.find(VersionContextWarning)
        expect(warning.length).toBe(1)

        element.unmount()
    })

    it('should not show version context warning if version context has changed from dropdown', () => {
        const history = createBrowserHistory()
        history.replace({ search: 'q=r:golang/oauth2+test+f:travis&c=test&from-context-toggle=true' })

        const element = mount(
            <BrowserRouter>
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
            </BrowserRouter>
        )

        const warning = element.find(VersionContextWarning)
        expect(warning.length).toBe(0)

        element.unmount()
    })

    it('should render correct components for file match and repository match', () => {
        const results: AggregateStreamingSearchResults = {
            ...streamingSearchResult,
            results: [RESULT, REPO_MATCH_RESULT] as GQL.SearchResult[],
        }
        const element = mount(
            <BrowserRouter>
                <StreamingSearchResults {...defaultProps} streamSearch={() => of(results)} />
            </BrowserRouter>
        )

        const listComponent = element.find<VirtualList<SearchResult>>(VirtualList)
        const renderedResultsList = listComponent.prop('items')
        expect(renderedResultsList.length).toBe(2)
        expect(listComponent.prop('renderItem')(renderedResultsList[0], undefined).type).toBe(FileMatch)
        expect(listComponent.prop('renderItem')(renderedResultsList[1], undefined).type).toBe(SearchResult)

        element.unmount()
    })

    it('should log view, query, and results fetched events', () => {
        const logSpy = sinon.spy()
        const logViewEventSpy = sinon.spy()
        const telemetryService = {
            ...NOOP_TELEMETRY_SERVICE,
            log: logSpy,
            logViewEvent: logViewEventSpy,
        }

        const element = mount(
            <BrowserRouter>
                <StreamingSearchResults {...defaultProps} telemetryService={telemetryService} />
            </BrowserRouter>
        )

        sinon.assert.calledOnceWithExactly(logViewEventSpy, 'SearchResults')
        sinon.assert.calledWith(logSpy, 'SearchResultsQueried')
        sinon.assert.calledWith(logSpy, 'SearchResultsFetched')

        element.unmount()
    })

    it('should log event when clicking on search result', () => {
        const logSpy = sinon.spy()
        const telemetryService = {
            ...NOOP_TELEMETRY_SERVICE,
            log: logSpy,
        }

        const element = mount(
            <BrowserRouter>
                <StreamingSearchResults {...defaultProps} telemetryService={telemetryService} />
            </BrowserRouter>
        )

        const item = element.find(FileMatch).first()
        act(() => item.prop('onSelect')())

        sinon.assert.calledWith(logSpy, 'SearchResultClicked')

        element.unmount()
    })

    it('should not show saved search modal on first load', () => {
        const element = mount(
            <BrowserRouter>
                <StreamingSearchResults {...defaultProps} />
            </BrowserRouter>
        )

        const modal = element.find(SavedSearchModal)
        expect(modal.length).toBe(0)
    })

    it('should open saved search modal when triggering event from infobar', () => {
        const element = mount(
            <BrowserRouter>
                <StreamingSearchResults {...defaultProps} />
            </BrowserRouter>
        )

        const infobar = element.find(SearchResultsInfoBar)
        act(() => infobar.prop('onSaveQueryClick')())
        element.update()

        const modal = element.find(SavedSearchModal)
        expect(modal.length).toBe(1)
    })

    it('should close saved search modal if close event triggers', () => {
        const element = mount(
            <BrowserRouter>
                <StreamingSearchResults {...defaultProps} />
            </BrowserRouter>
        )

        const infobar = element.find(SearchResultsInfoBar)
        act(() => infobar.prop('onSaveQueryClick')())
        element.update()

        let modal = element.find(SavedSearchModal)
        act(() => modal.prop('onDidCancel')())
        element.update()

        modal = element.find(SavedSearchModal)
        expect(modal.length).toBe(0)
    })

    it('should start a new search with added params when onSearchAgain event in triggered', () => {
        const submitSearchMock = jest.spyOn(helpers, 'submitSearch').mockImplementation(() => {})
        const element = mount(
            <BrowserRouter>
                <StreamingSearchResults {...defaultProps} />
            </BrowserRouter>
        )

        const progress = element.find(StreamingProgress)
        act(() => progress.prop('onSearchAgain')(['archived:yes', 'timeout:2m']))
        element.update()

        expect(helpers.submitSearch).toBeCalledTimes(1)
        const args = submitSearchMock.mock.calls[0][0]
        expect(args.query).toBe('r:golang/oauth2 test f:travis archived:yes timeout:2m')
    })
})
