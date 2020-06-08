import { RepoSettingsSideBarGroups } from '../../../repo/settings/RepoSettingsSidebar'
import { repoSettingsSideBarGroups } from '../../../repo/settings/sidebaritems'
import BrainIcon from 'mdi-react/BrainIcon'

export const enterpriseRepoSettingsSidebarGroups: RepoSettingsSideBarGroups = [
    ...repoSettingsSideBarGroups,
    {
        header: { label: 'Code intelligence', icon: BrainIcon },
        items: [
            {
                to: '/code-intelligence/uploads',
                label: 'Uploads',
            },
            {
                to: '/code-intelligence/indexes',
                label: 'Auto indexing',
            },
        ],
    },
    {
        to: '/permissions',
        exact: true,
        label: 'Permissions',
        condition: () => !!window.context.site['permissions.backgroundSync']?.enabled,
    },
]
