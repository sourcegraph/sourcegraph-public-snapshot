import CloseIcon from '@sourcegraph/icons/lib/Close'
import LoaderIcon from '@sourcegraph/icons/lib/Loader'
import * as React from 'react'
import { of } from 'rxjs/observable/of'
import { catchError } from 'rxjs/operators/catchError'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { startWith } from 'rxjs/operators/startWith'
import { tap } from 'rxjs/operators/tap'
import { withLatestFrom } from 'rxjs/operators/withLatestFrom'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { eventLogger } from '../../tracking/eventLogger'
import { inviteUser } from '../backend'

const emailInvitesEnabled = window.context.emailEnabled

const InvitedNotification: React.SFC<{
    className: string
    email: string
    acceptInviteURL: string
    onDismiss: () => void
}> = ({ className, email, acceptInviteURL, onDismiss }) => (
    <div className={`${className} invited-notification`}>
        {emailInvitesEnabled ? (
            <span className="invited-notification__message">Invite sent to {email}</span>
        ) : (
            <span className="invited-notification__message">
                Generated invite link. You must copy and send it to {email}:{' '}
                <a href={acceptInviteURL} target="_blank">
                    Invite link
                </a>
            </span>
        )}
        <button className="btn btn-icon">
            <CloseIcon title="Dismiss" onClick={onDismiss} />
        </button>
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
    email: string
    loading: boolean
    invited?: SubmittedInvite[]
    error?: Error
}

export class InviteForm extends React.PureComponent<Props, State> {
    public state: State = { loading: false, email: '' }

    private submits = new Subject<React.FormEvent<HTMLFormElement>>()
    private emailChanges = new Subject<string>()
    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        const orgChanges = this.componentUpdates.pipe(distinctUntilChanged((a, b) => a.orgID !== b.orgID))

        type Update = (prevState: State) => State

        this.subscriptions.add(this.emailChanges.subscribe(email => this.setState({ email })))

        this.subscriptions.add(
            this.submits
                .pipe(
                    tap(e => e.preventDefault()),
                    withLatestFrom(orgChanges, this.emailChanges),
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
                    mergeMap(([, { orgID }, email]) =>
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
                .subscribe(stateUpdate => this.setState(stateUpdate), err => console.error(err))
        )

        this.componentUpdates.next(this.props)
    }

    public componentWillReceiveProps(nextProps: Props): void {
        this.componentUpdates.next(nextProps)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <form className="invite-form" onSubmit={this.onSubmit}>
                <div className="invite-form__container">
                    <input
                        type="email"
                        className="form-control invite-form__email"
                        placeholder="newmember@yourcompany.com"
                        onChange={this.onEmailChange}
                        value={this.state.email}
                        required={true}
                        spellCheck={false}
                        size={30}
                    />
                    <button
                        type="submit"
                        disabled={this.state.loading}
                        className="btn btn-primary invite-form__submit-button"
                    >
                        {emailInvitesEnabled ? 'Invite' : 'Make invite link'}
                    </button>
                </div>
                {this.state.invited &&
                    this.state.invited.length > 0 && (
                        <div className="invite-form__alerts">
                            {this.state.invited.map(({ email, acceptInviteURL }, i) => (
                                <InvitedNotification
                                    key={i}
                                    className="alert alert-success invite-form__invited"
                                    email={email}
                                    acceptInviteURL={acceptInviteURL}
                                    // tslint:disable-next-line:jsx-no-lambda
                                    onDismiss={() => this.dismissNotification(i)}
                                />
                            ))}
                        </div>
                    )}
                <div className="invite-form__status">
                    {this.state.loading && <LoaderIcon className="icon-inline" />}
                    {this.state.error && <div className="text-danger">{this.state.error.message}</div>}
                </div>
            </form>
        )
    }

    private onEmailChange: React.ChangeEventHandler<HTMLInputElement> = e =>
        this.emailChanges.next(e.currentTarget.value)
    private onSubmit: React.FormEventHandler<HTMLFormElement> = e => this.submits.next(e)

    private dismissNotification = (i: number): void => {
        this.setState(prevState => ({ invited: (prevState.invited || []).filter((_, j) => i !== j) }))
    }
}
