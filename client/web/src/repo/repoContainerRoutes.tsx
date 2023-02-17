import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { ActionItemsBarProps } from '../extensions/components/ActionItemsBar'

import { RepoRevisionWrapper } from './components/RepoRevision'
import { RepoContainerRoute } from './RepoContainer'

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
const ActionItemsBar = lazyComponent<ActionItemsBarProps, 'ActionItemsBar'>(
    () => import('../extensions/components/ActionItemsBar'),
    'ActionItemsBar'
)

export const compareSpecPath = '/-/compare/*'

export const repoContainerRoutes: readonly RepoContainerRoute[] = [
    {
        path: '/-/commit/:revspec',
        render: context => (
            <RepoRevisionWrapper>
                <RepositoryCommitPage {...context} />
                {window.context.enableLegacyExtensions && (
                    <ActionItemsBar
                        extensionsController={context.extensionsController}
                        platformContext={context.platformContext}
                        useActionItemsBar={context.useActionItemsBar}
                        telemetryService={context.telemetryService}
                        source="commit"
                    />
                )}
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
        path: compareSpecPath,
        render: context => (
            <RepoRevisionWrapper>
                <RepositoryCompareArea {...context} />
                {window.context.enableLegacyExtensions && (
                    <ActionItemsBar
                        extensionsController={context.extensionsController}
                        platformContext={context.platformContext}
                        useActionItemsBar={context.useActionItemsBar}
                        telemetryService={context.telemetryService}
                        source="compare"
                    />
                )}
            </RepoRevisionWrapper>
        ),
    },
    {
        path: '/-/stats/contributors',
        render: context => <RepositoryStatsArea {...context} />,
    },
]
