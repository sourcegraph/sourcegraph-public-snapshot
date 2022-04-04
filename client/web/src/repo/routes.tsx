import React from 'react'

import { isErrorLike } from '@sourcegraph/common'
import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { ActionItemsBar } from '../extensions/components/ActionItemsBar'

import { RepoRevisionWrapper } from './components/RepoRevision'
import { RepoContainerRoute } from './RepoContainer'
import { RepoRevisionContainerRoute } from './RepoRevisionContainer'
import { RepositoryFileTreePageProps } from './RepositoryFileTreePage'
import { RepositoryBranchesTab } from './tree/BranchesTab'
import { RepositoryTagTab } from './tree/TagTab'

const RepositoryDocumentationPage = lazyComponent(
    () => import('./docs/RepositoryDocumentationPage'),
    'RepositoryDocumentationPage'
)
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

export const repoContainerRoutes: readonly RepoContainerRoute[] = [
    {
        path: '/-/commit/:revspec+',
        render: context => (
            <RepoRevisionWrapper>
                <RepositoryGitDataContainer {...context} repoName={context.repo.name}>
                    <RepositoryCommitPage {...context} />
                </RepositoryGitDataContainer>
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

// eslint-disable-next-line unicorn/prevent-abbreviations
export const RepoDocs: React.FunctionComponent<any> = ({
    useBreadcrumb,
    setBreadcrumb,
    settingsCascade,
    repo,
    history,
    location,
    isLightTheme,
    fetchHighlightedFileLineRanges,
    resolvedRev: { commitID },
    match,
}) => (
    <>
        {/*
            IMPORTANT: do NOT use `{...context}` expansion to pass props to page components
            here. Doing so adds other props that exist in `context` that are NOT required
            or specified by the component props, but TypeScript will NOT strip them out.
            For example, the navbarSearchQueryState - meaning every time a user types into
            the search input our React component props would change despite it being a field
            that we are absolutely not using in any way. See:
            https://github.com/sourcegraph/sourcegraph/issues/21200
        */}
        <RepositoryDocumentationPage
            useBreadcrumb={useBreadcrumb}
            setBreadcrumb={setBreadcrumb}
            settingsCascade={settingsCascade}
            repo={repo}
            history={history}
            location={location}
            isLightTheme={isLightTheme}
            fetchHighlightedFileLineRanges={fetchHighlightedFileLineRanges}
            pathID={match.params.pathID ? '/' + decodeURIComponent(match.params.pathID) : '/'}
            commitID={commitID}
        />
    </>
)

export const RepoContributors: React.FunctionComponent<any> = ({
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

export const RepoCommits: React.FunctionComponent<any> = ({
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
        render: (props: RepositoryFileTreePageProps) => <RepositoryFileTreePage {...props} />,
    })),
    {
        path: '/-/commits',
        render: RepoCommits,
    },
    {
        path: '/-/docs/:pathID*',
        condition: ({ settingsCascade }): boolean => {
            if (settingsCascade.final === null || isErrorLike(settingsCascade.final)) {
                return false
            }
            const settings: Settings = settingsCascade.final
            return settings.experimentalFeatures?.apiDocs !== false
        },
        render: RepoDocs,
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
