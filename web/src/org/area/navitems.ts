import FeatureSearchOutlineIcon from 'mdi-react/FeatureSearchOutlineIcon'
import SettingsIcon from 'mdi-react/SettingsIcon'
import UserIcon from 'mdi-react/UserIcon'
import { OrgAreaHeaderNavItem } from './OrgHeader'

export const orgAreaHeaderNavItems: ReadonlyArray<OrgAreaHeaderNavItem> = [
    {
        to: '',
        exact: true,
        label: 'Overview',
    },
    {
        to: '/members',
        label: 'Members',
        icon: UserIcon,
    },
    {
        to: '/settings',
        label: 'Settings',
        icon: SettingsIcon,
        condition: ({ org: { viewerCanAdminister } }) => viewerCanAdminister,
    },
    {
        to: '/searches',
        label: 'Saved searches',
        icon: FeatureSearchOutlineIcon,
        condition: ({ org: { viewerCanAdminister } }) => viewerCanAdminister,
    },
]
