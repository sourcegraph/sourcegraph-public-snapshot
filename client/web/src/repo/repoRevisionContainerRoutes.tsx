import { TraceSpanProvider } from '@sourcegraph/observability-client'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import { ActionItemsBarProps } from '../extensions/components/ActionItemsBar'

import { RepoRevisionContainerRoute } from './RepoRevisionContainer'

const RepositoryCommitsPage = lazyComponent(() => import('./commits/RepositoryCommitsPage'), 'RepositoryCommitsPage')

const RepositoryFileTreePage = lazyComponent(() => import('./RepositoryFileTreePage'), 'RepositoryFileTreePage')

const ActionItemsBar = lazyComponent<ActionItemsBarProps, 'ActionItemsBar'>(
    () => import('../extensions/components/ActionItemsBar'),
    'ActionItemsBar'
)

const routeToObjectType = {
    '/-/blob/*': 'blob',
    '/-/tree/*': 'tree',
    '': undefined,
} as const

export const commitsPath = '/-/commits/*'

export const repoRevisionContainerRoutes: readonly RepoRevisionContainerRoute[] = [
    ...Object.entries(routeToObjectType).map<RepoRevisionContainerRoute>(([routePath, objectType]) => ({
        path: routePath,
        render: props => (
            <TraceSpanProvider name="RepositoryFileTreePage" attributes={{ objectType }}>
                <RepositoryFileTreePage {...props} objectType={objectType} />
                {window.context.enableLegacyExtensions && (
                    <ActionItemsBar
                        repo={props.repo}
                        useActionItemsBar={props.useActionItemsBar}
                        extensionsController={props.extensionsController}
                        platformContext={props.platformContext}
                        telemetryService={props.telemetryService}
                        source={objectType === 'blob' ? 'blob' : undefined}
                    />
                )}
            </TraceSpanProvider>
        ),
    })),
    {
        path: commitsPath,
        render: ({ revision, repo, ...context }) =>
            repo ? <RepositoryCommitsPage {...context} repo={repo} revision={revision} /> : <LoadingSpinner />,
    },
]
