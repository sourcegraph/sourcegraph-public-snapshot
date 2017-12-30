import AddIcon from '@sourcegraph/icons/lib/Add'
import CityIcon from '@sourcegraph/icons/lib/City'
import * as React from 'react'
import { NavLink, RouteComponentProps } from 'react-router-dom'
import { Subscription } from 'rxjs/Subscription'
import { currentUser } from '../auth'
import { OrgAvatar } from '../org/OrgAvatar'

interface Props extends RouteComponentProps<{ orgName: string }> {}

interface State {
    orgs?: GQL.IOrg[]
}

/**
 * Sidebar for org pages
 */
export class OrgSidebar extends React.Component<Props, State> {
    public state: State = {}

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            currentUser.subscribe(user => {
                this.setState({ orgs: user ? user.orgs : undefined })
            })
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const org: GQL.IOrg | undefined =
            this.state.orgs && this.state.orgs.find(org => org.name === this.props.match.params.orgName)

        if (!this.state.orgs) {
            return <div className="sidebar org-sidebar" />
        } else if (!org) {
            return null
        }

        return (
            <div className="sidebar org-sidebar">
                <ul className="sidebar__items">
                    <div className="sidebar__header">
                        <div className="sidebar__header-icon">
                            <OrgAvatar org={org.name} />
                        </div>
                        <h5 className="sidebar__header-title">{org.name}</h5>
                    </div>
                    <li className="sidebar__item">
                        <NavLink
                            to={`/organizations/${org.name}/settings/profile`}
                            exact={true}
                            className="sidebar__item-link"
                            activeClassName="sidebar__item--active"
                        >
                            Profile
                        </NavLink>
                    </li>
                    <li className="sidebar__item">
                        <NavLink
                            to={`/organizations/${org.name}/settings/configuration`}
                            exact={true}
                            className="sidebar__item-link"
                            activeClassName="sidebar__item--active"
                        >
                            Configuration
                        </NavLink>
                    </li>
                    <ul>
                        <div className="sidebar__header">
                            <div className="sidebar__header-icon">
                                <CityIcon className="icon-inline" />
                            </div>
                            <h5 className="sidebar__header-title ui-title">Organizations</h5>
                        </div>
                        <ul>
                            {this.state.orgs &&
                                this.state.orgs.map(org => (
                                    <li className="sidebar__item" key={org.id}>
                                        <NavLink
                                            to={`/organizations/${org.name}/settings/profile`}
                                            className="sidebar__item-link"
                                            activeClassName="sidebar__item--active"
                                        >
                                            <div className="sidebar__item-icon">
                                                <OrgAvatar org={org.name} />
                                            </div>
                                            {org.name}
                                        </NavLink>
                                    </li>
                                ))}
                            <li className="sidebar__item">
                                <NavLink
                                    to="/organizations/new"
                                    className="sidebar__item-link"
                                    activeClassName="sidebar__item--active"
                                >
                                    <AddIcon className="icon-inline sidebar__item-icon" />Create new organization
                                </NavLink>
                            </li>
                        </ul>
                    </ul>
                </ul>
            </div>
        )
    }
}
