import FeatureSearchOutlineIcon from 'mdi-react/FeatureSearchOutlineIcon'
import PersonCardDetailsIcon from 'mdi-react/PersonCardDetailsOutlineIcon'
import { namespaceAreaHeaderNavItems } from '../../namespaces/navitems'
import { UserAreaTabsNavItem } from './UserAreaTabs'

export const userAreaHeaderNavItems: readonly UserAreaTabsNavItem[] = [
    {
        to: '',
        exact: true,
        label: 'Overview',
        icon: PersonCardDetailsIcon,
    },
    {
        to: '/searches',
        label: 'Saved searches',
        icon: FeatureSearchOutlineIcon,
        condition: ({ user: { viewerCanAdminister } }) => viewerCanAdminister,
    },
    ...namespaceAreaHeaderNavItems,
]
