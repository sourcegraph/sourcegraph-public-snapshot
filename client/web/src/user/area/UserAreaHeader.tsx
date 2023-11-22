import React, { useMemo } from 'react'

import classNames from 'classnames'
import { NavLink } from 'react-router-dom'

import { UserAvatar } from '@sourcegraph/shared/src/components/UserAvatar'
import { Icon, Link, PageHeader } from '@sourcegraph/wildcard'

import type { BatchChangesProps } from '../../batches'
import type { NavItemWithIconDescriptor } from '../../util/contributions'

import type { UserAreaRouteContext } from './UserArea'

import styles from './UserAreaHeader.module.scss'

interface Props extends UserAreaRouteContext {
    navItems: readonly UserAreaHeaderNavItem[]
    className?: string
}

export interface UserAreaHeaderContext extends BatchChangesProps, Pick<Props, 'user'> {
    isSourcegraphDotCom: boolean
    isCodyApp: boolean
}

export interface UserAreaHeaderNavItem extends NavItemWithIconDescriptor<UserAreaHeaderContext> {}

/**
 * Header for the user area.
 */
export const UserAreaHeader: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    url,
    navItems,
    className = '',
    ...props
}) => {
    /*
     * The path segment would always be recreated on rerenders, thus invalidating the loop over it in PageHeader.
     * As a result, the UserAvatar was always reinstanciated and rendered again, whenever the header rerenders
     * (every location change, for example). This prevents it from flickering.
     */
    const path = useMemo(
        () =>
            props.isCodyApp
                ? { text: 'Settings' }
                : {
                      text: (
                          <span className="align-middle">
                              {props.user.displayName ? (
                                  <>
                                      {props.user.displayName} ({props.user.username})
                                  </>
                              ) : (
                                  props.user.username
                              )}
                          </span>
                      ),
                      icon: () => <UserAvatar className={styles.avatar} user={props.user} />,
                  },
        [props.user, props.isCodyApp]
    )

    const filteredNavItems = navItems.filter(({ condition = () => true }) => condition(props))

    return (
        <div className={className}>
            <div className="container">
                <PageHeader
                    className="mb-3"
                    actions={props.isCodyApp && <Link to="/site-admin/configuration">Advanced settings</Link>}
                >
                    <PageHeader.Heading as="h2" styleAs="h1">
                        <PageHeader.Breadcrumb icon={path.icon}>{path.text}</PageHeader.Breadcrumb>
                    </PageHeader.Heading>
                </PageHeader>
                {filteredNavItems.length > 0 && (
                    <nav className="d-flex align-items-end justify-content-between" aria-label="User">
                        <ul className={classNames('nav nav-tabs w-100', styles.navigation)}>
                            {filteredNavItems.map(({ to, label, icon: ItemIcon }) => (
                                <li key={label} className="nav-item">
                                    <NavLink to={url + to} className={classNames('nav-link', styles.navigationLink)}>
                                        <span>
                                            {ItemIcon && <Icon as={ItemIcon} aria-hidden={true} />}{' '}
                                            <span className="text-content" data-tab-content={label}>
                                                {label}
                                            </span>
                                        </span>
                                    </NavLink>
                                </li>
                            ))}
                        </ul>
                    </nav>
                )}
            </div>
        </div>
    )
}
