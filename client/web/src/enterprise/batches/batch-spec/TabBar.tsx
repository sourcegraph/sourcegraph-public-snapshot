import { useMemo } from 'react'

import classNames from 'classnames'
import { uniqBy, upperFirst } from 'lodash'
import { NavLink as RouterLink } from 'react-router-dom'

import { Button } from '@sourcegraph/wildcard'

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
        <ul className={classNames('nav nav-tabs', styles.navList, className)}>
            {fullTabsConfig.map(({ key, isEnabled, handler }, index) => {
                const tabName = getTabName(key, index)

                return (
                    <li className="nav-item" key={key}>
                        {activeTabKey === key ? (
                            <span aria-disabled="true" className="nav-link active">
                                {tabName}
                            </span>
                        ) : !isEnabled ? (
                            <span
                                aria-disabled="true"
                                className={classNames('nav-link text-muted', styles.navLinkDisabled, styles.navLink)}
                                data-tab-content={tabName}
                            >
                                {tabName}
                            </span>
                        ) : !handler ? (
                            <span
                                aria-disabled="true"
                                className={classNames('nav-link', styles.navLink)}
                                data-tab-content={tabName}
                            >
                                {tabName}
                            </span>
                        ) : handler.type === 'link' ? (
                            <RouterLink
                                to={`${matchURL || ''}/${key}`}
                                role="button"
                                className={classNames('nav-link', styles.navLink)}
                                data-tab-content={tabName}
                            >
                                {tabName}
                            </RouterLink>
                        ) : (
                            <Button
                                className={classNames('nav-link', styles.button, styles.navLink)}
                                onClick={handler.onClick}
                                data-tab-content={tabName}
                            >
                                {tabName}
                            </Button>
                        )}
                    </li>
                )
            })}
        </ul>
    )
}
