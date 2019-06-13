import FeatureSearchOutlineIcon from 'mdi-react/FeatureSearchOutlineIcon'
import SettingsIcon from 'mdi-react/SettingsIcon'
import { NAMESPACE_AREA_HEADER_LINKS } from '../../namespaces/NamespaceAreaHeaderLinks'
import { UserAreaHeaderNavItem } from './UserAreaHeader'

export const userAreaHeaderNavItems: ReadonlyArray<UserAreaHeaderNavItem> = [
    {
        to: '',
        exact: true,
        label: 'Profile',
    },
    {
        to: '/settings',
        label: 'Settings',
        icon: SettingsIcon,
        condition: ({ user: { viewerCanAdminister } }) => viewerCanAdminister,
    },
    {
        to: '/searches',
        label: 'Saved searches',
        icon: FeatureSearchOutlineIcon,
        condition: ({ user: { viewerCanAdminister } }) => viewerCanAdminister,
    },
    ...NAMESPACE_AREA_HEADER_LINKS,
]
