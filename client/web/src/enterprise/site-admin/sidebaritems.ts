import BrainIcon from 'mdi-react/BrainIcon'
import BriefcaseIcon from 'mdi-react/BriefcaseIcon'
import PackageVariantIcon from 'mdi-react/PackageVariantIcon'

import { BatchChangesIcon } from '../../batches/icons'
import {
    apiConsoleGroup,
    analyticsGroup,
    configurationGroup as ossConfigurationGroup,
    maintenanceGroup as ossMaintenanceGroup,
    repositoriesGroup as ossRepositoriesGroup,
    usersGroup,
} from '../../site-admin/sidebaritems'
import { SiteAdminSideBarGroup, SiteAdminSideBarGroups } from '../../site-admin/SiteAdminSidebar'
import { SHOW_BUSINESS_FEATURES } from '../dotcom/productSubscriptions/features'

const configurationGroup: SiteAdminSideBarGroup = {
    ...ossConfigurationGroup,
    items: [
        ...ossConfigurationGroup.items,
        {
            label: 'License',
            to: '/site-admin/license',
        },
    ],
}

const maintenanceGroup: SiteAdminSideBarGroup = {
    ...ossMaintenanceGroup,
    items: [
        ...ossMaintenanceGroup.items,
        {
            label: 'Code Insights jobs',
            to: '/site-admin/code-insights-jobs'
        }
    ]
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
        {
            label: 'Incoming webhooks',
            to: '/site-admin/batch-changes/webhook-logs',
            condition: props => props.batchChangesWebhookLogsEnabled,
        },
        {
            label: 'Outgoing webhooks',
            to: '/site-admin/outbound-webhooks',
        },
    ],
    condition: ({ batchChangesEnabled }) => batchChangesEnabled,
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
    items: [
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
        },
    ],
}

const repositoriesGroup: SiteAdminSideBarGroup = {
    ...ossRepositoriesGroup,
    items: [
        ...ossRepositoriesGroup.items,
        {
            label: 'Incoming webhooks',
            to: '/site-admin/webhooks',
        },
    ],
}

export const enterpriseSiteAdminSidebarGroups: SiteAdminSideBarGroups = [
    analyticsGroup,
    configurationGroup,
    repositoriesGroup,
    codeIntelGroup,
    usersGroup,
    executorsGroup,
    maintenanceGroup,
    batchChangesGroup,
    businessGroup,
    apiConsoleGroup,
].filter(Boolean) as SiteAdminSideBarGroups
