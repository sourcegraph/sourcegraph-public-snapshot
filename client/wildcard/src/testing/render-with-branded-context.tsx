import { ReactNode } from 'react'

import { RenderResult, render } from '@testing-library/react'
import { MemoryHistory, createMemoryHistory } from 'history'
import { Router } from 'react-router-dom'
import { CompatRouter } from 'react-router-dom-v5-compat'

import { WildcardThemeContext, WildcardTheme } from '../hooks/useWildcardTheme'

export interface RenderWithBrandedContextResult extends RenderResult {
    history: MemoryHistory
}

interface RenderWithBrandedContextOptions {
    route?: string
    history?: MemoryHistory<unknown>
}

const wildcardTheme: WildcardTheme = {
    isBranded: true,
}

export function renderWithBrandedContext(
    children: ReactNode,
    { route = '/', history = createMemoryHistory({ initialEntries: [route] }) }: RenderWithBrandedContextOptions = {}
): RenderWithBrandedContextResult {
    return {
        ...render(
            <WildcardThemeContext.Provider value={wildcardTheme}>
                <Router history={history}>
                    <CompatRouter>{children}</CompatRouter>
                </Router>
            </WildcardThemeContext.Provider>
        ),
        history,
    }
}
