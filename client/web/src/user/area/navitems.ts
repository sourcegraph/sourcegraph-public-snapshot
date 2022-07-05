import { mdiCogOutline, mdiFeatureSearchOutline } from '@mdi/js'

import { namespaceAreaHeaderNavItems } from '../../namespaces/navitems'

import { UserAreaHeaderNavItem } from './UserAreaHeader'

export const userAreaHeaderNavItems: readonly UserAreaHeaderNavItem[] = [
    {
        to: '/settings',
        label: 'Settings',
        icon: mdiCogOutline,
        condition: ({ user: { viewerCanAdminister } }) => viewerCanAdminister,
    },
    {
        to: '/searches',
        label: 'Saved searches',
        icon: mdiFeatureSearchOutline,
        condition: ({ user: { viewerCanAdminister } }) => viewerCanAdminister,
    },
    ...namespaceAreaHeaderNavItems,
]
