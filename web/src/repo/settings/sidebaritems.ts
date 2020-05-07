import { RepoSettingsSideBarItems } from './RepoSettingsSidebar'

export const repoSettingsSidebarItems: RepoSettingsSideBarItems = [
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
]
