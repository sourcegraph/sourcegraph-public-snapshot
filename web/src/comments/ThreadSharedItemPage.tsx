import ErrorIcon from '@sourcegraph/icons/lib/Error'
import * as H from 'history'
import * as React from 'react'
import { Link, Redirect } from 'react-router-dom'
import reactive from 'rx-component'
import { merge } from 'rxjs/observable/merge'
import { map } from 'rxjs/operators/map'
import { scan } from 'rxjs/operators/scan'
import { Subject } from 'rxjs/Subject'
import { PageTitle } from '../components/PageTitle'
import { ToggleLineWrap } from '../repo/blob/actions/ToggleLineWrap'
import { FilePathBreadcrumb } from '../repo/FilePathBreadcrumb'
import { RepoHeader } from '../repo/RepoHeader'
import { RepoHeaderActionPortal } from '../repo/RepoHeaderActionPortal'
import { eventLogger } from '../tracking/eventLogger'
import { toEditorURL } from '../util/url'
import { ThreadRevisionAction } from './actions/ThreadRevisionAction'
import { CodeView } from './CodeView'
import { Comment } from './Comment'
import { CommentsInput } from './CommentsInput'
import { SecurityWidget } from './SecurityWidget'

/**
 * Common type for shared items and non-shared-item threads.
 */
type Item = GQL.ISharedItem | { thread: GQL.IThread }

function isThread(item: Item): item is { thread: GQL.IThread } {
    return (item as GQL.ISharedItem).public === undefined
}

function isSharedItem(item: Item): item is GQL.ISharedItem {
    return (item as GQL.ISharedItem).public !== undefined
}

interface Props {
    /**
     * The shared item or (wrapped) thread.
     */
    item: Item

    /**
     * If item is a shared item, this is its ULID.
     */
    ulid?: string

    /**
     * The currently authenticated user.
     */
    user: GQL.IUser | null

    location: H.Location
    history: H.History
    isLightTheme: boolean
}

interface State {
    item: Item
    ulid?: string
    highlightLastComment?: boolean
    highlightComment: GQLID | null
    location: H.Location
    history: H.History
    isLightTheme: boolean
    signedIn: boolean
    wrapCode: boolean
}

type Update = (s: State) => State

/**
 * Renders a shared item or thread. This component expects to be wrapped by another component that
 * already fetched the shared item or thread (CommentsPage or ThreadPage).
 */
export const ThreadSharedItemPage = reactive<Props>(props => {
    const threadUpdates = new Subject<GQL.IThread | GQL.ISharedItemThread>()
    const nextThreadUpdate = (updatedThread: GQL.IThread | GQL.ISharedItemThread) => threadUpdates.next(updatedThread)

    eventLogger.logViewEvent('SharedItem')

    const codeWrapUpdates = new Subject<boolean>()
    const nextWrapCodeChange = (codeWrap: boolean) => codeWrapUpdates.next(codeWrap)

    return merge(
        codeWrapUpdates.pipe(map((wrapCode): Update => state => ({ ...state, wrapCode }))),
        props.pipe(
            map(({ item, ulid, location, history, user, isLightTheme }): Update => state => ({
                ...state,
                item,
                ulid,
                location,
                highlightComment: new URLSearchParams(location.search).get('id'),
                history,
                isLightTheme,
                signedIn: !!user,
            }))
        ),

        threadUpdates.pipe(
            map((thread): Update => state => ({
                ...state,
                item:
                    state.item &&
                    ({
                        ...state.item,
                        thread,
                    } as Item),
                highlightLastComment: true,
            }))
        )
    ).pipe(
        scan<Update, State>((state: State, update: Update) => update(state), {} as State),
        map(
            ({
                item,
                ulid,
                highlightLastComment,
                highlightComment,
                location,
                history,
                isLightTheme,
                signedIn,
                wrapCode,
            }: State): JSX.Element | null => {
                const isPublic = isSharedItem(item) && item.public

                // If not logged in, redirect to sign in
                const newUrl = new URL(window.location.href)
                newUrl.pathname = '/sign-in'
                newUrl.searchParams.set('returnTo', window.location.href)
                const signInURL = newUrl.pathname + newUrl.search
                if (!isPublic && !signedIn) {
                    return <Redirect to={signInURL} />
                }

                const itemRepo = isThread(item) ? item.thread.repo.canonicalRemoteID : item.thread.repo.remoteUri

                const editorURL = toEditorURL(
                    itemRepo,
                    item.thread.branch || item.thread.repoRevision,
                    item.thread.file,
                    { line: item.thread.startLine },
                    item.thread.databaseID
                )
                const openEditor = () => {
                    eventLogger.log('OpenInNativeAppClicked')
                }

                const repoPath = item.thread.repo.repository ? item.thread.repo.repository.uri : itemRepo

                return (
                    <div className="comments-page">
                        <PageTitle title={getPageTitle(item)} />
                        {/* TODO(slimsag): future: do not disable breadcrumb _if_ the repository is public */}
                        <RepoHeader
                            repo={
                                item.thread.repo.repository || {
                                    uri: itemRepo,
                                    enabled: true,
                                    viewerCanAdminister: false,
                                }
                            }
                            disableLinks={!item.thread.repo.repository}
                            rev={item.thread.branch || item.thread.repoRevision}
                            filePath={item.thread.file}
                            location={location}
                            history={history}
                        />
                        <RepoHeaderActionPortal
                            position="left"
                            element={
                                <ThreadRevisionAction
                                    key="item.thread-revision"
                                    repoPath={repoPath}
                                    branch={item.thread.branch || undefined}
                                    rev={item.thread.repoRevision}
                                    link={!!item.thread.repo.repository}
                                />
                            }
                        />
                        <RepoHeaderActionPortal
                            position="right"
                            key="toggle-line-wrap"
                            element={<ToggleLineWrap key="toggle-line-wrap" onDidUpdate={nextWrapCodeChange} />}
                        />
                        {item.thread.file && (
                            <RepoHeaderActionPortal
                                position="nav"
                                element={
                                    <FilePathBreadcrumb
                                        key="path"
                                        repoPath={repoPath}
                                        rev={item.thread.repoRevision}
                                        filePath={item.thread.file}
                                        isDir={false}
                                    />
                                }
                            />
                        )}
                        {item &&
                            !item.thread.linesRevision && (
                                <div className="comments-page__no-revision">
                                    <ErrorIcon className="icon-inline comments-page__error-icon" />
                                    {item.thread.comments.length === 0
                                        ? 'This code snippet was created from code that was not pushed. File or line numbers may have changed since this snippet was created.'
                                        : 'This discussion was created on code that was not pushed. File or line numbers may have changed since this discussion was created.'}
                                </div>
                            )}
                        <div className="comments-page__content">
                            {item &&
                                !item.thread.lines && (
                                    <div className="comments-page__no-shared-code-container">
                                        <div className="comments-page__no-shared-code">
                                            The author of this discussion did not{' '}
                                            <a href="https://about.sourcegraph.com/docs/editor/share-code">
                                                share the code
                                            </a>.&nbsp;
                                            <a href={editorURL} target="sourcegraphapp" onClick={openEditor}>
                                                Open in Sourcegraph Editor
                                            </a>{' '}
                                            to see code.
                                        </div>
                                    </div>
                                )}
                            {item && <CodeView thread={item.thread} wrapCode={wrapCode} />}
                            {item &&
                                (item.thread.comments as (GQL.ISharedItemComment | GQL.IComment)[]).map(
                                    (comment, index) => (
                                        <Comment
                                            location={location}
                                            comment={comment}
                                            key={comment.id}
                                            forceTargeted={
                                                (highlightLastComment && index === item.thread.comments.length - 1) ||
                                                highlightComment === comment.id ||
                                                highlightComment === String(comment.databaseID)
                                            }
                                        />
                                    )
                                )}
                            {item &&
                                item.thread.comments.length === 0 && (
                                    <a
                                        className="btn btn-primary btn-block comments-page__reply-in-editor"
                                        href={editorURL}
                                        target="sourcegraphapp"
                                        onClick={openEditor}
                                    >
                                        Open in Sourcegraph Editor
                                    </a>
                                )}
                            {item &&
                                item.thread.comments.length !== 0 &&
                                (signedIn ? (
                                    <CommentsInput
                                        editorURL={editorURL}
                                        onOpenEditor={openEditor}
                                        threadID={item.thread.id}
                                        ulid={isSharedItem(item) ? ulid : undefined}
                                        onThreadUpdated={nextThreadUpdate}
                                        isLightTheme={isLightTheme}
                                    />
                                ) : (
                                    <Link className="btn btn-primary comments-page__sign-in" to={signInURL}>
                                        Sign in to comment
                                    </Link>
                                ))}
                            {item && <SecurityWidget isPublic={isPublic} />}
                        </div>
                    </div>
                )
            }
        )
    )
})

function getPageTitle(thread: Item): string | undefined {
    const title = isSharedItem(thread) && thread.comment ? thread.comment.title : thread.thread.title
    if (title === '') {
        return thread.thread.file
    }
    return title
}
