import React, { useEffect, useMemo } from 'react'

import { observeSystemIsLightTheme, ThemeProps } from '@sourcegraph/shared/src/theme'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

/**
 * Wrapper for the browser extension that listens to changes of the OS theme.
 */
export function ThemeWrapper({
    children,
}: {
    children: JSX.Element | null | ((props: ThemeProps) => JSX.Element | null)
}): JSX.Element | null {
    const isLightTheme = useObservable(useMemo(() => observeSystemIsLightTheme(), []))

    useEffect(() => {
        if (isLightTheme !== undefined) {
            document.body.classList.toggle('theme-light', isLightTheme)
            document.body.classList.toggle('theme-dark', !isLightTheme)
        }
    }, [isLightTheme])

    if (isLightTheme === undefined) {
        return null
    }

    if (typeof children === 'function') {
        const Children = children
        return <Children isLightTheme={isLightTheme} />
    }
    return children
}
