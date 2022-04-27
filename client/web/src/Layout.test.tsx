import { render } from '@testing-library/react'
import { createBrowserHistory } from 'history'
import { BrowserRouter } from 'react-router-dom'
import { NEVER } from 'rxjs'

import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { extensionsController } from '@sourcegraph/shared/src/testing/searchTestHelpers'

import { SearchPatternType } from './graphql-operations'
import { Layout, LayoutProps } from './Layout'
import { useNavbarQueryState } from './stores'

jest.mock('./theme', () => ({
    useThemeProps: () => ({
        isLightTheme: true,
        themePreference: 'system',
        onThemePreferenceChange: () => {},
    }),
}))

describe('Layout', () => {
    const defaultProps: LayoutProps = ({
        // Parsed query components
        patternType: SearchPatternType.literal,
        setPatternType: () => {},
        caseSensitive: false,
        setCaseSensitivity: () => {},

        // Other minimum props required to render
        routes: [],
        navbarSearchQueryState: { query: '' },
        onNavbarQueryChange: () => {},
        settingsCascade: {
            subjects: null,
            final: null,
        },
        keyboardShortcuts: [],
        extensionsController,
        platformContext: { forceUpdateTooltip: () => {}, settings: NEVER },
    } as unknown) as LayoutProps

    beforeEach(() => {
        const root = document.createElement('div')
        root.id = 'root'
        document.body.append(root)
    })

    afterEach(() => {
        document.querySelector('#root')?.remove()
    })

    it('should update patternType if different between URL and context', () => {
        const history = createBrowserHistory()
        history.replace({ search: 'q=r:golang/oauth2+test+f:travis&patternType=regexp' })

        useNavbarQueryState.setState({ searchPatternType: SearchPatternType.literal })

        render(
            <MockedTestProvider>
                <BrowserRouter>
                    <Layout {...defaultProps} history={history} location={history.location} />
                </BrowserRouter>
            </MockedTestProvider>
        )

        expect(useNavbarQueryState.getState().searchPatternType).toBe(SearchPatternType.regexp)
    })

    it('should not update patternType if query is empty', () => {
        const history = createBrowserHistory()
        history.replace({ search: 'q=&patternType=regexp' })

        useNavbarQueryState.setState({ searchPatternType: SearchPatternType.literal })

        render(
            <MockedTestProvider>
                <BrowserRouter>
                    <Layout {...defaultProps} history={history} location={history.location} />
                </BrowserRouter>
            </MockedTestProvider>
        )

        expect(useNavbarQueryState.getState().searchPatternType).toBe(SearchPatternType.literal)
    })

    it('should update caseSensitive if different between URL and context', () => {
        const history = createBrowserHistory()
        history.replace({ search: 'q=r:golang/oauth2+test+f:travis case:yes' })

        useNavbarQueryState.setState({ searchCaseSensitivity: false })

        render(
            <MockedTestProvider>
                <BrowserRouter>
                    <Layout {...defaultProps} history={history} location={history.location} />
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
                    <Layout {...defaultProps} history={history} location={history.location} />
                </BrowserRouter>
            </MockedTestProvider>
        )

        expect(useNavbarQueryState.getState().searchCaseSensitivity).toBe(false)
    })
})
