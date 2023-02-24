import React, { type FC, type PropsWithChildren } from 'react'

import { RouterProvider, createBrowserRouter } from 'react-router-dom'
import { RouterLink, setLinkComponent } from 'wildcard/src'

import { WildcardThemeContext, type WildcardTheme } from './wildcard'

setLinkComponent(RouterLink)

const WILDCARD_THEME: WildcardTheme = {
    isBranded: true,
}

/**
 * Creates a minimal context for rendering React components inside Svelte.
 */
export const ReactAdapter: FC<PropsWithChildren<{ route: string }>> = ({ route, children }) => (
    <WildcardThemeContext.Provider value={WILDCARD_THEME}>
        <RouterProvider
            router={createBrowserRouter([
                {
                    path: route,
                    element: <React.Suspense fallback={true}>{children}</React.Suspense>,
                },
            ])}
        />
    </WildcardThemeContext.Provider>
)
