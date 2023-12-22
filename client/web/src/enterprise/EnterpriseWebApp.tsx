import type { FC } from 'react'

import '../SourcegraphWebApp.scss'

import { logger } from '@sourcegraph/common'

import { LegacySourcegraphWebApp } from '../LegacySourcegraphWebApp'
import { orgAreaHeaderNavItems } from '../org/area/navitems'
import { orgAreaRoutes } from '../org/area/routes'
import { orgSettingsAreaRoutes } from '../org/settings/routes'
import { orgSettingsSideBarItems } from '../org/settings/sidebaritems'
import { repoSettingsAreaRoutes } from '../repo/settings/routes'
import { repoSettingsSideBarGroups } from '../repo/settings/sidebaritems'
import { routes } from '../routes'
import { siteAdminAreaRoutes } from '../site-admin/routes'
import { siteAdminSidebarGroups } from '../site-admin/sidebaritems'
import { SourcegraphWebApp } from '../SourcegraphWebApp'
import {
    type StaticAppConfig,
    type StaticHardcodedAppConfig,
    type StaticInjectedAppConfig,
    windowContextConfig,
} from '../staticAppConfig'
import type { AppShellInit } from '../storm/app-shell-init'
import { routes as stormRoutes } from '../storm/routes'
import { userAreaHeaderNavItems } from '../user/area/navitems'
import { userAreaRoutes } from '../user/area/routes'
import { userSettingsAreaRoutes } from '../user/settings/routes'
import { userSettingsSideBarItems } from '../user/settings/sidebaritems'

import { BrainDot } from './codeintel/dashboard/components/BrainDot'
import { enterpriseRepoContainerRoutes } from './repo/enterpriseRepoContainerRoutes'
import { enterpriseRepoRevisionContainerRoutes } from './repo/enterpriseRepoRevisionContainerRoutes'
import { siteAdminOverviewComponents } from './site-admin/overview/overviewComponents'

const injectedValuesConfig = {
    /**
     * Routes and nav links
     */
    siteAdminAreaRoutes,
    siteAdminSideBarGroups: siteAdminSidebarGroups,
    siteAdminOverviewComponents,
    userAreaHeaderNavItems,
    userAreaRoutes,
    userSettingsSideBarItems,
    userSettingsAreaRoutes,
    orgSettingsSideBarItems,
    orgSettingsAreaRoutes,
    orgAreaRoutes,
    orgAreaHeaderNavItems,
    repoContainerRoutes: enterpriseRepoContainerRoutes,
    repoRevisionContainerRoutes: enterpriseRepoRevisionContainerRoutes,
    repoSettingsAreaRoutes,
    repoSettingsSidebarGroups: repoSettingsSideBarGroups,
    routes,

    /**
     * Per feature injections
     */
    brainDot: BrainDot,
} satisfies StaticInjectedAppConfig

const hardcodedConfig = {
    codeIntelligenceEnabled: true,
    codeInsightsEnabled: true,
    searchContextsEnabled: true,
    notebooksEnabled: true,
    codeMonitoringEnabled: true,
    searchAggregationEnabled: true,
    ownEnabled: true,
} satisfies StaticHardcodedAppConfig

const staticAppConfig = {
    ...injectedValuesConfig,
    ...windowContextConfig,
    ...hardcodedConfig,
} satisfies StaticAppConfig

export const EnterpriseWebApp: FC<AppShellInit> = props => {
    if (window.context.experimentalFeatures.enableStorm) {
        const { graphqlClient, temporarySettingsStorage } = props

        logger.log('Storm 🌪️ is enabled for this page load.')

        return (
            <SourcegraphWebApp
                {...staticAppConfig}
                routes={stormRoutes}
                graphqlClient={graphqlClient}
                temporarySettingsStorage={temporarySettingsStorage}
            />
        )
    }

    return <LegacySourcegraphWebApp {...staticAppConfig} />
}
