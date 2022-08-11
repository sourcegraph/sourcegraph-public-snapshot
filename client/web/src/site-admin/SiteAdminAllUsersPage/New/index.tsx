import React, { useMemo, useEffect, useState, useCallback } from 'react'

/* TODO:
- Fix typos, linting, types, self-refactoring
- Figure out feature flagging
- Add link to billable events in bottom note
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
import { startCase } from 'lodash'
import { RouteComponentProps } from 'react-router'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { useMutation, useQuery } from '@sourcegraph/http-client'
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
    Alert,
} from '@sourcegraph/wildcard'

import { LineChart, Series } from '../../../charts'
import { CopyableText } from '../../../components/CopyableText'
import {
    SiteUsersLastActivePeriod,
    SiteUserOrderBy,
    UsersManagementChartResult,
    UsersManagementChartVariables,
    UsersManagementUsersListResult,
    UsersManagementUsersListVariables,
} from '../../../graphql-operations'
import { eventLogger } from '../../../tracking/eventLogger'
import { ChartContainer } from '../../analytics/components/ChartContainer'
import { HorizontalSelect } from '../../analytics/components/HorizontalSelect'
import { ToggleSelect } from '../../analytics/components/ToggleSelect'
import { ValueLegendList, ValueLegendListProps } from '../../analytics/components/ValueLegendList'
import { useChartFilters } from '../../analytics/useChartFilters'
import { StandardDatum } from '../../analytics/utils'
import { randomizeUserPassword, setUserIsSiteAdmin } from '../../backend'

import { Table } from './components/Table'
import {
    DELETE_USERS,
    DELETE_USERS_FOREVER,
    FORCE_SIGN_OUT_USERS,
    USERS_MANAGEMENT_CHART,
    USERS_MANAGEMENT_USERS_LIST,
} from './queries'

import styles from './index.module.scss'

type SiteUser = UsersManagementUsersListResult['site']['users']['nodes'][0]

export const UsersManagement: React.FunctionComponent<RouteComponentProps<{}>> = () => {
    useEffect(() => {
        eventLogger.logPageView('UsersManagement')
    }, [])
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
                <ChartSection />
                <TableSection />
            </Card>
            <Text className="font-italic text-center mt-2">
                All events are generated from entries in the event logs table and are updated every 24 hours..
            </Text>
        </>
    )
}

const ChartSection: React.FunctionComponent = () => {
    const { dateRange, aggregation, grouping } = useChartFilters({ name: 'Users', aggregation: 'uniqueUsers' })

    const { data, error, loading } = useQuery<UsersManagementChartResult, UsersManagementChartVariables>(
        USERS_MANAGEMENT_CHART,
        {
            variables: {
                dateRange: dateRange.value,
                grouping: grouping.value,
            },
        }
    )

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

    if (error) {
        throw error
    }

    if (loading || !data) {
        return <LoadingSpinner />
    }

    const groupingLabel = startCase(grouping.value.toLowerCase())

    return (
        <>
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
        </>
    )
}

const TableSection: React.FunctionComponent = () => {
    const [searchTxt, setSearchTxt] = useState<string>('')
    const [lastActivePeriod, setLastActivePeriod] = useState<SiteUsersLastActivePeriod>(SiteUsersLastActivePeriod.ALL)
    const [limit, setLimit] = useState(8) // TODO: fix to 25 before merging
    const [sortBy, setSortBy] = useState({ key: SiteUserOrderBy.EVENTS_COUNT, descending: false })

    const showMore = useCallback(() => setLimit(count => count * 2), [])

    const { data, previousData, refetch, variables, error, loading } = useQuery<
        UsersManagementUsersListResult,
        UsersManagementUsersListVariables
    >(USERS_MANAGEMENT_USERS_LIST, {
        variables: {
            first: limit,
            query: searchTxt || null,
            lastActivePeriod,
            orderBy: sortBy.key,
            descending: sortBy.descending,
        },
    })

    const refresh = useCallback(() => {
        refetch(variables).catch(console.error)
    }, [refetch, variables])
    const { handleDeleteUsers } = useDeleteUsers(refresh)
    const { handleDeleteUsersForever } = useDeleteUsersForever(refresh)
    const { handleForceSignOutUsers } = useForceSignOutUsers(refresh)
    const { handleRevokeSiteAdmin } = useRevokeSiteAdmin(refresh)
    const { handlePromoteToSiteAdmin } = usePromoteToSiteAdmin(refresh)
    const { handleResetUserPassword, data: resetPasswordData } = useResetUserPassword()

    const users = (data || previousData)?.site.users
    return (
        <div className="position-relative">
            <div className="mb-2 mt-4 pt-4 d-flex justify-content-between align-items-center text-nowrap">
                <H2>Users</H2>
                <div className="d-flex">
                    <HorizontalSelect<SiteUsersLastActivePeriod>
                        className="mr-4"
                        value={lastActivePeriod}
                        label="Last active"
                        onChange={value => {
                            setLastActivePeriod(value)
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
                        value={searchTxt}
                        onChange={event => setSearchTxt(event.target.value)}
                    />
                </div>
            </div>
            {resetPasswordData && (
                <>
                    {resetPasswordData.resetPasswordURL && (
                        <Alert className="mt-2" variant="success">
                            <Text>
                                Password was reset. You must manually send{' '}
                                <strong>{resetPasswordData?.user.username}</strong> this reset link:
                            </Text>
                            <CopyableText text={resetPasswordData.resetPasswordURL} size={40} />
                        </Alert>
                    )}
                    {resetPasswordData.resetPasswordURL === null && (
                        <Alert className="mt-2" variant="success">
                            Password was reset. The reset link was sent to the primary email of the user:
                        </Alert>
                    )}
                </>
            )}
            {loading && (
                <div
                    className={classNames(
                        'position-absolute w-100 h-100 d-flex justify-content-center align-items-center',
                        styles.loadingSpinnerContainer
                    )}
                >
                    <LoadingSpinner />
                </div>
            )}
            {error && <ErrorAlert error={error} />}
            {users && (
                <>
                    <Table
                        selectable={true}
                        sortBy={sortBy}
                        data={users.nodes}
                        onSortByChange={setSortBy}
                        getRowId={({ id }) => id}
                        actions={[
                            {
                                key: 'force-sign-out',
                                label: 'Force sign-out',
                                icon: mdiLogoutVariant,
                                onClick: handleForceSignOutUsers,
                            },
                            {
                                key: 'delete',
                                label: 'Delete',
                                icon: mdiArchive,
                                iconColor: 'danger',
                                onClick: handleDeleteUsers,
                            },
                            {
                                key: 'delete',
                                label: 'Delete forever',
                                icon: mdiDelete,
                                iconColor: 'danger',
                                labelColor: 'danger',
                                onClick: handleDeleteUsersForever,
                            },
                        ]}
                        columns={[
                            {
                                key: SiteUserOrderBy.USERNAME,
                                accessor: 'username',
                                header: 'User',
                                sortable: true,
                                render: function RenderUsernameAndEmail({ username, email }: SiteUser): JSX.Element {
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
                                render: function RenderActions(user: SiteUser): JSX.Element {
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
                                                        {!user.deletedAt && (
                                                            <>
                                                                <Button
                                                                    className="d-flex cursor-pointer"
                                                                    variant="link"
                                                                    as="li"
                                                                    outline={true}
                                                                    onClick={() => handleForceSignOutUsers([user])}
                                                                >
                                                                    <Icon
                                                                        svgPath={mdiLogoutVariant}
                                                                        aria-label="Force sign-out"
                                                                        size="md"
                                                                        className="text-muted"
                                                                    />
                                                                    <span className="ml-2">Force sign-out</span>
                                                                </Button>
                                                                {window.context.resetPasswordEnabled && (
                                                                    <Button
                                                                        className="d-flex cursor-pointer"
                                                                        variant="link"
                                                                        as="li"
                                                                        outline={true}
                                                                        onClick={() => handleResetUserPassword(user)}
                                                                    >
                                                                        <Icon
                                                                            svgPath={mdiLockReset}
                                                                            aria-label="Reset password"
                                                                            size="md"
                                                                            className="text-muted"
                                                                        />
                                                                        <span className="ml-2">Reset password</span>
                                                                    </Button>
                                                                )}
                                                                {user.siteAdmin ? (
                                                                    <Button
                                                                        className="d-flex cursor-pointer"
                                                                        variant="link"
                                                                        as="li"
                                                                        outline={true}
                                                                        onClick={() => handleRevokeSiteAdmin(user)}
                                                                    >
                                                                        <Icon
                                                                            svgPath={mdiClipboardMinus}
                                                                            aria-label="Revoke site admin"
                                                                            size="md"
                                                                            className="text-muted"
                                                                        />
                                                                        <span className="ml-2">Revoke site admin</span>
                                                                    </Button>
                                                                ) : (
                                                                    <Button
                                                                        className="d-flex cursor-pointer"
                                                                        variant="link"
                                                                        as="li"
                                                                        outline={true}
                                                                        onClick={() => handlePromoteToSiteAdmin(user)}
                                                                    >
                                                                        <Icon
                                                                            svgPath={mdiClipboardPlus}
                                                                            aria-label="Promote to site admin"
                                                                            size="md"
                                                                            className="text-muted"
                                                                        />
                                                                        <span className="ml-2">
                                                                            Promote to site admin
                                                                        </span>
                                                                    </Button>
                                                                )}
                                                                <Button
                                                                    className="d-flex cursor-pointer"
                                                                    variant="link"
                                                                    as="li"
                                                                    outline={true}
                                                                    onClick={() => handleDeleteUsers([user])}
                                                                >
                                                                    <Icon
                                                                        svgPath={mdiArchive}
                                                                        aria-label="Delete user"
                                                                        size="md"
                                                                        className="text-danger"
                                                                    />
                                                                    <span className="ml-2">Delete</span>
                                                                </Button>
                                                            </>
                                                        )}
                                                        <Button
                                                            className="d-flex cursor-pointer"
                                                            variant="link"
                                                            as="li"
                                                            outline={true}
                                                            onClick={() => handleDeleteUsersForever([user])}
                                                        >
                                                            <Icon
                                                                svgPath={mdiDelete}
                                                                aria-label="Delete user forever"
                                                                size="md"
                                                                className="text-danger"
                                                            />
                                                            <span className="ml-2 text-danger">Delete forever</span>
                                                        </Button>
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
                        note={
                            <Text as="span">
                                Note: Events is the count of all billable events which equate to billable usage.
                            </Text>
                        }
                    />
                    <div className="d-flex justify-content-between text-muted mb-4">
                        <Text>
                            Showing {users.nodes.length} of {users.totalCount} users
                        </Text>
                        {users.nodes.length !== users.totalCount ? (
                            <Button variant="link" onClick={showMore}>
                                Show More
                            </Button>
                        ) : null}
                    </div>
                </>
            )}
        </div>
    )
}

function useForceSignOutUsers(refetch: () => void) {
    const [forceSignOutUsers, { loading, error }] = useMutation(FORCE_SIGN_OUT_USERS)

    const handleForceSignOutUsers = useCallback(
        (users: SiteUser[]) => {
            if (confirm('Are you sure you want to force sign out the selected user(s)?')) {
                forceSignOutUsers({ variables: { userIDs: users.map(u => u.id) } })
                    .then(refetch)
                    .catch(console.error)
            }
        },
        [forceSignOutUsers, refetch]
    )

    return {
        handleForceSignOutUsers,
        loading,
        error,
    }
}

function useDeleteUsers(refetch: () => void) {
    const [deleteUsers, { loading, error }] = useMutation(DELETE_USERS)

    const handleDeleteUsers = useCallback(
        (users: SiteUser[]) => {
            if (confirm('Are you sure you want to delete the selected user(s)?')) {
                deleteUsers({ variables: { userIDs: users.map(u => u.id) } })
                    .then(refetch)
                    .catch(console.error)
            }
        },
        [deleteUsers, refetch]
    )
    return {
        handleDeleteUsers,
        loading,
        error,
    }
}

function useDeleteUsersForever(refetch: () => void) {
    const [deleteUsersForever, { loading, error }] = useMutation(DELETE_USERS_FOREVER)
    const handleDeleteUsersForever = useCallback(
        (users: SiteUser[]) => {
            if (confirm('Are you sure you want to delete the selected user(s)?')) {
                deleteUsersForever({ variables: { userIDs: users.map(u => u.id) } })
                    .then(refetch)
                    .catch(console.error)
            }
        },
        [deleteUsersForever, refetch]
    )
    return {
        handleDeleteUsersForever,
        loading,
        error,
    }
}

function usePromoteToSiteAdmin(refetch: () => void) {
    const handlePromoteToSiteAdmin = useCallback((user: SiteUser) => {
        if (confirm('Are you sure you want to promote the selected user to site admin?')) {
            setUserIsSiteAdmin(user.id, true).toPromise().then(refetch).catch(console.error)
        }
    }, [])

    return {
        handlePromoteToSiteAdmin,
    }
}

function useRevokeSiteAdmin(refetch: () => void) {
    const handleRevokeSiteAdmin = useCallback(
        (user: SiteUser) => {
            if (confirm('Are you sure you want to revoke the selected user from site admin?')) {
                setUserIsSiteAdmin(user.id, false).toPromise().then(refetch).catch(console.error)
            }
        },
        [refetch]
    )

    return {
        handleRevokeSiteAdmin,
    }
}

function useResetUserPassword() {
    const [data, setData] = useState<{ user: SiteUser; resetPasswordURL: string | null }>()
    const [loading, setLoading] = useState(false)
    const handleResetUserPassword = useCallback((user: SiteUser) => {
        if (confirm('Are you sure you want to reset the selected user password?')) {
            console.log('Reset user password', user)
            setLoading(true)
            randomizeUserPassword(user.id)
                .toPromise()
                .then(({ resetPasswordURL }) => setData({ resetPasswordURL, user }))
                .catch(console.error)
                .finally(() => setLoading(false))
        }
    }, [])
    return {
        handleResetUserPassword,
        data,
        loading,
    }
}
