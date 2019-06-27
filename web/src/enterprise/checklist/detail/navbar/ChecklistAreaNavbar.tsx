import FileMultipleIcon from 'mdi-react/FileMultipleIcon'
import React from 'react'
import { NavLink } from 'react-router-dom'
import { ChatIcon } from '../../../../../../shared/src/components/icons'
import { Checklist } from '../../checklist'

interface Props {
    checklist: Checklist
    areaURL: string
    className?: string
}

const NAV_LINK_CLASS_NAME = 'checklist-area-navbar__nav-link nav-link rounded-0 px-4'

/**
 * The navbar for a single checklist.
 */
export const ChecklistAreaNavbar: React.FunctionComponent<Props> = ({ checklist, areaURL, className = '' }) => (
    <div className={`checklist-area-navbar border-bottom ${className}`}>
        <div className="container px-0">
            <div className="nav nav-pills flex-nowrap">
                <div className="nav-item">
                    <NavLink
                        to={areaURL}
                        exact={true}
                        className={NAV_LINK_CLASS_NAME}
                        activeClassName="checklist-area-navbar__nav-link--active"
                    >
                        <ChatIcon className="icon-inline mr-1" /> Overview
                    </NavLink>
                </div>
                <div className="nav-item">
                    <NavLink
                        to={`${areaURL}/files`}
                        className={NAV_LINK_CLASS_NAME}
                        activeClassName="checklist-area-navbar__nav-link--active"
                    >
                        <FileMultipleIcon className="icon-inline mr-1" /> Files changed{' '}
                        <span className="badge badge-secondary ml-1">9</span>
                    </NavLink>
                </div>
            </div>
        </div>
    </div>
)
