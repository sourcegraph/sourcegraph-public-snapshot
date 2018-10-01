import * as React from 'react'
import { Link, NavLink, RouteComponentProps } from 'react-router-dom'
import { orgURL } from '../../org'
import { OrgAvatar } from '../../org/OrgAvatar'
import { NavItemWithIconDescriptor } from '../../util/contributions'
import { UserAvatar } from '../UserAvatar'
import { UserAreaRouteContext } from './UserArea'

interface UserAreaHeaderProps extends UserAreaRouteContext, RouteComponentProps<{}> {
    navItems: ReadonlyArray<UserAreaHeaderNavItem>
    className: string
}

export type UserAreaHeaderContext = Pick<UserAreaHeaderProps, 'user'>

export interface UserAreaHeaderNavItem extends NavItemWithIconDescriptor<UserAreaHeaderContext> {}

/**
 * Header for the user area.
 */
export const UserAreaHeader: React.SFC<UserAreaHeaderProps> = (props: UserAreaHeaderProps) => (
    <div className={`user-area-header area-header ${props.className}`}>
        <div className={`${props.className}-inner`}>
            {props.user && (
                <>
                    <h2 className="user-area-header__title">
                        {props.user.avatarURL && <UserAvatar className="user-area-header__avatar" user={props.user} />}
                        {props.user.displayName ? (
                            <>
                                {props.user.displayName}{' '}
                                <span className="user-area-header__title-subtitle">{props.user.username}</span>
                            </>
                        ) : (
                            props.user.username
                        )}
                    </h2>
                    <div className="area-header__nav">
                        <div className="area-header__nav-links">
                            {props.navItems.map(
                                ({ to, label, exact, icon: Icon, condition = () => true }) =>
                                    condition(props) && (
                                        <NavLink
                                            key={label}
                                            to={props.url + to}
                                            className="btn area-header__nav-link"
                                            activeClassName="area-header__nav-link--active"
                                            exact={exact}
                                        >
                                            {Icon && <Icon className="icon-inline" />} {label}
                                        </NavLink>
                                    )
                            )}
                        </div>
                        {props.user.organizations.nodes.length > 0 && (
                            <div className="area-header__nav-actions">
                                <small className="area-header__nav-actions-label">Organizations</small>
                                {props.user.organizations.nodes.map(org => (
                                    <Link
                                        className="area-header__nav-action"
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
