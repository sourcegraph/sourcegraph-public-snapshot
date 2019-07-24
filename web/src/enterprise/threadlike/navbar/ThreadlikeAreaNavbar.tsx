import SettingsIcon from 'mdi-react/SettingsIcon'
import React from 'react'
import { NavLink } from 'react-router-dom'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { ThreadlikePage } from '../ThreadlikeArea'

interface Props {
    thread: Pick<GQL.IThread, 'url'>
    pages: Pick<ThreadlikePage, 'title' | 'icon' | 'count' | 'path'>[]
    className?: string
}

const NAV_LINK_CLASS_NAME = 'threadlike-area-navbar__nav-link nav-link rounded-0 px-3'

/**
 * The navbar for a threadlike.
 */
export const ThreadlikeAreaNavbar: React.FunctionComponent<Props> = ({ thread, pages, className = '' }) => (
    <nav className={`threadlike-area-navbar border-bottom ${className}`}>
        <div className="container">
            <ul className="nav flex-nowrap">
                {pages.map(({ title, icon: Icon, count, path }, i) => (
                    <li key={i} className="threadlike-area-navbar__nav-item nav-item">
                        <NavLink
                            to={path ? `${thread.url}/${path}` : thread.url}
                            className={NAV_LINK_CLASS_NAME}
                            activeClassName="threadlike-area-navbar__nav-link--active"
                            aria-label={title}
                        >
                            {Icon && <Icon className="icon-inline" />} {title}{' '}
                            {count !== undefined && <span className="badge badge-secondary ml-1">{count}</span>}
                        </NavLink>
                    </li>
                ))}
                <li className="flex-1" />
                <li className="threadlike-area-navbar__nav-item nav-item">
                    <NavLink
                        to={`${thread.url}/settings`}
                        className={NAV_LINK_CLASS_NAME}
                        activeClassName="threadlike-area-navbar__nav-link--active"
                        aria-label="Settings"
                    >
                        <SettingsIcon className="icon-inline" />
                    </NavLink>
                </li>
            </ul>
        </div>
    </nav>
)
