import type { FC } from 'react'

import '../SourcegraphWebApp.scss'

import { logger } from '@sourcegraph/common'

import { LegacySourcegraphWebApp } from '../LegacySourcegraphWebApp'
import { SourcegraphWebApp } from '../SourcegraphWebApp'
import {
    type StaticAppConfig,
    type StaticHardcodedAppConfig,
    type StaticInjectedAppConfig,
    windowContextConfig,
} from '../staticAppConfig'
import type { AppShellInit } from '../storm/app-shell-init'
import { routes } from '../storm/routes'

import { APP_ROUTES } from './app/routes'
import { BrainDot } from './codeintel/dashboard/components/BrainDot'
import { enterpriseOrgAreaHeaderNavItems } from './organizations/navitems'
import { enterpriseOrganizationAreaRoutes } from './organizations/routes'
import { enterpriseOrgSettingsAreaRoutes } from './organizations/settings/routes'
import { enterpriseOrgSettingsSideBarItems } from './organizations/settings/sidebaritems'
import { enterpriseRepoContainerRoutes } from './repo/enterpriseRepoContainerRoutes'
import { enterpriseRepoRevisionContainerRoutes } from './repo/enterpriseRepoRevisionContainerRoutes'
import { enterpriseRepoHeaderActionButtons } from './repo/repoHeaderActionButtons'
import { enterpriseRepoSettingsAreaRoutes } from './repo/settings/routes'
import { enterpriseRepoSettingsSidebarGroups } from './repo/settings/sidebaritems'
import { enterpriseRoutes } from './routes'
import { enterpriseSiteAdminOverviewComponents } from './site-admin/overview/overviewComponents'
import { enterpriseSiteAdminAreaRoutes } from './site-admin/routes'
import { enterpriseSiteAdminSidebarGroups } from './site-admin/sidebaritems'
import { enterpriseUserAreaHeaderNavItems } from './user/navitems'
import { enterpriseUserAreaRoutes } from './user/routes'
import { enterpriseUserSettingsAreaRoutes } from './user/settings/routes'
import { enterpriseUserSettingsSideBarItems } from './user/settings/sidebaritems'

const injectedValuesConfig = {
    /**
     * Routes and nav links
     */
    siteAdminAreaRoutes: enterpriseSiteAdminAreaRoutes,
    siteAdminSideBarGroups: enterpriseSiteAdminSidebarGroups,
    siteAdminOverviewComponents: enterpriseSiteAdminOverviewComponents,
    userAreaHeaderNavItems: enterpriseUserAreaHeaderNavItems,
    userAreaRoutes: enterpriseUserAreaRoutes,
    userSettingsSideBarItems: enterpriseUserSettingsSideBarItems,
    userSettingsAreaRoutes: enterpriseUserSettingsAreaRoutes,
    orgSettingsSideBarItems: enterpriseOrgSettingsSideBarItems,
    orgSettingsAreaRoutes: enterpriseOrgSettingsAreaRoutes,
    orgAreaRoutes: enterpriseOrganizationAreaRoutes,
    orgAreaHeaderNavItems: enterpriseOrgAreaHeaderNavItems,
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
        const { graphqlClient, temporarySettingsStorage, telemetryRecorder } = props

        logger.log('Storm 🌪️ is enabled for this page load.')

        return (
            <SourcegraphWebApp
                {...staticAppConfig}
                routes={routes}
                graphqlClient={graphqlClient}
                temporarySettingsStorage={temporarySettingsStorage}
                telemetryRecorder={telemetryRecorder}
            />
        )
    }

    return <LegacySourcegraphWebApp {...staticAppConfig} />
}
