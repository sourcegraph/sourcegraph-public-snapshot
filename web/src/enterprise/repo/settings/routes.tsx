import * as React from 'react'
import { RepoSettingsCodeIntelligencePage } from './RepoSettingsCodeIntelligencePage'
import { RepoSettingsAreaRoute } from '../../../repo/settings/RepoSettingsArea'
import { repoSettingsAreaRoutes } from '../../../repo/settings/routes'
import { RepoSettingsLsifUploadPage } from './RepoSettingsLsifUploadPage'
import { RepoSettingsPermissionsPage } from './RepoSettingsPermissionsPage'

export const enterpriseRepoSettingsAreaRoutes: readonly RepoSettingsAreaRoute[] = [
    ...repoSettingsAreaRoutes,
    {
        path: '/permissions',
        exact: true,
        render: props => <RepoSettingsPermissionsPage {...props} />,
        condition: () => !!window.context.site['permissions.backgroundSync']?.enabled,
    },
    {
        path: '/code-intelligence',
        exact: true,
        render: props => <RepoSettingsCodeIntelligencePage {...props} />,
    },
    {
        path: '/code-intelligence/lsif-uploads/:id',
        exact: true,
        render: props => <RepoSettingsLsifUploadPage {...props} />,
    },
]
