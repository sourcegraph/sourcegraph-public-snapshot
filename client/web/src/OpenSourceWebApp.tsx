import React from 'react'

import './SourcegraphWebApp.scss'
import { orgAreaHeaderNavItems } from './org/area/navitems'
import { orgAreaRoutes } from './org/area/routes'
import { orgSettingsAreaRoutes } from './org/settings/routes'
import { orgSettingsSideBarItems } from './org/settings/sidebaritems'
import { repoContainerRoutes } from './repo/repoContainerRoutes'
import { repoHeaderActionButtons } from './repo/repoHeaderActionButtons'
import { repoRevisionContainerRoutes } from './repo/repoRevisionContainerRoutes'
import { repoSettingsAreaRoutes } from './repo/settings/routes'
import { repoSettingsSideBarGroups } from './repo/settings/sidebaritems'
import { routes } from './routes'
import { siteAdminAreaRoutes } from './site-admin/routes'
import { siteAdminSidebarGroups } from './site-admin/sidebaritems'
import { SourcegraphWebApp } from './SourcegraphWebApp'
import { userAreaHeaderNavItems } from './user/area/navitems'
import { userAreaRoutes } from './user/area/routes'
import { userSettingsAreaRoutes } from './user/settings/routes'
import { userSettingsSideBarItems } from './user/settings/sidebaritems'

// Entry point for the app without enterprise functionality.
// For more info see: https://docs.sourcegraph.com/admin/subscriptions#paid-subscriptions-for-sourcegraph-enterprise
export const OpenSourceWebApp: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <SourcegraphWebApp
        siteAdminAreaRoutes={siteAdminAreaRoutes}
        siteAdminSideBarGroups={siteAdminSidebarGroups}
        siteAdminOverviewComponents={[]}
        userAreaRoutes={userAreaRoutes}
        userAreaHeaderNavItems={userAreaHeaderNavItems}
        userSettingsSideBarItems={userSettingsSideBarItems}
        userSettingsAreaRoutes={userSettingsAreaRoutes}
        orgSettingsSideBarItems={orgSettingsSideBarItems}
        orgSettingsAreaRoutes={orgSettingsAreaRoutes}
        orgAreaRoutes={orgAreaRoutes}
        orgAreaHeaderNavItems={orgAreaHeaderNavItems}
        repoContainerRoutes={repoContainerRoutes}
        repoRevisionContainerRoutes={repoRevisionContainerRoutes}
        repoHeaderActionButtons={repoHeaderActionButtons}
        repoSettingsAreaRoutes={repoSettingsAreaRoutes}
        repoSettingsSidebarGroups={repoSettingsSideBarGroups}
        routes={routes}
        codeIntelligenceEnabled={false}
        batchChangesEnabled={false}
        searchContextsEnabled={false}
        notebooksEnabled={false}
        codeMonitoringEnabled={false}
        searchAggregationEnabled={false}
    />
)
