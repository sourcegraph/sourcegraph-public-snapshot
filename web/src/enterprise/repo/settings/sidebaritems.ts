import { RepoSettingsSideBarItems } from '../../../repo/settings/RepoSettingsSidebar'
import { repoSettingsSidebarItems } from '../../../repo/settings/sidebaritems'

export const enterpriseRepoSettingsSidebarItems: RepoSettingsSideBarItems = [
    ...repoSettingsSidebarItems,
    {
        to: '/code-intelligence',
        exact: true,
        label: 'Code intelligence',
    },
]
