import { ListboxGroup, ListboxGroupLabel, ListboxInput, ListboxList, ListboxPopover } from '@reach/listbox'
import { VisuallyHidden } from '@reach/visually-hidden'
import classnames from 'classnames'
import React, { useState } from 'react'

import { InsightDashboard, InsightsDashboardType } from '../../../../core/types'

import { MenuButton } from './components/menu-button/MenuButton'
import { SelectDashboardOption, SelectOption } from './components/select-option/SelectOption'
import styles from './DashboardSelect.module.scss'

const LABEL_ID = 'insights-dashboards--select'

export interface DashboardSelectProps {
    dashboards: InsightDashboard[]
}

/**
 * Renders dashboard select component for code insights dashboard page selection UI.
 */
export const DashboardSelect: React.FunctionComponent<DashboardSelectProps> = props => {
    const { dashboards } = props

    const [selectedDashboard, setSelectedDashboard] = useState<string>()

    const handleChange = (value: string): void => {
        setSelectedDashboard(value)
    }

    const organizationGroups = getDashboardOrganizationsGroups(dashboards)

    return (
        <div>
            <VisuallyHidden id={LABEL_ID}>Choose a dashboard</VisuallyHidden>

            <ListboxInput value={selectedDashboard} onChange={handleChange}>
                <MenuButton dashboards={dashboards} className={styles.selectButton} />

                <ListboxPopover className={classnames(styles.popover)} portal={true}>
                    <ListboxList className={classnames(styles.list, 'dropdown-menu')}>
                        <SelectOption
                            value={InsightsDashboardType.All}
                            label="All Insights"
                            className={styles.option}
                        />

                        <ListboxGroup>
                            <ListboxGroupLabel className={classnames(styles.groupLabel, 'text-muted')}>
                                Private
                            </ListboxGroupLabel>

                            {dashboards
                                .filter(dashboard => dashboard.type === InsightsDashboardType.Personal)
                                .map(dashboard => (
                                    <SelectDashboardOption
                                        key={dashboard.id}
                                        dashboard={dashboard}
                                        className={styles.option}
                                    />
                                ))}
                        </ListboxGroup>

                        {organizationGroups.map(group => (
                            <ListboxGroup key={group.id}>
                                <ListboxGroupLabel className={classnames(styles.groupLabel, 'text-muted')}>
                                    {group.name}
                                </ListboxGroupLabel>

                                {group.dashboards.map(dashboard => (
                                    <SelectDashboardOption
                                        key={dashboard.id}
                                        dashboard={dashboard}
                                        className={styles.option}
                                    />
                                ))}
                            </ListboxGroup>
                        ))}
                    </ListboxList>
                </ListboxPopover>
            </ListboxInput>
        </div>
    )
}

interface DashboardOrganizationGroup {
    id: string
    name: string
    dashboards: InsightDashboard[]
}

/**
 * Returns organization dashboards grouped by dashboard owner id
 */
const getDashboardOrganizationsGroups = (dashboards: InsightDashboard[]): DashboardOrganizationGroup[] => {
    const groupsDictionary = dashboards
        .filter(dashboard => dashboard.type === InsightsDashboardType.Organization)
        .reduce<Record<string, DashboardOrganizationGroup>>((store, dashboard) => {
            if (!store[dashboard.owner.id]) {
                store[dashboard.owner.id] = {
                    id: dashboard.owner.id,
                    name: dashboard.owner.name,
                    dashboards: [],
                }
            }

            store[dashboard.owner.id].dashboards.push(dashboard)

            return store
        }, {})

    return Object.values(groupsDictionary)
}
