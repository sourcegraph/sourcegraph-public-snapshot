import { ListboxGroup, ListboxGroupLabel, ListboxInput, ListboxList, ListboxPopover } from '@reach/listbox'
import { VisuallyHidden } from '@reach/visually-hidden'
import classNames from 'classnames'
import React from 'react'

import {
    InsightDashboard,
    InsightDashboardOwner,
    InsightsDashboardType,
    isGlobalDashboard,
    isOrganizationDashboard,
    isPersonalDashboard,
    RealInsightDashboard,
} from '../../../../../core/types'

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

    const organizationGroups = getDashboardOrganizationsGroups(dashboards)

    return (
        <div className={className}>
            <VisuallyHidden id={LABEL_ID}>Choose a dashboard</VisuallyHidden>

            <ListboxInput aria-labelledby={LABEL_ID} value={value ?? 'unknown'} onChange={handleChange}>
                <MenuButton dashboards={dashboards} />

                <ListboxPopover className={classNames(styles.popover)} portal={true}>
                    <ListboxList className={classNames(styles.list, 'dropdown-menu')}>
                        <SelectOption
                            value={InsightsDashboardType.All}
                            label="All Insights"
                            className={styles.option}
                        />

                        {dashboards.some(isPersonalDashboard) && (
                            <ListboxGroup>
                                <ListboxGroupLabel className={classNames(styles.groupLabel, 'text-muted')}>
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
                        )}

                        {dashboards.some(isGlobalDashboard) && (
                            <ListboxGroup>
                                <ListboxGroupLabel className={classNames(styles.groupLabel, 'text-muted')}>
                                    Global
                                </ListboxGroupLabel>

                                {dashboards.filter(isGlobalDashboard).map(dashboard => (
                                    <SelectDashboardOption
                                        key={dashboard.id}
                                        dashboard={dashboard}
                                        className={styles.option}
                                    />
                                ))}
                            </ListboxGroup>
                        )}

                        {organizationGroups.map(group => (
                            <ListboxGroup key={group.id}>
                                <ListboxGroupLabel className={classNames(styles.groupLabel, 'text-muted')}>
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
    dashboards: RealInsightDashboard[]
}

/**
 * Returns organization dashboards grouped by dashboard owner id
 */
const getDashboardOrganizationsGroups = (dashboards: InsightDashboard[]): DashboardOrganizationGroup[] => {
    const groupsDictionary = dashboards
        .filter(isOrganizationDashboard)
        .reduce<Record<string, DashboardOrganizationGroup>>((store, dashboard) => {
            const orgId = (dashboard.grants?.organizations && dashboard.grants?.organizations[0]) || ''
            const owner: InsightDashboardOwner = dashboard.owner || {
                id: orgId,
                name: orgId,
            }

            // This shouldn't happen. If we have made it this far, we should have a valid owner.
            if (owner.id === '') {
                return store
            }

            console.log('decoded name', atob(owner.id))

            if (!store[owner.id]) {
                store[owner.id] = {
                    id: owner.id,
                    name: owner.name,
                    dashboards: [],
                }
            }

            store[owner.id].dashboards.push(dashboard)

            return store
        }, {})

    return Object.values(groupsDictionary)
}
