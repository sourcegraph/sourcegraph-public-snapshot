import * as React from 'react'
import { RepoSettingsCodeIntelIndexesPage } from './RepoSettingsCodeIntelIndexesPage'
import { RepoSettingsAreaRoute } from '../../../repo/settings/RepoSettingsArea'
import { repoSettingsAreaRoutes } from '../../../repo/settings/routes'
import { RepoSettingsCodeIntelUploadPage } from './RepoSettingsCodeIntelUploadPage'
import { RepoSettingsCodeIntelIndexPage } from './RepoSettingsCodeIntelIndexPage'
import { RepoSettingsPermissionsPage } from './RepoSettingsPermissionsPage'
import { RepoSettingsCodeIntelUploadsPage } from './RepoSettingsCodeIntelUploadsPage'
import { Redirect, RouteComponentProps } from 'react-router'

export const enterpriseRepoSettingsAreaRoutes: readonly RepoSettingsAreaRoute[] = [
    ...repoSettingsAreaRoutes,
    {
        path: '/permissions',
        exact: true,
        render: props => <RepoSettingsPermissionsPage {...props} />,
        condition: () => !!window.context.site['permissions.backgroundSync']?.enabled,
    },
    {
        path: '/code-intelligence/uploads',
        exact: true,
        render: props => <RepoSettingsCodeIntelUploadsPage {...props} />,
    },
    {
        path: '/code-intelligence/lsif-uploads/:id',
        exact: true,
        render: ({
            match: {
                params: { id },
            },
        }: RouteComponentProps<{ id: string }>) => <Redirect to={`../uploads/${id}`} />,
    },
    {
        path: '/code-intelligence/uploads/:id',
        exact: true,
        render: props => <RepoSettingsCodeIntelUploadPage {...props} />,
    },
    {
        path: '/code-intelligence/indexes',
        exact: true,
        render: props => <RepoSettingsCodeIntelIndexesPage {...props} />,
    },
    {
        path: '/code-intelligence/indexes/:id',
        exact: true,
        render: props => <RepoSettingsCodeIntelIndexPage {...props} />,
    },
]
