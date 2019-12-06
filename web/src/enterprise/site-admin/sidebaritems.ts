import HeartIcon from 'mdi-react/HeartIcon'
import PuzzleIcon from 'mdi-react/PuzzleIcon'
import { primaryGroup, authGroup, otherGroup, siteAdminSidebarGroups } from '../../site-admin/sidebaritems'
import { SiteAdminSideBarGroup, SiteAdminSideBarGroups } from '../../site-admin/SiteAdminSidebar'
import { SHOW_BUSINESS_FEATURES } from '../dotcom/productSubscriptions/features'

const registryGroup: SiteAdminSideBarGroup = {
    header: {
        label: 'Registry',
        icon: PuzzleIcon,
    },
    items: [
        {
            label: 'Extensions',
            to: '/site-admin/registry/extensions',
        },
    ],
}

/**
 * Sidebar items that are only used on Sourcegraph.com.
 */
const dotcomGroup: SiteAdminSideBarGroup = {
    header: { label: 'Business', icon: HeartIcon },
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

export const enterpriseSiteAdminSidebarGroups: SiteAdminSideBarGroups = siteAdminSidebarGroups.reduce<
    SiteAdminSideBarGroups
>((enterpriseGroups, group) => {
    if (group === primaryGroup) {
        return [
            ...enterpriseGroups,
            {
                ...group,
                // Add LSIF jobs page to end of primary group
                items: [
                    ...group.items,
                    {
                        label: 'LSIF jobs',
                        to: '/site-admin/lsif-jobs',
                    },
                ],
            },
        ]
    }
    if (group === authGroup) {
        return [
            ...enterpriseGroups,
            // Extend auth group items
            {
                ...group,
                items: [
                    {
                        label: 'Providers',
                        to: '/site-admin/auth/providers',
                    },
                    {
                        label: 'External accounts',
                        to: '/site-admin/auth/external-accounts',
                    },
                    ...group.items,
                ],
            },
            // Insert registry group after auth group
            registryGroup,
        ]
    }
    if (group === otherGroup) {
        return [
            ...enterpriseGroups,
            // Insert dotcom group before other group (on Sourcegraph.com)
            dotcomGroup,
            // Extend other group items
            {
                ...group,
                items: [
                    {
                        label: 'License',
                        to: '/site-admin/license',
                    },
                    ...group.items,
                ],
            },
        ]
    }
    return [...enterpriseGroups, group]
}, [])
