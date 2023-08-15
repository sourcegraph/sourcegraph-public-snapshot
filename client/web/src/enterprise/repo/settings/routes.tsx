import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { RedirectRoute } from '../../../components/RedirectRoute'
import type { RepoSettingsAreaRoute } from '../../../repo/settings/RepoSettingsArea'
import { repoSettingsAreaRoutes } from '../../../repo/settings/routes'

import type { RepoSettingsLogsPageProps } from './RepoSettingsLogsPage'
import type { RepoSettingsPermissionsPageProps } from './RepoSettingsPermissionsPage'

const RepoSettingsPermissionsPage = lazyComponent<RepoSettingsPermissionsPageProps, 'RepoSettingsPermissionsPage'>(
    () => import('./RepoSettingsPermissionsPage'),
    'RepoSettingsPermissionsPage'
)

const RepoSettingsLogsPage = lazyComponent<RepoSettingsLogsPageProps, 'RepoSettingsLogsPage'>(
    () => import('./RepoSettingsLogsPage'),
    'RepoSettingsLogsPage'
)

export const enterpriseRepoSettingsAreaRoutes: readonly RepoSettingsAreaRoute[] = [
    ...repoSettingsAreaRoutes,
    {
        path: '/permissions',
        render: props => <RepoSettingsPermissionsPage {...props} />,
    },
    {
        path: '/logs',
        render: props => <RepoSettingsLogsPage {...props} />,
    },

    // Legacy routes
    {
        path: '/code-intelligence/lsif-uploads/:id',
        render: () => <RedirectRoute getRedirectURL={({ params }) => `../uploads/${params.id}`} />,
    },
    {
        path: '/code-intelligence/*',
        render: () => <RedirectRoute getRedirectURL={({ location }) => location.pathname.replace('/settings/', '/')} />,
    },
]
