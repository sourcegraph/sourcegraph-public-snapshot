import { RenderResult, render } from '@testing-library/react'
import { MemoryHistory, createMemoryHistory } from 'history'
import React, { ReactNode } from 'react'
import { Router } from 'react-router-dom'

import { WildcardThemeContext, WildcardTheme } from '@sourcegraph/wildcard'

interface RenderWithBrandedContextOptions {
    route?: string
    history?: MemoryHistory
}

export interface RenderWithBrandedContextResult extends RenderResult {
    history: MemoryHistory
}

const wildcardTheme: WildcardTheme = {
    isBranded: true,
}

export function renderWithBrandedContext(
    children: ReactNode,
    options: RenderWithBrandedContextOptions = {}
): RenderWithBrandedContextResult {
    const { route = '/', history = createMemoryHistory({ initialEntries: [route] }) } = options

    return {
        ...render(
            <WildcardThemeContext.Provider value={wildcardTheme}>
                <Router history={history}>{children}</Router>
            </WildcardThemeContext.Provider>
        ),
        history,
    }
}
