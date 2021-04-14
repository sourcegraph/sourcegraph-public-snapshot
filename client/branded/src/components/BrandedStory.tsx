import React, { useEffect } from 'react'
import { MemoryRouter, MemoryRouterProps } from 'react-router'
import { useDarkMode } from 'storybook-dark-mode'

import { ThemeProps } from '@sourcegraph/shared/src/theme'

import brandedStyles from '../global-styles/index.scss'

import { Tooltip } from './tooltip/Tooltip'

export interface WebStoryProps extends MemoryRouterProps {
    children: React.FunctionComponent<ThemeProps>
}

// Prepend global CSS styles to document head to keep them before CSS modules
function prependCSSToDocumentHead(css: string): HTMLStyleElement {
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
    const isLightTheme = !useDarkMode()

    useEffect(() => {
        const styleTag = prependCSSToDocumentHead(styles)

        return () => {
            styleTag.remove()
        }
    }, [styles])

    return (
        <MemoryRouter {...memoryRouterProps}>
            <Tooltip />
            <Children isLightTheme={isLightTheme} />
        </MemoryRouter>
    )
}
