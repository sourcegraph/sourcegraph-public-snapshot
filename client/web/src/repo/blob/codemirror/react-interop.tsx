import React, { useEffect } from 'react'

import { History } from 'history'
import { Router } from 'react-router'
import { CompatRouter } from 'react-router-dom-v5-compat'

import { WildcardThemeContext } from '@sourcegraph/wildcard'

/**
 * Creates the necessary context for React components to be rendered inside
 * CodeMirror.
 */
export const Container: React.FunctionComponent<
    React.PropsWithChildren<{ history: History; onRender?: () => void }>
> = ({ history, onRender, children }) => {
    useEffect(() => onRender?.())

    return (
        <WildcardThemeContext.Provider value={{ isBranded: true }}>
            <Router history={history}>
                <CompatRouter>{children}</CompatRouter>
            </Router>
        </WildcardThemeContext.Provider>
    )
}
