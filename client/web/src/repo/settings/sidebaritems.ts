import { mdiCogOutline } from '@mdi/js'

import { RepoSettingsSideBarGroups } from './RepoSettingsSidebar'

export const settingsGroup = {
    header: { label: 'Settings', icon: mdiCogOutline },
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
    ],
}

export const repoSettingsSideBarGroups: RepoSettingsSideBarGroups = [settingsGroup]
