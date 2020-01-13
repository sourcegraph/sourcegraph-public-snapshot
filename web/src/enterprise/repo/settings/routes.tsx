import * as React from 'react'
import { RepoSettingsCodeIntelligencePage } from './RepoSettingsCodeIntelligencePage'
import { RepoSettingsAreaRoute } from '../../../repo/settings/RepoSettingsArea'
import { repoSettingsAreaRoutes } from '../../../repo/settings/routes'
import { RepoSettingsLsifUploadPage } from './RepoSettingsLsifUploadPage'

export const enterpriseRepoSettingsAreaRoutes: readonly RepoSettingsAreaRoute[] = [
    ...repoSettingsAreaRoutes,
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
