import LoaderIcon from '@sourcegraph/icons/lib/Loader'
import { Base64 } from 'js-base64'
import * as React from 'react'
import { Redirect, RouteComponentProps } from 'react-router'
import reactive from 'rx-component'
import { merge, Observable, of, Subject } from 'rxjs'
import {
    bufferTime,
    catchError,
    concat,
    distinctUntilChanged,
    filter,
    map,
    mergeMap,
    scan,
    tap,
    withLatestFrom,
} from 'rxjs/operators'
import { refreshCurrentUser } from '../../auth'
import { Form } from '../../components/Form'
import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'
import { acceptUserInvite } from '../backend'

export interface Props extends RouteComponentProps<any> {}

interface State {
    email: string
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
                                    eventLogger.log('InviteAccepted', eventProps)
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
                    hasSubmitted: false,
                }),
                map(({ email, loading, error, orgName, hasSubmitted }) => (
                    <Form className="accept-invite-page" onSubmit={nextSubmitEvent}>
                        {!loading &&
                            !error &&
                            hasSubmitted &&
                            orgName && <Redirect to={`/organizations/${orgName}/settings`} />}
                        <PageTitle title="Accept invitation" />
                        <h2>
                            You were invited to join the <strong>{orgName}</strong> organization
                        </h2>

                        {error && <p className="form-text text-danger">{error.message}</p>}

                        <div className="form-group accept-invite-page__actions">
                            <button type="submit" className="btn btn-primary" disabled={loading}>
                                Accept invitation
                            </button>
                            {loading && <LoaderIcon className="icon-inline" />}
                        </div>
                    </Form>
                ))
            )
    )
})
