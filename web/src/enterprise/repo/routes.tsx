import React from 'react'
import { RepoContainerRoute } from '../../repo/RepoContainer'
import { RepoRevContainerRoute } from '../../repo/RepoRevContainer'
import { repoContainerRoutes, repoRevContainerRoutes } from '../../repo/routes'
import { RepositoryChecksArea } from '../checks/repo/RepositoryChecksArea'

export const enterpriseRepoContainerRoutes: ReadonlyArray<RepoContainerRoute> = [
    ...repoContainerRoutes,
    {
        path: '/-/checks',
        render: ({ repoHeaderContributionsLifecycleProps, ...context }) => (
            <RepositoryChecksArea
                {...context}
                repoHeaderContributionsLifecycleProps={repoHeaderContributionsLifecycleProps}
            />
        ),
    },
]

export const enterpriseRepoRevContainerRoutes: ReadonlyArray<RepoRevContainerRoute> = repoRevContainerRoutes
