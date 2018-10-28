import { RepoRevContainerRoute } from '@sourcegraph/webapp/dist/repo/RepoRevContainer'
import { repoRevContainerRoutes } from '@sourcegraph/webapp/dist/repo/routes'
import React from 'react'
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
