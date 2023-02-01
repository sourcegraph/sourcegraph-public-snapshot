import { ReactNode, useEffect, useLayoutEffect, useRef } from 'react'

import { RenderResult, render } from '@testing-library/react'
import { MemoryHistory, createMemoryHistory } from 'history'
import * as H from 'history'
import { Router } from 'react-router-dom'
import { CompatRouter, useLocation } from 'react-router-dom-v5-compat'

import { WildcardThemeContext, WildcardTheme } from '../hooks/useWildcardTheme'

export interface RenderWithBrandedContextResult extends RenderResult {
    history: MemoryHistory
}

interface RenderWithBrandedContextOptions {
    route?: string
    history?: MemoryHistory<unknown>
    onLocationChange?: (location: H.Location) => void
}

const wildcardTheme: WildcardTheme = {
    isBranded: true,
}

export function renderWithBrandedContext(
    children: ReactNode,
    {
        route = '/',
        history = createMemoryHistory({ initialEntries: [route] }),
        onLocationChange = (_location: H.Location) => {},
    }: RenderWithBrandedContextOptions = {}
): RenderWithBrandedContextResult {
    return {
        ...render(
            <WildcardThemeContext.Provider value={wildcardTheme}>
                <Router history={history}>
                    <CompatRouter>
                        {children}
                        <ExtractCurrentPathname onLocationChange={onLocationChange} />
                    </CompatRouter>
                </Router>
            </WildcardThemeContext.Provider>
        ),
        history,
    }
}

function ExtractCurrentPathname({ onLocationChange }: { onLocationChange: (location: H.Location) => void }): null {
    const onLocationChangeRef = useRef(onLocationChange)
    useLayoutEffect(() => {
        onLocationChangeRef.current = onLocationChange
    }, [onLocationChange])
    const location = useLocation()
    useEffect(() => {
        onLocationChangeRef.current(location)
    }, [location, onLocationChange])
    return null
}
