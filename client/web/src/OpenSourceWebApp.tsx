import { FC } from 'react'

import './SourcegraphWebApp.scss'
import { LegacySourcegraphWebApp } from './LegacySourcegraphWebApp'
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
import {
    StaticAppConfig,
    StaticHardcodedAppConfig,
    StaticInjectedAppConfig,
    windowContextConfig,
} from './staticAppConfig'
import { userAreaHeaderNavItems } from './user/area/navitems'
import { userAreaRoutes } from './user/area/routes'
import { userSettingsAreaRoutes } from './user/settings/routes'
import { userSettingsSideBarItems } from './user/settings/sidebaritems'

const injectedValuesConfig = {
    /**
     * Routes and nav links
     */
    siteAdminAreaRoutes,
    siteAdminSideBarGroups: siteAdminSidebarGroups,
    siteAdminOverviewComponents: [],
    userAreaRoutes,
    userAreaHeaderNavItems,
    userSettingsSideBarItems,
    userSettingsAreaRoutes,
    orgSettingsSideBarItems,
    orgSettingsAreaRoutes,
    orgAreaRoutes,
    orgAreaHeaderNavItems,
    repoContainerRoutes,
    repoRevisionContainerRoutes,
    repoHeaderActionButtons,
    repoSettingsAreaRoutes,
    repoSettingsSidebarGroups: repoSettingsSideBarGroups,
    routes,
} satisfies StaticInjectedAppConfig

const hardcodedConfig = {
    codeIntelligenceEnabled: false,
    searchContextsEnabled: false,
    notebooksEnabled: false,
    codeMonitoringEnabled: false,
    searchAggregationEnabled: false,
} satisfies StaticHardcodedAppConfig

const staticAppConfig = {
    ...injectedValuesConfig,
    ...windowContextConfig,
    ...hardcodedConfig,
} satisfies StaticAppConfig

// Entry point for the app without enterprise functionality.
// For more info see: https://docs.sourcegraph.com/admin/subscriptions#paid-subscriptions-for-sourcegraph-enterprise
export const OpenSourceWebApp: FC = () => <LegacySourcegraphWebApp {...staticAppConfig} />
