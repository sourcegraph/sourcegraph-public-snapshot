import FileDocumentBoxMultipleIcon from 'mdi-react/FileDocumentBoxMultipleIcon'
import FileMultipleIcon from 'mdi-react/FileMultipleIcon'
import React from 'react'
import { NavLink } from 'react-router-dom'
import * as sourcegraph from 'sourcegraph'

interface Props {
    task: sourcegraph.Diagnostic
    areaURL: string
    className?: string
}

const NAV_LINK_CLASS_NAME = 'task-area-navbar__nav-link nav-link rounded-0 px-4'

/**
 * The navbar for a single task.
 */
export const TaskAreaNavbar: React.FunctionComponent<Props> = ({ task, areaURL, className = '' }) => (
    <div className={`task-area-navbar border-bottom ${className}`}>
        <div className="container px-0">
            <div className="nav nav-pills flex-nowrap">
                <div className="nav-item">
                    <NavLink
                        to={areaURL}
                        exact={true}
                        className={NAV_LINK_CLASS_NAME}
                        activeClassName="task-area-navbar__nav-link--active"
                    >
                        <FileMultipleIcon className="icon-inline" /> Files changed
                        <span className="badge badge-secondary">9</span>
                    </NavLink>
                </div>
                <div className="nav-item">
                    <NavLink
                        to={`${areaURL}/changes`}
                        className={NAV_LINK_CLASS_NAME}
                        activeClassName="task-area-navbar__nav-link--active"
                    >
                        <FileDocumentBoxMultipleIcon className="icon-inline" /> Changes
                    </NavLink>
                </div>
            </div>
        </div>
    </div>
)
