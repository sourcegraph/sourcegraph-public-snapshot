import { FC } from 'react'

import { RouterProvider, createMemoryRouter, MemoryRouterProps } from 'react-router-dom'

import { MockedStoryProvider, MockedStoryProviderProps } from '@sourcegraph/shared/src/stories'
import { NOOP_TELEMETRY_SERVICE, TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeContext, ThemeSetting } from '@sourcegraph/shared/src/theme'
import { WildcardThemeContext } from '@sourcegraph/wildcard'
import { usePrependStyles, useStorybookTheme } from '@sourcegraph/wildcard/src/stories'

import { SourcegraphContext } from '../jscontext'
import { setExperimentalFeaturesForTesting } from '../stores/experimentalFeatures'

import { BreadcrumbSetters, BreadcrumbsProps, useBreadcrumbs } from './Breadcrumbs'

import webStyles from '../SourcegraphWebApp.scss'

// With `StoryStoreV7` stories are isolated and window value is not shared between them.
// Global variables should be updated for every story individually.
if (!window.context) {
    window.context = {} as SourcegraphContext & Mocha.SuiteFunction
}

export type WebStoryChildrenProps = BreadcrumbSetters &
    BreadcrumbsProps &
    TelemetryProps & {
        isLightTheme: boolean
    }

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
    initialEntries = ['/'],
    initialIndex = 1,
}) => {
    const isLightTheme = useStorybookTheme()
    const breadcrumbSetters = useBreadcrumbs()

    usePrependStyles('web-styles', webStyles)
    setExperimentalFeaturesForTesting()

    const routes = [
        {
            path,
            element: (
                <Children
                    {...breadcrumbSetters}
                    isLightTheme={isLightTheme}
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                />
            ),
        },
    ]

    const router = createMemoryRouter(routes, {
        initialEntries,
        initialIndex,
    })

    return (
        <MockedStoryProvider mocks={mocks} useStrictMocking={useStrictMocking}>
            <WildcardThemeContext.Provider value={{ isBranded: true }}>
                <ThemeContext.Provider value={{ themeSetting: isLightTheme ? ThemeSetting.Light : ThemeSetting.Dark }}>
                    <RouterProvider router={router} />
                </ThemeContext.Provider>
            </WildcardThemeContext.Provider>
        </MockedStoryProvider>
    )
}
