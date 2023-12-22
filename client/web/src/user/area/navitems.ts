import AccountIcon from 'mdi-react/AccountIcon'
import CogOutlineIcon from 'mdi-react/CogOutlineIcon'
import FeatureSearchOutlineIcon from 'mdi-react/FeatureSearchOutlineIcon'

import { namespaceAreaHeaderNavItems } from '../../namespaces/navitems'

import type { UserAreaHeaderNavItem } from './UserAreaHeader'

export const userAreaHeaderNavItems: readonly UserAreaHeaderNavItem[] = [
    {
        to: '/profile',
        label: 'Profile',
        icon: AccountIcon,
    },
    {
        to: '/settings',
        label: 'Settings',
        icon: CogOutlineIcon,
        condition: ({ user: { viewerCanAdminister } }) => viewerCanAdminister,
    },
    {
        to: '/searches',
        label: 'Saved searches',
        icon: FeatureSearchOutlineIcon,
        condition: ({ user: { viewerCanAdminister } }) => viewerCanAdminister,
    },
    ...namespaceAreaHeaderNavItems,
]
