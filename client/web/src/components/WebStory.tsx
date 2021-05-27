import React, { useLayoutEffect, useMemo } from 'react'
import { MemoryRouter, MemoryRouterProps, RouteComponentProps, withRouter } from 'react-router'
import { useDarkMode } from 'storybook-dark-mode'

import { prependCSSToDocumentHead } from '@sourcegraph/branded/src/components/BrandedStory'
import { Tooltip } from '@sourcegraph/branded/src/components/tooltip/Tooltip'
import { NOOP_TELEMETRY_SERVICE, TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import _webStyles from '../SourcegraphWebApp.scss'

import { BreadcrumbSetters, BreadcrumbsProps, useBreadcrumbs } from './Breadcrumbs'

export interface WebStoryProps extends MemoryRouterProps {
    children: React.FunctionComponent<
        ThemeProps & BreadcrumbSetters & BreadcrumbsProps & TelemetryProps & RouteComponentProps<any>
    >
}

/**
 * Wrapper component for webapp Storybook stories that provides light theme and react-router props.
 * Takes a render function as children that gets called with the props.
 */
export const WebStory: React.FunctionComponent<
    WebStoryProps & {
        webStyles?: string
        forceDarkMode?: boolean
    }
> = ({ children, webStyles = _webStyles, forceDarkMode, ...memoryRouterProps }) => {
    const darkMode = useDarkMode()
    const isLightTheme = forceDarkMode !== undefined ? !forceDarkMode : !darkMode
    const breadcrumbSetters = useBreadcrumbs()
    const Children = useMemo(() => withRouter(children), [children])

    useLayoutEffect(() => {
        const styleTag = prependCSSToDocumentHead(webStyles)

        return () => {
            styleTag.remove()
        }
    }, [webStyles])

    return (
        <MemoryRouter {...memoryRouterProps}>
            <Tooltip />
            <Children {...breadcrumbSetters} isLightTheme={isLightTheme} telemetryService={NOOP_TELEMETRY_SERVICE} />
        </MemoryRouter>
    )
}
