import {
    siteAdminGroup as ossSiteAdminGroup,
    maintenanceGroup,
    repositoriesGroup,
    usersGroup,
} from '../../site-admin/sidebaritems'
import { SiteAdminSideBarGroup, SiteAdminSideBarGroups } from '../../site-admin/SiteAdminSidebar'
import { SHOW_BUSINESS_FEATURES } from '../dotcom/productSubscriptions/features'

const siteAdminGroup: SiteAdminSideBarGroup = {
    ...ossSiteAdminGroup,
    items: [
        ...ossSiteAdminGroup.items,
        {
            label: 'License',
            to: '/site-admin/license',
        },
    ],
}

const featuresGroup: SiteAdminSideBarGroup = {
    header: { label: 'Features' },
    items: [
        {
            label: 'Batch Changes',
            to: '/site-admin/batch-changes',
            condition: ({ batchChangesEnabled }) => batchChangesEnabled,
        },
        {
            label: 'Extensions',
            to: '/site-admin/registry/extensions',
        },
        {
            label: 'Code intelligence',
            to: '/site-admin/code-intelligence/uploads',
        },
        {
            label: 'Auto indexing',
            to: '/site-admin/code-intelligence/indexes',
            condition: () => Boolean(window.context?.codeIntelAutoIndexingEnabled),
        },
    ],
}

const businessGroup: SiteAdminSideBarGroup = {
    header: { label: 'Business' },
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

export const enterpriseSiteAdminSidebarGroups: SiteAdminSideBarGroups = [
    siteAdminGroup,
    repositoriesGroup,
    usersGroup,
    featuresGroup,
    maintenanceGroup,
    businessGroup,
]
