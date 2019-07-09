import AlertCircleOutlineIcon from 'mdi-react/AlertCircleOutlineIcon'
import BellIcon from 'mdi-react/BellIcon'
import SettingsIcon from 'mdi-react/SettingsIcon'
import React from 'react'
import { NavLink } from 'react-router-dom'
import { ChecklistIcon } from '../../../../util/octicons'
import { CheckAreaContext } from '../CheckArea'

interface Props extends Pick<CheckAreaContext, 'check' | 'checkURL'> {
    className?: string
}

const NAV_LINK_CLASS_NAME = 'check-area-navbar__nav-link nav-link rounded-0 px-3'

/**
 * The navbar for a single check.
 */
export const CheckAreaNavbar: React.FunctionComponent<Props> = ({ checkURL, className = '' }) => (
    <nav className={`check-area-navbar border-bottom ${className}`}>
        <div className="container">
            <ul className="nav flex-nowrap">
                <li className="check-area-navbar__nav-item nav-item">
                    <NavLink
                        to={checkURL}
                        exact={true}
                        className={NAV_LINK_CLASS_NAME}
                        activeClassName="check-area-navbar__nav-link--active"
                    >
                        <BellIcon className="icon-inline" /> Notifications{' '}
                    </NavLink>
                </li>
                <li className="check-area-navbar__nav-item nav-item">
                    <NavLink
                        to={`${checkURL}/checks`}
                        className={NAV_LINK_CLASS_NAME}
                        activeClassName="check-area-navbar__nav-link--active"
                    >
                        <ChecklistIcon className="icon-inline" /> Checks
                    </NavLink>
                </li>
                <li className="check-area-navbar__nav-item nav-item">
                    <NavLink
                        to={`${checkURL}/issues`}
                        className={NAV_LINK_CLASS_NAME}
                        activeClassName="check-area-navbar__nav-link--active"
                    >
                        <AlertCircleOutlineIcon className="icon-inline" /> Issues
                    </NavLink>
                </li>
                <li className="flex-1" />
                <li className="check-area-navbar__nav-item nav-item">
                    <NavLink
                        to={`${checkURL}/settings`}
                        className={NAV_LINK_CLASS_NAME}
                        activeClassName="check-area-navbar__nav-link--active"
                        aria-label="Settings"
                    >
                        <SettingsIcon className="icon-inline" />
                    </NavLink>
                </li>
            </ul>
        </div>
    </nav>
)
