import type { RepoSettingsSideBarGroups } from '../../../repo/settings/RepoSettingsSidebar'
import { repoSettingsSideBarGroups, settingsGroup } from '../../../repo/settings/sidebaritems'

export const enterpriseRepoSettingsSidebarGroups: RepoSettingsSideBarGroups =
    repoSettingsSideBarGroups.reduce<RepoSettingsSideBarGroups>((enterpriseGroups, group) => {
        if (group === settingsGroup) {
            return [
                ...enterpriseGroups,
                // Extend settings group items
                {
                    ...group,
                    items: [
                        ...group.items,
                        {
                            to: '/logs',
                            exact: true,
                            label: 'Logs',
                        },
                        {
                            to: '/permissions',
                            exact: true,
                            label: 'Repo Permissions',
                        },
                    ],
                },
            ]
        }

        return [...enterpriseGroups, group]
    }, [])
