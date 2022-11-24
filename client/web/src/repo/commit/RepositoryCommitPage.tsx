import React, { useMemo, useCallback, useEffect } from 'react'

import classNames from 'classnames'
import { RouteComponentProps } from 'react-router'
import { Observable, ReplaySubject, Subject } from 'rxjs'
import { filter, map, withLatestFrom } from 'rxjs/operators'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { HoverMerged } from '@sourcegraph/client-api'
import { HoveredToken, createHoverifier, Hoverifier, HoverState } from '@sourcegraph/codeintellify'
import { isDefined, property } from '@sourcegraph/common'
import { gql, useQuery } from '@sourcegraph/http-client'
import { ActionItemAction } from '@sourcegraph/shared/src/actions/ActionItem'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { getHoverActions } from '@sourcegraph/shared/src/hover/actions'
import { HoverContext } from '@sourcegraph/shared/src/hover/HoverOverlay'
import { getModeFromPath } from '@sourcegraph/shared/src/languages'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import {
    FileSpec,
    ModeSpec,
    RepoSpec,
    ResolvedRevisionSpec,
    RevisionSpec,
    UIPositionSpec,
} from '@sourcegraph/shared/src/util/url'
import { LoadingSpinner, useObservable } from '@sourcegraph/wildcard'

import { getHover, getDocumentHighlights } from '../../backend/features'
import { FileDiffConnection } from '../../components/diff/FileDiffConnection'
import { FileDiffNode } from '../../components/diff/FileDiffNode'
import { FilteredConnectionQueryArguments } from '../../components/FilteredConnection'
import { PageTitle } from '../../components/PageTitle'
import { WebHoverOverlay } from '../../components/shared'
import {
    ExternalLinkFields,
    GitCommitFields,
    RepositoryCommitResult,
    RepositoryCommitVariables,
    RepositoryFields,
} from '../../graphql-operations'
import { GitCommitNode } from '../commits/GitCommitNode'
import { gitCommitFragment } from '../commits/RepositoryCommitsPage'
import { queryRepositoryComparisonFileDiffs, RepositoryComparisonDiff } from '../compare/RepositoryCompareDiffPage'

import styles from './RepositoryCommitPage.module.scss'

const COMMIT_QUERY = gql`
    query RepositoryCommit($repo: ID!, $revspec: String!) {
        node(id: $repo) {
            __typename
            ... on Repository {
                commit(rev: $revspec) {
                    __typename # Necessary for error handling to check if commit exists
                    ...GitCommitFields
                }
            }
        }
    }
    ${gitCommitFragment}
`

interface RepositoryCommitPageProps
    extends RouteComponentProps<{ revspec: string }>,
        TelemetryProps,
        PlatformContextProps,
        ExtensionsControllerProps,
        ThemeProps,
        SettingsCascadeProps {
    repo: RepositoryFields
    onDidUpdateExternalLinks: (externalLinks: ExternalLinkFields[] | undefined) => void
}

export type { DiffMode } from '@sourcegraph/shared/src/settings/temporary/diffMode'

/** Displays a commit. */
export const RepositoryCommitPage: React.FunctionComponent<RepositoryCommitPageProps> = props => {
    const { data, error, loading } = useQuery<RepositoryCommitResult, RepositoryCommitVariables>(COMMIT_QUERY, {
        variables: {
            repo: props.repo.id,
            revspec: props.match.params.revspec,
        },
    })

    const commit = useMemo(
        () => (data?.node && data?.node?.__typename === 'Repository' ? data?.node?.commit : undefined),
        [data]
    )

    const [diffMode, setDiffMode] = useTemporarySetting('repo.commitPage.diffMode', 'unified')

    /** Emits whenever the ref callback for the hover element is called */
    const hoverOverlayElements = useMemo(() => new Subject<HTMLElement | null>(), [])
    const nextOverlayElement = useCallback((element: HTMLElement | null): void => hoverOverlayElements.next(element), [
        hoverOverlayElements,
    ])

    /** Emits whenever the ref callback for the hover element is called */
    const repositoryCommitPageElements = useMemo(() => new Subject<HTMLElement | null>(), [])
    const nextRepositoryCommitPageElement = useCallback(
        (element: HTMLElement | null): void => repositoryCommitPageElements.next(element),
        [repositoryCommitPageElements]
    )

    // Subject that emits on every render. Source for `hoverOverlayRerenders`, used to
    // reposition hover overlay if needed when `Blob` rerenders
    const rerenders = useMemo(() => new ReplaySubject(1), [])
    useEffect(() => {
        rerenders.next()
    })

    const hoverifier: Hoverifier<
        RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec,
        HoverMerged,
        ActionItemAction
    > = useMemo(
        () =>
            createHoverifier<RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec, HoverMerged, ActionItemAction>({
                hoverOverlayElements,
                hoverOverlayRerenders: rerenders.pipe(
                    withLatestFrom(hoverOverlayElements, repositoryCommitPageElements),
                    map(([, hoverOverlayElement, repositoryCommitPageElement]) => ({
                        hoverOverlayElement,
                        // The root component element is guaranteed to be rendered after a componentDidUpdate
                        relativeElement: repositoryCommitPageElement!,
                    })),
                    // Can't reposition HoverOverlay if it wasn't rendered
                    filter(property('hoverOverlayElement', isDefined))
                ),
                getHover: hoveredToken => getHover(getLSPTextDocumentPositionParams(hoveredToken), props),
                getDocumentHighlights: hoveredToken =>
                    getDocumentHighlights(getLSPTextDocumentPositionParams(hoveredToken), props),
                getActions: context => getHoverActions(props, context),
            }),
        [hoverOverlayElements, props, repositoryCommitPageElements, rerenders]
    )
    useEffect(() => () => hoverifier.unsubscribe(), [hoverifier])

    const hoverState: Readonly<HoverState<HoverContext, HoverMerged, ActionItemAction>> =
        useObservable(hoverifier.hoverStateUpdates) || {}

    function getLSPTextDocumentPositionParams(
        hoveredToken: HoveredToken & RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec
    ): RepoSpec & RevisionSpec & ResolvedRevisionSpec & FileSpec & UIPositionSpec & ModeSpec {
        return {
            repoName: hoveredToken.repoName,
            revision: hoveredToken.revision,
            filePath: hoveredToken.filePath,
            commitID: hoveredToken.commitID,
            position: hoveredToken,
            mode: getModeFromPath(hoveredToken.filePath || ''),
        }
    }

    useEffect(() => {
        props.telemetryService.logViewEvent('RepositoryCommit')
    }, [props.telemetryService])

    useEffect(() => {
        if (commit) {
            props.onDidUpdateExternalLinks(commit.externalURLs)
        }

        return () => {
            props.onDidUpdateExternalLinks(undefined)
        }
    }, [commit, props])

    const queryDiffs = useCallback(
        (args: FilteredConnectionQueryArguments): Observable<RepositoryComparisonDiff['comparison']['fileDiffs']> =>
            // Non-null assertions here are safe because the query is only executed if the commit is defined.
            queryRepositoryComparisonFileDiffs({
                ...args,
                repo: props.repo.id,
                base: commitParentOrEmpty(commit!),
                head: commit!.oid,
            }),
        [commit, props.repo.id]
    )

    const { extensionsController } = props

    return (
        <div
            data-testid="repository-commit-page"
            className={classNames('p-3', styles.repositoryCommitPage)}
            ref={nextRepositoryCommitPageElement}
        >
            <PageTitle title={commit ? commit.subject : `Commit ${props.match.params.revspec}`} />
            {loading ? (
                <LoadingSpinner className="mt-2" />
            ) : error || !commit ? (
                <ErrorAlert className="mt-2" error={error ?? new Error('Commit not found')} />
            ) : (
                <>
                    <div className="border-bottom pb-2">
                        <div>
                            <GitCommitNode
                                node={commit}
                                expandCommitMessageBody={true}
                                showSHAAndParentsRow={true}
                                diffMode={diffMode}
                                onHandleDiffMode={setDiffMode}
                                className={styles.gitCommitNode}
                            />
                        </div>
                    </div>
                    <FileDiffConnection
                        listClassName="list-group list-group-flush"
                        noun="changed file"
                        pluralNoun="changed files"
                        queryConnection={queryDiffs}
                        nodeComponent={FileDiffNode}
                        nodeComponentProps={{
                            ...props,
                            extensionInfo:
                                extensionsController !== null
                                    ? {
                                          base: {
                                              repoName: props.repo.name,
                                              repoID: props.repo.id,
                                              revision: commitParentOrEmpty(commit),
                                              commitID: commitParentOrEmpty(commit),
                                          },
                                          head: {
                                              repoName: props.repo.name,
                                              repoID: props.repo.id,
                                              revision: commit.oid,
                                              commitID: commit.oid,
                                          },
                                          hoverifier,
                                          extensionsController,
                                      }
                                    : undefined,
                            lineNumbers: true,
                            diffMode,
                        }}
                        updateOnChange={`${props.repo.id}:${commit.oid}`}
                        defaultFirst={15}
                        hideSearch={true}
                        noSummaryIfAllNodesVisible={true}
                        history={props.history}
                        location={props.location}
                        cursorPaging={true}
                    />
                </>
            )}
            {hoverState.hoverOverlayProps && extensionsController !== null && (
                <WebHoverOverlay
                    {...props}
                    extensionsController={extensionsController}
                    {...hoverState.hoverOverlayProps}
                    nav={url => props.history.push(url)}
                    hoveredTokenElement={hoverState.hoveredTokenElement}
                    telemetryService={props.telemetryService}
                    hoverRef={nextOverlayElement}
                />
            )}
        </div>
    )
}

function commitParentOrEmpty(commit: GitCommitFields): string {
    // 4b825dc642cb6eb9a060e54bf8d69288fbee4904 is `git hash-object -t tree /dev/null`, which is used as the base
    // when computing the `git diff` of the root commit.
    return commit.parents.length > 0 ? commit.parents[0].oid : '4b825dc642cb6eb9a060e54bf8d69288fbee4904'
}
