import AccountIcon from 'mdi-react/AccountIcon'
import CogOutlineIcon from 'mdi-react/CogOutlineIcon'
import FeatureSearchOutlineIcon from 'mdi-react/FeatureSearchOutlineIcon'

import { CodyIcon } from '../../cody/components/CodyIcon'
import { namespaceAreaHeaderNavItems } from '../../namespaces/navitems'

import { UserAreaHeaderNavItem } from './UserAreaHeader'

export const userAreaHeaderNavItems: readonly UserAreaHeaderNavItem[] = [
    {
        to: '/app-settings',
        label: 'Repositories',
        icon: CodyIcon,
        condition: ({ isSourcegraphApp }) => isSourcegraphApp,
    },
    {
        to: '/profile',
        label: 'Profile',
        icon: AccountIcon,
        condition: ({ isSourcegraphApp }) => !isSourcegraphApp,
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
