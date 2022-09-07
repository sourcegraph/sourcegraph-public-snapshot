import { render } from '@testing-library/react'
import { createBrowserHistory } from 'history'
import { BrowserRouter } from 'react-router-dom'
import { CompatRouter } from 'react-router-dom-v5-compat'
import { NEVER } from 'rxjs'
import sinon from 'sinon'

import { SearchPatternType } from '@sourcegraph/search'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { useNavbarQueryState } from '../../stores'

import { SearchNavbarItem, SearchNavbarItemProps } from './SearchNavbarItem'

describe('SearchNavbarItem', () => {
    const defaultProps: Omit<SearchNavbarItemProps, 'history' | 'location'> = {
        // Parsed query components
        // Other minimum props required to render
        settingsCascade: {
            subjects: null,
            final: null,
        },
        authenticatedUser: null,
        isSourcegraphDotCom: false,
        globbing: false,
        isLightTheme: true,
        searchContextsEnabled: true,
        defaultSearchContextSpec: '',
        setSelectedSearchContextSpec: () => {},
        fetchAutoDefinedSearchContexts: () => NEVER,
        fetchSearchContexts: () => NEVER,
        getUserSearchContextNamespaces: () => [],
        telemetryService: NOOP_TELEMETRY_SERVICE,
        platformContext: { requestGraphQL: () => NEVER },
    }

    const origContext = window.context
    beforeEach(() => {
        const root = document.createElement('div')
        root.id = 'root'
        document.body.append(root)
        window.context = {
            enableLegacyExtensions: true,
        } as any
    })

    afterEach(() => {
        document.querySelector('#root')?.remove()
        window.context = origContext
    })

    it('should update patternType if different between URL and context', () => {
        const history = createBrowserHistory()
        history.replace({ search: 'q=r:golang/oauth2+test+f:travis&patternType=regexp' })

        useNavbarQueryState.setState({ searchPatternType: SearchPatternType.standard })

        render(
            <MockedTestProvider>
                <BrowserRouter>
                    <CompatRouter>
                        <SearchNavbarItem {...defaultProps} history={history} location={history.location} />
                    </CompatRouter>
                </BrowserRouter>
            </MockedTestProvider>
        )

        expect(useNavbarQueryState.getState().searchPatternType).toBe(SearchPatternType.regexp)
    })

    it('should not update patternType if query is empty', () => {
        const history = createBrowserHistory()
        history.replace({ search: 'q=&patternType=regexp' })

        useNavbarQueryState.setState({ searchPatternType: SearchPatternType.standard })

        render(
            <MockedTestProvider>
                <BrowserRouter>
                    <CompatRouter>
                        <SearchNavbarItem {...defaultProps} history={history} location={history.location} />
                    </CompatRouter>
                </BrowserRouter>
            </MockedTestProvider>
        )

        expect(useNavbarQueryState.getState().searchPatternType).toBe(SearchPatternType.standard)
    })

    it('should update caseSensitive if different between URL and context', () => {
        const history = createBrowserHistory()
        history.replace({ search: 'q=r:golang/oauth2+test+f:travis case:yes' })

        useNavbarQueryState.setState({ searchCaseSensitivity: false })

        render(
            <MockedTestProvider>
                <BrowserRouter>
                    <CompatRouter>
                        <SearchNavbarItem {...defaultProps} history={history} location={history.location} />
                    </CompatRouter>
                </BrowserRouter>
            </MockedTestProvider>
        )

        expect(useNavbarQueryState.getState().searchCaseSensitivity).toBe(true)
    })

    it('should not update caseSensitive if query is empty', () => {
        const history = createBrowserHistory()
        history.replace({ search: 'q=case:yes' })

        useNavbarQueryState.setState({ searchCaseSensitivity: false })

        render(
            <MockedTestProvider>
                <BrowserRouter>
                    <CompatRouter>
                        <SearchNavbarItem {...defaultProps} history={history} location={history.location} />
                    </CompatRouter>
                </BrowserRouter>
            </MockedTestProvider>
        )

        expect(useNavbarQueryState.getState().searchCaseSensitivity).toBe(false)
    })

    it('should update search context if different from the currently selected context', () => {
        const history = createBrowserHistory()
        history.replace({ search: 'q=context:me+test' })

        const setSearchContext = sinon.spy()

        render(
            <MockedTestProvider>
                <BrowserRouter>
                    <CompatRouter>
                        <SearchNavbarItem
                            {...defaultProps}
                            history={history}
                            location={history.location}
                            setSelectedSearchContextSpec={setSearchContext}
                            selectedSearchContextSpec="global"
                        />
                    </CompatRouter>
                </BrowserRouter>
            </MockedTestProvider>
        )

        sinon.assert.calledOnceWithExactly(setSearchContext, 'me')
    })
})
