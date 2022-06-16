import { ReactNode } from 'react'

import { RenderResult, render } from '@testing-library/react'
import { MemoryHistory, createMemoryHistory } from 'history'
import { Router } from 'react-router-dom'

import { WildcardThemeContext, WildcardTheme } from '@sourcegraph/wildcard'

export interface RenderWithBrandedContextResult extends RenderResult {
    history: MemoryHistory
}

const wildcardTheme: WildcardTheme = {
    isBranded: true,
}

export function renderWithBrandedContext(
    children: ReactNode,
    { route = '/', history = createMemoryHistory({ initialEntries: [route] }) } = {}
): RenderWithBrandedContextResult {
    return {
        ...render(
            <WildcardThemeContext.Provider value={wildcardTheme}>
                <Router history={history}>{children}</Router>
            </WildcardThemeContext.Provider>
        ),
        history,
    }
}
