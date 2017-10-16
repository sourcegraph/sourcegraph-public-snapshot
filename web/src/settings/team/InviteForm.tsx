import LoaderIcon from '@sourcegraph/icons/lib/Loader'
import * as React from 'react'
import reactive from 'rx-component'
import 'rxjs/add/observable/merge'
import 'rxjs/add/observable/of'
import 'rxjs/add/operator/catch'
import 'rxjs/add/operator/concat'
import 'rxjs/add/operator/concat'
import 'rxjs/add/operator/delay'
import 'rxjs/add/operator/do'
import 'rxjs/add/operator/map'
import 'rxjs/add/operator/mergeMap'
import 'rxjs/add/operator/scan'
import 'rxjs/add/operator/startWith'
import 'rxjs/add/operator/withLatestFrom'
import { Observable } from 'rxjs/Observable'
import { Subject } from 'rxjs/Subject'
import { events } from '../../tracking/events'
import { inviteUser } from '../backend'

export interface Props {
    orgID: number
}

interface State {
    orgID: number
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

    const orgID = props.map(({ orgID }) => orgID)

    return Observable.merge<Update>(

        orgID
            .map(orgID => (state: State): State => ({ ...state, orgID })),

        emailChanges
            .map(email => (state: State): State => ({ ...state, email })),

        submits
            .do(e => e.preventDefault())
            .withLatestFrom(orgID, emailChanges)
            .do(([, orgId, email]) => events.InviteOrgMemberClicked.log({
                organization: {
                    invite: {
                        user_email: email,
                    },
                    org_id: orgId,
                },
            }))
            .mergeMap(([, orgID, email]) =>
                inviteUser(email, orgID)
                    .mergeMap(() =>
                        // Reset email, reenable submit button, flash "invited" text
                        Observable.of((state: State): State => ({ ...state, loading: false, error: undefined, email: '', invited: true }))
                            // Hide "invited" text again after 1s
                            .concat(Observable.of<Update>(state => ({ ...state, invited: false })).delay(1000))
                    )
                    // Disable button while loading
                    .startWith<Update>((state: State): State => ({ ...state, loading: true }))
                    .catch(error => [(state: State): State => ({ ...state, loading: false, error })])
            )
    )
        .scan<Update, State>((state: State, update: Update) => update(state), { invited: false, loading: false, email: '' } as State)
        .map(({ loading, email, invited, error }) => (
            <form className='invite-form form-inline' onSubmit={nextSubmit}>
                <input
                    type='email'
                    className='ui-text-box invite-form__email'
                    placeholder='newmember@yourcompany.com'
                    onChange={nextEmailChange}
                    value={email}
                    required={true}
                    spellCheck={false}
                    size={30}
                />
                <button type='submit' disabled={loading} className='btn btn-primary invite-form__submit-button'>Invite</button>
                {loading && <LoaderIcon className='icon-inline' />}
                {error && <div className='text-error'><small>{error.message}</small></div>}
                <div className={'invite-form__invited-text' + (invited ? ' invite-form__invited-text--visible' : '')}><small>Invited!</small></div>
            </form>
        ))

})
