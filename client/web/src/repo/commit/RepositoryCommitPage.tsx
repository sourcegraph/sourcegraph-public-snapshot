import React, { useMemo, useCallback, useEffect } from 'react'

import classNames from 'classnames'
import { RouteComponentProps } from 'react-router'
import { Observable, ReplaySubject, Subject } from 'rxjs'
import { catchError, filter, map, withLatestFrom } from 'rxjs/operators'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { HoverMerged } from '@sourcegraph/client-api'
import { HoveredToken, createHoverifier, Hoverifier, HoverState } from '@sourcegraph/codeintellify'
import { createAggregateError, isErrorLike, isDefined, memoizeObservable, property, asError } from '@sourcegraph/common'
import { gql } from '@sourcegraph/http-client'
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
import { requestGraphQL } from '../../backend/graphql'
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
    Scalars,
} from '../../graphql-operations'
import { GitCommitNode } from '../commits/GitCommitNode'
import { gitCommitFragment } from '../commits/RepositoryCommitsPage'
import { queryRepositoryComparisonFileDiffs, RepositoryComparisonDiff } from '../compare/RepositoryCompareDiffPage'

import styles from './RepositoryCommitPage.module.scss'

const queryCommit = memoizeObservable(
    (args: { repo: Scalars['ID']; revspec: string }): Observable<GitCommitFields> =>
        requestGraphQL<RepositoryCommitResult, RepositoryCommitVariables>(
            gql`
                query RepositoryCommit($repo: ID!, $revspec: String!) {
                    node(id: $repo) {
                        __typename
                        ... on Repository {
                            commit(rev: $revspec) {
                                __typename # necessary so that isErrorLike(x) is false when x: GitCommitFields
                                ...GitCommitFields
                            }
                        }
                    }
                }
                ${gitCommitFragment}
            `,
            args
        ).pipe(
            map(({ data, errors }) => {
                if (!data || !data.node) {
                    throw createAggregateError(errors)
                }
                if (data.node.__typename !== 'Repository') {
                    throw new Error(`Node is a ${data.node.__typename}, not a Repository`)
                }
                if (!data.node.commit) {
                    // Filter out any revision not found errors, they usually come in multiples when searching for a commit, we want to replace all of them with 1 "Commit not found" error
                    // TODO: Figuring why should we use `errors` here since it is `undefined` in this place
                    const errorsWithoutRevisionError = errors?.filter(
                        (error: { message: string | string[] }) => !error.message.includes('revision not found')
                    )

                    const revisionErrorsFiltered =
                        errors && errorsWithoutRevisionError && errorsWithoutRevisionError.length < errors?.length

                    // If there are no other errors left (or there wasn't any errors to begin with throw out a Commit not found error
                    if (!errorsWithoutRevisionError || errorsWithoutRevisionError.length === 0) {
                        throw new Error('Commit not found')
                    }

                    // if we found at least 1 "revision nor found error" add "Commit not found" to the errors
                    if (revisionErrorsFiltered) {
                        throw createAggregateError([new Error('Commit not found'), ...errorsWithoutRevisionError])
                    }

                    // no "revision not found" errors, throw the other errors
                    throw createAggregateError(errorsWithoutRevisionError)
                }
                return data.node.commit
            })
        ),
    args => `${args.repo}:${args.revspec}`
)

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

export type DiffMode = 'split' | 'unified'

/** Displays a commit. */
export const RepositoryCommitPage: React.FunctionComponent<RepositoryCommitPageProps> = props => {
    const commitOrError = useObservable(
        useMemo(
            () =>
                queryCommit({ repo: props.repo.id, revspec: props.match.params.revspec }).pipe(
                    catchError(error => [asError(error)])
                ),
            [props.match.params.revspec, props.repo.id]
        )
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
        if (commitOrError && !isErrorLike(commitOrError)) {
            props.onDidUpdateExternalLinks(commitOrError.externalURLs)
        }

        return () => {
            props.onDidUpdateExternalLinks(undefined)
        }
    }, [commitOrError, props])

    const queryDiffs = useCallback(
        (args: FilteredConnectionQueryArguments): Observable<RepositoryComparisonDiff['comparison']['fileDiffs']> =>
            queryRepositoryComparisonFileDiffs({
                ...args,
                repo: props.repo.id,
                base: commitParentOrEmpty(commitOrError as GitCommitFields),
                head: (commitOrError as GitCommitFields).oid,
            }),
        [commitOrError, props.repo.id]
    )

    const { extensionsController } = props

    return (
        <div
            data-testid="repository-commit-page"
            className={classNames('p-3', styles.repositoryCommitPage)}
            ref={nextRepositoryCommitPageElement}
        >
            <PageTitle
                title={
                    commitOrError && !isErrorLike(commitOrError)
                        ? commitOrError.subject
                        : `Commit ${props.match.params.revspec}`
                }
            />
            {commitOrError === undefined ? (
                <LoadingSpinner className="mt-2" />
            ) : isErrorLike(commitOrError) ? (
                <ErrorAlert className="mt-2" error={commitOrError} />
            ) : (
                <>
                    <div className="border-bottom pb-2">
                        <div>
                            <GitCommitNode
                                node={commitOrError}
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
                                              revision: commitParentOrEmpty(commitOrError),
                                              commitID: commitParentOrEmpty(commitOrError),
                                          },
                                          head: {
                                              repoName: props.repo.name,
                                              repoID: props.repo.id,
                                              revision: commitOrError.oid,
                                              commitID: commitOrError.oid,
                                          },
                                          hoverifier,
                                          extensionsController,
                                      }
                                    : undefined,
                            lineNumbers: true,
                            diffMode,
                        }}
                        updateOnChange={`${props.repo.id}:${commitOrError.oid}`}
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
