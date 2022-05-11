import React, { useMemo } from 'react'

import { MemoryRouter, MemoryRouterProps, RouteComponentProps, withRouter } from 'react-router'

import { NOOP_TELEMETRY_SERVICE, TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { MockedStoryProvider, MockedStoryProviderProps, usePrependStyles, useTheme } from '@sourcegraph/storybook'
// Add root Tooltip for Storybook
// eslint-disable-next-line no-restricted-imports
import { Tooltip, WildcardThemeContext } from '@sourcegraph/wildcard'

import { BreadcrumbSetters, BreadcrumbsProps, useBreadcrumbs } from './Breadcrumbs'

import webStyles from '../SourcegraphWebApp.scss'

export interface WebStoryProps extends MemoryRouterProps, Pick<MockedStoryProviderProps, 'mocks' | 'useStrictMocking'> {
    children: React.FunctionComponent<
        React.PropsWithChildren<
            ThemeProps & BreadcrumbSetters & BreadcrumbsProps & TelemetryProps & RouteComponentProps<any>
        >
    >
}

/**
 * Wrapper component for webapp Storybook stories that provides light theme and react-router props.
 * Takes a render function as children that gets called with the props.
 */
export const WebStory: React.FunctionComponent<React.PropsWithChildren<WebStoryProps>> = ({
    children,
    mocks,
    useStrictMocking,
    ...memoryRouterProps
}) => {
    const isLightTheme = useTheme()
    const breadcrumbSetters = useBreadcrumbs()
    const Children = useMemo(() => withRouter(children), [children])

    usePrependStyles('web-styles', webStyles)

    return (
        <MockedStoryProvider mocks={mocks} useStrictMocking={useStrictMocking}>
            <WildcardThemeContext.Provider value={{ isBranded: true }}>
                <MemoryRouter {...memoryRouterProps}>
                    <Tooltip />
                    <Children
                        {...breadcrumbSetters}
                        isLightTheme={isLightTheme}
                        telemetryService={NOOP_TELEMETRY_SERVICE}
                    />
                </MemoryRouter>
            </WildcardThemeContext.Provider>
        </MockedStoryProvider>
    )
}
