import { TraceSpanProvider } from '@sourcegraph/observability-client'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import type { RepoRevisionContainerRoute } from './RepoRevisionContainer'

const RepositoryCommitsPage = lazyComponent(() => import('./commits/RepositoryCommitsPage'), 'RepositoryCommitsPage')
const RepositoryFileTreePage = lazyComponent(() => import('./RepositoryFileTreePage'), 'RepositoryFileTreePage')

const routeToObjectType = {
    '/-/blob/*': 'blob',
    '/-/tree/*': 'tree',
    ['*']: undefined,
} as const

export function createRepoRevisionContainerRoutes(
    PageComponent: typeof RepositoryFileTreePage
): RepoRevisionContainerRoute[] {
    return [
        ...Object.entries(routeToObjectType).map<RepoRevisionContainerRoute>(([routePath, objectType]) => ({
            path: routePath,
            render: props => (
                <TraceSpanProvider name="RepositoryFileTreePage" attributes={{ objectType }}>
                    <PageComponent {...props} objectType={objectType} globalContext={window.context} />
                </TraceSpanProvider>
            ),
        })),
        {
            path: '/-/commits/*',
            render: ({ revision, repo, ...context }) =>
                repo ? (
                    <RepositoryCommitsPage
                        {...context}
                        repo={repo}
                        revision={revision}
                        telemetryRecorder={context.platformContext.telemetryRecorder}
                    />
                ) : (
                    <LoadingSpinner />
                ),
        },
        {
            path: '/-/changelists/*',
            render: ({ revision, repo, ...context }) =>
                repo ? (
                    <RepositoryCommitsPage
                        {...context}
                        repo={repo}
                        revision={revision}
                        telemetryRecorder={context.platformContext.telemetryRecorder}
                    />
                ) : (
                    <LoadingSpinner />
                ),
        },
    ]
}

export const repoRevisionContainerRoutes: readonly RepoRevisionContainerRoute[] =
    createRepoRevisionContainerRoutes(RepositoryFileTreePage)
