import React from 'react'
import ReactDOMServer from 'react-dom/server'

import { WildcardThemeContext } from '../hooks/useWildcardTheme'

/**
 * Helper used to render components to string with branded styles.
 * Required to render Wildcard components as `text` field with `shepherd.js` library.
 * See related issue here: https://github.com/shipshapecode/react-shepherd/issues/612.
 */
export function renderBrandedToString(children: React.ReactElement): string {
    return ReactDOMServer.renderToString(
        <WildcardThemeContext.Provider value={{ isBranded: true }}>{children}</WildcardThemeContext.Provider>
    )
}
