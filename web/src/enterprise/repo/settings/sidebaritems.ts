import { RepoSettingsSideBarGroups } from '../../../repo/settings/RepoSettingsSidebar'
import { repoSettingsSideBarGroups, settingsGroup } from '../../../repo/settings/sidebaritems'
import BrainIcon from 'mdi-react/BrainIcon'

const codeIntelSettingsGroup = {
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
}

export const enterpriseRepoSettingsSidebarGroups: RepoSettingsSideBarGroups = repoSettingsSideBarGroups.reduce<
    RepoSettingsSideBarGroups
>((enterpriseGroups, group) => {
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
                        condition: () => !!window.context.site['permissions.backgroundSync']?.enabled,
                    },
                ],
            },
            // Insert code intel group after settings group
            codeIntelSettingsGroup,
        ]
    }

    return [...enterpriseGroups, group]
}, [])
