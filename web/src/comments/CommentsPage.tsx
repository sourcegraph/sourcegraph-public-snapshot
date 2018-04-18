import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import ErrorIcon from '@sourcegraph/icons/lib/Error'
import LockIcon from '@sourcegraph/icons/lib/Lock'
import * as H from 'history'
import { isEqual } from 'lodash-es'
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
import { EPERMISSIONDENIED, fetchSharedItem } from './backend'
import { ThreadSharedItemPage } from './ThreadSharedItemPage'

const SharedItemNotFound = () => (
    <HeroPage icon={DirectionalSignIcon} title="404: Not Found" subtitle="Sorry, we can&#39;t find anything here." />
)

interface Props extends RouteComponentProps<{ ulid: string }> {
    user: GQL.IUser | null
    isLightTheme: boolean
}

interface State {
    sharedItem?: GQL.ISharedItem | null
    ulid: string
    location: H.Location
    history: H.History
    isLightTheme: boolean
    error?: any
    user: GQL.IUser | null
}

type Update = (s: State) => State

/**
 * Renders a shared item (comment thread).
 */
export const CommentsPage = reactive<Props>(props => {
    eventLogger.logViewEvent('SharedItem')

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
            map(props => ({ ulid: props.match.params.ulid, isLightTheme: props.isLightTheme })),
            distinctUntilChanged((a, b) => isEqual(a, b)),
            mergeMap(({ ulid, isLightTheme }) =>
                fetchSharedItem({ ulid, isLightTheme }).pipe(
                    map((sharedItem): Update => state => ({ ...state, sharedItem, ulid })),
                    catchError((error): Update[] => {
                        console.error(error)
                        return [state => ({ ...state, error, ulid })]
                    })
                )
            )
        )
    ).pipe(
        scan<Update, State>((state: State, update: Update) => update(state), {} as State),
        map(({ sharedItem, ulid, location, history, error, user, isLightTheme }: State): JSX.Element | null => {
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

            return (
                <ThreadSharedItemPage
                    item={sharedItem}
                    ulid={ulid}
                    user={user}
                    location={location}
                    history={history}
                    isLightTheme={isLightTheme}
                />
            )
        })
    )
})
