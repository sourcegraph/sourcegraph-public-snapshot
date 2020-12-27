import React from 'react'
import { RepoContainerRoute } from '../../repo/RepoContainer'
import { RepoRevisionContainerRoute } from '../../repo/RepoRevisionContainer'
import { repoContainerRoutes, repoRevisionContainerRoutes } from '../../repo/routes'
import { lazyComponent } from '../../util/lazyComponent'

const SymbolsArea = lazyComponent(() => import('../symbols/SymbolsArea'), 'SymbolsArea')
const RepositoryDependenciesPage = lazyComponent(
    () => import('./network/RepositoryDependenciesPage'),
    'RepositoryDependenciesPage'
)

export const enterpriseRepoContainerRoutes: readonly RepoContainerRoute[] = repoContainerRoutes

export const enterpriseRepoRevisionContainerRoutes: readonly RepoRevisionContainerRoute[] = [
    ...repoRevisionContainerRoutes,
    {
        path: '/-/symbols',
        render: context => <SymbolsArea {...context} />,
    },
    {
        path: '/-/dependencies',
        render: context => <RepositoryDependenciesPage {...context} />,
    },
]
