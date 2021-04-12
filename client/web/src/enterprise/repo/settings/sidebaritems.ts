import BrainIcon from 'mdi-react/BrainIcon'

import { RepoSettingsSideBarGroups } from '../../../repo/settings/RepoSettingsSidebar'
import { repoSettingsSideBarGroups, settingsGroup } from '../../../repo/settings/sidebaritems'

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
            condition: () => Boolean(window.context?.codeIntelAutoIndexingEnabled),
        },
        {
            to: '/code-intelligence/index-configuration',
            label: 'Index configuration',
            condition: () => Boolean(window.context?.codeIntelAutoIndexingEnabled),
        },
    ],
}

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
                    ],
                },
                // Insert code intel group after settings group
                codeIntelSettingsGroup,
            ]
        }

        return [...enterpriseGroups, group]
    },
    []
)
