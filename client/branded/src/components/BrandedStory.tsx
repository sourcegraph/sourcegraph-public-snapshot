import React, { useLayoutEffect, useState } from 'react'
import { MemoryRouter, MemoryRouterProps } from 'react-router'
import { useDarkMode } from 'storybook-dark-mode'

import { ThemeProps } from '@sourcegraph/shared/src/theme'

import brandedStyles from '../global-styles/index.scss'

import { Tooltip } from './tooltip/Tooltip'

export interface WebStoryProps extends MemoryRouterProps {
    children: React.FunctionComponent<ThemeProps>
}

// Prepend global CSS styles to document head to keep them before CSS modules
export function prependCSSToDocumentHead(css: string): HTMLStyleElement {
    const styleTag = document.createElement('style')
    styleTag.textContent = css
    document.head.prepend(styleTag)

    return styleTag
}

/**
 * Wrapper component for webapp Storybook stories that provides light theme and react-router props.
 * Takes a render function as children that gets called with the props.
 */
export const BrandedStory: React.FunctionComponent<
    WebStoryProps & {
        styles?: string
    }
> = ({ children: Children, styles = brandedStyles, ...memoryRouterProps }) => {
    const [isLightTheme, setIsLightTheme] = useState(!useDarkMode())

    useLayoutEffect(() => {
        const styleTag = prependCSSToDocumentHead(styles)

        return () => {
            styleTag.remove()
        }
    }, [styles])

    useLayoutEffect(() => {
        const listener = ((event: CustomEvent<boolean>): void => {
            setIsLightTheme(event.detail)
        }) as EventListener
        window.addEventListener('theme-changed', listener)
        return () => window.removeEventListener('theme-changed', listener)
    }, [])

    return (
        <MemoryRouter {...memoryRouterProps}>
            <Tooltip />
            <Children isLightTheme={isLightTheme} />
        </MemoryRouter>
    )
}
