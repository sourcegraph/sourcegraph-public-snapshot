import AddIcon from '@sourcegraph/icons/lib/Add'
import FriendsIcon from '@sourcegraph/icons/lib/Friends'
import GearIcon from '@sourcegraph/icons/lib/Gear'
import KeyIcon from '@sourcegraph/icons/lib/Key'
import SignOut from '@sourcegraph/icons/lib/SignOut'
import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Subscription } from 'rxjs/Subscription'
import { currentUser } from '../auth'
import { events } from '../tracking/events'

interface Props {
    history: H.History
}

interface State {
    orgs?: GQL.IOrg[]
}

/**
 * Sidebar for settings pages
 */
export class SettingsSidebar extends React.Component<Props, State> {

    private subscriptions = new Subscription()

    constructor() {
        super()
        this.state = {}
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            currentUser.subscribe(
                user => {
                    // If not logged in, redirect
                    if (!user) {
                        // TODO currently we can't redirect here because the initial value will always be `null`
                        // this.props.history.push('/sign-in')
                        return
                    }
                    this.setState({ orgs: user.orgs })
                }
            )
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className='settings-sidebar'>
                <div className='settings-sidebar__header'>
                    <div className='settings-sidebar__header-icon'><GearIcon /></div>
                    <div className='settings-sidebar__header-title'>Settings</div>
                </div>
                <ul>
                    {/* <li className='settings-sidebar__item'>
                        <Link to='/settings' className='settings-sidebar__item-text'>Profile</Link>
                    </li> */}
                    <li className='settings-sidebar__item'>
                        <div className='settings-sidebar__item-text'>Your profile</div>
                        <ul>
                            <li className='settings-sidebar__item'>
                                <Link to='/settings/editor-auth' className='settings-sidebar__item-text'><KeyIcon /> Authenticate your editor</Link>
                            </li>

                            <li className='settings-sidebar__item'>
                                <a href='/-/sign-out' className='settings-sidebar__item-text' onClick={this.logTelemetryOnSignOut}><SignOut /> Sign out</a>
                            </li>
                        </ul>
                    </li>

                    <li className='settings-sidebar__item'>
                        <div className='settings-sidebar__item-text'>Your teams</div>
                        <ul>
                            {
                                this.state.orgs && this.state.orgs.map(org => (
                                    <li className='settings-sidebar__item' key={org.id}>
                                        <Link to={`/settings/teams/${org.name}`} className='settings-sidebar__item-text'>
                                            <FriendsIcon /> {org.name}
                                        </Link>
                                    </li>
                                ))
                            }
                            <li className='settings-sidebar__item'>
                                <Link to='/settings/teams/new' className='settings-sidebar__item-text'>
                                    <AddIcon /> Create new team
                                </Link>
                            </li>
                        </ul>
                    </li>
                </ul>
            </div>
        )
    }

    private logTelemetryOnSignOut(): void {
        events.SignOutClicked.log()
    }
}
