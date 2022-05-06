import { useMemo } from 'react'

import classNames from 'classnames'
import { uniqBy, upperFirst } from 'lodash'
import { NavLink as RouterLink } from 'react-router-dom'

import { Button } from '@sourcegraph/wildcard'

import styles from './TabBar.module.scss'

export type TabName = 'configuration' | 'batch spec' | 'execution' | 'preview'

export interface TabsConfig {
    name: TabName
    isEnabled: boolean
    disabledTooltip?: string
    handler?: { type: 'link'; to: string } | { type: 'button'; onClick: () => void }
}

const DEFAULT_TABS: TabsConfig[] = [
    { name: 'configuration', isEnabled: false },
    { name: 'batch spec', isEnabled: false },
    { name: 'execution', isEnabled: false },
    { name: 'preview', isEnabled: false },
]

interface TabBarProps {
    activeTabName: TabName
    tabsConfig: TabsConfig[]
}

export const TabBar: React.FunctionComponent<TabBarProps> = ({ activeTabName, tabsConfig }) => {
    // uniqBy removes duplicates by taking the first item it finds with a given 'name', so we spread the defaults last
    const fullTabsConfig = useMemo<TabsConfig[]>(() => uniqBy([...tabsConfig, ...DEFAULT_TABS], 'name'), [tabsConfig])
    return (
        <ul className="nav nav-tabs d-inline-flex d-sm-flex flex-nowrap text-nowrap">
            {fullTabsConfig.map(({ name, isEnabled, disabledTooltip, handler }, index) => {
                const tabName = `${index + 1}. ${upperFirst(name)}`

                return (
                    <li className="nav-item" key={name}>
                        {activeTabName === name ? (
                            <span aria-disabled="true" className="nav-link active">
                                {tabName}
                            </span>
                        ) : !isEnabled ? (
                            <span
                                aria-disabled="true"
                                className={classNames('nav-link text-muted', styles.navLinkDisabled, styles.navLink)}
                                data-tab-content={tabName}
                                data-tooltip={disabledTooltip}
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
                                to={handler.to}
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
