import AddIcon from '@sourcegraph/icons/lib/Add'
import CloseIcon from '@sourcegraph/icons/lib/Close'
import InvitationIcon from '@sourcegraph/icons/lib/Invitation'
import LoaderIcon from '@sourcegraph/icons/lib/Loader'
import { upperFirst } from 'lodash'
import * as React from 'react'
import { Observable } from 'rxjs/Observable'
import { merge } from 'rxjs/observable/merge'
import { of } from 'rxjs/observable/of'
import { catchError } from 'rxjs/operators/catchError'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { filter } from 'rxjs/operators/filter'
import { map } from 'rxjs/operators/map'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { startWith } from 'rxjs/operators/startWith'
import { tap } from 'rxjs/operators/tap'
import { withLatestFrom } from 'rxjs/operators/withLatestFrom'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { gql, mutateGraphQL } from '../../backend/graphql'
import * as GQL from '../../backend/graphqlschema'
import { Form } from '../../components/Form'
import { eventLogger } from '../../tracking/eventLogger'
import { createAggregateError } from '../../util/errors'

export function inviteUserToOrganization(
    usernameOrEmail: string,
    organization: GQL.ID
): Observable<GQL.IInviteUserResult> {
    return mutateGraphQL(
        gql`
            mutation InviteUserToOrganization($organization: ID!, $usernameOrEmail: String!) {
                inviteUserToOrganization(organization: $organization, usernameOrEmail: $usernameOrEmail) {
                    acceptInviteURL
                }
            }
        `,
        {
            usernameOrEmail,
            organization,
        }
    ).pipe(
        map(({ data, errors }) => {
            const eventData = {
                organization: {
                    invite: {
                        user_email: usernameOrEmail,
                    },
                    org_id: organization,
                },
            }
            if (!data || !data.inviteUserToOrganization || (errors && errors.length > 0)) {
                eventLogger.log('InviteOrgMemberFailed', eventData)
                throw createAggregateError(errors)
            }
            eventLogger.log('OrgMemberInvited', eventData)
            return data.inviteUserToOrganization
        })
    )
}

export function addUserToOrganization(usernameOrEmail: string, organization: GQL.ID): Observable<void> {
    return mutateGraphQL(
        gql`
            mutation AddUserToOrganization($organization: ID!, $usernameOrEmail: String!) {
                addUserToOrganization(organization: $organization, usernameOrEmail: $usernameOrEmail) {
                    alwaysNil
                }
            }
        `,
        {
            usernameOrEmail,
            organization,
        }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.addUserToOrganization || (errors && errors.length > 0)) {
                eventLogger.log('AddOrgMemberFailed')
                throw createAggregateError(errors)
            }
            eventLogger.log('OrgMemberAdded')
        })
    )
}

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
    authenticatedUser: GQL.IUser | null

    /** Called when the organization members list changes. */
    onDidUpdateOrganizationMembers: () => void
}

interface SubmittedInvite {
    email: string
    acceptInviteURL: string
}

interface State {
    email: string

    /** Loading state (undefined means not loading). */
    loading?: 'inviteUserToOrganization' | 'addUserToOrganization'

    invited?: SubmittedInvite[]
    error?: Error
}

export class InviteForm extends React.PureComponent<Props, State> {
    public state: State = { email: '' }

    private submits = new Subject<React.FormEvent<HTMLFormElement>>()
    private inviteClicks = new Subject<React.MouseEvent<HTMLButtonElement>>()
    private emailChanges = new Subject<string>()
    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        const orgChanges = this.componentUpdates.pipe(distinctUntilChanged((a, b) => a.orgID !== b.orgID))

        type Update = (prevState: State) => State

        this.subscriptions.add(this.emailChanges.subscribe(email => this.setState({ email })))

        // Invite clicks.
        this.subscriptions.add(
            merge(this.submits.pipe(filter(() => !this.viewerCanAddUserToOrganization)), this.inviteClicks)
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
                        inviteUserToOrganization(email, orgID).pipe(
                            tap(() => this.emailChanges.next('')),
                            mergeMap(({ acceptInviteURL }) =>
                                // Reset email, reenable submit button, flash "invited" text
                                of((state: State): State => ({
                                    ...state,
                                    loading: undefined,
                                    error: undefined,
                                    email: '',
                                    invited: [...(state.invited || []), { email, acceptInviteURL }],
                                }))
                            ),
                            // Disable button while loading
                            startWith<Update>((state: State): State => ({
                                ...state,
                                loading: 'inviteUserToOrganization',
                            })),
                            catchError(error => [(state: State): State => ({ ...state, loading: undefined, error })])
                        )
                    )
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), err => console.error(err))
        )

        // Adds.
        this.subscriptions.add(
            this.submits
                .pipe(filter(() => this.viewerCanAddUserToOrganization))
                .pipe(
                    tap(e => e.preventDefault()),
                    withLatestFrom(orgChanges, this.emailChanges),
                    mergeMap(([, { orgID }, email]) =>
                        addUserToOrganization(email, orgID).pipe(
                            tap(() => this.props.onDidUpdateOrganizationMembers()),
                            tap(() => this.emailChanges.next('')),
                            mergeMap(() =>
                                // Reset email, reenable submit button, flash "invited" text
                                of((state: State): State => ({
                                    ...state,
                                    loading: undefined,
                                    error: undefined,
                                    email: '',
                                }))
                            ),
                            // Disable button while loading
                            startWith<Update>((state: State): State => ({
                                ...state,
                                loading: 'addUserToOrganization',
                            })),
                            catchError(error => [(state: State): State => ({ ...state, loading: undefined, error })])
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

    private get viewerCanAddUserToOrganization(): boolean {
        return !!this.props.authenticatedUser && this.props.authenticatedUser.siteAdmin
    }

    public render(): JSX.Element | null {
        const viewerCanAddUserToOrganization = this.viewerCanAddUserToOrganization

        return (
            <div className="invite-form">
                <div className="card invite-form__container">
                    <div className="card-body">
                        <h4 className="card-title">Invite member</h4>
                        <Form className="form-inline" onSubmit={this.onSubmit}>
                            <label className="sr-only" htmlFor="invite-form__email">
                                Username or email address
                            </label>
                            <input
                                type="text"
                                className="form-control mb-2 mr-sm-2"
                                id="invite-form__email"
                                placeholder="Username or email address"
                                onChange={this.onEmailChange}
                                value={this.state.email}
                                autoComplete="off"
                                autoCapitalize="off"
                                autoCorrect="off"
                                required={true}
                                spellCheck={false}
                                size={30}
                            />
                            {viewerCanAddUserToOrganization && (
                                <button
                                    type="submit"
                                    disabled={!!this.state.loading}
                                    className="btn btn-primary mb-2 mr-sm-2"
                                    data-tooltip="Add existing user without sending invitation (site admins only)"
                                >
                                    {this.state.loading === 'addUserToOrganization' ? (
                                        <LoaderIcon className="icon-inline" />
                                    ) : (
                                        <AddIcon className="icon-inline" />
                                    )}{' '}
                                    Add
                                </button>
                            )}
                            <button
                                type={viewerCanAddUserToOrganization ? 'button' : 'submit'}
                                disabled={!!this.state.loading}
                                className={`btn mb-2  ${
                                    viewerCanAddUserToOrganization ? 'btn-secondary' : 'btn-primary'
                                }`}
                                data-tooltip={
                                    emailInvitesEnabled
                                        ? 'Send invitation email with link to join this organization'
                                        : 'Generate invitation link to manually send to user'
                                }
                                onClick={viewerCanAddUserToOrganization ? this.onInviteClick : undefined}
                            >
                                {this.state.loading === 'inviteUserToOrganization' ? (
                                    <LoaderIcon className="icon-inline" />
                                ) : (
                                    <InvitationIcon className="icon-inline" />
                                )}{' '}
                                {emailInvitesEnabled ? 'Invite' : 'Make invite link'}
                            </button>
                        </Form>
                    </div>
                </div>
                {this.state.invited &&
                    this.state.invited.map(({ email, acceptInviteURL }, i) => (
                        <InvitedNotification
                            key={i}
                            className="alert alert-success invite-form__alert"
                            email={email}
                            acceptInviteURL={acceptInviteURL}
                            // tslint:disable-next-line:jsx-no-lambda
                            onDismiss={() => this.dismissNotification(i)}
                        />
                    ))}
                {this.state.error && (
                    <div className="invite-form__alert alert alert-danger">
                        Error: {upperFirst(this.state.error.message)}
                    </div>
                )}
            </div>
        )
    }

    private onEmailChange: React.ChangeEventHandler<HTMLInputElement> = e =>
        this.emailChanges.next(e.currentTarget.value)
    private onSubmit: React.FormEventHandler<HTMLFormElement> = e => this.submits.next(e)
    private onInviteClick: React.MouseEventHandler<HTMLButtonElement> = e => this.inviteClicks.next(e)

    private dismissNotification = (i: number): void => {
        this.setState(prevState => ({ invited: (prevState.invited || []).filter((_, j) => i !== j) }))
    }
}
