import CloseIcon from '@sourcegraph/icons/lib/Close'
import * as React from 'react'
import { Redirect, RouteComponentProps } from 'react-router'
import { concat } from 'rxjs/operators/concat'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { filter } from 'rxjs/operators/filter'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { tap } from 'rxjs/operators/tap'
import { withLatestFrom } from 'rxjs/operators/withLatestFrom'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { currentUser } from '../../auth'
import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'
import { UserAvatar } from '../../user/UserAvatar'
import { removeUserFromOrg } from '../backend'
import { InviteForm } from './InviteForm'

interface Props extends RouteComponentProps<any> {
    org: GQL.IOrg
    user: GQL.IUser
}

interface State {
    /**
     * The org from props, possibly modified optimistically to reflect remote operations.
     */
    org: GQL.IOrg

    /** Whether the user just left the org */
    left: boolean
    error?: string
}

/**
 * The organizations members settings page
 */
export class OrgSettingsMembersPage extends React.PureComponent<Props, State> {
    private orgChanges = new Subject<GQL.IOrg>()
    private memberRemoves = new Subject<GQL.IOrgMember>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)

        this.state = {
            left: false,
            org: this.props.org,
        }
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            this.orgChanges
                .pipe(
                    distinctUntilChanged(),
                    tap(org => eventLogger.logViewEvent('OrgSettingsMembers', { organization: { org_name: org.name } }))
                )
                .subscribe(org => this.setState({ org }))
        )
        this.orgChanges.next(this.props.org)

        this.subscriptions.add(
            this.memberRemoves
                .pipe(
                    tap(member =>
                        eventLogger.log('RemoveOrgMemberClicked', {
                            organization: {
                                remove: {
                                    auth_id: member.user,
                                },
                                org_id: member.org.id,
                            },
                        })
                    ),
                    withLatestFrom(currentUser),
                    filter(([member, user]) => {
                        if (!user) {
                            return false
                        }
                        if (member.org.members.length === 1) {
                            return confirm(
                                [
                                    `You're the last member of ${member.org.displayName}. `,
                                    `Leaving will delete the ${member.org.displayName} organization. `,
                                    `Leave this organization?`,
                                ].join('')
                            )
                        }
                        if (user.authID === member.user.authID) {
                            return confirm(`Leave this organization?`)
                        }
                        return confirm(`Remove ${member.user.displayName} from this organization?`)
                    }),
                    mergeMap(([memberToRemove, user]) =>
                        removeUserFromOrg(memberToRemove.org.id, memberToRemove.user.id).pipe(
                            concat([
                                {
                                    left: memberToRemove.user.authID === user!.authID,
                                    org:
                                        this.state.org &&
                                        ({
                                            ...this.state.org,
                                            members: this.state.org.members.filter(
                                                member => member.user.authID !== memberToRemove.user.authID
                                            ),
                                        } as GQL.IOrg),
                                },
                            ])
                        )
                    )
                )
                .subscribe(
                    ({ left, org }) => this.setState({ left, org }),
                    err => this.setState({ error: err.message })
                )
        )
    }

    public componentWillReceiveProps(props: Props): void {
        if (props.org !== this.props.org) {
            this.orgChanges.next(props.org)
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        // If the current user just left the org, redirect to settings start page
        if (this.state.left) {
            return <Redirect to="/settings/profile" />
        }

        return (
            <div className="org-settings-members-page">
                <PageTitle title={`Members - ${this.props.org.name}`} />
                <div className="org-settings-members-page__header">
                    <h2>Members</h2>
                </div>
                <InviteForm orgID={this.props.org.id} />
                <table className="table table-hover org-settings-members-page__table">
                    <thead>
                        <tr>
                            <th className="org-settings-members-page__avatar-cell" />
                            <th>Name</th>
                            <th>Username</th>
                            <th>Email</th>
                            <th className="org-settings-members-page__actions-cell" />
                        </tr>
                    </thead>
                    <tbody>
                        {this.state.org.members.map(member => (
                            <tr key={member.id}>
                                <td className="org-settings-members-page__avatar-cell">
                                    <UserAvatar user={member.user} size={64} />
                                </td>
                                <td>{member.user.displayName}</td>
                                <td>{member.user.username}</td>
                                <td>{member.user.email}</td>
                                <td className="org-settings-members-page__actions-cell">
                                    {this.props.user && (
                                        <button
                                            className="btn btn-icon"
                                            title={this.props.user.authID === member.user.authID ? 'Leave' : 'Remove'}
                                            // tslint:disable-next-line:jsx-no-lambda
                                            onClick={() => this.memberRemoves.next({ ...member, org: this.state.org })}
                                        >
                                            <CloseIcon className="icon-inline" />
                                        </button>
                                    )}
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        )
    }
}
