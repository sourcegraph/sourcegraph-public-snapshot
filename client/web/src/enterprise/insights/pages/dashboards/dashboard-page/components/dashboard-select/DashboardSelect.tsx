import {
    ListboxGroup,
    ListboxGroupLabel,
    ListboxInput,
    ListboxList,
    ListboxOption,
    ListboxPopover,
} from '@reach/listbox'
import { VisuallyHidden } from '@reach/visually-hidden'
import classNames from 'classnames'
import React, { useContext, useMemo } from 'react'

import { AuthenticatedUser } from '@sourcegraph/web/src/auth'

import { CodeInsightsBackendContext } from '../../../../../core/backend/code-insights-backend-context'
import {
    InsightDashboard,
    InsightDashboardOwner,
    isGlobalDashboard,
    isOrganizationDashboard,
    isPersonalDashboard,
    isRealDashboard,
    isVirtualDashboard,
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
    user?: AuthenticatedUser | null
}

/**
 * Renders dashboard select component for the code insights dashboard page selection UI.
 */
export const DashboardSelect: React.FunctionComponent<DashboardSelectProps> = props => {
    const { value, dashboards, onSelect, className, user } = props
    const { getUiFeatures } = useContext(CodeInsightsBackendContext)

    const features = useMemo(() => getUiFeatures(), [getUiFeatures])

    const licensed = features.licensed

    if (!user) {
        return null
    }

    const handleChange = (value: string): void => {
        const dashboard = dashboards.find(dashboard => dashboard.id === value)

        if (dashboard) {
            onSelect(dashboard)
        }
    }

    const realDashboards = dashboards.filter(isRealDashboard)
    const organizationGroups = getDashboardOrganizationsGroups(realDashboards, user.organizations.nodes)

    return (
        <div className={className}>
            <VisuallyHidden id={LABEL_ID}>Choose a dashboard</VisuallyHidden>

            <ListboxInput aria-labelledby={LABEL_ID} value={value ?? 'unknown'} onChange={handleChange}>
                <MenuButton dashboards={dashboards} />

                <ListboxPopover className={classNames(styles.popover)} portal={true}>
                    <ListboxList className={classNames(styles.list, 'dropdown-menu')}>
                        {dashboards.filter(isVirtualDashboard).map(dashboard => (
                            <SelectOption
                                key={dashboard.id}
                                value={dashboard.id}
                                label={dashboard.title}
                                className={styles.option}
                            />
                        ))}

                        {realDashboards.some(isPersonalDashboard) && (
                            <ListboxGroup>
                                <ListboxGroupLabel className={classNames(styles.groupLabel, 'text-muted')}>
                                    Private
                                </ListboxGroupLabel>

                                {realDashboards.filter(isPersonalDashboard).map(dashboard => (
                                    <SelectDashboardOption
                                        key={dashboard.id}
                                        dashboard={dashboard}
                                        className={styles.option}
                                    />
                                ))}
                            </ListboxGroup>
                        )}

                        {realDashboards.some(isGlobalDashboard) && (
                            <ListboxGroup>
                                <ListboxGroupLabel className={classNames(styles.groupLabel, 'text-muted')}>
                                    Global
                                </ListboxGroupLabel>

                                {realDashboards.filter(isGlobalDashboard).map(dashboard => (
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

                        {!licensed && (
                            <ListboxGroup>
                                <ListboxOption
                                    className={classNames(styles.option, styles.limitedAccessWrapper)}
                                    value="na"
                                >
                                    <div className={classNames(styles.limitedAccess)}>
                                        <h3>Limited access</h3>
                                        <p>Unlock for unlimited dashboards custom dashboards.</p>
                                    </div>
                                </ListboxOption>
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
    dashboards: RealInsightDashboard[]
}

/**
 * Returns organization dashboards grouped by dashboard owner id
 */
const getDashboardOrganizationsGroups = (
    dashboards: RealInsightDashboard[],
    organizations: AuthenticatedUser['organizations']['nodes']
): DashboardOrganizationGroup[] => {
    // We need a map of the organization names when using the new GraphQL API
    const organizationsMap = organizations.reduce<Record<string, InsightDashboardOwner>>(
        (map, organization) => ({
            ...map,
            [organization.id]: {
                id: organization.id,
                name: organization.displayName ?? organization.name,
            },
        }),
        {}
    )

    const groupsDictionary = dashboards
        .map(dashboard => {
            const owner =
                ('owner' in dashboard && dashboard.owner) ||
                ('grants' in dashboard &&
                    dashboard.grants?.organizations &&
                    organizationsMap[dashboard.grants?.organizations[0]])
            // Grabbing the first organization to minimize changes with existing api
            // TODO: handle multiple organizations when settings API is deprecated

            if (!owner) {
                return dashboard
            }

            return {
                ...dashboard,
                owner,
            }
        })
        .filter(isOrganizationDashboard)
        .reduce<Record<string, DashboardOrganizationGroup>>((store, dashboard) => {
            if (!dashboard.owner) {
                // TODO: remove this check after settings api is deprecated
                throw new Error('`owner` is missing from the dashboard')
            }

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
