import LoaderIcon from '@sourcegraph/icons/lib/Loader'
import * as H from 'history'
import { Base64 } from 'js-base64'
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
import 'rxjs/add/operator/withLatestFrom'
import { Observable } from 'rxjs/Observable'
import { Subject } from 'rxjs/Subject'
import { fetchCurrentUser } from '../../auth'
import { events } from '../../tracking/events'
import { acceptUserInvite } from '../backend'

export interface Props {
    location: H.Location
}

interface State {
    email: string
    emailVerified: boolean
    loading: boolean
    hasSubmitted: boolean
    orgName?: string
    error?: Error
}

type Update = (state: State) => State

interface TokenPayload {
    email: string
    orgID: number
    orgName: string
}

export const AcceptInvitePage = reactive<Props>(props => {

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

    const tokenPayload: Observable<TokenPayload> = inviteToken
        .map(token => JSON.parse(Base64.decode(token.split('.')[1])))

    return Observable.merge<Update>(
        // Any update to these should cause a rerender
        tokenPayload.map(payload => (state: State): State => ({ ...state, orgName: payload.orgName, email: payload.email })),

        // Form submits
        submitEvents
            .do(event => event.preventDefault())
            // Don't submit if form is invalid
            // Feedback is done through CSS
            .filter(event => event.currentTarget.checkValidity())
            // Get latest state values
            .withLatestFrom(inviteToken, tokenPayload)
            .mergeMap(([, inviteToken, tokenPayload]) =>
                // Show loader
                Observable.of<Update>(state => ({ ...state, loading: true, email: tokenPayload.email }))
                    .concat(
                        acceptUserInvite({ inviteToken })
                            .do(status => {
                                const eventProps = {
                                    user_email: tokenPayload.email,
                                    org_name: tokenPayload.orgName,
                                }
                                if (status.emailVerified) {
                                    events.InviteAccepted.log(eventProps)
                                } else {
                                    events.AcceptInviteFailed.log(eventProps)
                                }
                            })
                            .mergeMap(status =>
                                // Reload user
                                fetchCurrentUser()
                                    // Redirect
                                    .concat([(state: State): State => ({
                                        ...state,
                                        loading: false,
                                        hasSubmitted: true,
                                        emailVerified: status.emailVerified,
                                    })]),
                            )
                            // Show error
                            .catch(error => {
                                console.error(error)
                                return [(state: State): State => ({ ...state, hasSubmitted: true, loading: false, error })]
                            }),
                    ),
            ),
    )
        // Buffer state updates in the same tick to avoid too many rerenders
        .bufferTime(0)
        .filter(updates => updates.length > 0)
        .map(updates => (state: State): State => updates.reduce((state, update) => update(state), state))
        .scan<Update, State>((state: State, update: Update) => update(state), {
            email: '',
            loading: false,
            emailVerified: true,
            hasSubmitted: false,
        })
        .map(({ email, loading, error, orgName, emailVerified, hasSubmitted }) => (
            <form className='accept-invite-page' onSubmit={nextSubmitEvent}>
                {!loading && !error && hasSubmitted && orgName && emailVerified && <Redirect to={`/settings/teams/${orgName}`} />}
                <h1>You were invited to join {orgName} on Sourcegraph!</h1>

                {error && <p className='form-text text-error'>{error.message}</p>}

                {/* TODO(john): provide action to re-send verification email */}
                {
                    hasSubmitted && !emailVerified &&
                        <p className='form-text text-error'>Please verify your email address to accept this invitation; check your inbox for a verification link.</p>
                }

                <div className='form-group'>
                    <label>Your company email</label>
                    <input
                        type='email'
                        className='ui-text-box'
                        placeholder='you@yourcompany.com'
                        required={true}
                        autoCorrect='off'
                        spellCheck={false}
                        value={email}
                        disabled={true}
                    />
                </div>

                <div className='form-group accept-invite-page__actions'>
                    <button type='submit' className='btn btn-primary' disabled={loading}>Accept Invite</button>
                    {loading && <LoaderIcon className='icon-inline' />}
                </div>

            </form>
        ))
})
