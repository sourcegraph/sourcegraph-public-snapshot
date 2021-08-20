import CogOutlineIcon from 'mdi-react/CogOutlineIcon'

import { RepoSettingsSideBarGroups } from './RepoSettingsSidebar'

export const settingsGroup = {
    header: { label: 'Settings', icon: CogOutlineIcon },
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
