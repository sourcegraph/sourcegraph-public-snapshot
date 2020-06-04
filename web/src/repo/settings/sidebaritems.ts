import { RepoSettingsSideBarGroups } from './RepoSettingsSidebar'
import GearIcon from 'mdi-react/GearIcon'

export const repoSettingsSideBarGroups: RepoSettingsSideBarGroups = [
    {
        header: { label: 'Settings', icon: GearIcon },
        items: [
            {
                to: '',
                exact: true,
                label: 'Options',
            },
            {
                to: '/index',
                exact: true,
                label: 'Indexing',
            },
            {
                to: '/mirror',
                exact: true,
                label: 'Mirroring',
            },
            {
                to: '/permissions',
                exact: true,
                label: 'Permissions',
                condition: () => !!window.context.site['permissions.backgroundSync']?.enabled,
            },
        ],
    },
]
