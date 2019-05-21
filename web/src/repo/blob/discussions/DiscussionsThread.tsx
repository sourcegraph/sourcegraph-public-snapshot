import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as H from 'history'
import { isEqual } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import * as React from 'react'
import { Redirect } from 'react-router'
import { combineLatest, Subject, Subscription, throwError } from 'rxjs'
import { catchError, delay, distinctUntilChanged, map, repeatWhen, startWith, switchMap, tap } from 'rxjs/operators'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError } from '../../../../../shared/src/util/errors'
import { addCommentToThread, fetchDiscussionThreadAndComments, updateComment } from '../../../discussions/backend'
import { DiscussionsComment } from '../../../discussions/DiscussionsComment'
import { eventLogger } from '../../../tracking/eventLogger'
import { formatHash } from '../../../util/url'
import { DiscussionsInput, TitleMode } from './DiscussionsInput'
import { DiscussionsNavbar } from './DiscussionsNavbar'

interface Props extends ExtensionsControllerProps {
    threadIDWithoutKind: string
    commentIDWithoutKind?: string
    filePath?: string
    showNavbar?: boolean
    history: H.History
    location: H.Location
    forceURL?: boolean
    className?: string
    commentClassName?: string

    /**
     * Do not show the first comment in the thread. This is useful when the first comment is treated
     * specially (e.g., as the description of a thread or check).
     */
    skipFirstComment?: boolean
}

interface State {
    loading: boolean
    error?: any
    thread?: GQL.IDiscussionThread
}

export class DiscussionsThread extends React.PureComponent<Props, State> {
    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)
        this.state = {
            loading: true,
        }
    }

    public componentDidMount(): void {
        eventLogger.logViewEvent('DiscussionsThread')

        // TODO(slimsag:discussions): ASAP: changing threadID manually in URL does not work. Can't click links to threads/comments effectively.
        this.subscriptions.add(
            combineLatest(this.componentUpdates.pipe(startWith(this.props)))
                .pipe(
                    distinctUntilChanged(([a], [b]) => a.threadIDWithoutKind === b.threadIDWithoutKind),
                    switchMap(([props]) =>
                        fetchDiscussionThreadAndComments(props.threadIDWithoutKind).pipe(
                            map(thread => ({ thread, error: undefined, loading: false })),
                            catchError(error => {
                                console.error(error)
                                return [{ error, loading: false }]
                            }),
                            repeatWhen(delay(2500))
                        )
                    )
                )
                .subscribe(
                    stateUpdate => this.setState(state => ({ ...state, ...stateUpdate })),
                    err => console.error(err)
                )
        )
        this.componentUpdates.next(this.props)
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        // TODO(slimsag:discussions): future: test error state + cleanup CSS

        const { error, loading, thread } = this.state
        const { location, commentIDWithoutKind } = this.props

        // If the thread is loaded, ensure that the URL hash is updated to
        // reflect the line that the discussion was created on.
        if (thread && this.props.forceURL) {
            // TODO(sqs): support multiple thread targets
            const target =
                thread.targets && thread.targets.nodes && thread.targets.nodes.length > 0
                    ? thread.targets.nodes[0]
                    : undefined

            const desiredHash = this.urlHashWithLine(
                thread,
                target,
                commentIDWithoutKind ? { idWithoutKind: commentIDWithoutKind } : undefined
            )
            if (!hashesEqual(desiredHash, location.hash)) {
                const discussionURL = location.pathname + location.search + desiredHash
                return <Redirect to={discussionURL} />
            }
        }

        return (
            <div className={`discussions-thread ${this.props.className || ''}`}>
                {this.props.showNavbar && (
                    <DiscussionsNavbar {...this.props} threadTitle={thread ? thread.title : undefined} />
                )}
                {loading && <LoadingSpinner className="icon-inline" />}
                {error && (
                    <div className="discussions-thread__error alert alert-danger">
                        <AlertCircleIcon className="icon-inline discussions-thread__error-icon" />
                        Error loading thread: {error.message}
                    </div>
                )}
                {thread && (
                    <div className="discussions-thread__comments">
                        {(this.props.skipFirstComment ? thread.comments.nodes.slice(1) : thread.comments.nodes).map(
                            node => (
                                <DiscussionsComment
                                    key={node.id}
                                    {...this.props}
                                    threadID={thread.id}
                                    comment={node}
                                    onReport={this.onCommentReport}
                                    onClearReports={this.onCommentClearReports}
                                    onDelete={this.onCommentDelete}
                                    extensionsController={this.props.extensionsController}
                                    className={this.props.commentClassName}
                                />
                            )
                        )}
                        <DiscussionsInput
                            className="mt-3"
                            key="input"
                            submitLabel="Comment"
                            titleMode={TitleMode.None}
                            onSubmit={this.onSubmit}
                            {...this.props}
                        />
                    </div>
                )}
            </div>
        )
    }

    /**
     * Produces a URL hash for linking to the given discussion thread and the
     * line that it was created on.
     * @param thread The thread to link to.
     */
    private urlHashWithLine(
        thread: Pick<GQL.IDiscussionThread, 'idWithoutKind'>,
        target: Pick<GQL.IDiscussionThreadTargetRepo, '__typename' | 'selection'> | undefined,
        comment?: Pick<GQL.IDiscussionComment, 'idWithoutKind'>
    ): string {
        const hash = new URLSearchParams()
        hash.set('tab', 'discussions')
        hash.set('threadID', thread.idWithoutKind)
        if (comment) {
            hash.set('commentID', comment.idWithoutKind)
        }

        return target && target.__typename === 'DiscussionThreadTargetRepo' && target.selection !== null
            ? formatHash(
                  {
                      line: target.selection.startLine + 1,
                      character: target.selection.startCharacter,
                      endLine:
                          // The 0th character means the selection ended at the end of the previous
                          // line.
                          (target.selection.endCharacter === 0
                              ? target.selection.endLine - 1
                              : target.selection.endLine) + 1,
                      endCharacter: target.selection.endCharacter,
                  },
                  hash
              )
            : '#' + hash.toString()
    }

    private onSubmit = (title: string, contents: string) => {
        eventLogger.log('RepliedToDiscussion')
        if (!this.state.thread) {
            throw new Error('no thread')
        }
        return addCommentToThread(this.state.thread.id, contents).pipe(
            tap(thread => this.setState({ thread })),
            map(thread => undefined),
            catchError(e => throwError(new Error('Error creating comment: ' + asError(e).message)))
        )
    }

    private onCommentReport = (comment: GQL.IDiscussionComment, reason: string) =>
        updateComment({ commentID: comment.id, report: reason }).pipe(
            tap(thread => this.setState({ thread })),
            map(thread => undefined)
        )

    private onCommentClearReports = (comment: GQL.IDiscussionComment) =>
        updateComment({ commentID: comment.id, clearReports: true }).pipe(
            tap(thread => this.setState({ thread })),
            map(thread => undefined)
        )

    private onCommentDelete = (comment: GQL.IDiscussionComment) =>
        // TODO: Support deleting the whole thread, and/or fix this when it is deleting the 1st comment
        // in a thread. See https://github.com/sourcegraph/sourcegraph/issues/429.
        updateComment({ commentID: comment.id, delete: true }).pipe(
            tap(thread => this.setState({ thread })),
            map(thread => undefined)
        )
}

/**
 * @returns Whether the 2 URI fragments contain the same keys and values (assuming they contain a
 * `#` then HTML-form-encoded keys and values like `a=b&c=d`).
 */
function hashesEqual(a: string, b: string): boolean {
    if (a.startsWith('#')) {
        a = a.slice(1)
    }
    if (b.startsWith('#')) {
        b = b.slice(1)
    }
    const canonicalize = (hash: string): string[] =>
        Array.from(new URLSearchParams(hash).entries())
            .map(([key, value]) => `${key}=${value}`)
            .sort()
    return isEqual(canonicalize(a), canonicalize(b))
}
