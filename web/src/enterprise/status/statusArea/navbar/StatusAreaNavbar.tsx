import SettingsIcon from 'mdi-react/SettingsIcon'
import React from 'react'
import { NavLink } from 'react-router-dom'
import { ChatIcon } from '../../../../../../shared/src/components/icons'
import { StatusIcon } from '../../icons'
import { StatusAreaContext } from '../StatusArea'

interface Props extends Pick<StatusAreaContext, 'statusURL'> {
    className?: string
}

const NAV_LINK_CLASS_NAME = 'status-area-navbar__nav-link nav-link rounded-0 px-3'

/**
 * The navbar for a single status.
 */
export const StatusAreaNavbar: React.FunctionComponent<Props> = ({ statusURL, className = '' }) => (
    <nav className={`status-area-navbar border-bottom ${className}`}>
        <div className="container">
            <ul className="nav flex-nowrap">
                <li className="status-area-navbar__nav-item nav-item">
                    <NavLink
                        to={statusURL}
                        exact={true}
                        className={NAV_LINK_CLASS_NAME}
                        activeClassName="status-area-navbar__nav-link--active"
                    >
                        <ChatIcon className="icon-inline" /> Notifications{' '}
                        <span className="badge badge-secondary ml-1">7</span>
                    </NavLink>
                </li>
                <li className="status-area-navbar__nav-item nav-item">
                    <NavLink
                        to={`${statusURL}/tasks`}
                        className={NAV_LINK_CLASS_NAME}
                        activeClassName="status-area-navbar__nav-link--active"
                    >
                        <StatusIcon className="icon-inline" /> Tasks
                    </NavLink>
                </li>
                <li className="flex-1" />
                <li className="status-area-navbar__nav-item nav-item">
                    <NavLink
                        to={`${statusURL}/settings`}
                        className={NAV_LINK_CLASS_NAME}
                        activeClassName="status-area-navbar__nav-link--active"
                        aria-label="Settings"
                    >
                        <SettingsIcon className="icon-inline" />
                    </NavLink>
                </li>
            </ul>
        </div>
    </nav>
)
