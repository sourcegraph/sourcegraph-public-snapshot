import * as React from 'react'
import { match } from 'react-router'
import reactive from 'rx-component'
import { merge } from 'rxjs/observable/merge'
import { map } from 'rxjs/operators/map'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { scan } from 'rxjs/operators/scan'
import { currentUser } from '../../auth'
import { eventLogger } from '../../tracking/eventLogger'
import { fetchAllUsers } from '../backend'
import { UserAvatar } from '../user/UserAvatar'

export interface Props {
    match: match<{ orgName: string }>
}

export interface State {
    user?: GQL.IUser
    users?: GQL.IUser[]
}

type Update = (s: State) => State

export const OrgAdmin = reactive<Props>(props => {
    eventLogger.logViewEvent('ServerAdmin')

    const today = new Date()
    let expiry = new Date()
    if (window.context.license) {
        expiry = new Date(window.context.license.Expiry)
    }
    const timeDiff = Math.abs(expiry.getTime() - today.getTime())
    const dateDiff = Math.ceil(timeDiff / (1000 * 3600 * 24))
    return merge<Update>(
        currentUser.pipe(
            mergeMap(user => {
                if (!user) {
                    return [(state: State): State => ({ ...state, user: undefined })]
                }

                return fetchAllUsers().pipe(
                    map(users => (state: State): State => ({
                        ...state,
                        user,
                        users: users || undefined,
                    }))
                )
            })
        )
    ).pipe(
        scan<Update, State>((state: State, update: Update) => update(state), {} as State),
        map(({ user, users }: State): JSX.Element | null => (
            <div className="org-admin">
                <h1>Server Admin Page</h1>
                {window.context.license &&
                    window.context.license.Expiry && (
                        <p className="alert alert-primary">
                            <b>Trial</b>. {dateDiff} days remaining. Contact{' '}
                            <a href="mailto:sales@sourcegraph.com">sales@sourcegraph.com</a> to purchase.
                        </p>
                    )}

                <table className="table table-hover org__table">
                    <thead>
                        <tr>
                            <th className="org__avatar-cell" />
                            <th>Name</th>
                            <th>Username</th>
                            <th>Page views</th>
                            <th>Search queries</th>
                        </tr>
                    </thead>
                    <tbody>
                        {users &&
                            users.map(user => (
                                <tr key={user.id}>
                                    <td className="org__avatar-cell">
                                        <UserAvatar user={user} size={64} />
                                    </td>
                                    <td>{user.displayName}</td>
                                    <td>{user.username}</td>
                                    <td>{user.activity.pageViews}</td>
                                    <td>{user.activity.searchQueries}</td>
                                </tr>
                            ))}
                    </tbody>
                </table>
                <div className="org-admin__section">
                    <h1>Help and support</h1>
                    <p>
                        Contact <a href="mailto:support@sourcegraph.com">support@sourcegraph.com</a> or{' '}
                        <a href="https://about.sourcegraph.com/docs/server/api">
                            view Sourcegraph Server documentation
                        </a>
                    </p>
                </div>
            </div>
        ))
    )
})
