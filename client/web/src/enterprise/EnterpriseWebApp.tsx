import { FC } from 'react'

import '../SourcegraphWebApp.scss'

import { LegacySourcegraphWebApp } from '../LegacySourcegraphWebApp'
import { SourcegraphWebApp } from '../SourcegraphWebApp'
import {
    StaticAppConfig,
    StaticHardcodedAppConfig,
    StaticInjectedAppConfig,
    windowContextConfig,
} from '../staticAppConfig'

import { CodeIntelligenceBadgeContent } from './codeintel/badge/components/CodeIntelligenceBadgeContent'
import { CodeIntelligenceBadgeMenu } from './codeintel/badge/components/CodeIntelligenceBadgeMenu'
import { useCodeIntel } from './codeintel/useCodeIntel'
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
    routes: enterpriseRoutes,

    /**
     * Per feature injections
     */
    useCodeIntel,
    codeIntelligenceBadgeMenu: CodeIntelligenceBadgeMenu,
    codeIntelligenceBadgeContent: CodeIntelligenceBadgeContent,
} satisfies StaticInjectedAppConfig

const hardcodedConfig = {
    codeIntelligenceEnabled: true,
    codeInsightsEnabled: true,
    searchContextsEnabled: true,
    notebooksEnabled: true,
    codeMonitoringEnabled: true,
    searchAggregationEnabled: true,
} satisfies StaticHardcodedAppConfig

const staticAppConfig = {
    ...injectedValuesConfig,
    ...windowContextConfig,
    ...hardcodedConfig,
} satisfies StaticAppConfig

export const EnterpriseWebApp: FC = () => {
    if (window.context.experimentalFeatures.enableStorm) {
        // eslint-disable-next-line no-console
        console.log('Storm 🌪️ is enabled for this page load.')

        return <SourcegraphWebApp {...staticAppConfig} />
    }

    return <LegacySourcegraphWebApp {...staticAppConfig} />
}
