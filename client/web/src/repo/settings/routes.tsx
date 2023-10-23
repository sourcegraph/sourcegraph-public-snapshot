import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { RedirectRoute } from '../../components/RedirectRoute'

import type { RepoSettingsAreaRoute } from './RepoSettingsArea'

export const repoSettingsAreaPath = '/-/settings/*'

export const repoSettingsAreaRoutes: readonly RepoSettingsAreaRoute[] = [
    {
        path: '',
        render: lazyComponent(() => import('./RepoSettingsMirrorPage'), 'RepoSettingsMirrorPage'),
    },
    {
        path: '/index',
        render: lazyComponent(() => import('./RepoSettingsIndexPage'), 'RepoSettingsIndexPage'),
    },
    {
        path: '/mirror',
        // The /mirror page used to be separate but we combined this one and the
        // '' route above, so we redirect here in case people still link to this
        // page.
        render: () => <RedirectRoute getRedirectURL={() => '..'} />,
    },
]
