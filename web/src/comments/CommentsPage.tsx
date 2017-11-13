import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import ErrorIcon from '@sourcegraph/icons/lib/Error'
import LockIcon from '@sourcegraph/icons/lib/Lock'
import * as H from 'history'
import * as React from 'react'
import { match } from 'react-router'
import { Link, Redirect } from 'react-router-dom'
import reactive from 'rx-component'
import 'rxjs/add/observable/merge'
import 'rxjs/add/operator/catch'
import 'rxjs/add/operator/distinctUntilChanged'
import 'rxjs/add/operator/map'
import 'rxjs/add/operator/mergeMap'
import 'rxjs/add/operator/scan'
import { Observable } from 'rxjs/Observable'
import { Subject } from 'rxjs/Subject'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { RepoNav } from '../repo/RepoNav'
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
}

interface State {
    sharedItem?: GQL.ISharedItem | null
    highlightLastComment?: boolean
    ulid: string
    location: H.Location
    history: H.History
    error?: any
}

type Update = (s: State) => State

/**
 * Renders a shared code comment's thread.
 */
export const CommentsPage = reactive<Props>(props => {
    const threadUpdates = new Subject<GQL.ISharedItemThread>()
    const nextThreadUpdate = (updatedThread: GQL.ISharedItemThread) => threadUpdates.next(updatedThread)

    eventLogger.logViewEvent('SharedItem')

    return Observable.merge(
        props.map(({ location, history }): Update => state => ({ ...state, location, history })),

        props
            .map(props => props.match.params.ulid)
            .distinctUntilChanged()
            .mergeMap(ulid =>
                fetchSharedItem(ulid)
                    .map((sharedItem): Update => state => ({ ...state, sharedItem, ulid, highlightLastComment: false }))
                    .catch((error): Update[] => {
                        console.error(error)
                        return [state => ({ ...state, error, ulid, highlightLastComment: false })]
                    })
            ),

        threadUpdates.map((thread): Update => state => ({
            ...state,
            sharedItem: state.sharedItem && {
                ...state.sharedItem,
                thread,
            },
            highlightLastComment: true,
        }))
    )
        .scan<Update, State>((state: State, update: Update) => update(state), {} as State)
        .map(({ sharedItem, highlightLastComment, ulid, location, history, error }: State): JSX.Element | null => {
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
            const signedIn = window.context.user
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
                        !sharedItem.thread.repoRevision && (
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
                                        <a href={editorURL} onClick={openEditor}>
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
        })
})

function getPageTitle(sharedItem: GQL.ISharedItem): string | undefined {
    const title = sharedItem.comment ? sharedItem.comment.title : sharedItem.thread.title
    if (title === '') {
        return sharedItem.thread.file
    }
    return title
}
