import React from 'react'

import { Redirect, RouteComponentProps } from 'react-router'

import { appendLineRangeQueryParameter } from '@sourcegraph/common'
import { TraceSpanProvider } from '@sourcegraph/observability-client'
import { getModeFromPath } from '@sourcegraph/shared/src/languages'
import { isLegacyFragment, parseQueryAndHash, toRepoURL } from '@sourcegraph/shared/src/util/url'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import { ErrorBoundary } from '../components/ErrorBoundary'
import { ActionItemsBar } from '../extensions/components/ActionItemsBar'
import { GettingStartedTour } from '../tour/GettingStartedTour'
import { formatHash, formatLineOrPositionOrRange } from '../util/url'

import { BlobProps } from './blob/Blob'
import { BlobPage } from './blob/BlobPage'
import { BlobStatusBarContainer } from './blob/ui/BlobStatusBarContainer'
import { RepoRevisionContainerContext } from './RepoRevisionContainer'
import { RepoRevisionSidebar } from './RepoRevisionSidebar'
import { TreePage } from './tree/TreePage'

export interface RepositoryFileTreePageProps
    extends RepoRevisionContainerContext,
        RouteComponentProps<{
            objectType: 'blob' | 'tree' | undefined
            filePath: string | undefined
            spec: string
        }>,
        Pick<BlobProps, 'onHandleFuzzyFinder'> {}

/** Dev feature flag to make benchmarking the file tree in isolation easier. */
const hideRepoRevisionContent = localStorage.getItem('hideRepoRevContent')

/** A page that shows a file or a directory (tree view) in a repository at the
 * current revision. */
export const RepositoryFileTreePage: React.FunctionComponent<
    React.PropsWithChildren<RepositoryFileTreePageProps>
> = props => {
    const { repo, resolvedRev, repoName, match, globbing, ...context } = props

    // The decoding depends on the pinned `history` version.
    // See https://github.com/sourcegraph/sourcegraph/issues/4408
    // and https://github.com/ReactTraining/history/issues/505
    const filePath = decodeURIComponent(match.params.filePath || '') // empty string is root
    // Redirect tree and blob routes pointing to the root to the repo page
    if (match.params.objectType && filePath.replace(/\/+$/g, '') === '') {
        return <Redirect to={toRepoURL({ repoName, revision: context.revision })} />
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
        commitID: resolvedRev?.commitID,
        filePath,
        globbing,
    }

    return (
        <>
            <RepoRevisionSidebar
                {...context}
                {...repoRevisionProps}
                repoID={repo?.id}
                repoName={repoName}
                className="repo-revision-container__sidebar"
                isDir={objectType === 'tree'}
                defaultBranch={resolvedRev?.defaultBranch || 'HEAD'}
            />
            {!hideRepoRevisionContent && (
                // Add `.blob-status-bar__container` because this is the
                // lowest common ancestor of Blob and the absolutely-positioned Blob status bar
                <BlobStatusBarContainer>
                    <GettingStartedTour.Info isSourcegraphDotCom={context.isSourcegraphDotCom} className="mr-3 mb-3" />
                    <ErrorBoundary location={context.location}>
                        {objectType === 'blob' ? (
                            <TraceSpanProvider name="BlobPage">
                                <BlobPage
                                    {...context}
                                    {...repoRevisionProps}
                                    repoID={repo?.id}
                                    repoName={repoName}
                                    repoUrl={repo?.url}
                                    mode={mode}
                                    repoHeaderContributionsLifecycleProps={
                                        context.repoHeaderContributionsLifecycleProps
                                    }
                                    onHandleFuzzyFinder={props.onHandleFuzzyFinder}
                                />
                            </TraceSpanProvider>
                        ) : resolvedRev ? (
                            // TODO: see if we can render without resolvedRev.commitID
                            <TreePage
                                {...props}
                                commitID={context.revision}
                                filePath={filePath}
                                globbing={globbing}
                                repo={repo}
                                repoName={repoName}
                                match={match}
                                useActionItemsBar={context.useActionItemsBar}
                                isSourcegraphDotCom={context.isSourcegraphDotCom}
                            />
                        ) : (
                            <LoadingSpinner />
                        )}
                    </ErrorBoundary>
                </BlobStatusBarContainer>
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
}
