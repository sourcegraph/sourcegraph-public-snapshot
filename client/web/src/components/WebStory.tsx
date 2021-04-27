import React, { useEffect, useMemo } from 'react'
import { MemoryRouter, MemoryRouterProps, RouteComponentProps, withRouter } from 'react-router'
import { useDarkMode } from 'storybook-dark-mode'

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
    }
> = ({ children, webStyles = _webStyles, ...memoryRouterProps }) => {
    const isLightTheme = !useDarkMode()
    const breadcrumbSetters = useBreadcrumbs()
    const Children = useMemo(() => withRouter(children), [children])

    useEffect(() => {
        const webappCss = document.createElement('style')
        webappCss.textContent = webStyles
        webappCss.setAttribute('id', 'webapp-css')

        const firstCss = document.head.querySelector('style')
        document.head.insertBefore(webappCss, firstCss)

        return () => {
            const webappCss = document.querySelector('#webapp-css')
            webappCss?.remove()
        }
    }, [webStyles])

    return (
        <>
            <MemoryRouter {...memoryRouterProps}>
                <Tooltip />
                <Children
                    {...breadcrumbSetters}
                    isLightTheme={isLightTheme}
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                />
            </MemoryRouter>
        </>
    )
}
