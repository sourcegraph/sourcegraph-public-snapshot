import React from 'react'

import { MemoryRouter, MemoryRouterProps } from 'react-router'

import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { MockedStoryProvider, MockedStoryProviderProps, usePrependStyles, useTheme } from '@sourcegraph/storybook'
// Add root Tooltip for Storybook
// eslint-disable-next-line no-restricted-imports
import { Tooltip, WildcardThemeContext } from '@sourcegraph/wildcard'

import brandedStyles from '../global-styles/index.scss'

export interface BrandedProps extends MemoryRouterProps, Pick<MockedStoryProviderProps, 'mocks' | 'useStrictMocking'> {
    children: React.FunctionComponent<React.PropsWithChildren<ThemeProps>>
    styles?: string
}

/**
 * Wrapper component for branded Storybook stories that provides light theme and react-router props.
 * Takes a render function as children that gets called with the props.
 */
export const BrandedStory: React.FunctionComponent<React.PropsWithChildren<BrandedProps>> = ({
    children: Children,
    styles = brandedStyles,
    mocks,
    useStrictMocking,
    ...memoryRouterProps
}) => {
    const isLightTheme = useTheme()
    usePrependStyles('branded-story-styles', styles)

    return (
        <MockedStoryProvider mocks={mocks} useStrictMocking={useStrictMocking}>
            <WildcardThemeContext.Provider value={{ isBranded: true }}>
                <MemoryRouter {...memoryRouterProps}>
                    <Tooltip />
                    <Children isLightTheme={isLightTheme} />
                </MemoryRouter>
            </WildcardThemeContext.Provider>
        </MockedStoryProvider>
    )
}
