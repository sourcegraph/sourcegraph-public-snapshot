import React from 'react'

import { RepoContainerRoute } from '../../repo/RepoContainer'
import { RepoRevisionContainerRoute } from '../../repo/RepoRevisionContainer'
import { repoContainerRoutes, repoRevisionContainerRoutes } from '../../repo/routes'
import { lazyComponent } from '../../util/lazyComponent'

const ContextArea = lazyComponent(() => import('../context/ContextArea'), 'ContextArea')

export const enterpriseRepoContainerRoutes: readonly RepoContainerRoute[] = repoContainerRoutes

export const enterpriseRepoRevisionContainerRoutes: readonly RepoRevisionContainerRoute[] = [
    ...repoRevisionContainerRoutes,
    {
        path: '/-/ctx',
        render: context => <ContextArea {...context} />,
    },
]
