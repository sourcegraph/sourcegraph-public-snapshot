import * as React from 'react'
import { RepoSettingsCodeIntelligencePage } from './RepoSettingsCodeIntelligencePage'
import { RepoSettingsAreaRoute } from './../../../repo/settings/RepoSettingsArea'
import { repoSettingsRoutes } from './../../../repo/settings/routes'

export const enterpriseRepoSettingsRoutes: readonly RepoSettingsAreaRoute[] = [
    ...repoSettingsRoutes,
    {
        path: '/code-intelligence',
        exact: true,
        render: props => <RepoSettingsCodeIntelligencePage {...props} />,
    },
]
