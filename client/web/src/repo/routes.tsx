import React from 'react'
import { Redirect, RouteComponentProps } from 'react-router'

import { getModeFromPath } from '@sourcegraph/shared/src/languages'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'
import {
    appendLineRangeQueryParameter,
    isLegacyFragment,
    parseQueryAndHash,
    toRepoURL,
} from '@sourcegraph/shared/src/util/url'

import { ErrorBoundary } from '../components/ErrorBoundary'
import { ActionItemsBar } from '../extensions/components/ActionItemsBar'
import { Settings } from '../schema/settings.schema'
import { lazyComponent } from '../util/lazyComponent'
import { formatHash, formatLineOrPositionOrRange } from '../util/url'

import { RepoContainerRoute } from './RepoContainer'
import { RepoRevisionContainerContext, RepoRevisionContainerRoute } from './RepoRevisionContainer'

const BlobPage = lazyComponent(() => import('./blob/BlobPage'), 'BlobPage')
const RepositoryDocumentationPage = lazyComponent(
    () => import('./docs/RepositoryDocumentationPage'),
    'RepositoryDocumentationPage'
)
const RepositoryCommitsPage = lazyComponent(() => import('./commits/RepositoryCommitsPage'), 'RepositoryCommitsPage')
const RepoRevisionSidebar = lazyComponent(() => import('./RepoRevisionSidebar'), 'RepoRevisionSidebar')
const TreePage = lazyComponent(() => import('./tree/TreePage'), 'TreePage')

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
            <div className="repo-revision-container">
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
            </div>
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
            <div className="repo-revision-container">
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
            </div>
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

/** Dev feature flag to make benchmarking the file tree in isolation easier. */
const hideRepoRevisionContent = localStorage.getItem('hideRepoRevContent')

export const repoRevisionContainerRoutes: readonly RepoRevisionContainerRoute[] = [
    ...['', '/-/:objectType(blob|tree)/:filePath*'].map(routePath => ({
        path: routePath,
        exact: routePath === '',
        render: ({
            repo,
            resolvedRev: { commitID, defaultBranch },
            match,
            patternType,
            setPatternType,
            caseSensitive,
            setCaseSensitivity,
            globbing,
            ...context
        }: RepoRevisionContainerContext &
            RouteComponentProps<{
                objectType: 'blob' | 'tree' | undefined
                filePath: string | undefined
            }>) => {
            // The decoding depends on the pinned `history` version.
            // See https://github.com/sourcegraph/sourcegraph/issues/4408
            // and https://github.com/ReactTraining/history/issues/505
            const filePath = decodeURIComponent(match.params.filePath || '') // empty string is root

            // Redirect tree and blob routes pointing to the root to the repo page
            if (match.params.objectType && filePath.replace(/\/+$/g, '') === '') {
                return <Redirect to={toRepoURL({ repoName: repo.name, revision: context.revision })} />
            }

            const objectType: 'blob' | 'tree' = match.params.objectType || 'tree'

            const mode = getModeFromPath(filePath)

            // Redirect OpenGrok-style line number hashes (#123, #123-321) to query parameter (?L123, ?L123-321)
            const hashLineNumberMatch = window.location.hash.match(/^#?(\d+)(-\d+)?$/)
            if (objectType === 'blob' && hashLineNumberMatch) {
                const startLineNumber = parseInt(hashLineNumberMatch[1], 10)
                const endLineNumber = hashLineNumberMatch[2] ? parseInt(hashLineNumberMatch[2].slice(1), 10) : undefined
                const url = appendLineRangeQueryParameter(
                    window.location.pathname + window.location.search,
                    `L${startLineNumber}` + (endLineNumber ? `-${endLineNumber}` : '')
                )
                return <Redirect to={url} />
            }

            // For blob pages with legacy URL fragment hashes like "#L17:19-21:23$foo:bar"
            // redirect to the modern URL fragment hashes like "#L17:19-21:23&tab=foo:bar"
            if (!hideRepoRevisionContent && objectType === 'blob' && isLegacyFragment(window.location.hash)) {
                const parsedQuery = parseQueryAndHash(window.location.search, window.location.hash)
                const hashParameters = new URLSearchParams()
                if (parsedQuery.viewState) {
                    hashParameters.set('tab', parsedQuery.viewState)
                }
                const range = formatLineOrPositionOrRange(parsedQuery)
                const url = appendLineRangeQueryParameter(
                    window.location.pathname + window.location.search,
                    range ? `L${range}` : undefined
                )
                return <Redirect to={url + formatHash(hashParameters)} />
            }

            const repoRevisionProps = {
                commitID,
                filePath,
                patternType,
                setPatternType,
                caseSensitive,
                setCaseSensitivity,
                globbing,
            }

            return (
                <>
                    <RepoRevisionSidebar
                        {...context}
                        {...repoRevisionProps}
                        repoID={repo.id}
                        repoName={repo.name}
                        className="repo-revision-container__sidebar"
                        isDir={objectType === 'tree'}
                        defaultBranch={defaultBranch || 'HEAD'}
                    />
                    {!hideRepoRevisionContent && (
                        // Add `.blob-status-bar__container` because this is the
                        // lowest common ancestor of Blob and the absolutely-positioned Blob status bar
                        <div className="repo-revision-container__content blob-status-bar__container">
                            <ErrorBoundary location={context.location}>
                                {objectType === 'blob' ? (
                                    <BlobPage
                                        {...context}
                                        {...repoRevisionProps}
                                        repoID={repo.id}
                                        repoName={repo.name}
                                        repoUrl={repo.url}
                                        mode={mode}
                                        repoHeaderContributionsLifecycleProps={
                                            context.repoHeaderContributionsLifecycleProps
                                        }
                                    />
                                ) : (
                                    <TreePage {...context} {...repoRevisionProps} repo={repo} />
                                )}
                            </ErrorBoundary>
                        </div>
                    )}
                    <ActionItemsBar
                        useActionItemsBar={context.useActionItemsBar}
                        location={context.location}
                        extensionsController={context.extensionsController}
                        platformContext={context.platformContext}
                        telemetryService={context.telemetryService}
                    />
                </>
            )
        },
    })),
    {
        path: '/-/commits',
        render: ({ resolvedRev: { commitID }, repoHeaderContributionsLifecycleProps, ...context }) => (
            <>
                <RepositoryCommitsPage
                    {...context}
                    commitID={commitID}
                    repoHeaderContributionsLifecycleProps={repoHeaderContributionsLifecycleProps}
                />
                <ActionItemsBar
                    useActionItemsBar={context.useActionItemsBar}
                    location={context.location}
                    extensionsController={context.extensionsController}
                    platformContext={context.platformContext}
                    telemetryService={context.telemetryService}
                />
            </>
        ),
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
        render: ({
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
            ...context
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
                <ActionItemsBar
                    useActionItemsBar={context.useActionItemsBar}
                    location={location}
                    extensionsController={context.extensionsController}
                    platformContext={context.platformContext}
                    telemetryService={context.telemetryService}
                />
            </>
        ),
    },
]
