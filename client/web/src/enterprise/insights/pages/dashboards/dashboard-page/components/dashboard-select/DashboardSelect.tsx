import React, { useContext, useEffect, useState } from 'react'

import { ListboxGroup, ListboxGroupLabel, ListboxInput, ListboxList, ListboxPopover } from '@reach/listbox'
import { VisuallyHidden } from '@reach/visually-hidden'
import classNames from 'classnames'

import { Input, Typography } from '@sourcegraph/wildcard'

import {
    CodeInsightsBackendContext,
    CustomInsightDashboard,
    InsightDashboard,
    isCustomDashboard,
    isGlobalDashboard,
    isOrganizationDashboard,
    isPersonalDashboard,
    isVirtualDashboard,
} from '../../../../../core'

import { MenuButton, SelectDashboardOption, SelectOption } from './components'

import styles from './DashboardSelect.module.scss'

export interface DashboardSelectProps {
    value: string | undefined
    dashboards: InsightDashboard[]
    className?: string
    onSelect: (dashboard: InsightDashboard) => void
}

/**
 * Renders dashboard select component for the code insights dashboard page selection UI.
 */
export const DashboardSelect: React.FunctionComponent<React.PropsWithChildren<DashboardSelectProps>> = props => {
    const { value, dashboards: rawDashboards, onSelect, className } = props
    const [filter, setFilter] = useState('')
    const [dashboards, setDashboards] = useState(rawDashboards)
    const {
        UIFeatures: { licensed },
    } = useContext(CodeInsightsBackendContext)

    const handleChange = (value: string): void => {
        const dashboard = dashboards.find(dashboard => dashboard.id === value)

        if (dashboard) {
            setFilter('')
            setDashboards(rawDashboards)
            onSelect(dashboard)
        }
    }

    const handleFilter: React.ChangeEventHandler<HTMLInputElement> = event => {
        setFilter(event.target.value)
    }

    useEffect(() => {
        if (filter === '') {
            setDashboards(rawDashboards)
            return
        }
        setDashboards(rawDashboards.filter(({ title }) => title.toLowerCase().includes(filter.toLowerCase())))
    }, [filter, rawDashboards])

    const customDashboards = dashboards.filter(isCustomDashboard)
    const organizationGroups = getDashboardOrganizationsGroups(customDashboards)

    return (
        <div className={className}>
            <VisuallyHidden id="insights-dashboards--select">Choose a dashboard</VisuallyHidden>

            <ListboxInput
                aria-labelledby="insights-dashboards--select"
                value={value ?? 'unknown'}
                onChange={handleChange}
            >
                <MenuButton dashboards={rawDashboards} />

                <ListboxPopover className={classNames(styles.popover)} portal={true}>
                    <ListboxList className={classNames(styles.list, 'dropdown-menu')}>
                        <Input
                            name="filter"
                            value={filter}
                            placeholder="Find dashboard..."
                            className="mx-1"
                            onChange={handleFilter}
                        />
                        {dashboards.filter(isVirtualDashboard).map(dashboard => (
                            <SelectOption
                                key={dashboard.id}
                                value={dashboard.id}
                                label={dashboard.title}
                                filter={filter}
                                className={styles.option}
                            />
                        ))}

                        {customDashboards.some(isPersonalDashboard) && (
                            <ListboxGroup>
                                <ListboxGroupLabel className={classNames(styles.groupLabel, 'text-muted')}>
                                    Private
                                </ListboxGroupLabel>

                                {customDashboards.filter(isPersonalDashboard).map(dashboard => (
                                    <SelectDashboardOption
                                        key={dashboard.id}
                                        dashboard={dashboard}
                                        filter={filter}
                                        className={styles.option}
                                    />
                                ))}
                            </ListboxGroup>
                        )}

                        {customDashboards.some(isGlobalDashboard) && (
                            <ListboxGroup>
                                <ListboxGroupLabel className={classNames(styles.groupLabel, 'text-muted')}>
                                    Global
                                </ListboxGroupLabel>

                                {customDashboards.filter(isGlobalDashboard).map(dashboard => (
                                    <SelectDashboardOption
                                        key={dashboard.id}
                                        dashboard={dashboard}
                                        filter={filter}
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
                                        filter={filter}
                                        className={styles.option}
                                    />
                                ))}
                            </ListboxGroup>
                        ))}

                        {!licensed && (
                            <ListboxGroup>
                                <hr />

                                <div className={classNames(styles.limitedAccess)}>
                                    <Typography.H3>Limited access</Typography.H3>
                                    <p>Unlock for unlimited custom dashboards.</p>
                                </div>
                            </ListboxGroup>
                        )}
                    </ListboxList>
                </ListboxPopover>
            </ListboxInput>
        </div>
    )
}

interface DashboardOrganizationGroup {
    id: string
    name: string
    dashboards: CustomInsightDashboard[]
}

/**
 * Returns organization dashboards grouped by dashboard owner id
 */
const getDashboardOrganizationsGroups = (dashboards: CustomInsightDashboard[]): DashboardOrganizationGroup[] => {
    const groupsDictionary = dashboards
        .filter(isOrganizationDashboard)
        .reduce<Record<string, DashboardOrganizationGroup>>((store, dashboard) => {
            for (const owner of dashboard.owners) {
                if (!store[owner.id]) {
                    store[owner.id] = {
                        id: owner.id,
                        name: owner.title,
                        dashboards: [],
                    }
                }

                store[owner.id].dashboards.push(dashboard)
            }

            return store
        }, {})

    return Object.values(groupsDictionary)
}
