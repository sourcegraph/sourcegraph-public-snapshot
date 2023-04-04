import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { RepoRevisionContainerRoute } from '../../repo/RepoRevisionContainer'
import { createRepoRevisionContainerRoutes } from '../../repo/repoRevisionContainerRoutes'

const EnterpriseRepositoryFileTreePage = lazyComponent(
    () => import('./EnterpriseRepositoryFileTreePage'),
    'EnterpriseRepositoryFileTreePage'
)

export const enterpriseRepoRevisionContainerRoutes: readonly RepoRevisionContainerRoute[] =
    createRepoRevisionContainerRoutes(EnterpriseRepositoryFileTreePage)
