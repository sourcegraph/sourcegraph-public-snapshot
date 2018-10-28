import React from 'react'
import { RepoRevContainerRoute } from '../../../src/repo/RepoRevContainer'
import { repoRevContainerRoutes } from '../../../src/repo/routes'
import { RepositoryGraphArea } from './graph/RepositoryGraphArea'

export const enterpriseRepoRevContainerRoutes: ReadonlyArray<RepoRevContainerRoute> = [
    ...repoRevContainerRoutes,
    {
        path: `/-/graph`,
        render: context => (
            <RepositoryGraphArea
                {...context}
                defaultBranch={context.resolvedRev.defaultBranch}
                commitID={context.resolvedRev.commitID}
                repoHeaderContributionsLifecycleProps={context.repoHeaderContributionsLifecycleProps}
            />
        ),
    },
]
