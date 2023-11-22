import CogOutlineIcon from 'mdi-react/CogOutlineIcon'

import type { RepoSettingsSideBarGroups } from './RepoSettingsSidebar'

export const settingsGroup = {
    header: { label: 'Settings', icon: CogOutlineIcon },
    items: [
        {
            to: '',
            exact: true,
            label: 'Mirroring',
        },
        {
            to: '/index',
            exact: true,
            label: 'Search Indexing',
        },
    ],
}

export const repoSettingsSideBarGroups: RepoSettingsSideBarGroups = [settingsGroup]
