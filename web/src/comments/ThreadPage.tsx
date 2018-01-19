import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import ErrorIcon from '@sourcegraph/icons/lib/Error'
import LockIcon from '@sourcegraph/icons/lib/Lock'
import * as H from 'history'
import isEqual from 'lodash/isEqual'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import reactive from 'rx-component'
import { merge } from 'rxjs/observable/merge'
import { catchError } from 'rxjs/operators/catchError'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { map } from 'rxjs/operators/map'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { scan } from 'rxjs/operators/scan'
import { HeroPage } from '../components/HeroPage'
import { eventLogger } from '../tracking/eventLogger'
import { EPERMISSIONDENIED, fetchThread } from './backend'
import { ThreadSharedItemPage } from './ThreadSharedItemPage'

const ThreadNotFound = () => (
    <HeroPage icon={DirectionalSignIcon} title="404: Not Found" subtitle="Sorry, we can&#39;t find anything here." />
)

interface Props extends RouteComponentProps<{ threadID: GQLID }> {
    user: GQL.IUser | null
    isLightTheme: boolean
}

interface State {
    thread?: GQL.IThread | null
    location: H.Location
    history: H.History
    isLightTheme: boolean
    error?: any
    user: GQL.IUser | null
}

type Update = (s: State) => State

/**
 * The page for a comment thread. Similar to CommentsPage, but for non-shared-item threads.
 */
export const ThreadPage = reactive<Props>(props => {
    eventLogger.logViewEvent('Thread')

    return merge(
        props.pipe(
            map(({ location, history, user, isLightTheme }): Update => state => ({
                ...state,
                location,
                history,
                user,
                isLightTheme,
            }))
        ),
        props.pipe(
            map(props => ({ threadID: props.match.params.threadID, isLightTheme: props.isLightTheme })),
            distinctUntilChanged((a, b) => isEqual(a, b)),
            mergeMap(({ threadID, isLightTheme }) =>
                fetchThread(threadID, isLightTheme).pipe(
                    map((thread): Update => state => ({ ...state, thread, highlightLastComment: false })),
                    catchError((error): Update[] => {
                        console.error(error)
                        return [state => ({ ...state, error, highlightLastComment: false })]
                    })
                )
            )
        )
    ).pipe(
        scan<Update, State>((state: State, update: Update) => update(state), {} as State),
        map(({ thread, location, history, error, user, isLightTheme }: State): JSX.Element | null => {
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

            return (
                <ThreadSharedItemPage
                    item={{ thread }}
                    user={user}
                    location={location}
                    history={history}
                    isLightTheme={isLightTheme}
                />
            )
        })
    )
})
