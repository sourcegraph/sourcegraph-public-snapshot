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
    React.PropsWithChildren<{ history: History; onMount?: () => void; onUnmount?: () => void; onRender?: () => void }>
> = ({ history, onMount, onUnmount, onRender, children }) => {
    useEffect(() => onRender?.())
    // This should only be called once when the component is mounted
    // eslint-disable-next-line react-hooks/exhaustive-deps
    useEffect(() => {
        onMount?.()
        console.log('MOUNT')
        return () => {
            console.log('UNMOUNT')
            onUnmount?.()
        }
    }, [])

    return (
        <WildcardThemeContext.Provider value={{ isBranded: true }}>
            <Router history={history}>
                <CompatRouter>{children}</CompatRouter>
            </Router>
        </WildcardThemeContext.Provider>
    )
}
