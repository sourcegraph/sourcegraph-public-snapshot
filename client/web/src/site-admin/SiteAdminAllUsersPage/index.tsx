import * as React from 'react'

import { mdiCog, mdiPlus } from '@mdi/js'
import * as H from 'history'
import { isEqual } from 'lodash'
import { RouteComponentProps } from 'react-router'
import { merge, of, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, switchMap } from 'rxjs/operators'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { asError, logger } from '@sourcegraph/common'
import { Button, Link, Alert, Icon, H2, Text, Tooltip } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { CopyableText } from '../../components/CopyableText'
import { FilteredConnection } from '../../components/FilteredConnection'
import { PageTitle } from '../../components/PageTitle'
import { UserNodeFields } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'
import { userURL } from '../../user'
import { setUserEmailVerified } from '../../user/settings/backend'
import {
    deleteUser,
    fetchAllUsers,
    randomizeUserPassword,
    setUserIsSiteAdmin,
    invalidateSessionsByID,
    setUserTag,
} from '../backend'

const CREATE_ORG_TAG = 'CreateOrg'

interface UserNodeProps {
    /**
     * The user to display in this list item.
     */
    node: UserNodeFields

    /**
     * The currently authenticated user.
     */
    authenticatedUser: AuthenticatedUser

    /**
     * Called when the user is updated by an action in this list item.
     */
    onDidUpdate?: () => void
    history: H.History
}

interface UserNodeState {
    loading: boolean
    errorDescription?: string
    resetPasswordURL?: string | null
}

const nukeDetails = `
- When deleting a user normally, the user and ALL associated data is marked as deleted in the DB and never served again. You could undo this by running DB commands manually.
- By deleting a user forever, the user and ALL associated data will be permanently removed from the DB (you CANNOT undo this). When deleting data at a user's request, "Delete forever" is used.

Beware this includes e.g. deleting extensions authored by the user, deleting ANY settings authored or updated by the user, etc.

For more information about what data is deleted, see https://github.com/sourcegraph/sourcegraph/blob/main/doc/admin/user_data_deletion.md

Are you ABSOLUTELY certain you wish to delete this user and all associated data?`

class UserNode extends React.PureComponent<UserNodeProps, UserNodeState> {
    public state: UserNodeState = {
        loading: false,
    }

    private emailVerificationClicks = new Subject<{ email: string; verified: boolean }>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.emailVerificationClicks
                .pipe(
                    distinctUntilChanged((a, b) => isEqual(a, b)),
                    switchMap(({ email, verified }) =>
                        merge(
                            of({
                                errorDescription: undefined,
                                resetPasswordURL: undefined,
                                loading: true,
                            }),
                            setUserEmailVerified(this.props.node.id, email, verified).pipe(
                                map(() => ({ loading: false })),
                                catchError(error => [{ loading: false, errorDescription: asError(error).message }])
                            )
                        )
                    )
                )
                .subscribe(
                    stateUpdate => {
                        this.setState(stateUpdate)
                        if (this.props.onDidUpdate) {
                            this.props.onDidUpdate()
                        }
                    },
                    error => logger.error(error)
                )
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const orgCreationLabel =
            window.context.sourcegraphDotComMode && this.props.node.tags?.includes(CREATE_ORG_TAG)
                ? 'Disable'
                : 'Enable'

        return (
            <li className="list-group-item py-2">
                <div className="d-flex align-items-center justify-content-between">
                    <div>
                        {window.context.sourcegraphDotComMode ? (
                            <strong>{this.props.node.username}</strong>
                        ) : (
                            <Link to={`/users/${this.props.node.username}`}>
                                <strong>{this.props.node.username}</strong>
                            </Link>
                        )}
                        <br />
                        <span className="text-muted">{this.props.node.displayName}</span>
                    </div>
                    <div>
                        {window.context.sourcegraphDotComMode && (
                            <>
                                <Tooltip content={`${orgCreationLabel} user tag to allow user to create organizations`}>
                                    <Button
                                        onClick={() => this.toggleOrgCreationTag(orgCreationLabel === 'Enable')}
                                        disabled={this.state.loading}
                                        variant="secondary"
                                        size="sm"
                                    >
                                        {orgCreationLabel} org creation
                                    </Button>
                                </Tooltip>{' '}
                            </>
                        )}
                        {!window.context.sourcegraphDotComMode && (
                                <Button
                                    to={`${userURL(this.props.node.username)}/settings`}
                                    variant="secondary"
                                    size="sm"
                                    as={Link}
                                >
                                    <Icon aria-hidden={true} svgPath={mdiCog} /> Settings
                                </Button>
                            ) &&
                            ' '}
                        {this.props.node.id !== this.props.authenticatedUser.id && (
                            <Tooltip content="Force the user to re-authenticate on their next request">
                                <Button
                                    onClick={this.invalidateSessions}
                                    disabled={this.state.loading}
                                    variant="secondary"
                                    size="sm"
                                >
                                    Force sign-out
                                </Button>
                            </Tooltip>
                        )}{' '}
                        {window.context.resetPasswordEnabled && (
                            <Button
                                onClick={this.randomizePassword}
                                disabled={this.state.loading || !!this.state.resetPasswordURL}
                                variant="secondary"
                                size="sm"
                            >
                                Reset password
                            </Button>
                        )}{' '}
                        {this.props.node.id !== this.props.authenticatedUser.id &&
                            (this.props.node.siteAdmin ? (
                                <Button
                                    onClick={this.demoteFromSiteAdmin}
                                    disabled={this.state.loading}
                                    variant="secondary"
                                    size="sm"
                                >
                                    Revoke site admin
                                </Button>
                            ) : (
                                <Button
                                    key="promote"
                                    onClick={this.promoteToSiteAdmin}
                                    disabled={this.state.loading}
                                    variant="secondary"
                                    size="sm"
                                >
                                    Promote to site admin
                                </Button>
                            ))}{' '}
                        {this.props.node.id !== this.props.authenticatedUser.id && (
                            <Button onClick={this.deleteUser} disabled={this.state.loading} variant="danger" size="sm">
                                Delete
                            </Button>
                        )}
                        {this.props.node.id !== this.props.authenticatedUser.id && (
                            <Button
                                className="ml-1"
                                onClick={this.nukeUser}
                                disabled={this.state.loading}
                                variant="danger"
                                size="sm"
                            >
                                Delete forever
                            </Button>
                        )}
                    </div>
                </div>
                {this.state.errorDescription && <ErrorAlert className="mt-2" error={this.state.errorDescription} />}
                {this.state.resetPasswordURL && (
                    <Alert className="mt-2" variant="success">
                        <Text>
                            Password was reset. You must manually send <strong>{this.props.node.username}</strong> this
                            reset link:
                        </Text>
                        <CopyableText text={this.state.resetPasswordURL} size={40} />
                    </Alert>
                )}
                {this.state.resetPasswordURL === null && (
                    <Alert className="mt-2" variant="success">
                        Password was reset. The reset link was sent to the primary email of the user:{' '}
                        <strong>{this.props.node.emails.find(item => item.isPrimary)?.email}</strong>
                    </Alert>
                )}
            </li>
        )
    }

    private promoteToSiteAdmin = (): void => this.setSiteAdmin(true)
    private demoteFromSiteAdmin = (): void => this.setSiteAdmin(false)

    private setSiteAdmin(siteAdmin: boolean): void {
        if (
            !window.confirm(
                siteAdmin
                    ? `Promote user ${this.props.node.username} to site admin?`
                    : `Revoke site admin status from user ${this.props.node.username}?`
            )
        ) {
            return
        }

        this.setState({
            errorDescription: undefined,
            loading: true,
        })

        setUserIsSiteAdmin(this.props.node.id, siteAdmin)
            .toPromise()
            .then(
                () => {
                    this.setState({ loading: false })
                    if (this.props.onDidUpdate) {
                        this.props.onDidUpdate()
                    }
                },
                error => this.setState({ loading: false, errorDescription: asError(error).message })
            )
    }

    private randomizePassword = (): void => {
        if (
            !window.confirm(
                `Reset the password for ${this.props.node.username} to a random password? The user must reset their password to sign in again.`
            )
        ) {
            return
        }

        this.setState({
            errorDescription: undefined,
            resetPasswordURL: undefined,
            loading: true,
        })

        randomizeUserPassword(this.props.node.id)
            .toPromise()
            .then(
                ({ resetPasswordURL }) => {
                    this.setState({
                        loading: false,
                        resetPasswordURL,
                    })
                },
                error => this.setState({ loading: false, errorDescription: asError(error).message })
            )
    }

    private invalidateSessions = (): void => {
        if (
            !window.confirm(
                `Revoke all active sessions for ${this.props.node.username}? The user will need to re-authenticate on their next request or visit to Sourcegraph.`
            )
        ) {
            return
        }

        this.setState({ loading: true })
        invalidateSessionsByID(this.props.node.id)
            .toPromise()
            .then(
                () => {
                    this.setState({
                        loading: false,
                    })
                },
                error => this.setState({ loading: false, errorDescription: asError(error).message })
            )
    }

    private deleteUser = (): void => this.doDeleteUser(false)
    private nukeUser = (): void => this.doDeleteUser(true)

    private doDeleteUser = (hard: boolean): void => {
        let message = `Delete the user ${this.props.node.username}?`
        if (hard) {
            message = `Delete the user ${this.props.node.username} forever?${nukeDetails}`
        }
        if (!window.confirm(message)) {
            return
        }

        this.setState({
            errorDescription: undefined,
            resetPasswordURL: undefined,
            loading: true,
        })

        deleteUser(this.props.node.id, hard)
            .toPromise()
            .then(
                () => {
                    this.setState({ loading: false })
                    if (this.props.onDidUpdate) {
                        this.props.onDidUpdate()
                    }
                },
                error => this.setState({ loading: false, errorDescription: asError(error).message })
            )
    }

    private toggleOrgCreationTag = (newValue: boolean): void => {
        this.setState({
            errorDescription: undefined,
            resetPasswordURL: undefined,
            loading: true,
        })

        setUserTag(this.props.node.id, CREATE_ORG_TAG, newValue)
            .toPromise()
            .then(() => {
                this.setState({ loading: false })
                if (this.props.onDidUpdate) {
                    this.props.onDidUpdate()
                }
            })
            .catch(error => {
                this.setState({ loading: false, errorDescription: asError(error).message })
            })
    }
}

export interface Props extends RouteComponentProps<{}> {
    authenticatedUser: AuthenticatedUser
    history: H.History
}

interface State {
    users?: UserNodeFields[]
    totalCount?: number
}

/**
 * A page displaying the users on this site.
 */
export class SiteAdminAllUsersPage extends React.Component<Props, State> {
    public state: State = {}

    private userUpdates = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminAllUsers')
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const nodeProps: Omit<UserNodeProps, 'node'> = {
            authenticatedUser: this.props.authenticatedUser,
            onDidUpdate: this.onDidUpdateUser,
            history: this.props.history,
        }

        return (
            <div className="site-admin-all-users-page">
                <PageTitle title="Users - Admin" />
                <div className="d-flex justify-content-between align-items-center mb-3">
                    <H2 className="mb-0">Users</H2>
                    <div>
                        <Button to="/site-admin/users/new" variant="primary" as={Link}>
                            <Icon aria-hidden={true} svgPath={mdiPlus} /> Create user account
                        </Button>
                    </div>
                </div>
                <FilteredConnection<UserNodeFields, Omit<UserNodeProps, 'node'>>
                    className="list-group list-group-flush mt-3"
                    noun="user"
                    pluralNoun="users"
                    queryConnection={fetchAllUsers}
                    nodeComponent={UserNode}
                    nodeComponentProps={nodeProps}
                    updates={this.userUpdates}
                    history={this.props.history}
                    location={this.props.location}
                />
            </div>
        )
    }

    private onDidUpdateUser = (): void => this.userUpdates.next()
}
