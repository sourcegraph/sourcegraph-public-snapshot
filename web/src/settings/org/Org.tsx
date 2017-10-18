import CloseIcon from '@sourcegraph/icons/lib/Close'
import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import * as React from 'react'
import { match, Redirect } from 'react-router'
import reactive from 'rx-component'
import 'rxjs/add/observable/combineLatest'
import 'rxjs/add/observable/merge'
import 'rxjs/add/operator/concat'
import 'rxjs/add/operator/distinctUntilChanged'
import 'rxjs/add/operator/do'
import 'rxjs/add/operator/filter'
import 'rxjs/add/operator/map'
import 'rxjs/add/operator/mergeMap'
import 'rxjs/add/operator/scan'
import 'rxjs/add/operator/withLatestFrom'
import { Observable } from 'rxjs/Observable'
import { Subject } from 'rxjs/Subject'
import { currentUser } from '../../auth'
import { HeroPage } from '../../components/HeroPage'
import { events } from '../../tracking/events'
import { fetchOrg, removeUserFromOrg } from '../backend'
import { UserAvatar } from '../user/UserAvatar'
import { InviteForm } from './InviteForm'
import { OrgSettingsForm } from './OrgSettingsForm'

const OrgNotFound = () => <HeroPage icon={DirectionalSignIcon} title='404: Not Found' subtitle='Sorry, the requested organization was not found.' />

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

    return Observable.merge<Update>(
        Observable.combineLatest(
            currentUser,
            props
                .map(props => props.match.params.orgName)
                .distinctUntilChanged()
        )
            .mergeMap(([user, orgName]) => {
                if (!user) {
                    return [(state: State): State => ({ ...state, user: undefined })]
                }
                // Find org ID from user auth state
                const org = user.orgs.find(org => org.name === orgName)
                if (!org) {
                    return [(state: State): State => ({ ...state, user, org })]
                }
                // Fetch the org by ID by ID
                return fetchOrg(org.id)
                    .map(org => (state: State): State => ({ ...state, user, org: org || undefined }))
                }
            ),

        memberRemoves
            .do(member => events.RemoveOrgMemberClicked.log({
                organization: {
                    remove: {
                        auth0_id: member.userID,
                    },
                    org_id: member.org.id,
                },
            }))
            .withLatestFrom(currentUser)
            .filter(([member, user]) => !!user && confirm(
                user.id === member.userID
                    ? `Leave this organization?`
                    : `Remove ${member.user.displayName} from this organization?`
            ))
            .mergeMap(([memberToRemove, user]) =>
                removeUserFromOrg(memberToRemove.org.id, memberToRemove.userID)
                    .concat([(state: State): State => ({
                        ...state,
                        left: memberToRemove.userID === user!.id,
                        org: state.org && {
                            ...state.org,
                            members: state.org.members.filter(member => member.userID !== memberToRemove.userID),
                        },
                    })])
            )
    )
        .scan<Update, State>((state: State, update: Update) => update(state), { left: false })
        .map(({ user, org, left }: State): JSX.Element | null => {
            // If the current user just left the org, redirect to settings start page
            if (left) {
                return <Redirect to='/settings' />
            }
            if (!user) {
                return <Redirect to='/sign-in' />
            }
            if (!org) {
                return <OrgNotFound />
            }
            return (
                <div className='org'>
                    <div className='org__header'>
                        <h1>{org.name}</h1>

                        <InviteForm orgID={org.id}/>
                    </div>

                    <h3>Members</h3>
                    <table className='table table-hover org__table'>
                        <thead>
                            <tr>
                                <th className='org__avatar-cell'></th>
                                <th>Name</th>
                                <th>Username</th>
                                <th>Email</th>
                                <th className='org__actions-cell'></th>
                            </tr>
                        </thead>
                        <tbody>
                            {
                                org.members.map(member => (
                                    <tr key={member.id}>
                                        <td className='org__avatar-cell'><UserAvatar user={member.user} size={64}/></td>
                                        <td>{member.user.displayName}</td>
                                        <td>{member.user.username}</td>
                                        <td>{member.user.email}</td>
                                        <td className='org__actions-cell'>
                                            <button
                                                className='btn btn-icon'
                                                title={user.id === member.userID ? 'Leave' : 'Remove'}
                                                // tslint:disable-next-line:jsx-no-lambda
                                                onClick={() => memberRemoves.next({ ...member, org })}
                                            >
                                                <CloseIcon className='icon-inline' />
                                            </button>
                                        </td>
                                    </tr>
                                ))
                            }
                        </tbody>
                    </table>

                    <OrgSettingsForm org={org}/>
                </div>
            )
        })

})
