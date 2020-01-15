import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import AddIcon from 'mdi-react/AddIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import EmailOpenOutlineIcon from 'mdi-react/EmailOpenOutlineIcon'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { merge, Observable, of, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, filter, map, mergeMap, startWith, tap, withLatestFrom } from 'rxjs/operators'
import { gql } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { createAggregateError } from '../../../../shared/src/util/errors'
import { mutateGraphQL } from '../../backend/graphql'
import { CopyableText } from '../../components/CopyableText'
import { DismissibleAlert } from '../../components/DismissibleAlert'
import { Form } from '../../components/Form'
import { eventLogger } from '../../tracking/eventLogger'
import { ErrorAlert } from '../../components/alerts'

function inviteUserToOrganization(
    username: string,
    organization: GQL.ID
): Observable<GQL.IInviteUserToOrganizationResult> {
    return mutateGraphQL(
        gql`
            mutation InviteUserToOrganization($organization: ID!, $username: String!) {
                inviteUserToOrganization(organization: $organization, username: $username) {
                    sentInvitationEmail
                    invitationURL
                }
            }
        `,
        {
            username,
            organization,
        }
    ).pipe(
        map(({ data, errors }) => {
            const eventData = {
                organization: {
                    invite: {
                        username,
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

function addUserToOrganization(username: string, organization: GQL.ID): Observable<void> {
    return mutateGraphQL(
        gql`
            mutation AddUserToOrganization($organization: ID!, $username: String!) {
                addUserToOrganization(organization: $organization, username: $username) {
                    alwaysNil
                }
            }
        `,
        {
            username,
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

const InvitedNotification: React.FunctionComponent<{
    className: string
    username: string
    sentInvitationEmail: boolean
    invitationURL: string
    onDismiss: () => void
}> = ({ className, username, sentInvitationEmail, invitationURL, onDismiss }) => (
    <div className={`${className} invited-notification`}>
        <div className="invited-notification__message">
            {sentInvitationEmail ? (
                <>
                    Invitation sent to {username}. You can also send {username} the invitation link directly:
                </>
            ) : (
                <>Generated invitation link. Copy and send it to {username}:</>
            )}
            <CopyableText text={invitationURL} size={40} className="mt-2" />
        </div>
        <button type="button" className="btn btn-icon" title="Dismiss" onClick={onDismiss}>
            <CloseIcon className="icon-inline" />
        </button>
    </div>
)

interface Props {
    orgID: string
    authenticatedUser: GQL.IUser | null

    /** Called when the organization members list changes. */
    onDidUpdateOrganizationMembers: () => void

    onOrganizationUpdate: () => void
}

interface SubmittedInvite extends Pick<GQL.IInviteUserToOrganizationResult, 'sentInvitationEmail' | 'invitationURL'> {
    username: string
}

interface State {
    username: string

    /** Loading state (undefined means not loading). */
    loading?: 'inviteUserToOrganization' | 'addUserToOrganization'

    invited?: SubmittedInvite[]
    error?: Error
}

export class InviteForm extends React.PureComponent<Props, State> {
    public state: State = { username: '' }

    private submits = new Subject<React.FormEvent<HTMLFormElement>>()
    private inviteClicks = new Subject<React.MouseEvent<HTMLButtonElement>>()
    private usernameChanges = new Subject<string>()
    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        const orgChanges = that.componentUpdates.pipe(distinctUntilChanged((a, b) => a.orgID !== b.orgID))

        type Update = (prevState: State) => State

        that.subscriptions.add(that.usernameChanges.subscribe(username => that.setState({ username })))

        // Invite clicks.
        that.subscriptions.add(
            merge(that.submits.pipe(filter(() => !that.viewerCanAddUserToOrganization)), that.inviteClicks)
                .pipe(
                    tap(e => e.preventDefault()),
                    withLatestFrom(orgChanges, that.usernameChanges),
                    tap(([, orgId, username]) =>
                        eventLogger.log('InviteOrgMemberClicked', {
                            organization: {
                                invite: {
                                    username,
                                },
                                org_id: orgId,
                            },
                        })
                    ),
                    mergeMap(([, { orgID }, username]) =>
                        inviteUserToOrganization(username, orgID).pipe(
                            tap(() => that.props.onOrganizationUpdate()),
                            tap(() => that.usernameChanges.next('')),
                            mergeMap(({ sentInvitationEmail, invitationURL }) =>
                                // Reset email, reenable submit button, flash "invited" text
                                of(
                                    (state: State): State => ({
                                        ...state,
                                        loading: undefined,
                                        error: undefined,
                                        username: '',
                                        invited: [
                                            ...(state.invited || []),
                                            { username, sentInvitationEmail, invitationURL },
                                        ],
                                    })
                                )
                            ),
                            // Disable button while loading
                            startWith<Update>(
                                (state: State): State => ({
                                    ...state,
                                    loading: 'inviteUserToOrganization',
                                })
                            ),
                            catchError(error => [(state: State): State => ({ ...state, loading: undefined, error })])
                        )
                    )
                )
                .subscribe(
                    stateUpdate => that.setState(stateUpdate),
                    err => console.error(err)
                )
        )

        // Adds.
        that.subscriptions.add(
            that.submits
                .pipe(filter(() => that.viewerCanAddUserToOrganization))
                .pipe(
                    tap(e => e.preventDefault()),
                    withLatestFrom(orgChanges, that.usernameChanges),
                    mergeMap(([, { orgID }, username]) =>
                        addUserToOrganization(username, orgID).pipe(
                            tap(() => that.props.onDidUpdateOrganizationMembers()),
                            tap(() => that.usernameChanges.next('')),
                            mergeMap(() =>
                                // Reset email, reenable submit button, flash "invited" text
                                of(
                                    (state: State): State => ({
                                        ...state,
                                        loading: undefined,
                                        error: undefined,
                                        username: '',
                                    })
                                )
                            ),
                            // Disable button while loading
                            startWith<Update>(
                                (state: State): State => ({
                                    ...state,
                                    loading: 'addUserToOrganization',
                                })
                            ),
                            catchError(error => [(state: State): State => ({ ...state, loading: undefined, error })])
                        )
                    )
                )
                .subscribe(
                    stateUpdate => that.setState(stateUpdate),
                    err => console.error(err)
                )
        )

        that.componentUpdates.next(that.props)
    }

    public componentDidUpdate(): void {
        that.componentUpdates.next(that.props)
    }

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
    }

    private get viewerCanAddUserToOrganization(): boolean {
        return !!that.props.authenticatedUser && that.props.authenticatedUser.siteAdmin
    }

    public render(): JSX.Element | null {
        const viewerCanAddUserToOrganization = that.viewerCanAddUserToOrganization

        return (
            <div className="invite-form">
                <div className="card invite-form__container">
                    <div className="card-body">
                        <h4 className="card-title">
                            {that.viewerCanAddUserToOrganization ? 'Add or invite member' : 'Invite member'}
                        </h4>
                        <Form className="form-inline align-items-start" onSubmit={that.onSubmit}>
                            <label className="sr-only" htmlFor="invite-form__username">
                                Username
                            </label>
                            <input
                                type="text"
                                className="form-control mb-2 mr-sm-2"
                                id="invite-form__username"
                                placeholder="Username"
                                onChange={that.onUsernameChange}
                                value={that.state.username}
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
                                    disabled={!!that.state.loading}
                                    className="btn btn-primary mb-2 mr-sm-2"
                                    data-tooltip="Add immediately without sending invitation (site admins only)"
                                >
                                    {that.state.loading === 'addUserToOrganization' ? (
                                        <LoadingSpinner className="icon-inline" />
                                    ) : (
                                        <AddIcon className="icon-inline" />
                                    )}{' '}
                                    Add member
                                </button>
                            )}
                            {(emailInvitesEnabled || !that.viewerCanAddUserToOrganization) && (
                                <div className="form-group flex-column mb-2 mr-sm-2">
                                    {/* eslint-disable-next-line react/button-has-type */}
                                    <button
                                        type={viewerCanAddUserToOrganization ? 'button' : 'submit'}
                                        disabled={!!that.state.loading}
                                        className={`btn ${
                                            viewerCanAddUserToOrganization ? 'btn-secondary' : 'btn-primary'
                                        }`}
                                        data-tooltip={
                                            emailInvitesEnabled
                                                ? 'Send invitation email with link to join this organization'
                                                : 'Generate invitation link to manually send to user'
                                        }
                                        onClick={viewerCanAddUserToOrganization ? that.onInviteClick : undefined}
                                    >
                                        {that.state.loading === 'inviteUserToOrganization' ? (
                                            <LoadingSpinner className="icon-inline" />
                                        ) : (
                                            <EmailOpenOutlineIcon className="icon-inline" />
                                        )}{' '}
                                        {emailInvitesEnabled
                                            ? that.viewerCanAddUserToOrganization
                                                ? 'Send invitation to join'
                                                : 'Send invitation'
                                            : 'Generate invitation link'}
                                    </button>
                                </div>
                            )}
                        </Form>
                    </div>
                </div>
                {that.props.authenticatedUser &&
                    that.props.authenticatedUser.siteAdmin &&
                    !window.context.emailEnabled && (
                        <DismissibleAlert className="alert-info" partialStorageKey="org-invite-email-config">
                            <p className=" mb-0">
                                Set <code>email.smtp</code> in{' '}
                                <Link to="/site-admin/configuration">site configuration</Link> to send email
                                notfications about invitations.
                            </p>
                        </DismissibleAlert>
                    )}
                {that.state.invited &&
                    that.state.invited.map(({ username, sentInvitationEmail, invitationURL }, i) => (
                        /* eslint-disable react/jsx-no-bind */
                        <InvitedNotification
                            key={i}
                            className="alert alert-success invite-form__alert"
                            username={username}
                            sentInvitationEmail={sentInvitationEmail}
                            invitationURL={invitationURL}
                            onDismiss={() => that.dismissNotification(i)}
                        />
                        /* eslint-enable react/jsx-no-bind */
                    ))}
                {that.state.error && <ErrorAlert className="invite-form__alert" error={that.state.error} />}
            </div>
        )
    }

    private onUsernameChange: React.ChangeEventHandler<HTMLInputElement> = e =>
        that.usernameChanges.next(e.currentTarget.value)
    private onSubmit: React.FormEventHandler<HTMLFormElement> = e => that.submits.next(e)
    private onInviteClick: React.MouseEventHandler<HTMLButtonElement> = e => that.inviteClicks.next(e)

    private dismissNotification = (i: number): void => {
        that.setState(prevState => ({ invited: (prevState.invited || []).filter((_, j) => i !== j) }))
    }
}
