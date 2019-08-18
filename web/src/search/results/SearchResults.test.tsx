import { createBrowserHistory } from 'history'
import React from 'react'
import { BrowserRouter } from 'react-router-dom'
import { cleanup, getAllByTestId, getByTestId, render, waitForElement } from 'react-testing-library'
import { noop } from 'rxjs'
import sinon from 'sinon'
import { setLinkComponent } from '../../../../shared/src/components/Link'
import {
    extensionsController,
    HIGHLIGHTED_FILE_LINES_REQUEST,
    NOOP_SETTINGS_CASCADE,
    OBSERVABLE_SEARCH_REQUEST,
} from '../testHelpers'
import { SearchResults, SearchResultsProps } from './SearchResults'

describe('SearchResults', () => {
    setLinkComponent((props: any) => <a {...props} />)

    afterAll(() => {
        setLinkComponent(null as any)
        cleanup()
    })

    const history = createBrowserHistory()
    history.replace({ search: 'q=r:golang/oauth2+test+f:travis' })

    const defaultProps: SearchResultsProps = {
        authenticatedUser: null,
        location: history.location,
        history,
        navbarSearchQuery: '',
        fetchHighlightedFileLines: HIGHLIGHTED_FILE_LINES_REQUEST,
        searchRequest: OBSERVABLE_SEARCH_REQUEST,
        isLightTheme: true,
        settingsCascade: NOOP_SETTINGS_CASCADE,
        extensionsController,
        isSourcegraphDotCom: false,
        platformContext: { forceUpdateTooltip: sinon.spy() },
        telemetryService: { log: noop, logViewEvent: noop },
        deployType: 'dev',
    }

    it('calls the search request once', () => {
        render(
            <BrowserRouter>
                <SearchResults {...defaultProps} />
            </BrowserRouter>
        )
        expect(OBSERVABLE_SEARCH_REQUEST.calledOnce)
    })

    it('displays exactly one filter bar and one repositories filter bar', async () => {
        const { container } = render(
            <BrowserRouter>
                <SearchResults {...defaultProps} />
            </BrowserRouter>
        )
        await waitForElement(() => getByTestId(container, 'filters-bar'))
        expect(getAllByTestId(container, 'filters-bar').length).toBe(1)
        expect(getAllByTestId(container, 'repo-filters-bar').length).toBe(1)
    })

    it('displays correct number of non-repository filters', () => {
        const { container } = render(
            <BrowserRouter>
                <SearchResults {...defaultProps} />
            </BrowserRouter>
        )
        const filtersBar = getByTestId(container, 'filters-bar')
        expect(getAllByTestId(filtersBar, 'filter-chip').length).toBe(2)
    })

    it('displays correct number of repository filters', () => {
        const { container } = render(
            <BrowserRouter>
                <SearchResults {...defaultProps} />
            </BrowserRouter>
        )
        const repoFiltersBar = getByTestId(container, 'repo-filters-bar')
        expect(getAllByTestId(repoFiltersBar, 'filter-chip').length).toBe(1)
    })
})
