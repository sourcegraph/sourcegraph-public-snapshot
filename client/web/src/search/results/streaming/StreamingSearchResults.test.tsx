import { mount } from 'enzyme'
import { createBrowserHistory } from 'history'
import React from 'react'
import { act } from 'react-dom/test-utils'
import { BrowserRouter } from 'react-router-dom'
import { NEVER, of } from 'rxjs'
import sinon from 'sinon'
import { FileMatch } from '../../../../../shared/src/components/FileMatch'
import { SearchPatternType } from '../../../../../shared/src/graphql-operations'
import { NOOP_TELEMETRY_SERVICE } from '../../../../../shared/src/telemetry/telemetryService'
import * as GQL from '../../../../../shared/src/graphql/schema'
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
import { VirtualList } from '../../../../../shared/src/components/VirtualList'
import { SearchResult } from '../../../components/SearchResult'

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
        setVersionContext: sinon.spy(),
        availableVersionContexts: [],
        previousVersionContext: null,

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

        fetchHighlightedFileLines: HIGHLIGHTED_FILE_LINES_REQUEST,
        isLightTheme: true,
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
                    availableVersionContexts={[{ name: 'test', revisions: [] }]}
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

    it('should call streaming search API with no version context if parameter is invalid', () => {
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
                    availableVersionContexts={[{ name: 'something', revisions: [] }]}
                />
            </BrowserRouter>
        )

        sinon.assert.calledOnce(searchSpy)
        sinon.assert.calledWith(
            searchSpy,
            'r:golang/oauth2 test f:travis  case:yes',
            'V2',
            SearchPatternType.regexp,
            undefined
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

    it('should update versionContext if different between URL and context', () => {
        const history = createBrowserHistory()
        history.replace({ search: 'q=r:golang/oauth2+test+f:travis&c=test' })

        const setVersionContextSpy = sinon.spy()

        const element = mount(
            <BrowserRouter>
                <StreamingSearchResults
                    {...defaultProps}
                    history={history}
                    location={history.location}
                    versionContext="other"
                    setVersionContext={setVersionContextSpy}
                    availableVersionContexts={[
                        { name: 'test', revisions: [] },
                        { name: 'other', revisions: [] },
                    ]}
                />
            </BrowserRouter>
        )

        sinon.assert.calledOnce(setVersionContextSpy)
        sinon.assert.calledWith(setVersionContextSpy, 'test')

        element.unmount()
    })

    it('should not update versionContext if same between URL and context', () => {
        const history = createBrowserHistory()
        history.replace({ search: 'q=r:golang/oauth2+test+f:travis&c=test' })

        const setVersionContextSpy = sinon.spy()

        const element = mount(
            <BrowserRouter>
                <StreamingSearchResults
                    {...defaultProps}
                    history={history}
                    location={history.location}
                    versionContext="test"
                    setVersionContext={setVersionContextSpy}
                    availableVersionContexts={[
                        { name: 'test', revisions: [] },
                        { name: 'other', revisions: [] },
                    ]}
                />
            </BrowserRouter>
        )

        sinon.assert.notCalled(setVersionContextSpy)

        element.unmount()
    })

    it('should clear versionContext if updating to clear', () => {
        const history = createBrowserHistory()
        history.replace({ search: 'q=r:golang/oauth2+test+f:travis' })

        const setVersionContextSpy = sinon.spy()

        const element = mount(
            <BrowserRouter>
                <StreamingSearchResults
                    {...defaultProps}
                    history={history}
                    location={history.location}
                    versionContext="test"
                    setVersionContext={setVersionContextSpy}
                    availableVersionContexts={[
                        { name: 'test', revisions: [] },
                        { name: 'other', revisions: [] },
                    ]}
                />
            </BrowserRouter>
        )

        sinon.assert.calledOnce(setVersionContextSpy)
        sinon.assert.calledWith(setVersionContextSpy, undefined)

        element.unmount()
    })

    it('should clear versionContext if updating to invalid one', () => {
        const history = createBrowserHistory()
        history.replace({ search: 'q=r:golang/oauth2+test+f:travis&c=test' })

        const setVersionContextSpy = sinon.spy()

        const element = mount(
            <BrowserRouter>
                <StreamingSearchResults
                    {...defaultProps}
                    history={history}
                    location={history.location}
                    versionContext="other"
                    setVersionContext={setVersionContextSpy}
                    availableVersionContexts={[{ name: 'other', revisions: [] }]}
                />
            </BrowserRouter>
        )

        sinon.assert.calledOnce(setVersionContextSpy)
        sinon.assert.calledWith(setVersionContextSpy, undefined)

        element.unmount()
    })

    it('should retain clear versionContext if updating to invalid one', () => {
        const history = createBrowserHistory()
        history.replace({ search: 'q=r:golang/oauth2+test+f:travis&c=test' })

        const setVersionContextSpy = sinon.spy()

        const element = mount(
            <BrowserRouter>
                <StreamingSearchResults
                    {...defaultProps}
                    history={history}
                    location={history.location}
                    versionContext={undefined}
                    setVersionContext={setVersionContextSpy}
                    availableVersionContexts={[{ name: 'other', revisions: [] }]}
                />
            </BrowserRouter>
        )

        sinon.assert.notCalled(setVersionContextSpy)

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

        const renderedResultsList = element.find(VirtualList).prop('items')
        expect(renderedResultsList.length).toBe(2)
        expect(renderedResultsList[0].type).toBe(FileMatch)
        expect(renderedResultsList[1].type).toBe(SearchResult)

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

        sinon.assert.calledOnce(logSpy)
        sinon.assert.calledWith(logSpy, 'SearchResultClicked')

        element.unmount()
    })
})
