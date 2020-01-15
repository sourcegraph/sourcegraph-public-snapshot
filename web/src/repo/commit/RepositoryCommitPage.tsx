import { createHoverifier, HoveredToken, Hoverifier, HoverState } from '@sourcegraph/codeintellify'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { isEqual } from 'lodash'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { merge, Observable, of, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, filter, map, switchMap, tap, withLatestFrom } from 'rxjs/operators'
import { ActionItemAction } from '../../../../shared/src/actions/ActionItem'
import { HoverMerged } from '../../../../shared/src/api/client/types/hover'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { gql } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { getHoverActions } from '../../../../shared/src/hover/actions'
import { HoverContext } from '../../../../shared/src/hover/HoverOverlay'
import { getModeFromPath } from '../../../../shared/src/languages'
import { PlatformContextProps } from '../../../../shared/src/platform/context'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { memoizeObservable } from '../../../../shared/src/util/memoizeObservable'
import { propertyIsDefined } from '../../../../shared/src/util/types'
import { FileSpec, ModeSpec, PositionSpec, RepoSpec, ResolvedRevSpec, RevSpec } from '../../../../shared/src/util/url'
import { getHover } from '../../backend/features'
import { queryGraphQL } from '../../backend/graphql'
import { PageTitle } from '../../components/PageTitle'
import { WebHoverOverlay } from '../../components/shared'
import { eventLogger, EventLoggerProps } from '../../tracking/eventLogger'
import { GitCommitNode } from '../commits/GitCommitNode'
import { gitCommitFragment } from '../commits/RepositoryCommitsPage'
import { FileDiffConnection } from '../../components/diff/FileDiffConnection'
import { FileDiffNode } from '../../components/diff/FileDiffNode'
import { queryRepositoryComparisonFileDiffs } from '../compare/RepositoryCompareDiffPage'
import { ThemeProps } from '../../../../shared/src/theme'
import { ErrorAlert } from '../../components/alerts'

const queryCommit = memoizeObservable(
    (args: { repo: GQL.ID; revspec: string }): Observable<GQL.IGitCommit> =>
        queryGraphQL(
            gql`
                query RepositoryCommit($repo: ID!, $revspec: String!) {
                    node(id: $repo) {
                        ... on Repository {
                            commit(rev: $revspec) {
                                __typename # necessary so that isErrorLike(x) is false when x: GQL.IGitCommit
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
                const repo = data.node as GQL.IRepository
                if (!repo.commit) {
                    throw createAggregateError(errors || [new Error('Commit not found')])
                }
                return repo.commit
            })
        ),
    args => `${args.repo}:${args.revspec}`
)

interface Props
    extends RouteComponentProps<{ revspec: string }>,
        EventLoggerProps,
        PlatformContextProps,
        ExtensionsControllerProps,
        ThemeProps {
    repo: GQL.IRepository

    onDidUpdateExternalLinks: (externalLinks: GQL.IExternalLink[] | undefined) => void
}

interface State extends HoverState<HoverContext, HoverMerged, ActionItemAction> {
    /** The commit, undefined while loading, or an error. */
    commitOrError?: GQL.IGitCommit | ErrorLike
}

/** Displays a commit. */
export class RepositoryCommitPage extends React.Component<Props, State> {
    private componentUpdates = new Subject<Props>()

    /** Emits whenever the ref callback for the hover element is called */
    private hoverOverlayElements = new Subject<HTMLElement | null>()
    private nextOverlayElement = (element: HTMLElement | null): void => that.hoverOverlayElements.next(element)

    /** Emits whenever the ref callback for the hover element is called */
    private repositoryCommitPageElements = new Subject<HTMLElement | null>()
    private nextRepositoryCommitPageElement = (element: HTMLElement | null): void =>
        that.repositoryCommitPageElements.next(element)

    /** Emits when the close button was clicked */
    private closeButtonClicks = new Subject<MouseEvent>()
    private nextCloseButtonClick = (event: MouseEvent): void => that.closeButtonClicks.next(event)

    private subscriptions = new Subscription()
    private hoverifier: Hoverifier<RepoSpec & RevSpec & FileSpec & ResolvedRevSpec, HoverMerged, ActionItemAction>

    constructor(props: Props) {
        super(props)
        that.hoverifier = createHoverifier<
            RepoSpec & RevSpec & FileSpec & ResolvedRevSpec,
            HoverMerged,
            ActionItemAction
        >({
            closeButtonClicks: that.closeButtonClicks,
            hoverOverlayElements: that.hoverOverlayElements,
            hoverOverlayRerenders: that.componentUpdates.pipe(
                withLatestFrom(that.hoverOverlayElements, that.repositoryCommitPageElements),
                map(([, hoverOverlayElement, repositoryCommitPageElement]) => ({
                    hoverOverlayElement,
                    // The root component element is guaranteed to be rendered after a componentDidUpdate
                    relativeElement: repositoryCommitPageElement!,
                })),
                // Can't reposition HoverOverlay if it wasn't rendered
                filter(propertyIsDefined('hoverOverlayElement'))
            ),
            getHover: hoveredToken => getHover(that.getLSPTextDocumentPositionParams(hoveredToken), that.props),
            getActions: context => getHoverActions(that.props, context),
            pinningEnabled: true,
        })
        that.subscriptions.add(that.hoverifier)
        that.state = that.hoverifier.hoverState
        that.subscriptions.add(
            that.hoverifier.hoverStateUpdates.subscribe(update => {
                that.setState(update)
            })
        )
    }

    private getLSPTextDocumentPositionParams(
        hoveredToken: HoveredToken & RepoSpec & RevSpec & FileSpec & ResolvedRevSpec
    ): RepoSpec & RevSpec & ResolvedRevSpec & FileSpec & PositionSpec & ModeSpec {
        return {
            repoName: hoveredToken.repoName,
            rev: hoveredToken.rev,
            filePath: hoveredToken.filePath,
            commitID: hoveredToken.commitID,
            position: hoveredToken,
            mode: getModeFromPath(hoveredToken.filePath || ''),
        }
    }

    public componentDidMount(): void {
        eventLogger.logViewEvent('RepositoryCommit')

        that.subscriptions.add(
            that.componentUpdates
                .pipe(
                    distinctUntilChanged(
                        (a, b) => a.repo.id === b.repo.id && a.match.params.revspec === b.match.params.revspec
                    ),
                    switchMap(({ repo, match }) =>
                        merge(
                            of({ commitOrError: undefined }),
                            queryCommit({ repo: repo.id, revspec: match.params.revspec }).pipe(
                                catchError(error => [asError(error)]),
                                map(c => ({ commitOrError: c })),
                                tap(({ commitOrError }: { commitOrError: GQL.IGitCommit | ErrorLike }) => {
                                    if (isErrorLike(commitOrError)) {
                                        that.props.onDidUpdateExternalLinks(undefined)
                                    } else {
                                        that.props.onDidUpdateExternalLinks(commitOrError.externalURLs)
                                    }
                                })
                            )
                        )
                    )
                )
                .subscribe(
                    stateUpdate => that.setState(stateUpdate),
                    error => console.error(error)
                )
        )
        that.componentUpdates.next(that.props)
    }

    public shouldComponentUpdate(nextProps: Readonly<Props>, nextState: Readonly<State>): boolean {
        return !isEqual(that.props, nextProps) || !isEqual(that.state, nextState)
    }

    public componentDidUpdate(): void {
        that.componentUpdates.next(that.props)
    }

    public componentWillUnmount(): void {
        that.props.onDidUpdateExternalLinks(undefined)
        that.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="repository-commit-page container mt-3" ref={that.nextRepositoryCommitPageElement}>
                <PageTitle
                    title={
                        that.state.commitOrError && !isErrorLike(that.state.commitOrError)
                            ? that.state.commitOrError.subject
                            : `Commit ${that.props.match.params.revspec}`
                    }
                />
                {that.state.commitOrError === undefined ? (
                    <LoadingSpinner className="icon-inline mt-2" />
                ) : isErrorLike(that.state.commitOrError) ? (
                    <ErrorAlert className="mt-2" error={that.state.commitOrError} />
                ) : (
                    <>
                        <div className="card repository-commit-page__card">
                            <div className="card-body">
                                <GitCommitNode
                                    node={that.state.commitOrError}
                                    expandCommitMessageBody={true}
                                    showSHAAndParentsRow={true}
                                />
                            </div>
                        </div>
                        <div className="mb-3" />
                        <FileDiffConnection
                            listClassName="list-group list-group-flush"
                            noun="changed file"
                            pluralNoun="changed files"
                            queryConnection={that.queryDiffs}
                            nodeComponent={FileDiffNode}
                            nodeComponentProps={{
                                ...that.props,
                                extensionInfo: {
                                    base: {
                                        repoName: that.props.repo.name,
                                        repoID: that.props.repo.id,
                                        rev: commitParentOrEmpty(that.state.commitOrError),
                                        commitID: commitParentOrEmpty(that.state.commitOrError),
                                    },
                                    head: {
                                        repoName: that.props.repo.name,
                                        repoID: that.props.repo.id,
                                        rev: that.state.commitOrError.oid,
                                        commitID: that.state.commitOrError.oid,
                                    },
                                    hoverifier: that.hoverifier,
                                    extensionsController: that.props.extensionsController,
                                },
                                lineNumbers: true,
                            }}
                            updateOnChange={`${that.props.repo.id}:${that.state.commitOrError.oid}`}
                            defaultFirst={25}
                            hideSearch={true}
                            noSummaryIfAllNodesVisible={true}
                            history={that.props.history}
                            location={that.props.location}
                        />
                    </>
                )}
                {that.state.hoverOverlayProps && (
                    <WebHoverOverlay
                        {...that.props}
                        {...that.state.hoverOverlayProps}
                        telemetryService={that.props.telemetryService}
                        hoverRef={that.nextOverlayElement}
                        onCloseButtonClick={that.nextCloseButtonClick}
                    />
                )}
            </div>
        )
    }

    private queryDiffs = (args: { first?: number }): Observable<GQL.IFileDiffConnection> =>
        queryRepositoryComparisonFileDiffs({
            ...args,
            repo: that.props.repo.id,
            base: commitParentOrEmpty(that.state.commitOrError as GQL.IGitCommit),
            head: (that.state.commitOrError as GQL.IGitCommit).oid,
        })
}

function commitParentOrEmpty(commit: GQL.IGitCommit): string {
    // 4b825dc642cb6eb9a060e54bf8d69288fbee4904 is `git hash-object -t tree /dev/null`, which is used as the base
    // when computing the `git diff` of the root commit.
    return commit.parents.length > 0 ? commit.parents[0].oid : '4b825dc642cb6eb9a060e54bf8d69288fbee4904'
}
