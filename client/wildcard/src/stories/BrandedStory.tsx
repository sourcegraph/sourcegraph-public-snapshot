import React from 'react'

import { MemoryRouter, type MemoryRouterProps } from 'react-router-dom'

import { WildcardThemeContext } from '../hooks'

import { usePrependStyles, useStorybookTheme } from './hooks'

import brandedStyles from '../global-styles/index.scss'

export interface BrandedProps extends Omit<MemoryRouterProps, 'children'> {
    children: React.FunctionComponent<
        React.PropsWithChildren<{
            isLightTheme: boolean
        }>
    >
    styles?: string
}

/**
 * Wrapper component for branded Storybook stories that provides light theme and react-router props.
 * Takes a render function as children that gets called with the props.
 */
export const BrandedStory: React.FunctionComponent<BrandedProps> = ({
    children: Children,
    styles = brandedStyles,
    ...memoryRouterProps
}) => {
    const isLightTheme = useStorybookTheme()
    usePrependStyles('branded-story-styles', styles)

    return (
        <WildcardThemeContext.Provider value={{ isBranded: true }}>
            <MemoryRouter {...memoryRouterProps}>
                <Children isLightTheme={isLightTheme} />
            </MemoryRouter>
        </WildcardThemeContext.Provider>
    )
}
