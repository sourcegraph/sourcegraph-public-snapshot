import FeatureSearchOutlineIcon from 'mdi-react/FeatureSearchOutlineIcon'
import CogOutlineIcon from 'mdi-react/CogOutlineIcon'
import { namespaceAreaHeaderNavItems } from '../../namespaces/navitems'
import { OrgAreaHeaderNavItem } from './OrgHeader'

export const orgAreaHeaderNavItems: readonly OrgAreaHeaderNavItem[] = [
    {
        to: '/settings',
        label: 'Settings',
        icon: CogOutlineIcon,
        condition: ({ org: { viewerCanAdminister } }) => viewerCanAdminister,
    },
    {
        to: '/searches',
        label: 'Saved searches',
        icon: FeatureSearchOutlineIcon,
        condition: ({ org: { viewerCanAdminister } }) => viewerCanAdminister,
    },
    ...namespaceAreaHeaderNavItems,
]
