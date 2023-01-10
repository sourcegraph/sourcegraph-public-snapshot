import React from 'react'

import { MemoryRouter, MemoryRouterProps } from 'react-router'
import { CompatRouter } from 'react-router-dom-v5-compat'

import { WildcardThemeContext } from '../hooks/useWildcardTheme'

import { usePrependStyles, useTheme } from './hooks'

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
    const isLightTheme = useTheme()
    usePrependStyles('branded-story-styles', styles)

    return (
        <WildcardThemeContext.Provider value={{ isBranded: true }}>
            <MemoryRouter {...memoryRouterProps}>
                <CompatRouter>
                    <Children isLightTheme={isLightTheme} />
                </CompatRouter>
            </MemoryRouter>
        </WildcardThemeContext.Provider>
    )
}
