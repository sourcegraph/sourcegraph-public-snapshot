import CloseIcon from '@sourcegraph/icons/lib/Close'
import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import * as React from 'react'
import { match, Redirect } from 'react-router'
import reactive from 'rx-component'
import { BehaviorSubject } from 'rxjs/BehaviorSubject'
import { combineLatest } from 'rxjs/observable/combineLatest'
import { merge } from 'rxjs/observable/merge'
import { concat } from 'rxjs/operators/concat'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { filter } from 'rxjs/operators/filter'
import { map } from 'rxjs/operators/map'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { scan } from 'rxjs/operators/scan'
import { tap } from 'rxjs/operators/tap'
import { withLatestFrom } from 'rxjs/operators/withLatestFrom'
import { Subject } from 'rxjs/Subject'
import { currentUser } from '../../auth'
import { HeroPage } from '../../components/HeroPage'
import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'
import { fetchOrg, removeUserFromOrg } from '../backend'
import { UserAvatar } from '../user/UserAvatar'
import { InviteForm } from './InviteForm'
import { OrgSettingsFile } from './OrgSettingsFile'
import { OrgSettingsForm } from './OrgSettingsForm'

const OrgNotFound = () => (
    <HeroPage
        icon={DirectionalSignIcon}
        title="404: Not Found"
        subtitle="Sorry, the requested organization was not found."
    />
)

export interface Props {
    match: match<{ orgName: string }>
}

interface State {
    org?: GQL.IOrg
    user?: GQL.IUser
    /** Whether the user just left the org */
    left: boolean
}

type Update = (s: State) => State

/**
 * The organizations settings page
 */
export const Org = reactive<Props>(props => {
    const memberRemoves = new Subject<GQL.IOrgMember>()

    const orgChanges = props.pipe(
        map(props => props.match.params.orgName),
        distinctUntilChanged(),
        tap(orgName => eventLogger.logViewEvent('OrgProfile', { organization: { org_name: orgName } }))
    )

    const settingsCommits = new BehaviorSubject<void>(void 0)
    const nextSettingsCommit = () => settingsCommits.next(void 0)

    return merge<Update>(
        combineLatest(currentUser, orgChanges, settingsCommits).pipe(
            mergeMap(([user, orgName]) => {
                if (!user) {
                    return [(state: State): State => ({ ...state, user: undefined })]
                }
                // Find org ID from user auth state
                const org = user.orgs.find(org => org.name === orgName)
                if (!org) {
                    return [(state: State): State => ({ ...state, user, org })]
                }
                // Fetch the org by ID by ID
                return fetchOrg(org.id).pipe(
                    map(org => (state: State): State => ({ ...state, user, org: org || undefined }))
                )
            })
        ),

        memberRemoves.pipe(
            tap(member =>
                eventLogger.log('RemoveOrgMemberClicked', {
                    organization: {
                        remove: {
                            auth0_id: member.user,
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
                            `You're the last member of ${member.org.displayName}.`,
                            `Leaving will delete the ${member.org.displayName} organization.`,
                            `Leave this organization?`,
                        ].join('')
                    )
                }
                if (user.auth0ID === member.user.auth0ID) {
                    return confirm(`Leave this organization?`)
                }
                return confirm(`Remove ${member.user.displayName} from this organization?`)
            }),
            mergeMap(([memberToRemove, user]) =>
                removeUserFromOrg(memberToRemove.org.id, memberToRemove.user.auth0ID).pipe(
                    concat([
                        (state: State): State => ({
                            ...state,
                            left: memberToRemove.user.auth0ID === user!.auth0ID,
                            org: state.org && {
                                ...state.org,
                                members: state.org.members.filter(
                                    member => member.user.auth0ID !== memberToRemove.user.auth0ID
                                ),
                            },
                        }),
                    ])
                )
            )
        )
    ).pipe(
        scan<Update, State>((state: State, update: Update) => update(state), { left: false }),
        map(({ user, org, left }: State): JSX.Element | null => {
            // If the current user just left the org, redirect to settings start page
            if (left) {
                return <Redirect to="/settings" />
            }
            if (!user) {
                return <Redirect to="/sign-in" />
            }
            if (!org) {
                return <OrgNotFound />
            }
            return (
                <div className="org">
                    <PageTitle title={org.name} />
                    <div className="org__header">
                        <h1>{org.name}</h1>

                        <InviteForm orgID={org.id} />
                    </div>
                    <h3>Members</h3>
                    <table className="table table-hover org__table">
                        <thead>
                            <tr>
                                <th className="org__avatar-cell" />
                                <th>Name</th>
                                <th>Username</th>
                                <th>Email</th>
                                <th className="org__actions-cell" />
                            </tr>
                        </thead>
                        <tbody>
                            {org.members.map(member => (
                                <tr key={member.id}>
                                    <td className="org__avatar-cell">
                                        <UserAvatar user={member.user} size={64} />
                                    </td>
                                    <td>{member.user.displayName}</td>
                                    <td>{member.user.username}</td>
                                    <td>{member.user.email}</td>
                                    <td className="org__actions-cell">
                                        <button
                                            className="btn btn-icon"
                                            title={user.auth0ID === member.user.auth0ID ? 'Leave' : 'Remove'}
                                            // tslint:disable-next-line:jsx-no-lambda
                                            onClick={() => memberRemoves.next({ ...member, org })}
                                        >
                                            <CloseIcon className="icon-inline" />
                                        </button>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                    <OrgSettingsForm org={org} />
                    <OrgSettingsFile orgID={org.id} settings={org.latestSettings} onDidCommit={nextSettingsCommit} />
                </div>
            )
        })
    )
})
