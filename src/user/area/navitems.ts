import SettingsIcon from 'mdi-react/SettingsIcon'
import TuneVerticalIcon from 'mdi-react/TuneVerticalIcon'
import { UserAreaHeaderNavItem } from './UserAreaHeader'

export const userAreaHeaderNavItems: ReadonlyArray<UserAreaHeaderNavItem> = [
    {
        to: '',
        exact: true,
        label: 'Profile',
    },
    {
        to: '/settings',
        exact: true,
        label: 'Settings',
        icon: SettingsIcon,
        condition: ({ user: { viewerCanAdminister } }) => viewerCanAdminister,
    },
    {
        to: '/account',
        label: 'Account',
        icon: TuneVerticalIcon,
        condition: ({ user: { viewerCanAdminister } }) => viewerCanAdminister,
    },
]
