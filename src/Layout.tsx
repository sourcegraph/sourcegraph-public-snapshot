import { ClientConnection } from '@sourcegraph/extensions-client-common/lib/messaging'
import React from 'react'
import { Redirect, Route, RouteComponentProps, Switch } from 'react-router'
import * as GQL from './backend/graphqlschema'
import { LinkExtension } from './extension/Link'
import { ExtensionsDocumentsProps, ExtensionsEnvironmentProps } from './extensions/environment/ExtensionsEnvironment'
import { ExtensionAreaRoute } from './extensions/extension/ExtensionArea'
import { ExtensionAreaHeaderNavItem } from './extensions/extension/ExtensionAreaHeader'
import { ExtensionsAreaRoute } from './extensions/ExtensionsArea'
import { ExtensionsAreaHeaderActionButton } from './extensions/ExtensionsAreaHeader'
import {
    ConfigurationCascadeProps,
    ExtensionsControllerProps,
    ExtensionsProps,
} from './extensions/ExtensionsClientCommonContext'
import { GlobalAlerts } from './global/GlobalAlerts'
import { GlobalDebug } from './global/GlobalDebug'
import { KeybindingsProps } from './keybindings'
import { IntegrationsToast } from './marketing/IntegrationsToast'
import { GlobalNavbar } from './nav/GlobalNavbar'
import { RepoHeaderActionButton } from './repo/RepoHeader'
import { RepoRevContainerRoute } from './repo/RepoRevContainer'
import { routes } from './routes'
import { parseSearchURLQuery } from './search'
import { SiteAdminAreaRoute } from './site-admin/SiteAdminArea'
import { SiteAdminSideBarGroups } from './site-admin/SiteAdminSidebar'
import { UserAccountAreaRoute } from './user/account/UserAccountArea'
import { UserAccountSidebarItems } from './user/account/UserAccountSidebar'
import { UserAreaRoute } from './user/area/UserArea'
import { UserAreaHeaderNavItem } from './user/area/UserAreaHeader'

export interface LayoutProps
    extends RouteComponentProps<any>,
        ConfigurationCascadeProps,
        ExtensionsProps,
        ExtensionsEnvironmentProps,
        ExtensionsControllerProps,
        ExtensionsDocumentsProps,
        KeybindingsProps {
    extensionAreaRoutes: ReadonlyArray<ExtensionAreaRoute>
    extensionAreaHeaderNavItems: ReadonlyArray<ExtensionAreaHeaderNavItem>
    extensionsAreaRoutes: ReadonlyArray<ExtensionsAreaRoute>
    extensionsAreaHeaderActionButtons: ReadonlyArray<ExtensionsAreaHeaderActionButton>
    siteAdminAreaRoutes: ReadonlyArray<SiteAdminAreaRoute>
    siteAdminSideBarGroups: SiteAdminSideBarGroups
    siteAdminOverviewComponents: ReadonlyArray<React.ComponentType>
    userAreaHeaderNavItems: ReadonlyArray<UserAreaHeaderNavItem>
    userAreaRoutes: ReadonlyArray<UserAreaRoute>
    userAccountSideBarItems: UserAccountSidebarItems
    userAccountAreaRoutes: ReadonlyArray<UserAccountAreaRoute>
    repoRevContainerRoutes: ReadonlyArray<RepoRevContainerRoute>
    repoHeaderActionButtons: ReadonlyArray<RepoHeaderActionButton>

    authenticatedUser: GQL.IUser | null

    /**
     * The subject GraphQL node ID of the viewer, which is used to look up the viewer's configuration settings.
     * This is either the site's GraphQL node ID (for anonymous users) or the authenticated user's GraphQL node ID.
     */
    viewerSubject: Pick<GQL.IConfigurationSubject, 'id' | 'viewerCanAdminister'>

    clientConnection: Promise<ClientConnection>

    isLightTheme: boolean
    onThemeChange: () => void
    onMainPage: (mainPage: boolean) => void
    isMainPage: boolean
    navbarSearchQuery: string
    onNavbarQueryChange: (query: string) => void

    children?: never
}

export const Layout: React.SFC<LayoutProps> = props => {
    const isSearchHomepage = props.location.pathname === '/search' && !parseSearchURLQuery(props.location.search)

    const needsSiteInit = window.context.showOnboarding
    const isSiteInit = props.location.pathname === '/site-admin/init'

    // Force light theme on site init page.
    if (isSiteInit && !props.isLightTheme) {
        props.onThemeChange()
    }

    // Remove trailing slash (which is never valid in any of our URLs).
    if (props.location.pathname !== '/' && props.location.pathname.endsWith('/')) {
        return <Redirect to={{ ...props.location, pathname: props.location.pathname.slice(0, -1) }} />
    }

    return (
        <div className="layout">
            <GlobalAlerts isSiteAdmin={!!props.authenticatedUser && props.authenticatedUser.siteAdmin} />
            {!needsSiteInit &&
                !isSiteInit &&
                !!props.authenticatedUser && <IntegrationsToast history={props.history} />}
            {!isSiteInit && <GlobalNavbar {...props} lowProfile={isSearchHomepage} />}
            {needsSiteInit && !isSiteInit && <Redirect to="/site-admin/init" />}
            <Switch>
                {routes.map((route, i) => {
                    const isFullWidth = !route.forceNarrowWidth
                    return (
                        <Route
                            {...route}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            component={undefined}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <div
                                    className={[
                                        'layout__app-router-container',
                                        `layout__app-router-container--${isFullWidth ? 'full-width' : 'restricted'}`,
                                    ].join(' ')}
                                >
                                    {route.render({ ...props, ...routeComponentProps })}
                                    {!!props.authenticatedUser && (
                                        <LinkExtension authenticatedUser={props.authenticatedUser} />
                                    )}
                                </div>
                            )}
                        />
                    )
                })}
            </Switch>
            <GlobalDebug {...props} />
        </div>
    )
}
