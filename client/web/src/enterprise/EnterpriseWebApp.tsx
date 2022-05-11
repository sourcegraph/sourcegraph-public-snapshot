import React from 'react'

import '../SourcegraphWebApp.scss'
import { KEYBOARD_SHORTCUTS } from '@sourcegraph/shared/src/keyboardShortcuts/keyboardShortcuts'

import { SourcegraphWebApp } from '../SourcegraphWebApp'

import { CodeIntelligenceBadgeContent } from './codeintel/badge/components/CodeIntelligenceBadgeContent'
import { CodeIntelligenceBadgeMenu } from './codeintel/badge/components/CodeIntelligenceBadgeMenu'
import { enterpriseExtensionAreaHeaderNavItems } from './extensions/extension/extensionAreaHeaderNavItems'
import { enterpriseExtensionAreaRoutes } from './extensions/extension/routes'
import { enterpriseExtensionsAreaHeaderActionButtons } from './extensions/extensionsAreaHeaderActionButtons'
import { enterpriseExtensionsAreaRoutes } from './extensions/routes'
import { enterpriseOrgAreaHeaderNavItems } from './organizations/navitems'
import { enterpriseOrganizationAreaRoutes } from './organizations/routes'
import { enterpriseRepoHeaderActionButtons } from './repo/repoHeaderActionButtons'
import { enterpriseRepoContainerRoutes, enterpriseRepoRevisionContainerRoutes } from './repo/routes'
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

export const EnterpriseWebApp: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <SourcegraphWebApp
        extensionAreaRoutes={enterpriseExtensionAreaRoutes}
        extensionAreaHeaderNavItems={enterpriseExtensionAreaHeaderNavItems}
        extensionsAreaRoutes={enterpriseExtensionsAreaRoutes}
        extensionsAreaHeaderActionButtons={enterpriseExtensionsAreaHeaderActionButtons}
        siteAdminAreaRoutes={enterpriseSiteAdminAreaRoutes}
        siteAdminSideBarGroups={enterpriseSiteAdminSidebarGroups}
        siteAdminOverviewComponents={enterpriseSiteAdminOverviewComponents}
        userAreaHeaderNavItems={enterpriseUserAreaHeaderNavItems}
        userAreaRoutes={enterpriseUserAreaRoutes}
        userSettingsSideBarItems={enterpriseUserSettingsSideBarItems}
        userSettingsAreaRoutes={enterpriseUserSettingsAreaRoutes}
        orgAreaRoutes={enterpriseOrganizationAreaRoutes}
        orgAreaHeaderNavItems={enterpriseOrgAreaHeaderNavItems}
        repoContainerRoutes={enterpriseRepoContainerRoutes}
        repoRevisionContainerRoutes={enterpriseRepoRevisionContainerRoutes}
        repoHeaderActionButtons={enterpriseRepoHeaderActionButtons}
        repoSettingsAreaRoutes={enterpriseRepoSettingsAreaRoutes}
        repoSettingsSidebarGroups={enterpriseRepoSettingsSidebarGroups}
        routes={enterpriseRoutes}
        keyboardShortcuts={KEYBOARD_SHORTCUTS}
        codeIntelligenceEnabled={true}
        codeIntelligenceBadgeMenu={CodeIntelligenceBadgeMenu}
        codeIntelligenceBadgeContent={CodeIntelligenceBadgeContent}
        codeInsightsEnabled={true}
        batchChangesEnabled={window.context.batchChangesEnabled}
        searchContextsEnabled={true}
    />
)
