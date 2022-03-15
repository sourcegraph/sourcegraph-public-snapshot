import React from 'react'

import ReactDOMServer from 'react-dom/server'

import { WildcardThemeContext, WildcardTheme } from '@sourcegraph/wildcard'

const wildcardTheme: WildcardTheme = {
    isBranded: true,
}

/**
 * Helper used to render components to string with branded styles.
 * Required to render Wildcard components as `text` field with `shepherd.js` library.
 * See related issue here: https://github.com/shipshapecode/react-shepherd/issues/612.
 */
export function renderBrandedToString(children: React.ReactNode): string {
    return ReactDOMServer.renderToString(
        <WildcardThemeContext.Provider value={wildcardTheme}>{children}</WildcardThemeContext.Provider>
    )
}
