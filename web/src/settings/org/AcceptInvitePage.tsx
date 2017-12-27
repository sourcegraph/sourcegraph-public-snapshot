import LoaderIcon from '@sourcegraph/icons/lib/Loader'
import * as H from 'history'
import { Base64 } from 'js-base64'
import * as React from 'react'
import { Redirect } from 'react-router'
import reactive from 'rx-component'
import { Observable } from 'rxjs/Observable'
import { merge } from 'rxjs/observable/merge'
import { of } from 'rxjs/observable/of'
import { bufferTime } from 'rxjs/operators/bufferTime'
import { catchError } from 'rxjs/operators/catchError'
import { concat } from 'rxjs/operators/concat'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { filter } from 'rxjs/operators/filter'
import { map } from 'rxjs/operators/map'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { scan } from 'rxjs/operators/scan'
import { tap } from 'rxjs/operators/tap'
import { withLatestFrom } from 'rxjs/operators/withLatestFrom'
import { Subject } from 'rxjs/Subject'
import { refreshCurrentUser } from '../../auth'
import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'
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

    eventLogger.logViewEvent('AcceptInvite')

    /** The token in the query params */
    const inviteToken: Observable<string> = props.pipe(
        map(({ location }) => location),
        distinctUntilChanged(),
        map(location => {
            const token = new URLSearchParams(location.search).get('token')
            if (!token) {
                throw new Error('No invite token in URL')
            }
            return token
        })
    )

    const tokenPayload: Observable<TokenPayload> = inviteToken.pipe(
        map(token => JSON.parse(Base64.decode(token.split('.')[1])))
    )

    return (
        merge<Update>(
            // Any update to these should cause a rerender
            tokenPayload.pipe(
                map(payload => (state: State): State => ({
                    ...state,
                    orgName: payload.orgName,
                    email: payload.email,
                }))
            ),

            // Form submits
            submitEvents.pipe(
                tap(event => event.preventDefault()),
                // Don't submit if form is invalid
                // Feedback is done through CSS
                filter(event => event.currentTarget.checkValidity()),
                // Get latest state values
                withLatestFrom(inviteToken, tokenPayload),
                mergeMap(([, inviteToken, tokenPayload]) =>
                    // Show loader
                    of<Update>(state => ({ ...state, loading: true, email: tokenPayload.email })).pipe(
                        concat(
                            acceptUserInvite({ inviteToken }).pipe(
                                tap(status => {
                                    const eventProps = {
                                        org_id: tokenPayload.orgID,
                                        user_email: tokenPayload.email,
                                        org_name: tokenPayload.orgName,
                                    }
                                    if (status.emailVerified) {
                                        eventLogger.log('InviteAccepted', eventProps)
                                    } else {
                                        eventLogger.log('AcceptInviteFailed', eventProps)
                                    }
                                }),
                                mergeMap(status =>
                                    // Reload user
                                    refreshCurrentUser()
                                        // Redirect
                                        .pipe(
                                            concat([
                                                (state: State): State => ({
                                                    ...state,
                                                    loading: false,
                                                    hasSubmitted: true,
                                                    emailVerified: status.emailVerified,
                                                }),
                                            ])
                                        )
                                ),
                                // Show error
                                catchError(error => {
                                    console.error(error)
                                    return [
                                        (state: State): State => ({
                                            ...state,
                                            hasSubmitted: true,
                                            loading: false,
                                            error,
                                        }),
                                    ]
                                })
                            )
                        )
                    )
                )
            )
        )
            // Buffer state updates in the same tick to avoid too many rerenders
            .pipe(
                bufferTime(0),
                filter(updates => updates.length > 0),
                map(updates => (state: State): State => updates.reduce((state, update) => update(state), state)),
                scan<Update, State>((state: State, update: Update) => update(state), {
                    email: '',
                    loading: false,
                    emailVerified: true,
                    hasSubmitted: false,
                }),
                map(({ email, loading, error, orgName, emailVerified, hasSubmitted }) => (
                    <form className="accept-invite-page" onSubmit={nextSubmitEvent}>
                        {!loading &&
                            !error &&
                            hasSubmitted &&
                            orgName &&
                            emailVerified && <Redirect to={`/settings/orgs/${orgName}`} />}
                        <PageTitle title="Accept invite" />
                        <h1>You were invited to join {orgName} on Sourcegraph!</h1>

                        {error && <p className="form-text text-error">{error.message}</p>}

                        {/* TODO(john): provide action to re-send verification email */}
                        {hasSubmitted &&
                            !emailVerified && (
                                <p className="form-text text-error">
                                    Please verify your email address to accept this invitation; check your inbox for a
                                    verification link.
                                </p>
                            )}

                        <div className="form-group">
                            <label>Your company email</label>
                            <input
                                type="email"
                                className="form-control"
                                placeholder="you@yourcompany.com"
                                required={true}
                                autoCorrect="off"
                                spellCheck={false}
                                value={email}
                                disabled={true}
                            />
                        </div>

                        <div className="form-group accept-invite-page__actions">
                            <button type="submit" className="btn btn-primary" disabled={loading}>
                                Accept Invite
                            </button>
                            {loading && <LoaderIcon className="icon-inline" />}
                        </div>
                    </form>
                ))
            )
    )
})
