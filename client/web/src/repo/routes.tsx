import React from 'react'

import { RouteComponentProps } from 'react-router-dom'

import { TraceSpanProvider } from '@sourcegraph/observability-client'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import { ActionItemsBarProps } from '../extensions/components/ActionItemsBar'

import type { RepositoryCommitsPageProps } from './commits/RepositoryCommitsPage'
import { RepoRevisionWrapper } from './components/RepoRevision'
import { RepoContainerRoute } from './RepoContainer'
import { RepoRevisionContainerContext, RepoRevisionContainerRoute } from './RepoRevisionContainer'
import { RepositoryFileTreePageProps } from './RepositoryFileTreePage'
import { RepositoryTagTab } from './tree/TagTab'

const RepositoryCommitsPage = lazyComponent(() => import('./commits/RepositoryCommitsPage'), 'RepositoryCommitsPage')

const RepositoryFileTreePage = lazyComponent(() => import('./RepositoryFileTreePage'), 'RepositoryFileTreePage')

const RepositoryCommitPage = lazyComponent(() => import('./commit/RepositoryCommitPage'), 'RepositoryCommitPage')
const RepositoryBranchesArea = lazyComponent(
    () => import('./branches/RepositoryBranchesArea'),
    'RepositoryBranchesArea'
)

const RepositoryReleasesArea = lazyComponent(
    () => import('./releases/RepositoryReleasesArea'),
    'RepositoryReleasesArea'
)
const RepoSettingsArea = lazyComponent(() => import('./settings/RepoSettingsArea'), 'RepoSettingsArea')
const RepositoryCompareArea = lazyComponent(() => import('./compare/RepositoryCompareArea'), 'RepositoryCompareArea')
const RepositoryStatsArea = lazyComponent(() => import('./stats/RepositoryStatsArea'), 'RepositoryStatsArea')
const RepositoryBranchesTab = lazyComponent(() => import('./tree/BranchesTab'), 'RepositoryBranchesTab')
const ActionItemsBar = lazyComponent<ActionItemsBarProps, 'ActionItemsBar'>(
    () => import('../extensions/components/ActionItemsBar'),
    'ActionItemsBar'
)

export const compareSpecPath = '/-/compare/:spec*'

export const repoContainerRoutes: readonly RepoContainerRoute[] = [
    {
        path: '/-/commit/:revspec+',
        render: context => (
            <RepoRevisionWrapper>
                <RepositoryCommitPage {...context} />
                {window.context.enableLegacyExtensions && (
                    <ActionItemsBar
                        extensionsController={context.extensionsController}
                        platformContext={context.platformContext}
                        useActionItemsBar={context.useActionItemsBar}
                        location={context.location}
                        telemetryService={context.telemetryService}
                        source="commit"
                    />
                )}
            </RepoRevisionWrapper>
        ),
    },
    {
        path: '/-/branches',
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
                        location={context.location}
                        telemetryService={context.telemetryService}
                        source="compare"
                    />
                )}
            </RepoRevisionWrapper>
        ),
    },
    {
        path: '/-/stats',
        render: context => <RepositoryStatsArea {...context} />,
    },
    {
        path: '/-/settings',
        render: context => <RepoSettingsArea {...context} />,
    },
]

export const RepoContributors: React.FunctionComponent<
    React.PropsWithChildren<RepoRevisionContainerContext & RouteComponentProps>
> = ({ useBreadcrumb, setBreadcrumb, repo, history, location, match, globbing, repoName }) => (
    <RepositoryStatsArea
        useBreadcrumb={useBreadcrumb}
        setBreadcrumb={setBreadcrumb}
        repo={repo}
        repoName={repoName}
        history={history}
        location={location}
        match={match}
        globbing={globbing}
    />
)

export const RepoCommits: React.FunctionComponent<
    Omit<RepositoryCommitsPageProps, 'repo'> & Pick<RepoRevisionContainerContext, 'repo'> & RouteComponentProps
> = ({ revision, repo, ...context }) =>
    repo ? <RepositoryCommitsPage {...context} repo={repo} revision={revision} /> : <LoadingSpinner />

const blobPath = '/-/:objectType(blob)/:filePath*'
const treePath = '/-/:objectType(tree)/:filePath*'
export const commitsPath = '/-/commits/:filePath*'

export const repoRevisionContainerRoutes: readonly RepoRevisionContainerRoute[] = [
    ...[
        '',
        blobPath,
        treePath,
        '/-/docs/tab/:pathID*',
        '/-/commits/tab',
        '/-/branch/tab',
        '/-/tag/tab',
        '/-/contributors/tab',
        '/-/compare/tab/:spec*',
    ].map(routePath => ({
        path: routePath,
        exact: routePath === '',
        render: (props: RepositoryFileTreePageProps) => (
            <TraceSpanProvider
                name="RepositoryFileTreePage"
                attributes={{
                    objectType: props.match.params.objectType,
                }}
            >
                <RepositoryFileTreePage {...props} />
                {window.context.enableLegacyExtensions && (
                    <ActionItemsBar
                        repo={props.repo}
                        useActionItemsBar={props.useActionItemsBar}
                        location={props.location}
                        extensionsController={props.extensionsController}
                        platformContext={props.platformContext}
                        telemetryService={props.telemetryService}
                        source={routePath === blobPath ? 'blob' : undefined}
                    />
                )}
            </TraceSpanProvider>
        ),
    })),
    {
        path: commitsPath,
        render: RepoCommits,
    },
    {
        path: '/-/branch',
        render: ({ repo }) => <RepositoryBranchesTab repo={repo} />,
    },
    {
        path: '/-/tag',
        render: ({ repo }) => <RepositoryTagTab repo={repo} />,
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
                        location={context.location}
                        telemetryService={context.telemetryService}
                        source="compare"
                    />
                )}
            </RepoRevisionWrapper>
        ),
    },
    {
        path: '/-/contributors',
        render: RepoContributors,
    },
]
