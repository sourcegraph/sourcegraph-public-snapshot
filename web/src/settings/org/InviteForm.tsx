import LoaderIcon from '@sourcegraph/icons/lib/Loader'
import * as React from 'react'
import reactive from 'rx-component'
import { merge } from 'rxjs/observable/merge'
import { of } from 'rxjs/observable/of'
import { catchError } from 'rxjs/operators/catchError'
import { concat } from 'rxjs/operators/concat'
import { delay } from 'rxjs/operators/delay'
import { map } from 'rxjs/operators/map'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { scan } from 'rxjs/operators/scan'
import { startWith } from 'rxjs/operators/startWith'
import { tap } from 'rxjs/operators/tap'
import { withLatestFrom } from 'rxjs/operators/withLatestFrom'
import { Subject } from 'rxjs/Subject'
import { eventLogger } from '../../tracking/eventLogger'
import { inviteUser } from '../backend'

export interface Props {
    orgID: string
}

interface State {
    orgID: string
    email: string
    loading: boolean
    invited: boolean
    error?: Error
}

type Update = (state: State) => State

export const InviteForm = reactive<Props>(props => {
    const submits = new Subject<React.FormEvent<HTMLFormElement>>()
    const nextSubmit = (e: React.FormEvent<HTMLFormElement>) => submits.next(e)

    const emailChanges = new Subject<string>()
    const nextEmailChange = (event: React.ChangeEvent<HTMLInputElement>) => emailChanges.next(event.currentTarget.value)

    const orgID = props.pipe(map(({ orgID }) => orgID))

    return merge<Update>(
        orgID.pipe(map(orgID => (state: State): State => ({ ...state, orgID }))),

        emailChanges.pipe(map(email => (state: State): State => ({ ...state, email }))),

        submits.pipe(
            tap(e => e.preventDefault()),
            withLatestFrom(orgID, emailChanges),
            tap(([, orgId, email]) =>
                eventLogger.log('InviteOrgMemberClicked', {
                    organization: {
                        invite: {
                            user_email: email,
                        },
                        org_id: orgId,
                    },
                })
            ),
            mergeMap(([, orgID, email]) =>
                inviteUser(email, orgID).pipe(
                    mergeMap(() =>
                        // Reset email, reenable submit button, flash "invited" text
                        of((state: State): State => ({
                            ...state,
                            loading: false,
                            error: undefined,
                            email: '',
                            invited: true,
                        }))
                            // Hide "invited" text again after 1s
                            .pipe(concat(of<Update>(state => ({ ...state, invited: false })), delay(1000)))
                    ),
                    // Disable button while loading
                    startWith<Update>((state: State): State => ({ ...state, loading: true })),
                    catchError(error => [(state: State): State => ({ ...state, loading: false, error })])
                )
            )
        )
    ).pipe(
        scan<Update, State>((state: State, update: Update) => update(state), {
            invited: false,
            loading: false,
            email: '',
        } as State),
        map(({ loading, email, invited, error }) => (
            <form className="invite-form" onSubmit={nextSubmit}>
                <div className="invite-form__container">
                    <input
                        type="email"
                        className="ui-text-box invite-form__email"
                        placeholder="newmember@yourcompany.com"
                        onChange={nextEmailChange}
                        value={email}
                        required={true}
                        spellCheck={false}
                        size={30}
                    />
                    <button type="submit" disabled={loading} className="btn btn-primary invite-form__submit-button">
                        Invite
                    </button>
                </div>
                {loading && <LoaderIcon className="icon-inline" />}
                {error && (
                    <div className="text-error">
                        <small>{error.message}</small>
                    </div>
                )}
                <div className={'invite-form__invited-text' + (invited ? ' invite-form__invited-text--visible' : '')}>
                    <small>Invited!</small>
                </div>
            </form>
        ))
    )
})
