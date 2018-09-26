import { authGroup, otherGroup, siteAdminSidebarGroups } from '@sourcegraph/webapp/dist/site-admin/sidebaritems'
import { SiteAdminSideBarGroup, SiteAdminSideBarGroups } from '@sourcegraph/webapp/dist/site-admin/SiteAdminSidebar'
import PuzzleIcon from 'mdi-react/PuzzleIcon'

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
    header: { label: 'Sourcegraph.com' },
    items: [
        {
            label: 'Generate license',
            to: '/site-admin/dotcom/generate-license',
        },
    ],
    condition: () => (window as any).context.sourcegraphDotComMode,
}

export const enterpriseSiteAdminSidebarGroups: SiteAdminSideBarGroups = siteAdminSidebarGroups.reduce<
    SiteAdminSideBarGroups
>((enterpriseGroups, group) => {
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
                        label: 'External Accounts',
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
