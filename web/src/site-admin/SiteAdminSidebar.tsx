import FeedIcon from '@sourcegraph/icons/lib/Feed'
import LockIcon from '@sourcegraph/icons/lib/Lock'
import PuzzleIcon from '@sourcegraph/icons/lib/Puzzle'
import ServerIcon from '@sourcegraph/icons/lib/Server'
import * as H from 'history'
import * as React from 'react'
import { NavLink } from 'react-router-dom'
import { Subscription } from 'rxjs'
import * as GQL from '../backend/graphqlschema'
import { SIDEBAR_BUTTON_CLASS, SIDEBAR_CARD_CLASS, SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS } from '../components/Sidebar'
import { USE_PLATFORM } from '../extensions/environment/ExtensionsEnvironment'

interface Props {
    history: H.History
    location: H.Location
    className: string
    user: GQL.IUser
}

interface State {}

/**
 * Sidebar for the site admin area.
 */
export class SiteAdminSidebar extends React.Component<Props, State> {
    private subscriptions = new Subscription()

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className={`site-admin-sidebar ${this.props.className}`}>
                <div className={SIDEBAR_CARD_CLASS}>
                    <div className="card-header">
                        <ServerIcon className="icon-inline" /> Site admin
                    </div>
                    <div className="list-group list-group-flush">
                        <NavLink to="/site-admin" className={SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS} exact={true}>
                            Overview
                        </NavLink>
                        <NavLink
                            to="/site-admin/configuration"
                            className={SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS}
                            exact={true}
                        >
                            Configuration
                        </NavLink>
                        <NavLink
                            to="/site-admin/repositories"
                            className={SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS}
                            exact={true}
                        >
                            Repositories
                        </NavLink>
                    </div>
                </div>
                <div className={SIDEBAR_CARD_CLASS}>
                    <div className="list-group list-group-flush">
                        <NavLink to="/site-admin/users" className={SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS} exact={true}>
                            Users
                        </NavLink>
                        <NavLink
                            to="/site-admin/organizations"
                            className={SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS}
                            exact={true}
                        >
                            Organizations
                        </NavLink>
                        <NavLink
                            to="/site-admin/global-settings"
                            className={SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS}
                            exact={true}
                        >
                            Global settings
                        </NavLink>
                        <NavLink
                            to="/site-admin/code-intelligence"
                            className={SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS}
                            exact={true}
                        >
                            Code intelligence
                        </NavLink>
                    </div>
                </div>
                <div className={SIDEBAR_CARD_CLASS}>
                    <div className="card-header">
                        <LockIcon className="icon-inline" /> Auth
                    </div>
                    <div className="list-group list-group-flush">
                        <NavLink
                            to="/site-admin/auth/providers"
                            className={SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS}
                            exact={true}
                        >
                            Providers
                        </NavLink>
                        <NavLink
                            to="/site-admin/auth/external-accounts"
                            className={SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS}
                            exact={true}
                        >
                            External accounts
                        </NavLink>
                        <NavLink to="/site-admin/tokens" className={SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS} exact={true}>
                            Access tokens
                        </NavLink>
                    </div>
                </div>
                {USE_PLATFORM && (
                    <div className={SIDEBAR_CARD_CLASS}>
                        <div className="card-header">
                            <PuzzleIcon className="icon-inline" /> Registry
                        </div>
                        <div className="list-group list-group-flush">
                            <NavLink
                                to="/site-admin/registry/extensions"
                                className={SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS}
                                exact={true}
                            >
                                Extensions
                            </NavLink>
                        </div>
                    </div>
                )}
                <div className={SIDEBAR_CARD_CLASS}>
                    <div className="list-group list-group-flush">
                        <NavLink to="/site-admin/updates" className={SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS} exact={true}>
                            Updates
                        </NavLink>
                        <NavLink
                            to="/site-admin/analytics"
                            className={SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS}
                            exact={true}
                        >
                            Analytics
                        </NavLink>
                        <NavLink to="/site-admin/surveys" className={SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS} exact={true}>
                            User surveys
                        </NavLink>
                        <NavLink to="/site-admin/pings" className={SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS} exact={true}>
                            Pings
                        </NavLink>
                    </div>
                </div>
                <NavLink to="/api/console" className={SIDEBAR_BUTTON_CLASS}>
                    <FeedIcon className="icon-inline" />
                    API console
                </NavLink>
                <a href="/-/debug/" className={SIDEBAR_BUTTON_CLASS}>
                    Instrumentation
                </a>
            </div>
        )
    }
}
