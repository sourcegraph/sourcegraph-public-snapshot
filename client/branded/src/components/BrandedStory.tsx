import React from 'react'

import { MemoryRouter, MemoryRouterProps } from 'react-router'
import { CompatRouter } from 'react-router-dom-v5-compat'

import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { MockedStoryProvider, MockedStoryProviderProps, usePrependStyles, useTheme } from '@sourcegraph/storybook'
import { WildcardThemeContext } from '@sourcegraph/wildcard'

import brandedStyles from '../global-styles/index.scss'

export interface BrandedProps
    extends Omit<MemoryRouterProps, 'children'>,
        Pick<MockedStoryProviderProps, 'mocks' | 'useStrictMocking'> {
    children: React.FunctionComponent<React.PropsWithChildren<ThemeProps>>
    styles?: string
}

/**
 * Wrapper component for branded Storybook stories that provides light theme and react-router props.
 * Takes a render function as children that gets called with the props.
 */
export const BrandedStory: React.FunctionComponent<BrandedProps> = ({
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
                    <CompatRouter>
                        <Children isLightTheme={isLightTheme} />
                    </CompatRouter>
                </MemoryRouter>
            </WildcardThemeContext.Provider>
        </MockedStoryProvider>
    )
}
