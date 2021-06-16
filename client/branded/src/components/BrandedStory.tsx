import React, { useLayoutEffect, useState } from 'react'
import { MemoryRouter, MemoryRouterProps } from 'react-router'
import { useDarkMode } from 'storybook-dark-mode'

import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { useStyles } from '@sourcegraph/storybook/src/hooks/useStyles'

import brandedStyles from '../global-styles/index.scss'

import { Tooltip } from './tooltip/Tooltip'

export interface WebStoryProps extends MemoryRouterProps {
    children: React.FunctionComponent<ThemeProps>
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
    useStyles(styles)

    useLayoutEffect(() => {
        const listener = ((event: CustomEvent<boolean>): void => {
            setIsLightTheme(event.detail)
        }) as EventListener
        document.body.addEventListener('chromatic-light-theme-toggled', listener)
        return () => document.body.removeEventListener('chromatic-light-theme-toggled', listener)
    }, [])

    return (
        <MemoryRouter {...memoryRouterProps}>
            <Tooltip />
            <Children isLightTheme={isLightTheme} />
        </MemoryRouter>
    )
}
