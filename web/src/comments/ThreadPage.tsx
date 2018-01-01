import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import ErrorIcon from '@sourcegraph/icons/lib/Error'
import LockIcon from '@sourcegraph/icons/lib/Lock'
import * as H from 'history'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Redirect } from 'react-router-dom'
import reactive from 'rx-component'
import { merge } from 'rxjs/observable/merge'
import { catchError } from 'rxjs/operators/catchError'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { map } from 'rxjs/operators/map'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { scan } from 'rxjs/operators/scan'
import { withLatestFrom } from 'rxjs/operators/withLatestFrom'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { RepoNav } from '../repo/RepoNav'
import { colorTheme, ColorTheme } from '../settings/theme'
import { eventLogger } from '../tracking/eventLogger'
import { EPERMISSIONDENIED, fetchThread } from './backend'
import { CodeView } from './CodeView'
import { Comment } from './Comment'
import { SecurityWidget } from './SecurityWidget'

const ThreadNotFound = () => (
    <HeroPage icon={DirectionalSignIcon} title="404: Not Found" subtitle="Sorry, we can&#39;t find anything here." />
)

interface Props extends RouteComponentProps<{ threadID: GQLID }> {
    user: GQL.IUser | null
}

interface State {
    thread?: GQL.IThread | null
    threadID: string
    location: H.Location
    history: H.History
    colorTheme: ColorTheme
    error?: any
    signedIn: boolean
}

type Update = (s: State) => State

/**
 * The page for a comment thread.
 *
 * TODO(sqs): this is duplicated from CommentsPage, with some things omitted.
 */
export const ThreadPage = reactive<Props>(props => {
    eventLogger.logViewEvent('Thread')

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
            map(props => props.match.params.threadID),
            distinctUntilChanged(),
            withLatestFrom(colorTheme),
            mergeMap(([threadID, colorTheme]) =>
                fetchThread(threadID, colorTheme === 'light').pipe(
                    map((thread): Update => state => ({ ...state, thread, threadID, highlightLastComment: false })),
                    catchError((error): Update[] => {
                        console.error(error)
                        return [state => ({ ...state, error, threadID, highlightLastComment: false })]
                    })
                )
            )
        )
    ).pipe(
        scan<Update, State>((state: State, update: Update) => update(state), {} as State),
        map(({ thread, threadID, location, history, colorTheme, error, signedIn }: State): JSX.Element | null => {
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
            if (thread === undefined) {
                // TODO(slimsag): future: add loading screen
                return null
            }
            if (thread === null) {
                return <ThreadNotFound />
            }

            // If not logged in, redirect to sign in
            const newUrl = new URL(window.location.href)
            newUrl.pathname = '/sign-in'
            newUrl.searchParams.set('returnTo', window.location.href)
            const signInURL = newUrl.pathname + newUrl.search
            if (!signedIn) {
                return <Redirect to={signInURL} />
            }

            return (
                <div className="comments-page">
                    <PageTitle title={thread.title} />
                    <RepoNav
                        repoPath={thread.repo.canonicalRemoteID}
                        rev={thread.branch || thread.repoRevision}
                        filePath={thread.file}
                        isDirectory={false}
                        hideCopyLink={true}
                        breadcrumbDisabled={false}
                        revSwitcherDisabled={false}
                        line={thread && thread.startLine}
                        location={location}
                        history={history}
                    />
                    {thread &&
                        !thread.linesRevision && (
                            <div className="comments-page__no-revision">
                                <ErrorIcon className="icon-inline comments-page__error-icon" />
                                {thread.comments.length === 0
                                    ? 'This code snippet was created from code that was not pushed. File or line numbers may have changed since this snippet was created.'
                                    : 'This discussion was created on code that was not pushed. File or line numbers may have changed since this discussion was created.'}
                            </div>
                        )}
                    <div className="comments-page__content">
                        {thread && CodeView({ thread })}
                        {thread &&
                            thread.comments.map((comment, index) => (
                                <Comment location={location} comment={comment} key={comment.id} forceTargeted={false} />
                            ))}
                        {thread && <SecurityWidget sharedItem={{ public: false }} />}
                    </div>
                </div>
            )
        })
    )
})
