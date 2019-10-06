import React from 'react'
import { NavLink } from 'react-router-dom'
import { OverviewPagesAreaPage } from '../OverviewPagesArea'

interface Props {
    areaUrl: string
    pages: Pick<OverviewPagesAreaPage<never>, 'title' | 'icon' | 'count' | 'path' | 'exact'>[]
    className?: string
}

const NAV_LINK_CLASS_NAME =
    'overview-pages-area-navbar__nav-link nav-link rounded-0 px-3 text-nowrap d-flex align-items-center'

/**
 * The navbar for {@link OverviewPagesArea}.
 */
export const OverviewPagesAreaNavbar: React.FunctionComponent<Props> = ({ areaUrl, pages, className = '' }) => (
    <nav className={`overview-pages-area-navbar border-bottom ${className}`}>
        <div className="container">
            {/* eslint-disable-next-line react/forbid-dom-props */}
            <ul className="nav flex-nowrap" style={{ overflowX: 'auto' }}>
                {pages.map(({ title, icon: Icon, count, path, exact }) => (
                    <li key={path} className="overview-pages-area-navbar__nav-item nav-item">
                        <NavLink
                            to={path ? `${areaUrl}${path}` : areaUrl}
                            exact={exact}
                            className={NAV_LINK_CLASS_NAME}
                            activeClassName="overview-pages-area-navbar__nav-link--active"
                            aria-label={title}
                        >
                            {Icon && <Icon className="icon-inline mr-2" />} {title}{' '}
                            {count !== undefined && <span className="badge badge-secondary ml-2">{count}</span>}
                        </NavLink>
                    </li>
                ))}
            </ul>
        </div>
    </nav>
)
