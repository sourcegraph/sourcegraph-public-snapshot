import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import ErrorIcon from '@sourcegraph/icons/lib/Error'
import LockIcon from '@sourcegraph/icons/lib/Lock'
import * as H from 'history'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import reactive from 'rx-component'
import { merge } from 'rxjs/observable/merge'
import { catchError } from 'rxjs/operators/catchError'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { map } from 'rxjs/operators/map'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { scan } from 'rxjs/operators/scan'
import { withLatestFrom } from 'rxjs/operators/withLatestFrom'
import { HeroPage } from '../components/HeroPage'
import { colorTheme, ColorTheme } from '../settings/theme'
import { eventLogger } from '../tracking/eventLogger'
import { EPERMISSIONDENIED, fetchThread } from './backend'
import { ThreadSharedItemPage } from './ThreadSharedItemPage'

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
            withLatestFrom(colorTheme),
            map(([{ location, history, user }, colorTheme]): Update => state => ({
                ...state,
                location,
                history,
                colorTheme,
                user,
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
        map(({ thread, threadID, location, history, colorTheme, error, user }: State): JSX.Element | null => {
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

            return <ThreadSharedItemPage item={{ thread }} user={user} location={location} history={history} />
        })
    )
})
