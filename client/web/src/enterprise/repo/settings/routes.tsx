import * as React from 'react'
import { Redirect, RouteComponentProps } from 'react-router'
import { RepoSettingsAreaRoute } from '../../../repo/settings/RepoSettingsArea'
import { repoSettingsAreaRoutes } from '../../../repo/settings/routes'
import { lazyComponent } from '../../../util/lazyComponent'
import { CodeIntelIndexConfigurationPageProps } from '../../codeintel/configuration/CodeIntelIndexConfigurationPage'
import { CodeIntelIndexPageProps } from '../../codeintel/detail/CodeIntelIndexPage'
import { CodeIntelUploadPageProps } from '../../codeintel/detail/CodeIntelUploadPage'
import { CodeIntelIndexesPageProps } from '../../codeintel/list/CodeIntelIndexesPage'
import { CodeIntelUploadsPageProps } from '../../codeintel/list/CodeIntelUploadsPage'
import { RepoSettingsPermissionsPageProps } from './RepoSettingsPermissionsPage'

const RepoSettingsPermissionsPage = lazyComponent<RepoSettingsPermissionsPageProps, 'RepoSettingsPermissionsPage'>(
    () => import('./RepoSettingsPermissionsPage'),
    'RepoSettingsPermissionsPage'
)
const CodeIntelUploadsPage = lazyComponent<CodeIntelUploadsPageProps, 'CodeIntelUploadsPage'>(
    () => import('../../codeintel/list/CodeIntelUploadsPage'),
    'CodeIntelUploadsPage'
)
const CodeIntelUploadPage = lazyComponent<CodeIntelUploadPageProps, 'CodeIntelUploadPage'>(
    () => import('../../codeintel/detail/CodeIntelUploadPage'),
    'CodeIntelUploadPage'
)
const CodeIntelIndexConfigurationPage = lazyComponent<
    CodeIntelIndexConfigurationPageProps,
    'CodeIntelIndexConfigurationPage'
>(() => import('../../codeintel/configuration/CodeIntelIndexConfigurationPage'), 'CodeIntelIndexConfigurationPage')

const CodeIntelIndexesPage = lazyComponent<CodeIntelIndexesPageProps, 'CodeIntelIndexesPage'>(
    () => import('../../codeintel/list/CodeIntelIndexesPage'),
    'CodeIntelIndexesPage'
)
const CodeIntelIndexPage = lazyComponent<CodeIntelIndexPageProps, 'CodeIntelIndexPage'>(
    () => import('../../codeintel/detail/CodeIntelIndexPage'),
    'CodeIntelIndexPage'
)

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
        path: '/code-intelligence/index-configuration',
        exact: true,
        render: props => <CodeIntelIndexConfigurationPage {...props} />,
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
