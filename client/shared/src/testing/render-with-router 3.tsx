import { RenderResult, render } from '@testing-library/react'
import { MemoryHistory, createMemoryHistory } from 'history'
import React, { ReactNode } from 'react'
import { Router } from 'react-router-dom'

export interface RenderWithRouterResult extends RenderResult {
    history: MemoryHistory
}

export function renderWithRouter(
    children: ReactNode,
    { route = '/', history = createMemoryHistory({ initialEntries: [route] }) } = {}
): RenderWithRouterResult {
    return {
        ...render(<Router history={history}>{children}</Router>),
        history,
    }
}
