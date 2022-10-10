import BrainIcon from 'mdi-react/BrainIcon'
import BriefcaseIcon from 'mdi-react/BriefcaseIcon'
import PuzzleOutlineIcon from 'mdi-react/PuzzleOutlineIcon'

import { BatchChangesIcon } from '../../batches/icons'
import {
    apiConsoleGroup,
    configurationGroup as ossConfigurationGroup,
    maintenanceGroup as ossMaintenanceGroup,
    overviewGroup,
    repositoriesGroup,
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
            to: '/site-admin/executors',
            label: 'Executors',
            condition: () => Boolean(window.context?.executorsEnabled),
        },
    ],
}

const extensionsGroup: SiteAdminSideBarGroup = {
    header: {
        label: 'Extensions',
        icon: PuzzleOutlineIcon,
    },
    items: [
        {
            label: 'Extensions',
            to: '/site-admin/registry/extensions',
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
            to: '/site-admin/code-graph/uploads',
            label: 'Uploads',
        },
        {
            to: '/site-admin/code-graph/indexes',
            label: 'Auto-indexing',
            condition: () => Boolean(window.context?.codeIntelAutoIndexingEnabled),
        },
        {
            to: '/site-admin/code-graph/configuration',
            label: 'Configuration',
        },
        {
            to: '/site-admin/code-graph/inference-configuration',
            label: 'Inference'
        }
    ],
}

export const enterpriseSiteAdminSidebarGroups: SiteAdminSideBarGroups = [
    overviewGroup,
    configurationGroup,
    repositoriesGroup,
    codeIntelGroup,
    usersGroup,
    maintenanceGroup,
    window.context.enableLegacyExtensions ? extensionsGroup : undefined,
    batchChangesGroup,
    businessGroup,
    apiConsoleGroup,
].filter(Boolean) as SiteAdminSideBarGroups
