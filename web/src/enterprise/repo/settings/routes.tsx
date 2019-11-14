import * as React from 'react'
import { RepoSettingsCodeIntelligencePage } from './RepoSettingsCodeIntelligencePage'
import { RepoSettingsAreaRoute } from '../../../repo/settings/RepoSettingsArea'
import { repoSettingsAreaRoutes } from '../../../repo/settings/routes'

export const enterpriseRepoSettingsAreaRoutes: readonly RepoSettingsAreaRoute[] = [
    ...repoSettingsAreaRoutes,
    {
        path: '/code-intelligence',
        exact: true,
        render: props => <RepoSettingsCodeIntelligencePage {...props} />,
    },
]
