import React from 'react'
import { NavLink } from 'react-router-dom'
import { OverviewPagesAreaPage } from '../OverviewPagesArea'

interface Props {
    areaUrl: string
    pages: Pick<
        OverviewPagesAreaPage<never>,
        'title' | 'icon' | 'count' | 'path' | 'exact' | 'navbarDividerBefore' | 'hideInNavbar'
    >[]
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
            <ul className="nav flex-nowrap" style={{ overflowX: 'auto' }}>
                {pages
                    .filter(page => !page.hideInNavbar)
                    .map(({ title, icon: Icon, count, path, exact, navbarDividerBefore }) => (
                        <React.Fragment key={path}>
                            {navbarDividerBefore && <li className="border-right my-3 mx-3 pr-3" role="divider" />}
                            <li className="overview-pages-area-navbar__nav-item nav-item">
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
                        </React.Fragment>
                    ))}
            </ul>
        </div>
    </nav>
)
