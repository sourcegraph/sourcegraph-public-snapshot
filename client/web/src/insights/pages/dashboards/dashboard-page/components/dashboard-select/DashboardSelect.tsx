import { ListboxGroup, ListboxGroupLabel, ListboxInput, ListboxList, ListboxPopover } from '@reach/listbox'
import { VisuallyHidden } from '@reach/visually-hidden'
import classnames from 'classnames'
import React from 'react'

import {
    InsightDashboard,
    InsightsDashboardType,
    isPersonalDashboard,
} from '../../../../../core/types'
import { getGroupedOrganizationDashboards } from './utils/get-grouped-org-dashboards';

import { MenuButton } from './components/menu-button/MenuButton'
import { SelectDashboardOption, SelectOption } from './components/select-option/SelectOption'
import styles from './DashboardSelect.module.scss'

const LABEL_ID = 'insights-dashboards--select'

export interface DashboardSelectProps {
    value: string | undefined
    dashboards: InsightDashboard[]

    onSelect: (dashboard: InsightDashboard) => void
    className?: string
}

/**
 * Renders dashboard select component for the code insights dashboard page selection UI.
 */
export const DashboardSelect: React.FunctionComponent<DashboardSelectProps> = props => {
    const { value, dashboards, onSelect, className } = props

    const handleChange = (value: string): void => {
        const dashboard = dashboards.find(dashboard => dashboard.id === value)

        if (dashboard) {
            onSelect(dashboard)
        }
    }

    const organizationGroups = getGroupedOrganizationDashboards(dashboards)

    return (
        <div className={className}>
            <VisuallyHidden id={LABEL_ID}>Choose a dashboard</VisuallyHidden>

            <ListboxInput aria-labelledby={LABEL_ID} value={value} onChange={handleChange}>
                <MenuButton dashboards={dashboards} />

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

                            {dashboards.filter(isPersonalDashboard).map(dashboard => (
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
