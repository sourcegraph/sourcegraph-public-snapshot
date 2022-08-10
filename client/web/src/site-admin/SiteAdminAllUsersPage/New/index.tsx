import React, { useMemo, useEffect, useState } from 'react'

/* TODO:
- Add link to billable events in bottom note

- Integrate actions with APIs
- Fix typos, linting, types, self-refactoring
- Figure out feature flagging
*/

import {
    mdiAccount,
    mdiPlus,
    mdiDownload,
    mdiLogoutVariant,
    mdiArchive,
    mdiDelete,
    mdiDotsHorizontal,
    mdiLockReset,
    mdiClipboardMinus,
    mdiClipboardPlus,
} from '@mdi/js'
import classNames from 'classnames'
import { format as formatDate } from 'date-fns'
import { startCase, isEqual } from 'lodash'
import { RouteComponentProps } from 'react-router'

import { useQuery } from '@sourcegraph/http-client'
import {
    H1,
    H2,
    Card,
    LoadingSpinner,
    Text,
    Icon,
    Position,
    PopoverTrigger,
    PopoverContent,
    Popover,
    Input,
    Button,
    Link,
} from '@sourcegraph/wildcard'

import { LineChart, Series } from '../../../charts'
import {
    UsersManagementResult,
    UsersManagementVariables,
    SiteUsersLastActivePeriod,
    SiteUserOrderBy,
    AnalyticsDateRange,
    AnalyticsGrouping,
} from '../../../graphql-operations'
import { eventLogger } from '../../../tracking/eventLogger'
import { ChartContainer } from '../../analytics/components/ChartContainer'
import { HorizontalSelect } from '../../analytics/components/HorizontalSelect'
import { ToggleSelect } from '../../analytics/components/ToggleSelect'
import { ValueLegendList, ValueLegendListProps } from '../../analytics/components/ValueLegendList'
import { useChartFilters } from '../../analytics/useChartFilters'
import { StandardDatum } from '../../analytics/utils'

import { Table } from './components/Table'
import { USERS_MANAGEMENT } from './queries'

import styles from './index.module.scss'

export const UsersManagement: React.FunctionComponent<RouteComponentProps<{}>> = () => {
    const { data, previousData, error, loading, variables, refetch, called } = useQuery<
        UsersManagementResult,
        UsersManagementVariables
    >(USERS_MANAGEMENT, {
        variables: {
            first: 3,
            dateRange: AnalyticsDateRange.LAST_THREE_MONTHS,
            grouping: AnalyticsGrouping.WEEKLY,
            usersQuery: null,
            usersLastActivePeriod: SiteUsersLastActivePeriod.ALL,
            usersOrderBy: SiteUserOrderBy.EVENTS_COUNT,
            usersOrderDescending: false,
        },
    })

    if (error) {
        throw error
    }

    if ((loading && !called) || !(data || previousData)) {
        return <LoadingSpinner />
    }

    return <Content data={data || previousData} variables={variables} refetch={refetch} />
}

interface ContentProps {
    data: UsersManagementResult
    variables: UsersManagementVariables
    refetch: (variables: UsersManagementVariables) => any
}

const Content: React.FunctionComponent<ContentProps> = ({ data, variables, refetch }) => {
    const { dateRange, aggregation, grouping } = useChartFilters({ name: 'Users', aggregation: 'uniqueUsers' })
    const [usersQuery, setUsersQuery] = useState<string>('')
    const [usersLastActivePeriod, setUsersLastActivePeriod] = useState<SiteUsersLastActivePeriod>(
        SiteUsersLastActivePeriod.ALL
    )
    const [first, setFirst] = useState(3)

    const showMore = () => setFirst(count => count * 2)

    useEffect(() => {
        eventLogger.logPageView('UsersManagement')
    }, [])

    useEffect(() => {
        const newVariables = {
            ...variables,
            first,
            dateRange: dateRange.value,
            grouping: grouping.value,
            usersQuery,
            usersLastActivePeriod,
        }
        if (!isEqual(variables, newVariables)) {
            refetch(newVariables)
        }
    }, [
        first,
        dateRange.value,
        aggregation.selected,
        grouping.value,
        usersQuery,
        usersLastActivePeriod,
        variables,
        refetch,
    ])

    const [activities, legends] = useMemo(() => {
        if (!data) {
            return []
        }
        const { users } = data.site.analytics
        const activities: Series<StandardDatum>[] = [
            {
                id: 'activity',
                name: aggregation.selected === 'count' ? 'Activities' : 'Active users',
                color: aggregation.selected === 'count' ? 'var(--cyan)' : 'var(--purple)',
                data: users.activity.nodes.map(
                    node => ({
                        date: new Date(node.date),
                        value: node[aggregation.selected],
                    }),
                    dateRange.value
                ),
                getXValue: ({ date }) => date,
                getYValue: ({ value }) => value,
            },
        ]

        const legends: ValueLegendListProps['items'] = [
            {
                value: users.activity.summary.totalUniqueUsers,
                description: 'Active users',
                color: 'var(--purple)',
                tooltip: 'Users using the application in the selected timeframe.',
            },
            {
                value: data.users.totalCount,
                description: 'Registered Users',
                color: 'var(--body-color)',
                position: 'right',
                tooltip: 'The number of users who have created an account.',
            },
            {
                value: data.site.productSubscription.license?.userCount ?? 0,
                description: 'User licenses',
                color: 'var(--body-color)',
                position: 'right',
                tooltip: 'The number of user licenses your current account is provisioned for.',
            },
            {
                value: data.site.adminUsers?.totalCount ?? 0,
                description: 'Administrators',
                color: 'var(--body-color)',
                position: 'right',
                tooltip: 'The number of users with site admin permissions.',
            },
        ]

        return [activities, legends]
    }, [data, aggregation.selected, dateRange.value])

    const groupingLabel = startCase(grouping.value.toLowerCase())

    return (
        <>
            <div className="d-flex justify-content-between align-items-center mb-4 mt-2">
                <H1 className="d-flex align-items-center mb-0">
                    <Icon
                        svgPath={mdiAccount}
                        aria-label="user administration avatar icon"
                        size="md"
                        className={styles.linkColor}
                    />{' '}
                    User administration
                </H1>
                <div>
                    <Button
                        href="/site-admin/usage-statistics/archive"
                        download="true"
                        className="mr-4"
                        variant="secondary"
                        outline={true}
                        as="a"
                    >
                        <Icon svgPath={mdiDownload} aria-label="Download usage stats" className="mr-1" />
                        Download usage stats
                    </Button>
                    <Button to="/site-admin/users/new" variant="primary" as={Link}>
                        <Icon svgPath={mdiPlus} aria-label="create user" className="mr-1" />
                        Create User
                    </Button>
                </div>
            </div>
            <Card className="p-3">
                <div className="d-flex justify-content-end align-items-stretch mb-2 text-nowrap">
                    <HorizontalSelect<typeof dateRange.value> {...dateRange} />
                </div>
                {legends && <ValueLegendList className="mb-3" items={legends} />}
                {activities && (
                    <div>
                        <ChartContainer
                            title={
                                aggregation.selected === 'count'
                                    ? `${groupingLabel} activity`
                                    : `${groupingLabel} unique users`
                            }
                            labelX="Time"
                            labelY={aggregation.selected === 'count' ? 'Activity' : 'Unique users'}
                        >
                            {width => <LineChart width={width} height={300} series={activities} />}
                        </ChartContainer>
                        <div className="d-flex justify-content-end align-items-stretch mb-4 text-nowrap">
                            <HorizontalSelect<typeof grouping.value> {...grouping} className="mr-4" />
                            <ToggleSelect<typeof aggregation.selected> {...aggregation} />
                        </div>
                    </div>
                )}
                <div className="mb-2 mt-4 pt-4 d-flex justify-content-between align-items-center text-nowrap">
                    <H2>Users</H2>
                    <div className="d-flex">
                        <HorizontalSelect<SiteUsersLastActivePeriod>
                            className="mr-4"
                            value={usersLastActivePeriod}
                            label="Last active"
                            onChange={value => {
                                setUsersLastActivePeriod(value)
                                eventLogger.log(`UserManagementLastActive${value}`)
                            }}
                            items={[
                                { value: SiteUsersLastActivePeriod.ALL, label: 'All' },
                                { value: SiteUsersLastActivePeriod.TODAY, label: 'Today' },
                                { value: SiteUsersLastActivePeriod.THIS_WEEK, label: 'This week' },
                                { value: SiteUsersLastActivePeriod.THIS_MONTH, label: 'This month' },
                            ]}
                        />
                        <Input
                            placeholder="Search username or name"
                            value={usersQuery}
                            onChange={event => setUsersQuery(event.target.value)}
                        />
                    </div>
                </div>
                <Table
                    selectable={true}
                    initialSortColumn={SiteUserOrderBy.EVENTS_COUNT}
                    initialSortDirection="asc"
                    onSortChange={(usersOrderBy, usersOrderDirection) =>
                        refetch({ ...variables, usersOrderBy, usersOrderDescending: usersOrderDirection === 'desc' })
                    }
                    getRowId={({ id }) => id}
                    actions={[
                        {
                            key: 'force-sign-out',
                            label: 'Force sign-out',
                            icon: mdiLogoutVariant,
                            onClick: () => '',
                        },
                        {
                            key: 'delete',
                            label: 'Delete',
                            icon: mdiArchive,
                            iconColor: 'danger',
                            onClick: () => '',
                        },
                        {
                            key: 'delete',
                            label: 'Delete forever',
                            icon: mdiDelete,
                            iconColor: 'danger',
                            labelColor: 'danger',
                            onClick: () => '',
                        },
                    ]}
                    columns={[
                        {
                            key: SiteUserOrderBy.USERNAME,
                            accessor: 'username',
                            header: 'User',
                            sortable: true,
                            render: function RenderUsernameAndEmail({
                                username,
                                email,
                            }: typeof data.site.users.nodes[0]): JSX.Element {
                                return (
                                    <div className="d-flex flex-column p-2">
                                        <Text className={classNames(styles.linkColor, 'mb-0')}>{username}</Text>
                                        <Text className="mb-0">{email}</Text>
                                    </div>
                                )
                            },
                        },
                        {
                            key: SiteUserOrderBy.SITE_ADMIN,
                            accessor: item => (item.siteAdmin ? 'Yes' : 'No'),
                            header: { label: 'Site Admin', align: 'right' },
                            sortable: true,
                            align: 'center',
                        },
                        {
                            key: SiteUserOrderBy.EVENTS_COUNT,
                            accessor: 'eventsCount',
                            header: {
                                label: 'Events',
                                align: 'right',
                                tooltip:
                                    '"Events" count is based on event_logs table which stores only last 93 days of logs.',
                            },
                            sortable: true,
                            align: 'right',
                        },
                        {
                            key: SiteUserOrderBy.LAST_ACTIVE_AT,
                            accessor: item =>
                                item.lastActiveAt ? formatDate(new Date(item.lastActiveAt), 'dd/mm/yyyy') : '',
                            header: {
                                label: 'Last Active',
                                align: 'right',
                                tooltip:
                                    '"Last Active" is based on event_logs table which stores only last 93 days of logs.',
                            },
                            sortable: true,
                            align: 'right',
                        },
                        {
                            key: SiteUserOrderBy.CREATED_AT,
                            accessor: item => formatDate(new Date(item.createdAt), 'dd/mm/yyyy'),
                            header: { label: 'Created', align: 'right' },
                            sortable: true,
                            align: 'right',
                        },
                        {
                            key: SiteUserOrderBy.DELETED_AT,
                            accessor: item =>
                                item.deletedAt ? formatDate(new Date(item.deletedAt), 'dd/mm/yyyy') : '',
                            header: { label: 'Deleted', align: 'right' },
                            sortable: true,
                            align: 'right',
                        },
                        {
                            key: 'actions',
                            render: function RenderActions({
                                siteAdmin,
                            }: typeof data.site.users.nodes[0]): JSX.Element {
                                return (
                                    <Popover>
                                        <div className="d-flex justify-content-center">
                                            <PopoverTrigger
                                                as={Icon}
                                                svgPath={mdiDotsHorizontal}
                                                className="cursor-pointer"
                                            />
                                            <PopoverContent position={Position.bottom}>
                                                <ul className="list-unstyled mb-0">
                                                    <li className="d-flex p-2 cursor-pointer">
                                                        <Icon
                                                            svgPath={mdiLogoutVariant}
                                                            aria-label="Force sign-out"
                                                            size="md"
                                                            className="text-muted"
                                                        />
                                                        <span className="ml-2">Force sign-out</span>
                                                    </li>
                                                    {window.context.resetPasswordEnabled && (
                                                        <li className="d-flex p-2 cursor-pointer">
                                                            <Icon
                                                                svgPath={mdiLockReset}
                                                                aria-label="Reset password"
                                                                size="md"
                                                                className="text-muted"
                                                            />
                                                            <span className="ml-2">Reset password</span>
                                                        </li>
                                                    )}
                                                    {siteAdmin ? (
                                                        <li className="d-flex p-2 cursor-pointer">
                                                            <Icon
                                                                svgPath={mdiClipboardMinus}
                                                                aria-label="Revoke site admin"
                                                                size="md"
                                                                className="text-muted"
                                                            />
                                                            <span className="ml-2">Revoke site admin</span>
                                                        </li>
                                                    ) : (
                                                        <li className="d-flex p-2 cursor-pointer">
                                                            <Icon
                                                                svgPath={mdiClipboardPlus}
                                                                aria-label="Promote to site admin"
                                                                size="md"
                                                                className="text-muted"
                                                            />
                                                            <span className="ml-2">Promote to site admin</span>
                                                        </li>
                                                    )}
                                                    <li className="d-flex p-2 cursor-pointer">
                                                        <Icon
                                                            svgPath={mdiArchive}
                                                            aria-label="Delete user"
                                                            size="md"
                                                            className="text-danger"
                                                        />
                                                        <span className="ml-2">Delete</span>
                                                    </li>
                                                    <li className="d-flex p-2 cursor-pointer">
                                                        <Icon
                                                            svgPath={mdiDelete}
                                                            aria-label="Delete user forever"
                                                            size="md"
                                                            className="text-danger"
                                                        />
                                                        <span className="ml-2 text-danger">Delete forever</span>
                                                    </li>
                                                </ul>
                                            </PopoverContent>
                                        </div>
                                    </Popover>
                                )
                            },
                            header: { label: 'Actions', align: 'right' },
                            align: 'center',
                        },
                    ]}
                    data={data.site.users.nodes}
                    note={
                        <Text as="span">
                            Note: Events is the count of all billable events which equate to billable usage.
                        </Text>
                    }
                />
                <div className="d-flex justify-content-between text-muted mb-4">
                    <Text>
                        Showing {data.site.users.nodes.length} of {data.site.users.totalCount} users
                    </Text>
                    {data.site.users.nodes.length !== data.site.users.totalCount ? (
                        <Button variant="link" onClick={showMore}>
                            Show More
                        </Button>
                    ) : null}
                </div>
            </Card>
            <Text className="font-italic text-center mt-2">
                All events are generated from entries in the event logs table and are updated every 24 hours..
            </Text>
        </>
    )
}
