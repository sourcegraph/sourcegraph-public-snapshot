import FileMultipleIcon from 'mdi-react/FileMultipleIcon'
import SettingsIcon from 'mdi-react/SettingsIcon'
import SourceBranchIcon from 'mdi-react/SourceBranchIcon'
import SourceCommitIcon from 'mdi-react/SourceCommitIcon'
import React from 'react'
import { NavLink } from 'react-router-dom'
import { ChatIcon } from '../../../../../../shared/src/components/icons'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { TasksIcon } from '../../../tasks/icons'
import { ChecklistIcon, DiffIcon, GitCommitIcon } from '../../icons'

interface Props {
    thread: GQL.IDiscussionThread
    areaURL: string
    className?: string
}

const NAV_LINK_CLASS_NAME = 'changeset-area-navbar__nav-link nav-link rounded-0 px-3'

/**
 * The navbar for a single changeset.
 */
export const ChangesetAreaNavbar: React.FunctionComponent<Props> = ({ thread, areaURL, className = '' }) => (
    <nav className={`changeset-area-navbar border-bottom ${className}`}>
        <div className="container">
            <ul className="nav flex-nowrap">
                <li className="changeset-area-navbar__nav-item nav-item">
                    <NavLink
                        to={areaURL}
                        exact={true}
                        className={NAV_LINK_CLASS_NAME}
                        activeClassName="changeset-area-navbar__nav-link--active"
                    >
                        <ChatIcon className="icon-inline" /> Discussion{' '}
                        <span className="badge badge-secondary ml-1">7</span>
                    </NavLink>
                </li>
                <li className="changeset-area-navbar__nav-item nav-item">
                    <NavLink
                        to={`${areaURL}/tasks`}
                        className={NAV_LINK_CLASS_NAME}
                        activeClassName="changeset-area-navbar__nav-link--active"
                    >
                        <TasksIcon className="icon-inline" /> Tasks{' '}
                        <span className="badge badge-secondary ml-1">5</span>
                    </NavLink>
                </li>
                <li className="changeset-area-navbar__nav-item nav-item">
                    <NavLink
                        to={`${areaURL}/commits`}
                        className={NAV_LINK_CLASS_NAME}
                        activeClassName="changeset-area-navbar__nav-link--active"
                    >
                        <GitCommitIcon className="icon-inline" /> Commits{' '}
                        <span className="badge badge-secondary ml-1">5</span>
                    </NavLink>
                </li>
                <li className="changeset-area-navbar__nav-item nav-item">
                    <NavLink
                        to={`${areaURL}/changes`}
                        className={NAV_LINK_CLASS_NAME}
                        activeClassName="changeset-area-navbar__nav-link--active"
                    >
                        <DiffIcon className="icon-inline" /> Changes{' '}
                        <span className="badge badge-secondary ml-1">9</span>
                    </NavLink>
                </li>
                <li className="flex-1" />
                <li className="changeset-area-navbar__nav-item nav-item">
                    <NavLink
                        to={`${areaURL}/settings`}
                        className={NAV_LINK_CLASS_NAME}
                        activeClassName="changeset-area-navbar__nav-link--active"
                        aria-label="Settings"
                    >
                        <SettingsIcon className="icon-inline" />
                    </NavLink>
                </li>
            </ul>
        </div>
    </nav>
)
