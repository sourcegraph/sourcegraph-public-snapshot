import React, { HTMLAttributes, useCallback } from 'react'

import classNames from 'classnames'

import { SidebarTabID } from '@sourcegraph/shared/src/settings/temporary/searchSidebar'
import { Button, Icon, Tooltip } from '@sourcegraph/wildcard'

import styles from './SidebarButtonStrip.module.scss'

export interface SidebarTab {
    tab: SidebarTabID
    icon: string
    name: string
}

interface SidebarButtonStripProps extends HTMLAttributes<HTMLDivElement> {
    tabs: SidebarTab[]
    selectedTab: SidebarTabID | null | undefined
    onSelectedTabs: (id: SidebarTabID | null) => void
}

export const SidebarButtonStrip: React.FunctionComponent<SidebarButtonStripProps> = props => {
    const { tabs, selectedTab, className, onSelectedTabs } = props

    const onClickTab = useCallback(
        (tab: SidebarTabID) => {
            if (selectedTab === tab) {
                onSelectedTabs(null) // Close the sidebar if clicking the currently selected tab
            } else {
                onSelectedTabs(tab)
            }
        },
        [selectedTab, onSelectedTabs]
    )

    if (tabs.length === 0) {
        return null
    }

    return (
        <div className={classNames(styles.strip, className)}>
            {tabs.map(({ tab, icon, name }) => (
                <Tooltip key={tab} content={name} placement="left">
                    <Button
                        onClick={() => onClickTab(tab)}
                        role="tab"
                        aria-selected={tab === selectedTab}
                        variant="icon"
                        className={styles.button}
                    >
                        <Icon svgPath={icon} aria-label={name} className={styles.icon} />
                    </Button>
                </Tooltip>
            ))}
        </div>
    )
}
