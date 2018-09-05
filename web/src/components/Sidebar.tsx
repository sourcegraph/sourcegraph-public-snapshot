import React from 'react'
import { NavLink } from 'react-router-dom'

export const SIDEBAR_CARD_CLASS = 'card mb-3'

export const SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS = 'list-group-item list-group-item-action py-2'

export const SIDEBAR_BUTTON_CLASS = 'btn btn-secondary d-block w-100 my-2'

type Icon = React.ComponentType<{ className?: string }>

export const SideBarNavItem: React.SFC<{ to: string; exact: boolean }> = ({ children, to, exact }) => (
    <NavLink to={to} exact={exact} className={SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS}>
        {children}
    </NavLink>
)

export const SideBarGroupHeader: React.SFC<{ icon: Icon; label: string; children?: undefined }> = ({
    icon: Icon,
    label,
}) => (
    <div className="card-header">
        <Icon className="icon-inline" /> {label}
    </div>
)

export const SideBarGroup: React.SFC<{}> = ({ children }) => (
    <div className={SIDEBAR_CARD_CLASS}>
        <div className="list-group list-group-flush">{children}</div>
    </div>
)
