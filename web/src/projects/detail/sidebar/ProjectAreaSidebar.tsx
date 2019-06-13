import FolderTextIcon from 'mdi-react/FolderTextIcon'
import HomeVariantIcon from 'mdi-react/HomeVariantIcon'
import SettingsIcon from 'mdi-react/SettingsIcon'
import React from 'react'
import { Link, NavLink } from 'react-router-dom'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { CollapsibleSidebar } from '../../../components/collapsibleSidebar/CollapsibleSidebar'
import { ChecksIcon } from '../../../enterprise/checks/icons'
import { ThreadsIcon } from '../../../enterprise/threads/icons'
import { NavItemWithIconDescriptor } from '../../../util/contributions'
import { ProjectAvatar } from '../../components/ProjectAvatar'
import { LabelIcon } from '../../icons'

interface Props extends ExtensionsControllerProps {
    project: GQL.IProject
    areaURL: string
    className?: string
}

const LINKS: NavItemWithIconDescriptor[] = [
    { to: '', label: 'Project', icon: HomeVariantIcon, exact: true },
    { to: '/tree', label: 'Repository', icon: FolderTextIcon },
    { to: '/checks', label: 'Checks', icon: ChecksIcon },
    { to: '/threads', label: 'Threads', icon: ThreadsIcon },
    { to: '/labels', label: 'Labels', icon: LabelIcon },
    { to: '/settings', label: 'Settings', icon: SettingsIcon },
]

/**
 * The sidebar for the project area (for a single project).
 */
export const ProjectAreaSidebar: React.FunctionComponent<Props> = ({ project, areaURL, className = '' }) => (
    <CollapsibleSidebar
        localStorageKey="project-area__sidebar"
        side="left"
        className={`project-area-sidebar d-flex flex-column ${className}`}
        collapsedClassName="project-area-sidebar--collapsed"
        expandedClassName="project-area-sidebar--expanded"
    >
        {expanded => (
            <>
                <h3>
                    <Link
                        to={areaURL}
                        className="project-area-sidebar__nav-link pt-3 pb-2 d-block text-decoration-none shadow-none text-body px-3 text-truncate h5 mb-0"
                        data-tooltip={expanded ? '' : project.name}
                    >
                        <ProjectAvatar project={project} className={expanded ? 'mr-3' : ''} />
                        {expanded && project.name}
                    </Link>
                </h3>
                <ul className="list-group list-group-flush">
                    {LINKS.map(({ to, label, icon: Icon, exact }, i) => (
                        <li key={i} className="nav-item">
                            <NavLink
                                to={areaURL + to}
                                exact={exact}
                                className="project-area-sidebar__nav-link nav-link p-3 text-nowrap d-flex align-items-center"
                                activeClassName="project-area-sidebar__nav-link--active"
                                data-tooltip={expanded ? '' : label}
                            >
                                {Icon && <Icon className="icon-inline mr-3" />} {expanded && label}
                            </NavLink>
                        </li>
                    ))}
                </ul>
            </>
        )}
    </CollapsibleSidebar>
)
