import React, { useLayoutEffect, useMemo, useState } from 'react'
import { MemoryRouter, MemoryRouterProps, RouteComponentProps, withRouter } from 'react-router'
import { useDarkMode } from 'storybook-dark-mode'

import { Tooltip } from '@sourcegraph/branded/src/components/tooltip/Tooltip'
import { NOOP_TELEMETRY_SERVICE, TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { useStyles } from '@sourcegraph/storybook/src/hooks/useStyles'

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
    const [isLightTheme, setIsLightTheme] = useState(!useDarkMode())
    const breadcrumbSetters = useBreadcrumbs()
    const Children = useMemo(() => withRouter(children), [children])
    useStyles(webStyles)

    useLayoutEffect(() => {
        const listener = ((event: CustomEvent<boolean>): void => {
            setIsLightTheme(event.detail)
        }) as EventListener
        document.body.addEventListener('chromatic-light-theme-toggled', listener)
        return () => document.body.removeEventListener('chromatic-light-theme-toggled', listener)
    }, [])

    return (
        <MemoryRouter {...memoryRouterProps}>
            <Tooltip />
            <Children {...breadcrumbSetters} isLightTheme={isLightTheme} telemetryService={NOOP_TELEMETRY_SERVICE} />
        </MemoryRouter>
    )
}
