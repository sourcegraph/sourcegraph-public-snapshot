import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { RepoSettingsAreaRoute } from './RepoSettingsArea'

export const repoSettingsAreaPath = '/-/settings/*'

export const repoSettingsAreaRoutes: readonly RepoSettingsAreaRoute[] = [
    {
        path: '',
        render: lazyComponent(() => import('./RepoSettingsOptionsPage'), 'RepoSettingsOptionsPage'),
    },
    {
        path: '/index',
        render: lazyComponent(() => import('./RepoSettingsIndexPage'), 'RepoSettingsIndexPage'),
    },
    {
        path: '/mirror',
        render: lazyComponent(() => import('./RepoSettingsMirrorPage'), 'RepoSettingsMirrorPage'),
    },
]
