import React, { useCallback } from 'react'

import { mdiFilterOutline, mdiPoll } from '@mdi/js'
import classNames from 'classnames'

import { SidebarTabID } from '@sourcegraph/shared/src/settings/temporary/searchSidebar'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { Button, Icon, Tooltip } from '@sourcegraph/wildcard'

import styles from './SidebarButtonStrip.module.scss'

interface SidebarTab {
    tab: SidebarTabID
    icon: string
    name: string
}

const tabs: SidebarTab[] = [
    { tab: SidebarTabID.FILTERS, icon: mdiFilterOutline, name: 'Filters' },
    { tab: SidebarTabID.INSIGHTS, icon: mdiPoll, name: 'Insights' },
]

export const SidebarButtonStrip: React.FunctionComponent<{ className?: string }> = ({ className }) => {
    const [selectedTab, setSelectedTab] = useTemporarySetting('search.sidebar.selectedTab', SidebarTabID.FILTERS)

    const onClickTab = useCallback(
        (tab: SidebarTabID) => {
            if (selectedTab === tab) {
                setSelectedTab(null) // Close the sidebar if clicking the currently selected tab
            } else {
                setSelectedTab(tab)
            }
        },
        [selectedTab, setSelectedTab]
    )

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
