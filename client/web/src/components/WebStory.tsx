import { FC } from 'react'

import { MemoryRouter, MemoryRouterProps } from 'react-router'
import { Routes, Route, CompatRouter } from 'react-router-dom-v5-compat'

import { MockedStoryProvider, MockedStoryProviderProps } from '@sourcegraph/shared/src/stories'
import { NOOP_TELEMETRY_SERVICE, TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { WildcardThemeContext } from '@sourcegraph/wildcard'
import { usePrependStyles, useTheme } from '@sourcegraph/wildcard/src/stories'

import { SourcegraphContext } from '../jscontext'
import { setExperimentalFeaturesForTesting } from '../stores/experimentalFeatures'

import { BreadcrumbSetters, BreadcrumbsProps, useBreadcrumbs } from './Breadcrumbs'

import webStyles from '../SourcegraphWebApp.scss'

// With `StoryStoreV7` stories are isolated and window value is not shared between them.
// Global variables should be updated for every story individually.
if (!window.context) {
    window.context = {} as SourcegraphContext & Mocha.SuiteFunction
}

export type WebStoryChildrenProps = ThemeProps & BreadcrumbSetters & BreadcrumbsProps & TelemetryProps

export interface WebStoryProps
    extends Omit<MemoryRouterProps, 'children'>,
        Pick<MockedStoryProviderProps, 'mocks' | 'useStrictMocking'> {
    children: FC<WebStoryChildrenProps>
    path?: string
}

/**
 * Wrapper component for webapp Storybook stories that provides light theme and react-router props.
 * Takes a render function as children that gets called with the props.
 */
export const WebStory: FC<WebStoryProps> = ({
    children: Children,
    mocks,
    path = '*',
    useStrictMocking,
    ...memoryRouterProps
}) => {
    const isLightTheme = useTheme()
    const breadcrumbSetters = useBreadcrumbs()

    usePrependStyles('web-styles', webStyles)
    setExperimentalFeaturesForTesting()

    return (
        <MockedStoryProvider mocks={mocks} useStrictMocking={useStrictMocking}>
            <WildcardThemeContext.Provider value={{ isBranded: true }}>
                <MemoryRouter {...memoryRouterProps}>
                    <CompatRouter>
                        <Routes>
                            <Route
                                path={path}
                                element={
                                    <Children
                                        {...breadcrumbSetters}
                                        isLightTheme={isLightTheme}
                                        telemetryService={NOOP_TELEMETRY_SERVICE}
                                    />
                                }
                            />
                        </Routes>
                    </CompatRouter>
                </MemoryRouter>
            </WildcardThemeContext.Provider>
        </MockedStoryProvider>
    )
}
