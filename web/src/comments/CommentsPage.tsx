import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import ErrorIcon from '@sourcegraph/icons/lib/Error'
import LockIcon from '@sourcegraph/icons/lib/Lock'
import * as H from 'history'
import * as React from 'react'
import { match } from 'react-router'
import { Link, Redirect } from 'react-router-dom'
import reactive from 'rx-component'
import { merge } from 'rxjs/observable/merge'
import { catchError } from 'rxjs/operators/catchError'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { map } from 'rxjs/operators/map'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { scan } from 'rxjs/operators/scan'
import { withLatestFrom } from 'rxjs/operators/withLatestFrom'
import { Subject } from 'rxjs/Subject'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { RepoNav } from '../repo/RepoNav'
import { colorTheme, ColorTheme } from '../settings/theme'
import { eventLogger } from '../tracking/eventLogger'
import { toEditorURL } from '../util/url'
import { EPERMISSIONDENIED, fetchSharedItem } from './backend'
import { CodeView } from './CodeView'
import { Comment } from './Comment'
import { CommentsInput } from './CommentsInput'
import { SecurityWidget } from './SecurityWidget'

const SharedItemNotFound = () => (
    <HeroPage icon={DirectionalSignIcon} title="404: Not Found" subtitle="Sorry, we can&#39;t find anything here." />
)

interface Props {
    match: match<{ ulid: string }>
    location: H.Location
    history: H.History
    user: GQL.IUser | null
}

interface State {
    sharedItem?: GQL.ISharedItem | null
    highlightLastComment?: boolean
    ulid: string
    location: H.Location
    history: H.History
    colorTheme: ColorTheme
    error?: any
    signedIn: boolean
}

type Update = (s: State) => State

/**
 * Renders a shared code comment's thread.
 */
export const CommentsPage = reactive<Props>(props => {
    const threadUpdates = new Subject<GQL.ISharedItemThread>()
    const nextThreadUpdate = (updatedThread: GQL.ISharedItemThread) => threadUpdates.next(updatedThread)

    eventLogger.logViewEvent('SharedItem')

    return merge(
        props.pipe(
            withLatestFrom(colorTheme),
            map(([{ location, history, user }, colorTheme]): Update => state => ({
                ...state,
                location,
                history,
                colorTheme,
                signedIn: !!user,
            }))
        ),
        props.pipe(
            map(props => props.match.params.ulid),
            distinctUntilChanged(),
            withLatestFrom(colorTheme),
            mergeMap(([ulid, colorTheme]) =>
                fetchSharedItem(ulid, colorTheme === 'light').pipe(
                    map((sharedItem): Update => state => ({ ...state, sharedItem, ulid, highlightLastComment: false })),
                    catchError((error): Update[] => {
                        console.error(error)
                        return [state => ({ ...state, error, ulid, highlightLastComment: false })]
                    })
                )
            )
        ),

        threadUpdates.pipe(
            map((thread): Update => state => ({
                ...state,
                sharedItem: state.sharedItem && {
                    ...state.sharedItem,
                    thread,
                },
                highlightLastComment: true,
            }))
        )
    ).pipe(
        scan<Update, State>((state: State, update: Update) => update(state), {} as State),
        map(
            ({
                sharedItem,
                highlightLastComment,
                ulid,
                location,
                history,
                colorTheme,
                error,
                signedIn,
            }: State): JSX.Element | null => {
                if (error) {
                    if (error.code === EPERMISSIONDENIED) {
                        return (
                            <HeroPage
                                icon={LockIcon}
                                title="Permission denied."
                                subtitle={'You must be a member of the organization to view this page.'}
                            />
                        )
                    }
                    return <HeroPage icon={ErrorIcon} title="Something went wrong." subtitle={error.message} />
                }
                if (sharedItem === undefined) {
                    // TODO(slimsag): future: add loading screen
                    return null
                }
                if (sharedItem === null) {
                    return <SharedItemNotFound />
                }

                // If not logged in, redirect to sign in
                const newUrl = new URL(window.location.href)
                newUrl.pathname = '/sign-in'
                newUrl.searchParams.set('returnTo', window.location.href)
                const signInURL = newUrl.pathname + newUrl.search
                if (!sharedItem.public && !signedIn) {
                    return <Redirect to={signInURL} />
                }

                const editorURL = toEditorURL(
                    sharedItem.thread.repo.remoteUri,
                    sharedItem.thread.branch || sharedItem.thread.repoRevision,
                    sharedItem.thread.file,
                    { line: sharedItem.thread.startLine },
                    sharedItem.thread.id
                )
                const openEditor = () => {
                    eventLogger.log('OpenInNativeAppClicked')
                }

                return (
                    <div className="comments-page">
                        <PageTitle title={getPageTitle(sharedItem)} />
                        {/* TODO(slimsag): future: do not disable breadcrumb _if_ the repository is public */}
                        <RepoNav
                            repoPath={sharedItem.thread.repo.remoteUri}
                            rev={sharedItem.thread.branch || sharedItem.thread.repoRevision}
                            filePath={sharedItem.thread.file}
                            isDirectory={false}
                            hideCopyLink={true}
                            customEditorURL={editorURL}
                            breadcrumbDisabled={true}
                            revSwitcherDisabled={true}
                            line={sharedItem && sharedItem.thread.startLine}
                            location={location}
                            history={history}
                        />
                        {sharedItem &&
                            !sharedItem.thread.linesRevision && (
                                <div className="comments-page__no-revision">
                                    <ErrorIcon className="icon-inline comments-page__error-icon" />
                                    {sharedItem.thread.comments.length === 0
                                        ? 'This code snippet was created from code that was not pushed. File or line numbers may have changed since this snippet was created.'
                                        : 'This discussion was created on code that was not pushed. File or line numbers may have changed since this discussion was created.'}
                                </div>
                            )}
                        <div className="comments-page__content">
                            {sharedItem &&
                                !sharedItem.thread.lines && (
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
                            {sharedItem && CodeView(sharedItem)}
                            {sharedItem &&
                                sharedItem.thread.comments.map((comment, index) => (
                                    <Comment
                                        location={location}
                                        comment={comment}
                                        key={comment.id}
                                        forceTargeted={
                                            (highlightLastComment && index === sharedItem.thread.comments.length - 1) ||
                                            false
                                        }
                                    />
                                ))}
                            {sharedItem &&
                                sharedItem.thread.comments.length === 0 && (
                                    <a
                                        className="btn btn-primary btn-block comments-page__reply-in-editor"
                                        href={editorURL}
                                        target="sourcegraphapp"
                                        onClick={openEditor}
                                    >
                                        Open in Sourcegraph Editor
                                    </a>
                                )}
                            {sharedItem &&
                                sharedItem.thread.comments.length !== 0 &&
                                (signedIn ? (
                                    <CommentsInput
                                        editorURL={editorURL}
                                        onOpenEditor={openEditor}
                                        threadID={sharedItem.thread.id}
                                        ulid={ulid}
                                        onThreadUpdated={nextThreadUpdate}
                                    />
                                ) : (
                                    <Link className="btn btn-primary comments-page__sign-in" to={signInURL}>
                                        Sign in to comment
                                    </Link>
                                ))}
                            {sharedItem && <SecurityWidget sharedItem={sharedItem} />}
                        </div>
                    </div>
                )
            }
        )
    )
})

function getPageTitle(sharedItem: GQL.ISharedItem): string | undefined {
    const title = sharedItem.comment ? sharedItem.comment.title : sharedItem.thread.title
    if (title === '') {
        return sharedItem.thread.file
    }
    return title
}
