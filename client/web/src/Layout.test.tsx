import { render } from '@testing-library/react'
import { createBrowserHistory } from 'history'
import { BrowserRouter } from 'react-router-dom'
import { CompatRouter } from 'react-router-dom-v5-compat'
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
        patternType: SearchPatternType.standard,
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
        platformContext: { settings: NEVER },
    } as unknown) as LayoutProps

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
                        <Layout {...defaultProps} history={history} location={history.location} />
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
                        <Layout {...defaultProps} history={history} location={history.location} />
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
                        <Layout {...defaultProps} history={history} location={history.location} />
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
                        <Layout {...defaultProps} history={history} location={history.location} />
                    </CompatRouter>
                </BrowserRouter>
            </MockedTestProvider>
        )

        expect(useNavbarQueryState.getState().searchCaseSensitivity).toBe(false)
    })
})
