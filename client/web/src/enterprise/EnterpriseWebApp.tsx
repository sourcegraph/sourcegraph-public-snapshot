import type { FC } from 'react'

import '../SourcegraphWebApp.scss'

import { logger } from '@sourcegraph/common'

import { LegacySourcegraphWebApp } from '../LegacySourcegraphWebApp'
import { orgAreaHeaderNavItems } from '../org/area/navitems'
import { orgAreaRoutes } from '../org/area/routes'
import { orgSettingsAreaRoutes } from '../org/settings/routes'
import { orgSettingsSideBarItems } from '../org/settings/sidebaritems'
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
import { routes } from '../storm/routes'
import { userAreaHeaderNavItems } from '../user/area/navitems'
import { userAreaRoutes } from '../user/area/routes'
import { userSettingsAreaRoutes } from '../user/settings/routes'
import { userSettingsSideBarItems } from '../user/settings/sidebaritems'

import { APP_ROUTES } from './app/routes'
import { BrainDot } from './codeintel/dashboard/components/BrainDot'
import { enterpriseRepoContainerRoutes } from './repo/enterpriseRepoContainerRoutes'
import { enterpriseRepoRevisionContainerRoutes } from './repo/enterpriseRepoRevisionContainerRoutes'
import { enterpriseRepoHeaderActionButtons } from './repo/repoHeaderActionButtons'
import { enterpriseRepoSettingsAreaRoutes } from './repo/settings/routes'
import { enterpriseRepoSettingsSidebarGroups } from './repo/settings/sidebaritems'
import { enterpriseRoutes } from './routes'
import { enterpriseSiteAdminOverviewComponents } from './site-admin/overview/overviewComponents'

const injectedValuesConfig = {
    /**
     * Routes and nav links
     */
    siteAdminAreaRoutes,
    siteAdminSideBarGroups: siteAdminSidebarGroups,
    siteAdminOverviewComponents: enterpriseSiteAdminOverviewComponents,
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
    repoHeaderActionButtons: enterpriseRepoHeaderActionButtons,
    repoSettingsAreaRoutes: enterpriseRepoSettingsAreaRoutes,
    repoSettingsSidebarGroups: enterpriseRepoSettingsSidebarGroups,
    routes: windowContextConfig.isCodyApp ? APP_ROUTES : enterpriseRoutes,

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
                routes={routes}
                graphqlClient={graphqlClient}
                temporarySettingsStorage={temporarySettingsStorage}
            />
        )
    }

    return <LegacySourcegraphWebApp {...staticAppConfig} />
}
