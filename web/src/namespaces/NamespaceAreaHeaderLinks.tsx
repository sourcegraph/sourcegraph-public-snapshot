import React from 'react'
import { NavLink } from 'react-router-dom'
import { ProjectIcon } from '../projects/icons'
import { NavItemWithIconDescriptor } from '../util/contributions'

interface Props {
    url: string
}

/**
 * Link data for the headers of namespace areas.
 */
export const NAMESPACE_AREA_HEADER_LINKS: readonly Pick<
    NavItemWithIconDescriptor,
    Exclude<keyof NavItemWithIconDescriptor, 'condition'>
>[] = [
    {
        to: '/namespace/projects',
        label: 'Projects',
        icon: ProjectIcon,
    },
]

/**
 * Links for the headers of namespace areas.
 */
export const NamespaceAreaHeaderLinks: React.FunctionComponent<Props> = ({ url }) => (
    <>
        {NAMESPACE_AREA_HEADER_LINKS.map(({ to, label, exact, icon: Icon }) => (
            <li key={to} className="nav-item">
                <NavLink
                    to={url + to}
                    className="btn area-header__nav-link"
                    activeClassName="area-header__nav-link--active"
                    exact={exact}
                >
                    {Icon && <Icon className="icon-inline" />} {label}
                </NavLink>
            </li>
        ))}
    </>
)
