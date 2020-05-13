import * as React from 'react'
import { RepoSettingsIndexPage } from './RepoSettingsIndexPage'
import { RepoSettingsMirrorPage } from './RepoSettingsMirrorPage'
import { RepoSettingsOptionsPage } from './RepoSettingsOptionsPage'
import { RepoSettingsPermissionsPage } from './RepoSettingsPermissionsPage'
import { RepoSettingsAreaRoute } from './RepoSettingsArea'

export const repoSettingsAreaRoutes: readonly RepoSettingsAreaRoute[] = [
    {
        path: '',
        exact: true,
        render: props => <RepoSettingsOptionsPage {...props} />,
    },
    {
        path: '/index',
        exact: true,
        render: props => <RepoSettingsIndexPage {...props} />,
    },
    {
        path: '/mirror',
        exact: true,
        render: props => <RepoSettingsMirrorPage {...props} />,
    },
    {
        path: '/permissions',
        exact: true,
        render: props => <RepoSettingsPermissionsPage {...props} />,
        condition: () => !!window.context.site['permissions.backgroundSync']?.enabled,
    },
]
