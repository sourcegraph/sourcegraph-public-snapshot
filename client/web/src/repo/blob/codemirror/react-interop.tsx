import React, { useEffect } from 'react'

import { History } from 'history'
import { Router } from 'react-router'
import { CompatRouter } from 'react-router-dom-v5-compat'

import { WildcardThemeContext } from '@sourcegraph/wildcard'

import { globalHistory } from '../../../util/globalHistory'

interface CodeMirrorContainerProps {
    history: History
    onMount?: () => void
    onRender?: () => void
}

/**
 * Creates the necessary context for React components to be rendered inside
 * CodeMirror.
 */
export const CodeMirrorContainer: React.FunctionComponent<React.PropsWithChildren<CodeMirrorContainerProps>> = ({
    onMount,
    onRender,
    children,
}) => {
    useEffect(() => onRender?.())
    // This should only be called once when the component is mounted
    // eslint-disable-next-line react-hooks/exhaustive-deps
    useEffect(() => onMount?.(), [])

    return (
        <WildcardThemeContext.Provider value={{ isBranded: true }}>
            <Router history={globalHistory}>
                <CompatRouter>{children}</CompatRouter>
            </Router>
        </WildcardThemeContext.Provider>
    )
}
