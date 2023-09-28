import BrainIcon from 'mdi-react/BrainIcon'
import BriefcaseIcon from 'mdi-react/BriefcaseIcon'
import PackageVariantIcon from 'mdi-react/PackageVariantIcon'

import { BatchChangesIcon } from '../../batches/icons'
import { CodyPageIcon } from '../../cody/chat/CodyPageIcon'
import {
    apiConsoleGroup,
    analyticsGroup,
    configurationGroup as ossConfigurationGroup,
    maintenanceGroup as ossMaintenanceGroup,
    repositoriesGroup as ossRepositoriesGroup,
    usersGroup as ossUsersGroup,
} from '../../site-admin/sidebaritems'
import type { SiteAdminSideBarGroup, SiteAdminSideBarGroups } from '../../site-admin/SiteAdminSidebar'
import { SHOW_BUSINESS_FEATURES } from '../dotcom/productSubscriptions/features'

const configurationGroup: SiteAdminSideBarGroup = {
    ...ossConfigurationGroup,
    items: [
        ...ossConfigurationGroup.items,
        {
            label: 'License',
            to: '/site-admin/license',

            condition: ({ isCodyApp }) => !isCodyApp,
        },
        {
            label: 'Incoming webhooks',
            to: '/site-admin/webhooks/incoming',
            condition: ({ isCodyApp }) => !isCodyApp,
        },
        {
            label: 'Outgoing webhooks',
            to: '/site-admin/webhooks/outgoing',
            condition: ({ isCodyApp }) => !isCodyApp,
        },
    ],
}

const maintenanceGroup: SiteAdminSideBarGroup = {
    ...ossMaintenanceGroup,
    items: [
        ...ossMaintenanceGroup.items,
        {
            label: 'Code Insights jobs',
            to: '/site-admin/code-insights-jobs',
            condition: ({ isCodyApp, codeInsightsEnabled }) => !isCodyApp && codeInsightsEnabled,
        },
    ],
}

const executorsGroup: SiteAdminSideBarGroup = {
    header: {
        label: 'Executors',
        icon: PackageVariantIcon,
    },
    condition: () => Boolean(window.context?.executorsEnabled),
    items: [
        {
            to: '/site-admin/executors',
            label: 'Instances',
            exact: true,
        },
        {
            to: '/site-admin/executors/secrets',
            label: 'Secrets',
        },
    ],
}

export const batchChangesGroup: SiteAdminSideBarGroup = {
    header: {
        label: 'Batch Changes',
        icon: BatchChangesIcon,
    },
    items: [
        {
            label: 'Settings',
            to: '/site-admin/batch-changes',
        },
        {
            label: 'Batch specs',
            to: '/site-admin/batch-changes/specs',
            condition: props => props.batchChangesExecutionEnabled,
        },
    ],
    condition: ({ batchChangesEnabled, isCodyApp }) => batchChangesEnabled && !isCodyApp,
}

const businessGroup: SiteAdminSideBarGroup = {
    header: { label: 'Business', icon: BriefcaseIcon },
    items: [
        {
            label: 'Customers',
            to: '/site-admin/dotcom/customers',
            condition: () => SHOW_BUSINESS_FEATURES,
        },
        {
            label: 'Subscriptions',
            to: '/site-admin/dotcom/product/subscriptions',
            condition: () => SHOW_BUSINESS_FEATURES,
        },
        {
            label: 'License key lookup',
            to: '/site-admin/dotcom/product/licenses',
            condition: () => SHOW_BUSINESS_FEATURES,
        },
    ],
    condition: () => SHOW_BUSINESS_FEATURES,
}

const codeIntelGroup: SiteAdminSideBarGroup = {
    header: { label: 'Code graph', icon: BrainIcon },
    condition: ({ isCodyApp }) => !isCodyApp,
    items: [
        {
            to: '/site-admin/code-graph/dashboard',
            label: 'Dashboard',
        },
        {
            to: '/site-admin/code-graph/indexes',
            label: 'Precise indexes',
        },
        {
            to: '/site-admin/code-graph/configuration',
            label: 'Configuration',
        },
        {
            to: '/site-admin/code-graph/inference-configuration',
            label: 'Inference',
            condition: () => window.context?.codeIntelAutoIndexingEnabled,
        },
        {
            to: '/site-admin/code-graph/ranking',
            label: 'Ranking',
        },
        {
            label: 'Ownership signals',
            to: '/site-admin/own-signal-page',
            condition: ({ isCodyApp }) => !isCodyApp,
        },
    ],
}

export const codyGroup: SiteAdminSideBarGroup = {
    header: { label: 'Cody', icon: CodyPageIcon },
    items: [
        {
            label: 'Embeddings jobs',
            to: '/site-admin/embeddings',
            exact: true,
            condition: () => window.context?.embeddingsEnabled,
        },
        {
            label: 'Embeddings policies',
            to: '/site-admin/embeddings/configuration',
            condition: () => window.context?.embeddingsEnabled,
        },
    ],
    condition: () => window.context?.codyEnabled,
}

const usersGroup: SiteAdminSideBarGroup = {
    ...ossUsersGroup,
    items: [
        ...ossUsersGroup.items,
        {
            label: 'Roles',
            to: '/site-admin/roles',
        },
        {
            label: 'Permissions',
            to: '/site-admin/permissions-syncs',
        },
    ],
}

const repositoriesGroup: SiteAdminSideBarGroup = {
    ...ossRepositoriesGroup,
    items: [
        {
            label: 'GitHub Apps',
            to: '/site-admin/github-apps',
        },
        ...ossRepositoriesGroup.items,
    ],
}

export const enterpriseSiteAdminSidebarGroups: SiteAdminSideBarGroups = [
    analyticsGroup,
    configurationGroup,
    repositoriesGroup,
    codeIntelGroup,
    codyGroup,
    usersGroup,
    executorsGroup,
    maintenanceGroup,
    batchChangesGroup,
    businessGroup,
    apiConsoleGroup,
].filter(Boolean) as SiteAdminSideBarGroups
