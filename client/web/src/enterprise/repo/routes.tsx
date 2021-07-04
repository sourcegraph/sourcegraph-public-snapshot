import React from 'react'

import { RepoContainerRoute } from '../../repo/RepoContainer'
import { RepoRevisionContainerRoute } from '../../repo/RepoRevisionContainer'
import { repoContainerRoutes, repoRevisionContainerRoutes } from '../../repo/routes'
import { lazyComponent } from '../../util/lazyComponent'

const GuideArea = lazyComponent(() => import('../guide/GuideArea'), 'GuideArea')

const UsageArea = lazyComponent(() => import('../usage/UsageArea'), 'UsageArea')

export const enterpriseRepoContainerRoutes: readonly RepoContainerRoute[] = repoContainerRoutes

export const enterpriseRepoRevisionContainerRoutes: readonly RepoRevisionContainerRoute[] = [
    ...repoRevisionContainerRoutes,
    {
        path: '/-/guide',
        render: GuideArea,
    },
    {
        path: '/-/usage',
        render: UsageArea,
    },
]
