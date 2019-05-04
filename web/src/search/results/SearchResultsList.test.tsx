import { createBrowserHistory } from 'history'
import * as React from 'react'
import { BrowserRouter } from 'react-router-dom'
import { cleanup, getAllByTestId, getByTestId, queryByTestId, render } from 'react-testing-library'
import sinon from 'sinon'
import { setLinkComponent } from '../../../../shared/src/components/Link'
import { NOOP_TELEMETRY_SERVICE } from '../../../../shared/src/telemetry/telemetryService'
import {
    extensionsController,
    HIGHLIGHTED_FILE_LINES_REQUEST,
    MULTIPLE_SEARCH_REQUEST,
    SEARCH_REQUEST,
} from '../testHelpers'
import { SearchResultsList, SearchResultsListProps } from './SearchResultsList'

describe('SearchResultsList', () => {
    setLinkComponent((props: any) => <a {...props} />)

    afterAll(() => {
        setLinkComponent(null as any)
        cleanup()
    })

    const history = createBrowserHistory()
    history.replace({ search: 'q=r:golang/oauth2+test+f:travis' })

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
        onDidCreateSavedQuery: sinon.spy(),
        onSaveQueryClick: sinon.spy(),
        didSave: false,

        fetchHighlightedFileLines: HIGHLIGHTED_FILE_LINES_REQUEST,

        isLightTheme: true,
        settingsCascade: {
            subjects: null,
            final: null,
        },
        extensionsController: { executeCommand: sinon.spy(), services: extensionsController.services },
        platformContext: { forceUpdateTooltip: sinon.spy() },
        telemetryService: NOOP_TELEMETRY_SERVICE,
    }

    it('displays loading text when results is undefined', () => {
        const { container } = render(
            <BrowserRouter>
                <SearchResultsList {...defaultProps} resultsOrError={undefined} />
            </BrowserRouter>
        )

        expect(queryByTestId(container, 'loading-container')).toBeTruthy()
    })

    it('shows error message when the search GraphQL request returns an error', () => {
        const { container } = render(
            <BrowserRouter>
                <SearchResultsList {...defaultProps} resultsOrError={{ message: 'test error', code: 'error' }} />
            </BrowserRouter>
        )
        expect(getByTestId(container, 'search-results-list-error')).toBeTruthy()
    })

    it('renders the search results info bar when there are results', () => {
        const { container } = render(
            <BrowserRouter>
                <SearchResultsList {...defaultProps} />
            </BrowserRouter>
        )
        expect(getByTestId(container, 'results-info-bar')).toBeTruthy()
    })

    it('renders one search result', () => {
        const { container } = render(
            <BrowserRouter>
                <SearchResultsList {...defaultProps} />
            </BrowserRouter>
        )
        expect(getByTestId(container, 'result-container')).toBeTruthy()
        expect(getAllByTestId(container, 'result-container').length).toBe(1)
    })

    it('does not display the loading indicator or error message if there are results', () => {
        const { container } = render(
            <BrowserRouter>
                <SearchResultsList {...defaultProps} />
            </BrowserRouter>
        )
        expect(queryByTestId(container, 'loading-container')).toBeFalsy()
        expect(queryByTestId(container, 'search-results-list-error')).toBeFalsy()
    })

    it('renders correct number of search results if there are multiple', () => {
        const props = { ...defaultProps, resultsOrError: MULTIPLE_SEARCH_REQUEST() }
        const { container } = render(
            <BrowserRouter>
                <SearchResultsList {...props} />
            </BrowserRouter>
        )
        expect(getByTestId(container, 'result-container')).toBeTruthy()
        expect(getAllByTestId(container, 'result-container').length).toBe(3)
    })
})
