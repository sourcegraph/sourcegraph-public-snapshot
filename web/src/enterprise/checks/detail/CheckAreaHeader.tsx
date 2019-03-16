import CheckboxMultipleMarkedOutlineIcon from 'mdi-react/CheckboxMultipleMarkedOutlineIcon'
import HistoryIcon from 'mdi-react/HistoryIcon'
import SettingsIcon from 'mdi-react/SettingsIcon'
import React from 'react'
import { NavLink } from 'react-router-dom'
import { ChatIcon } from '../../../../../shared/src/components/icons'
import { Check } from '../data'

interface Props {
    check: Check
    areaURL: string
}

/**
 * The header for the check area (for a single check).
 */
export const CheckAreaHeader: React.FunctionComponent<Props> = ({ check, areaURL }) => (
    <div className="check-header border-top-0 border-bottom simple-area-header">
        <div className="container">
            <h1 className="font-weight-normal mt-3">
                <CheckboxMultipleMarkedOutlineIcon className="icon-inline text-muted small" /> {check.title}
            </h1>
            <div className="area-header__nav mt-4">
                <div className="area-header__nav-links">
                    <NavLink
                        to={areaURL}
                        className="btn area-header__nav-link"
                        activeClassName="area-header__nav-link--active"
                        exact={true}
                    >
                        <ChatIcon className="icon-inline" /> Conversation
                    </NavLink>
                    <NavLink
                        to={`${areaURL}/activity`}
                        className="btn area-header__nav-link"
                        activeClassName="area-header__nav-link--active"
                        exact={true}
                    >
                        <HistoryIcon className="icon-inline" /> Activity
                    </NavLink>
                    <NavLink
                        to={`${areaURL}/manage`}
                        className="btn area-header__nav-link"
                        activeClassName="area-header__nav-link--active"
                        exact={true}
                    >
                        <SettingsIcon className="icon-inline" /> Manage
                    </NavLink>
                </div>
            </div>
        </div>
    </div>
)
