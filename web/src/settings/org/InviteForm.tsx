import CloseIcon from '@sourcegraph/icons/lib/Close'
import LoaderIcon from '@sourcegraph/icons/lib/Loader'
import * as React from 'react'
import reactive from 'rx-component'
import { merge } from 'rxjs/observable/merge'
import { of } from 'rxjs/observable/of'
import { catchError } from 'rxjs/operators/catchError'
import { map } from 'rxjs/operators/map'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { scan } from 'rxjs/operators/scan'
import { startWith } from 'rxjs/operators/startWith'
import { tap } from 'rxjs/operators/tap'
import { withLatestFrom } from 'rxjs/operators/withLatestFrom'
import { Subject } from 'rxjs/Subject'
import { eventLogger } from '../../tracking/eventLogger'
import { inviteUser } from '../backend'

const InvitedNotification: React.SFC<{
    className: string
    email: string
    acceptInviteURL: string
    children: React.ReactChild
}> = ({ className, email, acceptInviteURL, children }) =>
    emailInvitesEnabled ? (
        <div className={`${className} invited-notification`}>
            <span className="invited-notification__message">Invite sent to {email}</span>
            {children}
        </div>
    ) : (
        <div className={`${className} invited-notification`}>
            <span className="invited-notification__message">
                Generated invite link. You must copy and send it to {email}:{' '}
                <a href={acceptInviteURL} target="_blank" className="invited-notification__link">
                    Invite link
                </a>
            </span>
            {children}
        </div>
    )

export interface Props {
    orgID: string
}

interface SubmittedInvite {
    email: string
    acceptInviteURL: string
}

interface State {
    orgID: string
    email: string
    loading: boolean
    invited?: SubmittedInvite[]
    error?: Error
}

type Update = (state: State) => State

const emailInvitesEnabled = window.context.emailEnabled

export const InviteForm = reactive<Props>(props => {
    const submits = new Subject<React.FormEvent<HTMLFormElement>>()
    const nextSubmit = (e: React.FormEvent<HTMLFormElement>) => submits.next(e)

    const emailChanges = new Subject<string>()
    const nextEmailChange = (event: React.ChangeEvent<HTMLInputElement>) => emailChanges.next(event.currentTarget.value)

    const notificationDismissals = new Subject<number>()

    const orgID = props.pipe(map(({ orgID }) => orgID))

    return merge<Update>(
        orgID.pipe(map(orgID => (state: State): State => ({ ...state, orgID }))),

        emailChanges.pipe(map(email => (state: State): State => ({ ...state, email }))),

        notificationDismissals.pipe(
            map(i => (state: State): State => ({
                ...state,
                invited: (state.invited || []).filter((_, j) => i !== j),
            }))
        ),

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
                    mergeMap(({ acceptInviteURL }) =>
                        // Reset email, reenable submit button, flash "invited" text
                        of((state: State): State => ({
                            ...state,
                            loading: false,
                            error: undefined,
                            email: '',
                            invited: [...(state.invited || []), { email, acceptInviteURL }],
                        }))
                    ),
                    // Disable button while loading
                    startWith<Update>((state: State): State => ({ ...state, loading: true })),
                    catchError(error => [(state: State): State => ({ ...state, loading: false, error })])
                )
            )
        )
    ).pipe(
        scan<Update, State>((state: State, update: Update) => update(state), {
            loading: false,
            email: '',
        } as State),
        map(({ loading, email, invited, error }) => (
            <form className="invite-form" onSubmit={nextSubmit}>
                <div className="invite-form__container">
                    <input
                        type="email"
                        className="form-control invite-form__email"
                        placeholder="newmember@yourcompany.com"
                        onChange={nextEmailChange}
                        value={email}
                        required={true}
                        spellCheck={false}
                        size={30}
                    />
                    <button type="submit" disabled={loading} className="btn btn-primary invite-form__submit-button">
                        {emailInvitesEnabled ? 'Invite' : 'Make invite link'}
                    </button>
                </div>
                <div className="invite-form__status">
                    {loading && <LoaderIcon className="icon-inline" />}
                    {error && <div className="text-error">{error.message}</div>}
                </div>
                {invited &&
                    invited.length > 0 && (
                        <div className="invite-form__alerts">
                            {invited.map(({ email, acceptInviteURL }, i) => (
                                <InvitedNotification
                                    key={i}
                                    className="alert alert-success invite-form__invited"
                                    email={email}
                                    acceptInviteURL={acceptInviteURL}
                                >
                                    <button className="btn btn-icon">
                                        <CloseIcon
                                            title="Dismiss"
                                            // tslint:disable-next-line:jsx-no-lambda
                                            onClick={() => notificationDismissals.next(i)}
                                        />
                                    </button>
                                </InvitedNotification>
                            ))}
                        </div>
                    )}
            </form>
        ))
    )
})
