import { useMemo } from 'react'

import classNames from 'classnames'
import { uniqBy, upperFirst } from 'lodash'
import { NavLink as RouterLink } from 'react-router-dom'

import { Link } from '@sourcegraph/wildcard'

import styles from './TabBar.module.scss'

export const TAB_KEYS = ['configuration', 'spec', 'execution', 'preview'] as const
export type TabKey = typeof TAB_KEYS[number]

export interface TabsConfig {
    key: TabKey
    isEnabled: boolean
    handler?: { type: 'link' } | { type: 'button'; onClick: () => void }
}

const DEFAULT_TABS: TabsConfig[] = [
    { key: 'configuration', isEnabled: false },
    { key: 'spec', isEnabled: false },
    { key: 'execution', isEnabled: false },
    { key: 'preview', isEnabled: false },
]

const getTabName = (key: TabKey, index: number): string => {
    const lowerCaseName = key === 'spec' ? 'batch spec' : key
    return `${index + 1}. ${upperFirst(lowerCaseName)}`
}

interface TabBarProps {
    activeTabKey: TabKey
    tabsConfig: TabsConfig[]
    matchURL?: string
    className?: string
}

export const TabBar: React.FunctionComponent<React.PropsWithChildren<TabBarProps>> = ({
    activeTabKey,
    tabsConfig,
    matchURL,
    className,
}) => {
    // uniqBy removes duplicates by taking the first item it finds with a given 'name', so we spread the defaults last
    const fullTabsConfig = useMemo<TabsConfig[]>(() => uniqBy([...tabsConfig, ...DEFAULT_TABS], 'key'), [tabsConfig])
    return (
        <nav className="overflow-auto flex-shrink-0" aria-label="Steps">
            <div
                className={classNames('nav nav-tabs d-inline-flex d-sm-flex flex-nowrap text-nowrap', className)}
                role="tablist"
            >
                {fullTabsConfig.map(({ key, isEnabled, handler }, index) => {
                    const tabName = getTabName(key, index)

                    return (
                        <div className="nav-item" key={key}>
                            {!isEnabled ? (
                                <span
                                    aria-disabled={true}
                                    className={classNames(
                                        'nav-link text-muted',
                                        styles.navLinkDisabled,
                                        styles.navLink
                                    )}
                                    role="tab"
                                    data-tab-content={tabName}
                                >
                                    {tabName}
                                </span>
                            ) : !handler ? (
                                <span
                                    className={classNames('nav-link', styles.navLink, activeTabKey === key && 'active')}
                                    role="tab"
                                    aria-selected={activeTabKey === key}
                                    data-tab-content={tabName}
                                >
                                    {tabName}
                                </span>
                            ) : handler.type === 'link' ? (
                                <RouterLink
                                    to={`${matchURL || ''}/${key}`}
                                    className={classNames('nav-link', styles.navLink, activeTabKey === key && 'active')}
                                    aria-selected={activeTabKey === key}
                                    role="tab"
                                    data-tab-content={tabName}
                                >
                                    {tabName}
                                </RouterLink>
                            ) : (
                                <Link
                                    to=""
                                    className={classNames('nav-link', styles.navLink, activeTabKey === key && 'active')}
                                    aria-selected={activeTabKey === key}
                                    onClick={event => {
                                        event.preventDefault()
                                        handler.onClick()
                                    }}
                                    role="tab"
                                    data-tab-content={tabName}
                                >
                                    {tabName}
                                </Link>
                            )}
                        </div>
                    )
                })}
            </div>
        </nav>
    )
}
