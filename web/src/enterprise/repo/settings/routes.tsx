import * as React from 'react'
import { RepoSettingsPermissionsPage } from './RepoSettingsPermissionsPage'
import { CodeIntelIndexesPage } from '../../codeintel/CodeIntelIndexesPage'
import { CodeIntelIndexPage } from '../../codeintel/CodeIntelIndexPage'
import { CodeIntelUploadPage } from '../../codeintel/CodeIntelUploadPage'
import { CodeIntelUploadsPage } from '../../codeintel/CodeIntelUploadsPage'
import { Redirect, RouteComponentProps } from 'react-router'
import { repoSettingsAreaRoutes } from '../../../repo/settings/routes'
import { RepoSettingsAreaRoute } from '../../../repo/settings/RepoSettingsArea'

export const enterpriseRepoSettingsAreaRoutes: readonly RepoSettingsAreaRoute[] = [
    ...repoSettingsAreaRoutes,
    {
        path: '/permissions',
        exact: true,
        render: props => <RepoSettingsPermissionsPage {...props} />,
    },
    {
        path: '/code-intelligence/uploads',
        exact: true,
        render: props => <CodeIntelUploadsPage {...props} />,
    },
    {
        path: '/code-intelligence/uploads/:id',
        exact: true,
        render: props => <CodeIntelUploadPage {...props} />,
    },
    {
        path: '/code-intelligence/indexes',
        exact: true,
        render: props => <CodeIntelIndexesPage {...props} />,
    },
    {
        path: '/code-intelligence/indexes/:id',
        exact: true,
        render: props => <CodeIntelIndexPage {...props} />,
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
]
