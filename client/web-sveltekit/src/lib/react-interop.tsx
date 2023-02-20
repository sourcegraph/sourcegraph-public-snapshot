import React, { type FC, type PropsWithChildren, type ReactElement } from 'react'
import { WildcardThemeContext, type WildcardTheme } from './wildcard'
import { Router } from 'react-router'
import { RouterLink, setLinkComponent } from 'wildcard/src'
import { createRoot, type Root } from 'react-dom/client'
import { CompatRouter, Routes, Route } from 'react-router-dom-v5-compat'
import { createBrowserHistory } from 'history'
import { onDestroy } from 'svelte'

setLinkComponent(RouterLink)

const WILDCARD_THEME: WildcardTheme = {
    isBranded: true,
}

/**
 * Creates a minimal context for rendering React components inside Svelte.
 * Notes:
 * - createBrowserHistory needs to be called every time a component is mounted, otherwise
 *   the history object doesn't seem to know about the latest path.
 */
export const ReactAdapter: FC<PropsWithChildren<{ route: string }>> = ({ route, children }) => (
    <WildcardThemeContext.Provider value={WILDCARD_THEME}>
        <Router history={createBrowserHistory()}>
            <CompatRouter>
                <Routes>
                    <Route path={route} element={<React.Suspense fallback={true}>{children}</React.Suspense>} />
                </Routes>
            </CompatRouter>
        </Router>
    </WildcardThemeContext.Provider>
)
