import LoaderIcon from '@sourcegraph/icons/lib/Loader'
import * as H from 'history'
import * as React from 'react'
import { Redirect } from 'react-router'
import reactive from 'rx-component'
import 'rxjs/add/observable/merge'
import 'rxjs/add/observable/of'
import 'rxjs/add/operator/bufferTime'
import 'rxjs/add/operator/catch'
import 'rxjs/add/operator/concat'
import 'rxjs/add/operator/distinctUntilChanged'
import 'rxjs/add/operator/do'
import 'rxjs/add/operator/filter'
import 'rxjs/add/operator/map'
import 'rxjs/add/operator/mergeMap'
import 'rxjs/add/operator/scan'
import 'rxjs/add/operator/startWith'
import 'rxjs/add/operator/take'
import 'rxjs/add/operator/withLatestFrom'
import { Observable } from 'rxjs/Observable'
import { Subject } from 'rxjs/Subject'
import { currentUser, fetchCurrentUser } from '../../auth'
import { acceptUserInvite } from '../backend'
import { VALID_USERNAME_REGEXP } from '../validation'

export interface Props {
    location: H.Location
}

interface State {
    email: string
    username: string
    displayName: string
    loading: boolean
    newOrgName?: string
    error?: Error
}

type Update = (state: State) => State

export const AcceptInvitePage = reactive<Props>(props => {

    const emailChangeEvents = new Subject<React.ChangeEvent<HTMLInputElement>>()
    const nextEmailChangeEvent = (event: React.ChangeEvent<HTMLInputElement>) => emailChangeEvents.next(event)

    const usernameChangeEvents = new Subject<React.ChangeEvent<HTMLInputElement>>()
    const nextUsernameChangeEvent = (event: React.ChangeEvent<HTMLInputElement>) => usernameChangeEvents.next(event)

    const displayNameChangeEvents = new Subject<React.ChangeEvent<HTMLInputElement>>()
    const nextDisplayNameChangeEvent = (event: React.ChangeEvent<HTMLInputElement>) => displayNameChangeEvents.next(event)

    const submitEvents = new Subject<React.FormEvent<HTMLFormElement>>()
    const nextSubmitEvent = (event: React.FormEvent<HTMLFormElement>) => submitEvents.next(event)

    /** The token in the query params */
    const inviteToken: Observable<string> = props
        .map(({ location }) => location)
        .distinctUntilChanged()
        .map(location => {
            const token = new URLSearchParams(location.search).get('token')
            if (!token) {
                throw new Error('No invite token in URL')
            }
            return token
        })

    /** The current email, either from auth or from updates to the input */
    const email: Observable<string> = currentUser
        .map(user => user && user.email)
        .filter((email: string | null): email is string => !!email)
        .take(1)
        .concat(emailChangeEvents.map(event => event.currentTarget.value))

    const username: Observable<string> = usernameChangeEvents
        .map(event => event.currentTarget.value)
        .startWith('')

    const displayName: Observable<string> = displayNameChangeEvents
        .map(event => event.currentTarget.value)
        .startWith('')

    return Observable.merge<Update>(
        // Any update to these should cause a rerender
        email.map(email => (state: State): State => ({ ...state, email })),
        username.map(username => (state: State): State => ({ ...state, username })),
        displayName.map(displayName => (state: State): State => ({ ...state, displayName })),

        // Form submits
        submitEvents
            .do(event => event.preventDefault())
            // Don't submit if form is invalid
            // Feedback is done through CSS
            .filter(event => event.currentTarget.checkValidity())
            // Get latest state values
            .withLatestFrom(inviteToken, email, username, displayName)
            .mergeMap(([, inviteToken, email, username, displayName]) =>
                // Show loader
                Observable.of<Update>(state => ({ ...state, loading: true }))
                    .concat(
                        acceptUserInvite({ inviteToken, username, email, displayName })
                            .mergeMap(orgMember =>
                                // Reload user
                                fetchCurrentUser()
                                    // Redirect
                                    .concat([(state: State): State => ({ ...state, loading: false, newOrgName: orgMember.org.name })])
                            )
                            // Show error
                            .catch(error => {
                                console.error(error)
                                return [(state: State): State => ({ ...state, loading: false, error })]
                            })
                    )
            )
    )
        // Buffer state updates in the same tick to avoid too many rerenders
        .bufferTime(0)
        .filter(updates => updates.length > 0)
        .map(updates => (state: State): State => updates.reduce((state, update) => update(state), state))

        .scan<Update, State>((state: State, update: Update) => update(state), { username: '', email: '', displayName: '', loading: false })
        .do(console.log.bind(console))
        .map(({ email, username, displayName, loading, error, newOrgName }) => (
            <form className='accept-invite-page' onSubmit={nextSubmitEvent}>
                {newOrgName && <Redirect to={`/settings/team/${newOrgName}`} />}
                <h1>You were invited to join a Sourcegraph team!</h1>

                {error && <p className='form-text text-error'>{error.message}</p>}

                <div className='form-group'>
                    <label>Your new username</label>
                    <input
                        type='text'
                        className='ui-text-box'
                        placeholder='yourusername'
                        pattern={VALID_USERNAME_REGEXP.toString().slice(1, -1)}
                        required={true}
                        autoCorrect='off'
                        value={username}
                        onChange={nextUsernameChangeEvent}
                        disabled={loading}
                    />
                    <small className='form-text'>A team name consists of letters, numbers, hyphens (-) and may not begin or end with a hyphen</small>
                </div>

                <div className='form-group'>
                    <label>Your display name</label>
                    <input
                        type='text'
                        className='ui-text-box'
                        placeholder='Your Name'
                        required={true}
                        autoCorrect='off'
                        value={displayName}
                        onChange={nextDisplayNameChangeEvent}
                        disabled={loading}
                    />
                </div>

                <div className='form-group'>
                    <label>Your company email</label>
                    <input
                        type='email'
                        className='ui-text-box'
                        placeholder='you@yourcompany.com'
                        required={true}
                        autoCorrect='off'
                        value={email}
                        onChange={nextEmailChangeEvent}
                        disabled={loading}
                    />
                </div>

                <div className='form-group accept-invite-page__actions'>
                    <button type='submit' className='btn btn-primary' disabled={loading}>Accept Invite</button>
                    {loading && <LoaderIcon className='icon-inline' />}
                </div>

            </form>
        ))
})
