import { RepoSettingsSideBarGroups } from '../../../repo/settings/RepoSettingsSidebar'
import { repoSettingsSideBarGroups, settingsGroup } from '../../../repo/settings/sidebaritems'

export const enterpriseRepoSettingsSidebarGroups: RepoSettingsSideBarGroups = repoSettingsSideBarGroups.reduce<RepoSettingsSideBarGroups>(
    (enterpriseGroups, group) => {
        if (group === settingsGroup) {
            return [
                ...enterpriseGroups,
                // Extend settings group items
                {
                    ...group,
                    items: [
                        ...group.items,
                        {
                            to: '/permissions',
                            exact: true,
                            label: 'Permissions',
                        },
                        {
                            to: '/code-intelligence',
                            exact: true,
                            label: 'Code intelligence',
                        },
                    ],
                },
            ]
        }

        return [...enterpriseGroups, group]
    },
    []
)
