import type { RouteObject } from 'react-router-dom'

import type { SearchContextProps } from '@sourcegraph/shared/src/search'

import type { BatchChangesProps } from './batches'
import type { CodeIntelligenceProps } from './codeintel'
import type { CodeMonitoringProps } from './codeMonitoring'
import type { CodeInsightsProps } from './insights/types'
import type { NotebookProps } from './notebooks'
import type { OrgAreaRoute } from './org/area/OrgArea'
import type { OrgAreaHeaderNavItem } from './org/area/OrgHeader'
import type { OrgSettingsAreaRoute } from './org/settings/OrgSettingsArea'
import type { OrgSettingsSidebarItems } from './org/settings/OrgSettingsSidebar'
import type { OwnConfigProps } from './own/OwnConfigProps'
import type { RepoContainerRoute } from './repo/RepoContainer'
import type { RepoRevisionContainerRoute } from './repo/RepoRevisionContainer'
import type { RepoSettingsAreaRoute } from './repo/settings/RepoSettingsArea'
import type { RepoSettingsSideBarGroup } from './repo/settings/RepoSettingsSidebar'
import type { SearchAggregationProps } from './search'
import type { SiteAdminAreaRoute } from './site-admin/SiteAdminArea'
import type { SiteAdminSideBarGroups } from './site-admin/SiteAdminSidebar'
import type { UserAreaRoute } from './user/area/UserArea'
import type { UserAreaHeaderNavItem } from './user/area/UserAreaHeader'
import type { UserSettingsAreaRoute } from './user/settings/UserSettingsArea'
import type { UserSettingsSidebarItems } from './user/settings/UserSettingsSidebar'

export interface StaticAppConfig
    extends StaticHardcodedAppConfig,
        StaticInjectedAppConfig,
        StaticWindowContextComputedAppConfig {}

/**
 * Primitive configuration values we hardcode at the tip of the React tree.
 */
export interface StaticHardcodedAppConfig
    extends Pick<SearchAggregationProps, 'searchAggregationEnabled'>,
        Pick<CodeMonitoringProps, 'codeMonitoringEnabled'>,
        Pick<NotebookProps, 'notebooksEnabled'>,
        Pick<SearchContextProps, 'searchContextsEnabled'>,
        Pick<CodeInsightsProps, 'codeInsightsEnabled'>,
        Pick<CodeIntelligenceProps, 'codeIntelligenceEnabled'>,
        Pick<OwnConfigProps, 'ownEnabled'> {}

/**
 * Non-primitive values (components, objects) we inject at the tip of the React tree.
 */
export interface StaticInjectedAppConfig {
    siteAdminAreaRoutes: readonly SiteAdminAreaRoute[]
    siteAdminSideBarGroups: SiteAdminSideBarGroups
    siteAdminOverviewComponents: readonly React.ComponentType<React.PropsWithChildren<unknown>>[]
    userAreaHeaderNavItems: readonly UserAreaHeaderNavItem[]
    userAreaRoutes: readonly UserAreaRoute[]
    userSettingsSideBarItems: UserSettingsSidebarItems
    userSettingsAreaRoutes: readonly UserSettingsAreaRoute[]
    orgSettingsSideBarItems: OrgSettingsSidebarItems
    orgSettingsAreaRoutes: readonly OrgSettingsAreaRoute[]
    orgAreaHeaderNavItems: readonly OrgAreaHeaderNavItem[]
    orgAreaRoutes: readonly OrgAreaRoute[]
    repoContainerRoutes: readonly RepoContainerRoute[]
    repoRevisionContainerRoutes: readonly RepoRevisionContainerRoute[]
    repoSettingsAreaRoutes: readonly RepoSettingsAreaRoute[]
    repoSettingsSidebarGroups: readonly RepoSettingsSideBarGroup[]
    routes: RouteObject[]
}

/**
 * Static values we compute based on the `window.context`.
 *
 * Static in the sense that there are no other ways to change
 * these values except by refetching the entire original value (window.contexxt)
 */
export interface StaticWindowContextComputedAppConfig extends Pick<BatchChangesProps, 'batchChangesEnabled'> {
    isSourcegraphDotCom: boolean
    isCodyApp: boolean
    needsRepositoryConfiguration: boolean
    batchChangesWebhookLogsEnabled: boolean
}

/**
 * This configuration object is universal for both versions of the app:
 * enterprise and open source.
 */
export const windowContextConfig = {
    isSourcegraphDotCom: window.context.sourcegraphDotComMode,
    isCodyApp: window.context.codyAppMode,
    needsRepositoryConfiguration: window.context.needsRepositoryConfiguration,
    batchChangesWebhookLogsEnabled: window.context.batchChangesWebhookLogsEnabled,
    batchChangesEnabled: window.context.batchChangesEnabled,
} satisfies StaticWindowContextComputedAppConfig
