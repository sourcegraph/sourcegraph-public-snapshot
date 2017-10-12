import AddIcon from '@sourcegraph/icons/lib/Add'
import FriendsIcon from '@sourcegraph/icons/lib/Friends'
import GearIcon from '@sourcegraph/icons/lib/Gear'
import KeyIcon from '@sourcegraph/icons/lib/Key'
import SignOutIcon from '@sourcegraph/icons/lib/SignOut'
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
    editorBeta: boolean
}

/**
 * Sidebar for settings pages
 */
export class SettingsSidebar extends React.Component<Props, State> {

    private subscriptions = new Subscription()

    constructor() {
        super()
        this.state = {
            editorBeta: false
        }
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
                    if (user.tags) {
                        for (const tag of user.tags) {
                            if (tag.name === 'editor-beta') {
                                this.setState({ editorBeta: true })
                                break
                            }
                        }
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
                    <div className='settings-sidebar__header-icon'><GearIcon className='icon-inline' /></div>
                    <div className='settings-sidebar__header-title ui-title'>Settings</div>
                </div>
                <ul className='settings-sidebar__items'>
                    {this.state.editorBeta &&
                        <ul>
                            <li className='settings-sidebar__item'>
                                <Link to='/settings/editor-auth' className='settings-sidebar__item-link'>
                                    <KeyIcon className='icon-inline settings-sidebar__item-icon' /> Editor authentication
                                </Link>
                            </li>
                            <li className='settings-sidebar__item'>
                                <div className='settings-sidebar__item-header'>
                                    <FriendsIcon className='icon-inline settings-sidebar__item-icon'/> Your teams
                                </div>
                                <ul>
                                    {
                                        this.state.orgs && this.state.orgs.map(org => (
                                            <li className='settings-sidebar__item' key={org.id}>
                                                <Link to={`/settings/teams/${org.name}`} className='settings-sidebar__item-link'>
                                                    {org.name}
                                                </Link>
                                            </li>
                                        ))
                                    }
                                    <li className='settings-sidebar__item'>
                                        <Link to='/settings/teams/new' className='settings-sidebar__item-link'>
                                            <AddIcon className='icon-inline settings-sidebar__item-icon'/> Create new team
                                        </Link>
                                    </li>
                                </ul>
                            </li>
                        </ul>
                    }
                    <li className='settings-sidebar__item'>
                        <a href='/-/sign-out' className='settings-sidebar__item-link' onClick={this.logTelemetryOnSignOut}>
                            <SignOutIcon className='icon-inline settings-sidebar__item-icon' /> Sign out
                        </a>
                    </li>
                </ul>
            </div>
        )
    }

    private logTelemetryOnSignOut(): void {
        events.SignOutClicked.log()
    }
}
