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

// Work around the issue that react router can not match nested splats when the URL contains spaces
// by expanding the repo matcher to an optional path of up to 10 segments.
//
// We don't rely on the route param names anyway and use `parseBrowserRepoURL`
// instead to parse the repo name.
//
// More info about the issue
// https://github.com/remix-run/react-router/pull/10028
//
// This splat should be used for all routes inside of `RepoContainer`.
export const repoSplat =
    '/:repo_one?/:repo_two?/:repo_three?/:repo_four?/:repo_five?/:repo_six?/:repo_seven?/:repo_eight?/:repo_nine?/:repo_ten?'

const routeToObjectType = {
    [repoSplat + '/-/blob/*']: 'blob',
    [repoSplat + '/-/tree/*']: 'tree',
    ['*']: undefined,
} as const

export const commitsPath = repoSplat + '/-/commits/*'

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
