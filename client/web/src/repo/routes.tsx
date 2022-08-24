import React from 'react'

import { TraceSpanProvider } from '@sourcegraph/observability-client'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { ActionItemsBarProps } from '../extensions/components/ActionItemsBar'

import { ToggleBlameAction } from './actions/ToggleBlameAction'
import { RepoRevisionWrapper } from './components/RepoRevision'
import { RepoContainerRoute } from './RepoContainer'
import { RepoHeaderContributionPortal } from './RepoHeaderContributionPortal'
import { RepoRevisionContainerRoute } from './RepoRevisionContainer'
import { RepositoryFileTreePageProps } from './RepositoryFileTreePage'
import { RepositoryTagTab } from './tree/TagTab'

const RepositoryCommitsPage = lazyComponent(() => import('./commits/RepositoryCommitsPage'), 'RepositoryCommitsPage')

const RepositoryFileTreePage = lazyComponent(() => import('./RepositoryFileTreePage'), 'RepositoryFileTreePage')

const RepositoryGitDataContainer = lazyComponent(
    () => import('./RepositoryGitDataContainer'),
    'RepositoryGitDataContainer'
)
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

export const repoContainerRoutes: readonly RepoContainerRoute[] = [
    {
        path: '/-/commit/:revspec+',
        render: context => (
            <RepoRevisionWrapper>
                <RepositoryGitDataContainer {...context} repoName={context.repo.name}>
                    <RepositoryCommitPage {...context} />
                </RepositoryGitDataContainer>
                <RepoHeaderContributionPortal
                    position="right"
                    id="toggle-blame"
                    repoHeaderContributionsLifecycleProps={context.repoHeaderContributionsLifecycleProps}
                >
                    {({ actionType }) => <ToggleBlameAction key="toggle-blame" actionType={actionType} />}
                </RepoHeaderContributionPortal>
                <ActionItemsBar
                    extensionsController={context.extensionsController}
                    platformContext={context.platformContext}
                    useActionItemsBar={context.useActionItemsBar}
                    location={context.location}
                    telemetryService={context.telemetryService}
                />
            </RepoRevisionWrapper>
        ),
    },
    {
        path: '/-/branches',
        render: context => (
            <RepositoryGitDataContainer {...context} repoName={context.repo.name}>
                <RepositoryBranchesArea {...context} />
            </RepositoryGitDataContainer>
        ),
    },
    {
        path: '/-/tags',
        render: context => (
            <RepositoryGitDataContainer {...context} repoName={context.repo.name}>
                <RepositoryReleasesArea {...context} />
            </RepositoryGitDataContainer>
        ),
    },
    {
        path: '/-/compare/:spec*',
        render: context => (
            <RepoRevisionWrapper>
                <RepositoryGitDataContainer {...context} repoName={context.repo.name}>
                    <RepositoryCompareArea {...context} />
                </RepositoryGitDataContainer>
                <RepoHeaderContributionPortal
                    position="right"
                    id="toggle-blame"
                    repoHeaderContributionsLifecycleProps={context.repoHeaderContributionsLifecycleProps}
                >
                    {({ actionType }) => <ToggleBlameAction key="toggle-blame" actionType={actionType} />}
                </RepoHeaderContributionPortal>
                <ActionItemsBar
                    extensionsController={context.extensionsController}
                    platformContext={context.platformContext}
                    useActionItemsBar={context.useActionItemsBar}
                    location={context.location}
                    telemetryService={context.telemetryService}
                />
            </RepoRevisionWrapper>
        ),
    },
    {
        path: '/-/stats',
        render: context => (
            <RepositoryGitDataContainer {...context} repoName={context.repo.name}>
                <RepositoryStatsArea {...context} />
            </RepositoryGitDataContainer>
        ),
    },
    {
        path: '/-/settings',
        render: context => (
            <RepositoryGitDataContainer {...context} repoName={context.repo.name}>
                <RepoSettingsArea {...context} />
            </RepositoryGitDataContainer>
        ),
    },
]

export const RepoContributors: React.FunctionComponent<React.PropsWithChildren<any>> = ({
    useBreadcrumb,
    setBreadcrumb,
    repo,
    history,
    location,
    match,
    globbing,
}) => (
    <>
        <RepositoryStatsArea
            useBreadcrumb={useBreadcrumb}
            setBreadcrumb={setBreadcrumb}
            repo={repo}
            history={history}
            location={location}
            match={match}
            globbing={globbing}
        />
    </>
)

export const RepoCommits: React.FunctionComponent<React.PropsWithChildren<any>> = ({
    resolvedRev: { commitID },
    repoHeaderContributionsLifecycleProps,
    ...context
}) => (
    <>
        <RepositoryCommitsPage
            {...context}
            commitID={commitID}
            repoHeaderContributionsLifecycleProps={repoHeaderContributionsLifecycleProps}
        />
    </>
)

export const repoRevisionContainerRoutes: readonly RepoRevisionContainerRoute[] = [
    ...[
        '',
        '/-/:objectType(blob|tree)/:filePath*',
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
            </TraceSpanProvider>
        ),
    })),
    {
        path: '/-/commits',
        render: RepoCommits,
    },
    {
        path: '/-/branch',
        render: ({ repo, location, history }) => (
            <RepositoryBranchesTab repo={repo} location={location} history={history} />
        ),
    },
    {
        path: '/-/tag',
        render: ({ repo, location, history }) => <RepositoryTagTab repo={repo} location={location} history={history} />,
    },
    {
        path: '/-/compare/:spec*',
        render: context => (
            <RepoRevisionWrapper>
                <RepositoryGitDataContainer {...context} repoName={context.repo.name}>
                    <RepositoryCompareArea {...context} />
                </RepositoryGitDataContainer>
                <RepoHeaderContributionPortal
                    position="right"
                    id="toggle-blame"
                    repoHeaderContributionsLifecycleProps={context.repoHeaderContributionsLifecycleProps}
                >
                    {({ actionType }) => <ToggleBlameAction key="toggle-blame" actionType={actionType} />}
                </RepoHeaderContributionPortal>
                <ActionItemsBar
                    extensionsController={context.extensionsController}
                    platformContext={context.platformContext}
                    useActionItemsBar={context.useActionItemsBar}
                    location={context.location}
                    telemetryService={context.telemetryService}
                />
            </RepoRevisionWrapper>
        ),
    },
    {
        path: '/-/contributors',
        render: RepoContributors,
    },
]
