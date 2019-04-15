import SettingsIcon from 'mdi-react/SettingsIcon'
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
]
