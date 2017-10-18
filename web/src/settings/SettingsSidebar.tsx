import AddIcon from '@sourcegraph/icons/lib/Add'
import ChartIcon from '@sourcegraph/icons/lib/Chart'
import CityIcon from '@sourcegraph/icons/lib/City'
import GearIcon from '@sourcegraph/icons/lib/Gear'
import KeyIcon from '@sourcegraph/icons/lib/Key'
import SignOutIcon from '@sourcegraph/icons/lib/SignOut'
import UserIcon from '@sourcegraph/icons/lib/User'
import * as H from 'history'
import * as React from 'react'
import { NavLink } from 'react-router-dom'
import { Subscription } from 'rxjs/Subscription'
import { currentUser } from '../auth'
import { events } from '../tracking/events'
import { OrgAvatar } from './org/OrgAvatar'
import { UserAvatar } from './user/UserAvatar'

interface Props {
    history: H.History
    location: H.Location
}

interface State {
    editorBeta: boolean
    currentUser?: GQL.IUser
    orgs?: GQL.IOrg[]
}

/**
 * Sidebar for settings pages
 */
export class SettingsSidebar extends React.Component<Props, State> {

    private subscriptions = new Subscription()

    constructor() {
        super()
        this.state = {
            editorBeta: false,
        }
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            currentUser.subscribe(
                user => {
                    // If not logged in, redirect
                    if (!user || !user.hasSourcegraphUser) {
                        // TODO currently we can't redirect here because the initial value will always be `null`
                        // this.props.history.push('/sign-in')
                        return
                    }
                    this.setState({ orgs: user.orgs, currentUser: user, editorBeta: !!user && user.tags.some(tag => tag.name === 'editor-beta')})
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
                <div className='settings-sidebar__header settings-sidebar__header-account-settings'>
                    <div className='settings-sidebar__header-icon'><GearIcon className='icon-inline' /></div>
                    <div className='settings-sidebar__header-title ui-title'>Account Settings</div>
                </div>
                <ul className='settings-sidebar__items'>
                    <div className='settings-sidebar__header'>
                        <div className='settings-sidebar__header-icon'><UserIcon className='icon-inline' /></div>
                        <div className='settings-sidebar__header-title ui-title'>Profile</div>
                    </div>
                    <li className='settings-sidebar__item'>
                        <NavLink
                            to='/settings'
                            className={`settings-sidebar__item-link`}
                            activeClassName={`${this.props.location && this.props.location.pathname === '/settings' && 'settings-sidebar__item--active'}`}
                        >
                            <div className='settings-sidebar__profile'>
                                <div className='settings-sidebar__profile-avatar-column'>
                                    <UserAvatar user={this.state.currentUser}/>
                                </div>
                                <div className='settings-sidebar__profile-content'>
                                    <div className='settings-sidebar__profile-row'>
                                        {this.state.currentUser ? this.state.currentUser.displayName : ''}
                                    </div>
                                    <div className='settings-sidebar__profile-row' title={this.state.currentUser ? this.state.currentUser.email  : ''}>
                                        {this.state.currentUser ? this.state.currentUser.email : ''}
                                    </div>
                                </div>
                            </div>
                        </NavLink>
                    </li>
                    {this.state.editorBeta &&
                        <ul>
                            <div className='settings-sidebar__header'>
                                <div className='settings-sidebar__header-icon'><CityIcon className='icon-inline' /></div>
                                <div className='settings-sidebar__header-title ui-title'>Organizations</div>
                            </div>
                            <ul>
                                {
                                    this.state.orgs &&
                                        this.state.orgs.map(org => (
                                            <li className='settings-sidebar__item' key={org.id}>
                                                <NavLink
                                                    to={`/settings/orgs/${org.name}`}
                                                    className='settings-sidebar__item-link'
                                                    activeClassName='settings-sidebar__item--active'
                                                >
                                                    <div className='settings-sidebar__profile-avatar-column'>
                                                        <OrgAvatar org={org.name}/>
                                                    </div>
                                                    {org.name}
                                                </NavLink>
                                            </li>
                                        ))
                                }
                                <li className='settings-sidebar__item'>
                                    <NavLink
                                        to='/settings/orgs/new'
                                        className='settings-sidebar__item-link'
                                        activeClassName='settings-sidebar__item--active'
                                    >
                                        <AddIcon className='icon-inline settings-sidebar__item-icon'/>Create new organization
                                    </NavLink>
                                </li>
                            </ul>
                            <div className='settings-sidebar__header'>
                                <div className='settings-sidebar__header-icon'><ChartIcon className='icon-inline' /></div>
                                <div className='settings-sidebar__header-title ui-title'>Connections</div>
                            </div>
                            <li className='settings-sidebar__item'>
                                <NavLink
                                    to='/settings/editor-auth'
                                    className='settings-sidebar__item-link'
                                    activeClassName='settings-sidebar__item--active'
                                >
                                    <KeyIcon className='icon-inline settings-sidebar__item-icon' />Editor authentication
                                </NavLink>
                            </li>
                        </ul>
                    }
                    <li className='settings-sidebar__item'>
                        <a href='/-/sign-out' className='settings-sidebar__item-link' onClick={this.logTelemetryOnSignOut}>
                            <SignOutIcon className='icon-inline settings-sidebar__item-icon' />Sign out
                        </a>
                    </li>
                    {this.state.editorBeta &&
                        <li className='settings-sidebar__item settings-sidebar__download-editor'>
                            <a className='settings-sidebar__download-editor-button btn' target='_blank' href='https://about.sourcegraph.com/beta/201708'>
                                Download Editor
                            </a>
                        </li>
                    }
                </ul>
            </div>
        )
    }

    private logTelemetryOnSignOut(): void {
        events.SignOutClicked.log()
    }
}
