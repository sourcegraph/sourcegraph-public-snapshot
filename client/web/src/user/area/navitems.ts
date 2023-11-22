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
        condition: ({ isCodyApp }) => !isCodyApp,
    },
    {
        to: '/settings',
        label: 'Settings',
        icon: CogOutlineIcon,
        condition: ({ user: { viewerCanAdminister }, isCodyApp }) => viewerCanAdminister && !isCodyApp,
    },
    {
        to: '/searches',
        label: 'Saved searches',
        icon: FeatureSearchOutlineIcon,
        condition: ({ user: { viewerCanAdminister }, isCodyApp }) => viewerCanAdminister && !isCodyApp,
    },
    ...namespaceAreaHeaderNavItems,
]
