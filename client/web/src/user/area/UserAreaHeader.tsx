import AccountCircleIcon from 'mdi-react/AccountCircleIcon'
import * as React from 'react'
import { Link, NavLink } from 'react-router-dom'
import { PageHeader } from '../../components/PageHeader'
import { orgURL } from '../../org'
import { OrgAvatar } from '../../org/OrgAvatar'
import { NavItemWithIconDescriptor } from '../../util/contributions'
import { UserAvatar } from '../UserAvatar'
import { UserAreaRouteContext } from './UserArea'

interface Props extends UserAreaRouteContext {
    navItems: readonly UserAreaHeaderNavItem[]
    className?: string
}

export interface UserAreaHeaderContext extends Pick<Props, 'user'> {
    isSourcegraphDotCom: boolean
}

export interface UserAreaHeaderNavItem extends NavItemWithIconDescriptor<UserAreaHeaderContext> {}

/**
 * Header for the user area.
 */
export const UserAreaHeader: React.FunctionComponent<Props> = ({ url, navItems, className = '', ...props }) => (
    <div className={`user-area-header ${className}`}>
        <div className="container">
            {props.user && (
                <>
                    <PageHeader
                        path={[
                            {
                                text: (
                                    <span className="align-middle">
                                        {props.user.displayName || props.user.username}
                                    </span>
                                ),
                                icon: () => <UserAvatar className="user-area-header__avatar" user={props.user} />,
                            },
                        ]}
                        className="mb-4"
                    />
                    <div className="d-flex align-items-end justify-content-between">
                        <ul className="nav nav-tabs border-bottom-0">
                            {navItems.map(
                                ({ to, label, exact, icon: Icon, condition = () => true }) =>
                                    condition(props) && (
                                        <li key={label} className="nav-item">
                                            <NavLink
                                                to={url + to}
                                                className="nav-link"
                                                activeClassName="active"
                                                exact={exact}
                                            >
                                                {Icon && <Icon className="icon-inline" />} {label}
                                            </NavLink>
                                        </li>
                                    )
                            )}
                        </ul>
                        <div className="flex-1" />
                        {props.user.organizations.nodes.length > 0 && (
                            <div className="d-flex align-items-center pb-1">
                                <small className="mr-2">Organizations</small>
                                {props.user.organizations.nodes.map(org => (
                                    <Link
                                        className="mr-2"
                                        key={org.id}
                                        to={orgURL(org.name)}
                                        data-tooltip={org.displayName || org.name}
                                    >
                                        <OrgAvatar org={org.name} />
                                    </Link>
                                ))}
                            </div>
                        )}
                    </div>
                </>
            )}
        </div>
    </div>
)
