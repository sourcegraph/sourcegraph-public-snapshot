import React from 'react'
import { SiteAdminRepositoriesOverviewCard } from '../repositories/SiteAdminRepositoriesOverviewCard'
import { SiteAdminUsersOverviewCard } from '../users/SiteAdminUsersOverviewCard'
import { SiteAdminUsageStatisticsOverviewCard } from '../usageStatistics/SiteAdminUsageStatisticsOverviewCard'
import { ActivationProps } from '../../../../shared/src/components/activation/Activation'
import { SiteAdminActivationChecklistOverviewCard } from '../activationChecklist/SiteAdminActivationChecklistOverviewCard'
import H from 'history'

export interface SiteAdminOverviewComponent {
    component: React.ComponentType<ActivationProps & { history: H.History }>
    noCardClass?: boolean
    fullWidth?: boolean
}

/**
 * Additional components to render on the SiteAdminOverviewPage.
 */
export const siteAdminOverviewComponents: readonly SiteAdminOverviewComponent[] = [
    { component: SiteAdminRepositoriesOverviewCard },
    { component: SiteAdminUsersOverviewCard },
    { component: SiteAdminUsageStatisticsOverviewCard },
    { component: SiteAdminActivationChecklistOverviewCard },
]
