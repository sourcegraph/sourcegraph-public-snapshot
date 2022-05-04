import * as React from 'react'

import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { Route, RouteComponentProps, Switch } from 'react-router'

import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import * as GQL from '@sourcegraph/shared/src/schema'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { AuthenticatedUser } from '../auth'
import { useBreadcrumbs, BreadcrumbSetters } from '../components/Breadcrumbs'
import { HeroPage } from '../components/HeroPage'
import { Page } from '../components/Page'
import { RouteDescriptor } from '../util/contributions'

import { ExtensionAreaRoute } from './extension/ExtensionArea'
import { ExtensionAreaHeaderNavItem } from './extension/ExtensionAreaHeader'
import { ExtensionsAreaHeader, ExtensionsAreaHeaderActionButton } from './ExtensionsAreaHeader'

import styles from './ExtensionsArea.module.scss'

const NotFoundPage: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" />
)

export interface ExtensionsAreaRoute extends RouteDescriptor<ExtensionsAreaRouteContext> {}

/**
 * Properties passed to all page components in the extensions area.
 */
export interface ExtensionsAreaRouteContext
    extends SettingsCascadeProps,
        PlatformContextProps,
        ThemeProps,
        TelemetryProps,
        BreadcrumbSetters {
    /** The currently authenticated user. */
    authenticatedUser: AuthenticatedUser | null

    /** The subject whose extensions and configuration to display. */
    subject: Pick<GQL.ISettingsSubject, 'id' | 'viewerCanAdminister'>
    extensionAreaRoutes: readonly ExtensionAreaRoute[]
    extensionAreaHeaderNavItems: readonly ExtensionAreaHeaderNavItem[]
    isSourcegraphDotCom: boolean
}

interface ExtensionsAreaProps
    extends RouteComponentProps<{ extensionID: string }>,
        SettingsCascadeProps,
        PlatformContextProps,
        ThemeProps,
        TelemetryProps {
    routes: readonly ExtensionsAreaRoute[]

    /**
     * The currently authenticated user.
     */
    authenticatedUser: AuthenticatedUser | null

    viewerSubject: Pick<GQL.ISettingsSubject, 'id' | 'viewerCanAdminister'>
    extensionAreaRoutes: readonly ExtensionAreaRoute[]
    extensionsAreaHeaderActionButtons: readonly ExtensionsAreaHeaderActionButton[]
    extensionAreaHeaderNavItems: readonly ExtensionAreaHeaderNavItem[]
    isSourcegraphDotCom: boolean
}

/**
 * The extensions area.
 */
export const ExtensionsArea: React.FunctionComponent<React.PropsWithChildren<ExtensionsAreaProps>> = props => {
    const { breadcrumbs, ...rootBreadcrumbSetters } = useBreadcrumbs()

    const childBreadcrumbSetters = rootBreadcrumbSetters.useBreadcrumb(
        React.useMemo(() => ({ link: { to: '/extensions', label: 'Extensions' }, key: 'Extensions' }), [])
    )

    const context: ExtensionsAreaRouteContext = {
        authenticatedUser: props.authenticatedUser,
        settingsCascade: props.settingsCascade,
        platformContext: props.platformContext,
        subject: props.viewerSubject,
        extensionAreaRoutes: props.extensionAreaRoutes,
        extensionAreaHeaderNavItems: props.extensionAreaHeaderNavItems,
        isLightTheme: props.isLightTheme,
        isSourcegraphDotCom: props.isSourcegraphDotCom,
        telemetryService: props.telemetryService,
        ...childBreadcrumbSetters,
    }

    return (
        <Page className={styles.extensionsArea}>
            <ExtensionsAreaHeader
                {...props}
                {...context}
                actionButtons={props.extensionsAreaHeaderActionButtons}
                isPrimaryHeader={props.location.pathname === props.match.path}
            />
            <Switch>
                {props.routes.map(
                    ({ path, exact, condition = () => true, render }) =>
                        condition(context) && (
                            <Route
                                key="hardcoded-key"
                                path={props.match.url + path}
                                exact={exact}
                                render={routeComponentProps => render({ ...context, ...routeComponentProps })}
                            />
                        )
                )}
                <Route key="hardcoded-key" component={NotFoundPage} />
            </Switch>
        </Page>
    )
}
