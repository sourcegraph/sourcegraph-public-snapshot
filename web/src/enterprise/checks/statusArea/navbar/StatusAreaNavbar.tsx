import AlertCircleOutlineIcon from 'mdi-react/AlertCircleOutlineIcon'
import BellIcon from 'mdi-react/BellIcon'
import SettingsIcon from 'mdi-react/SettingsIcon'
import React from 'react'
import { NavLink } from 'react-router-dom'
import { ChecklistIcon } from '../../../../util/octicons'
import { CheckAreaContext } from '../CheckArea'

interface Props extends Pick<CheckAreaContext, 'status' | 'statusURL'> {
    className?: string
}

const NAV_LINK_CLASS_NAME = 'status-area-navbar__nav-link nav-link rounded-0 px-3'

/**
 * The navbar for a single status.
 */
export const CheckAreaNavbar: React.FunctionComponent<Props> = ({ statusURL, className = '' }) => (
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
                        <BellIcon className="icon-inline" /> Notifications{' '}
                    </NavLink>
                </li>
                <li className="status-area-navbar__nav-item nav-item">
                    <NavLink
                        to={`${statusURL}/checks`}
                        className={NAV_LINK_CLASS_NAME}
                        activeClassName="status-area-navbar__nav-link--active"
                    >
                        <ChecklistIcon className="icon-inline" /> Checks
                    </NavLink>
                </li>
                <li className="status-area-navbar__nav-item nav-item">
                    <NavLink
                        to={`${statusURL}/issues`}
                        className={NAV_LINK_CLASS_NAME}
                        activeClassName="status-area-navbar__nav-link--active"
                    >
                        <AlertCircleOutlineIcon className="icon-inline" /> Issues
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
