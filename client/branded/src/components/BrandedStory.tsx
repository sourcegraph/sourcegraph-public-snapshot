import React from 'react'
import { MemoryRouter, MemoryRouterProps } from 'react-router'
import { ThemeProps } from '../../../shared/src/theme'
import brandedStyles from '../global-styles/index.scss'
import { Tooltip } from './tooltip/Tooltip'
import { useDarkMode } from 'storybook-dark-mode'

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
> = ({ children, styles = brandedStyles, ...memoryRouterProps }) => {
    const isLightTheme = !useDarkMode()
    const Children = children
    return (
        <MemoryRouter {...memoryRouterProps}>
            <Tooltip />
            <Children isLightTheme={isLightTheme} />
            <style title="Webapp CSS">{styles}</style>
        </MemoryRouter>
    )
}
