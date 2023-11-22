import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { canWriteRepoMetadata } from '../util/rbac'

import { RepositoryChangelistPage } from './commit/RepositoryCommitPage'
import { RepoRevisionWrapper } from './components/RepoRevision'
import type { RepoContainerRoute } from './RepoContainer'

const RepositoryCommitPage = lazyComponent(() => import('./commit/RepositoryCommitPage'), 'RepositoryCommitPage')
const RepositoryBranchesArea = lazyComponent(
    () => import('./branches/RepositoryBranchesArea'),
    'RepositoryBranchesArea'
)

const RepositoryReleasesArea = lazyComponent(
    () => import('./releases/RepositoryReleasesArea'),
    'RepositoryReleasesArea'
)
const RepositoryCompareArea = lazyComponent(() => import('./compare/RepositoryCompareArea'), 'RepositoryCompareArea')
const RepositoryStatsArea = lazyComponent(() => import('./stats/RepositoryStatsArea'), 'RepositoryStatsArea')

export const compareSpecPath = '/-/compare/*'

const RepositoryMetadataPage = lazyComponent(() => import('./RepoMetadataPage'), 'RepoMetadataPage')

export const repoContainerRoutes: readonly RepoContainerRoute[] = [
    {
        path: '/-/commit/:revspec',
        render: context => (
            <RepoRevisionWrapper>
                <RepositoryCommitPage {...context} />
            </RepoRevisionWrapper>
        ),
    },
    {
        path: '/-/changelist/:changelistID',
        render: context => (
            <RepoRevisionWrapper>
                <RepositoryChangelistPage {...context} />
            </RepoRevisionWrapper>
        ),
    },
    {
        path: '/-/branches/*',
        render: context => <RepositoryBranchesArea {...context} />,
    },
    {
        path: '/-/tags',
        render: context => <RepositoryReleasesArea {...context} />,
    },
    {
        path: '/-/versions',
        render: context => <RepositoryReleasesArea {...context} />,
    },
    {
        path: compareSpecPath,
        render: context => (
            <RepoRevisionWrapper>
                <RepositoryCompareArea {...context} />
            </RepoRevisionWrapper>
        ),
    },
    {
        path: '/-/stats/contributors',
        render: context => <RepositoryStatsArea {...context} />,
    },
    {
        path: '/-/metadata',
        condition: ({ authenticatedUser }) => canWriteRepoMetadata(authenticatedUser),
        render: context => <RepositoryMetadataPage {...context} />,
    },
]
