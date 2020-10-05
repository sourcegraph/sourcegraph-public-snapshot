import * as React from 'react'
import { NavLink } from 'react-router-dom'
import { NavItemWithIconDescriptor } from '../../util/contributions'
import { UserAreaRouteContext } from './UserArea'

interface Props extends UserAreaRouteContext {
    navItems: readonly UserAreaTabsNavItem[]
    size: 'small' | 'large'
    className?: string
}

interface UserAreaTabsContext extends Pick<Props, 'user'> {
    isSourcegraphDotCom: boolean
}

export interface UserAreaTabsNavItem extends NavItemWithIconDescriptor<UserAreaTabsContext> {}

/**
 * Tabs for the user area.
 */
export const UserAreaTabs: React.FunctionComponent<Props> = ({ url, navItems, size, className = '', ...props }) => (
    <ul className={`nav nav-pills ${className}`}>
        {navItems.map(
            ({ to, label, exact, icon: Icon, condition = () => true }) =>
                condition(props) && (
                    <li key={label} className="nav-item">
                        <NavLink to={url + to} className="nav-link" activeClassName="active" exact={exact}>
                            {Icon && <Icon className="icon-inline" />} {label}
                        </NavLink>
                    </li>
                )
        )}
    </ul>
)
