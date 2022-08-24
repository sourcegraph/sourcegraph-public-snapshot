import React, { useState, useCallback, useMemo } from 'react'

import {
    mdiLogoutVariant,
    mdiArchive,
    mdiDelete,
    mdiLockReset,
    mdiClipboardMinus,
    mdiClipboardPlus,
    mdiClose,
} from '@mdi/js'
import classNames from 'classnames'
import { formatDistanceToNowStrict, startOfDay, endOfDay } from 'date-fns'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { useMutation, useQuery } from '@sourcegraph/http-client'
import { H2, LoadingSpinner, Text, Button, Alert, useDebounce, Link, Tooltip, Icon } from '@sourcegraph/wildcard'

import { CopyableText } from '../../../../components/CopyableText'
import {
    SiteUserOrderBy,
    UsersManagementUsersListResult,
    UsersManagementUsersListVariables,
} from '../../../../graphql-operations'
import { useURLSyncedState } from '../../../../hooks'
import { randomizeUserPassword, setUserIsSiteAdmin } from '../../../backend'
import { DELETE_USERS, DELETE_USERS_FOREVER, FORCE_SIGN_OUT_USERS, USERS_MANAGEMENT_USERS_LIST } from '../queries'

import { Table } from './Table'

import styles from '../index.module.scss'

type SiteUser = UsersManagementUsersListResult['site']['users']['nodes'][0]

const LIMIT = 25
interface UsersListProps {
    onActionEnd?: () => void
}

function parseDateRangeQueryParameter(rawValue: string): [Date, Date] | null | undefined {
    if (!rawValue) {
        return
    }
    const value = JSON.parse(rawValue)

    if (value === null) {
        return null
    }

    if (Array.isArray(value) && value.length === 2) {
        return [new Date(value[0]), new Date(value[1])]
    }
    return null
}

const DEFAULT_FILTERS = {
    searchText: '',
    limit: LIMIT.toString(),
    orderBy: SiteUserOrderBy.EVENTS_COUNT,
    descending: 'false',
    isAdmin: '',
    eventsCount: '',
    lastActiveAt: '',
    createdAt: '',
    deletedAt: '',
}

export const UsersList: React.FunctionComponent<UsersListProps> = ({ onActionEnd }) => {
    const [filters, setFilters] = useURLSyncedState(DEFAULT_FILTERS)
    const debouncedSearchText = useDebounce(filters.searchText, 300)
    const showMore = useCallback(() => setFilters({ limit: (Number(filters.limit) + LIMIT).toString() }), [
        filters.limit,
        setFilters,
    ])

    const lastActiveAtFilter = useMemo(() => parseDateRangeQueryParameter(filters.lastActiveAt), [filters.lastActiveAt])
    const deletedAtFilter = useMemo(() => parseDateRangeQueryParameter(filters.deletedAt), [filters.deletedAt])
    const createdAtFilter = useMemo(() => parseDateRangeQueryParameter(filters.createdAt), [filters.createdAt])

    const { data, previousData, refetch, variables, error, loading } = useQuery<
        UsersManagementUsersListResult,
        UsersManagementUsersListVariables
    >(USERS_MANAGEMENT_USERS_LIST, {
        variables: {
            first: Number(filters.limit),
            query: debouncedSearchText || null,
            lastActiveAt:
                lastActiveAtFilter === undefined
                    ? null
                    : {
                          isNull: lastActiveAtFilter === null,
                          after: lastActiveAtFilter === null ? null : startOfDay(lastActiveAtFilter[0]),
                          before: lastActiveAtFilter === null ? null : endOfDay(lastActiveAtFilter[1]),
                      },
            deletedAt:
                deletedAtFilter === undefined
                    ? null
                    : {
                          isNull: deletedAtFilter === null,
                          after: deletedAtFilter === null ? null : startOfDay(deletedAtFilter[0]),
                          before: deletedAtFilter === null ? null : endOfDay(deletedAtFilter[1]),
                      },
            createdAt: !createdAtFilter
                ? null
                : {
                      after: startOfDay(createdAtFilter[0]),
                      before: endOfDay(createdAtFilter[1]),
                  },
            eventsCount: !filters.eventsCount
                ? null
                : {
                      max: Number(filters.eventsCount),
                  },
            orderBy: filters.orderBy,
            descending: filters.descending ? (JSON.parse(filters.descending) as boolean) : null,
            siteAdmin: filters.isAdmin ? (JSON.parse(filters.isAdmin) as boolean) : null,
        },
    })

    const handleActionEnd = useCallback(
        (error?: any) => {
            if (!error) {
                // reload data
                refetch(variables).catch(console.error)
                onActionEnd?.()
            }
        },
        [onActionEnd, refetch, variables]
    )

    const {
        handleDeleteUsers,
        handleDeleteUsersForever,
        handleForceSignOutUsers,
        handleRevokeSiteAdmin,
        handlePromoteToSiteAdmin,
        handleResetUserPassword,
        notification,
        handleDismissNotification,
    } = useUserListActions(handleActionEnd)

    const users = (data || previousData)?.site.users
    return (
        <div className="position-relative">
            <div className="mb-4 mt-4 pt-4 d-flex justify-content-between align-items-center text-nowrap">
                <H2>Users</H2>
                <Button size="sm" onClick={() => setFilters(DEFAULT_FILTERS)} outline={true} variant="secondary">
                    Clear all filters
                    <Icon aria-hidden={true} className="ml-1" svgPath={mdiClose} />
                </Button>
            </div>

            {notification && (
                <Alert
                    className="mt-2 d-flex justify-content-between align-items-center"
                    variant={notification.isError ? 'danger' : 'success'}
                >
                    {notification.text}
                    <Button variant="secondary" outline={true} onClick={handleDismissNotification}>
                        <Icon aria-label="Close notification" svgPath={mdiClose} />
                    </Button>
                </Alert>
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
                        sortBy={{ key: filters.orderBy, descending: JSON.parse(filters.descending) as boolean }}
                        data={users.nodes}
                        onSortByChange={value => {
                            setFilters({
                                orderBy: value.key as SiteUserOrderBy,
                                descending: value.descending.toString(),
                            })
                        }}
                        getRowId={({ id }) => id}
                        actions={[
                            {
                                key: 'force-sign-out',
                                label: 'Force sign-out',
                                icon: mdiLogoutVariant,
                                onClick: handleForceSignOutUsers,
                                bulk: true,
                                condition: users => users.some(user => !user.deletedAt),
                            },
                            {
                                key: 'reset-password',
                                label: 'Reset password',
                                icon: mdiLockReset,
                                onClick: handleResetUserPassword,
                                condition: ([user]) => !user?.deletedAt,
                            },
                            {
                                key: 'revoke-site-admin',
                                label: 'Revoke site admin',
                                icon: mdiClipboardMinus,
                                onClick: handleRevokeSiteAdmin,
                                condition: ([user]) => user?.siteAdmin && !user?.deletedAt,
                            },
                            {
                                key: 'promote-to-site-admin',
                                label: 'Promote to site admin',
                                icon: mdiClipboardPlus,
                                onClick: handlePromoteToSiteAdmin,
                                condition: ([user]) => !user?.siteAdmin && !user?.deletedAt,
                            },
                            {
                                key: 'delete',
                                label: 'Delete',
                                icon: mdiArchive,
                                iconColor: 'danger',
                                onClick: handleDeleteUsers,
                                bulk: true,
                                condition: users => users.some(user => !user.deletedAt),
                            },
                            {
                                key: 'delete-forever',
                                label: 'Delete forever',
                                icon: mdiDelete,
                                iconColor: 'danger',
                                labelColor: 'danger',
                                onClick: handleDeleteUsersForever,
                                bulk: true,
                            },
                        ]}
                        columns={[
                            {
                                key: SiteUserOrderBy.USERNAME,
                                accessor: 'username',
                                header: 'User',
                                sortable: true,
                                render: RenderUsernameAndEmail,
                                filter: {
                                    type: 'text',
                                    placeholder: 'Username, email, display name',
                                    onChange: value => {
                                        setFilters({ searchText: value })
                                    },
                                    value: filters.searchText,
                                },
                            },
                            {
                                key: SiteUserOrderBy.SITE_ADMIN,
                                accessor: item => (item.siteAdmin ? 'Yes' : 'No'),
                                header: { label: 'Is Admin', align: 'left' },
                                sortable: true,
                                align: 'center',
                                filter: {
                                    type: 'select',
                                    options: [
                                        { value: 'null', label: 'All' },
                                        { value: 'true', label: 'Yes' },
                                        { value: 'false', label: 'No' },
                                    ],
                                    onChange: value => {
                                        setFilters({ isAdmin: value })
                                    },
                                    value: filters.isAdmin,
                                },
                            },
                            {
                                key: SiteUserOrderBy.EVENTS_COUNT,
                                accessor: 'eventsCount',
                                header: {
                                    label: 'Events',
                                    align: 'left',
                                    tooltip:
                                        '"Events" count is cached and updated every 12 hours. It is based on event logs table and available for the last 93 days.',
                                },
                                sortable: true,
                                align: 'right',
                                filter: {
                                    type: 'select',
                                    options: [
                                        { value: '', label: 'All' },
                                        { value: '0', label: '= 0' },
                                        { value: '10', label: '< 10' },
                                        { value: '100', label: '< 100' },
                                        { value: '1000', label: '< 1000' },
                                    ],
                                    onChange: value => {
                                        setFilters({ eventsCount: value })
                                    },
                                    value: filters.eventsCount,
                                },
                            },
                            {
                                key: SiteUserOrderBy.LAST_ACTIVE_AT,
                                accessor: item =>
                                    item.lastActiveAt
                                        ? formatDistanceToNowStrict(new Date(item.lastActiveAt), { addSuffix: true })
                                        : '',
                                header: {
                                    label: 'Last Active',
                                    align: 'left',
                                    tooltip:
                                        '"Last Active" is cached and updated every 12 hours. It is based on event logs table and available for the last 93 days.',
                                },
                                sortable: true,
                                align: 'right',
                                filter: {
                                    type: 'date-range',
                                    placeholder: 'Select',
                                    nullLabel: 'Never',
                                    onChange: value => {
                                        setFilters({ lastActiveAt: JSON.stringify(value) })
                                    },
                                    value: lastActiveAtFilter,
                                },
                            },
                            {
                                key: SiteUserOrderBy.CREATED_AT,
                                accessor: item =>
                                    formatDistanceToNowStrict(new Date(item.createdAt), { addSuffix: true }),
                                header: { label: 'Created', align: 'left' },
                                sortable: true,
                                align: 'right',
                                filter: {
                                    type: 'date-range',
                                    placeholder: 'Select',
                                    onChange: value => {
                                        setFilters({ createdAt: JSON.stringify(value) })
                                    },
                                    value: parseDateRangeQueryParameter(filters.createdAt),
                                },
                            },
                            {
                                key: SiteUserOrderBy.DELETED_AT,
                                accessor: item =>
                                    item.deletedAt
                                        ? formatDistanceToNowStrict(new Date(item.deletedAt), { addSuffix: true })
                                        : '',
                                header: { label: 'Deleted', align: 'left' },
                                sortable: true,
                                align: 'right',
                                filter: {
                                    type: 'date-range',
                                    placeholder: 'Select',
                                    nullLabel: 'Not Deleted',
                                    onChange: value => {
                                        setFilters({ deletedAt: JSON.stringify(value) })
                                    },
                                    value: deletedAtFilter,
                                },
                            },
                        ]}
                        note={
                            <Text as="span">
                                {/* TODO: Fix link */}
                                Note: Events is the count of <Link to="#">all billable events</Link> which equate to
                                billable usage.
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

function RenderUsernameAndEmail({ username, email, displayName, deletedAt }: SiteUser): JSX.Element {
    return (
        <Tooltip content={username}>
            <div className={classNames('d-flex flex-column p-2', styles.usernameColumn)}>
                {!deletedAt ? (
                    <Link to={`/users/${username}`} className="text-truncate">
                        @{username}
                    </Link>
                ) : (
                    <Text className="mb-0 text-truncate">@{username}</Text>
                )}
                <Text className="mb-0 text-truncate">{displayName}</Text>
                <Text className="mb-0 text-truncate">{email}</Text>
            </div>
        </Tooltip>
    )
}

type ActionHandler = (users: SiteUser[]) => void

interface UseUserListActionReturnType {
    handleForceSignOutUsers: ActionHandler
    handleDeleteUsers: ActionHandler
    handleDeleteUsersForever: ActionHandler
    handlePromoteToSiteAdmin: ActionHandler
    handleRevokeSiteAdmin: ActionHandler
    notification: { text: React.ReactNode; isError?: boolean } | undefined
    handleDismissNotification: () => void
    handleResetUserPassword: ActionHandler
}

const getUsernames = (users: SiteUser[]): string => users.map(user => user.username).join(', ')

function useUserListActions(onEnd: (error?: any) => void): UseUserListActionReturnType {
    const [forceSignOutUsers] = useMutation(FORCE_SIGN_OUT_USERS)
    const [deleteUsers] = useMutation(DELETE_USERS)
    const [deleteUsersForever] = useMutation(DELETE_USERS_FOREVER)

    const [notification, setNotification] = useState<UseUserListActionReturnType['notification']>()

    const handleDismissNotification = useCallback(() => setNotification(undefined), [])

    const onError = useCallback(
        (error: any) => {
            setNotification({
                text: (
                    <Text as="span">
                        Something went wrong :(!
                        <Text as="pre" className="m-1" size="small">
                            {error?.message}
                        </Text>
                    </Text>
                ),
                isError: true,
            })
            console.error(error)
            onEnd(error)
        },
        [onEnd]
    )

    const createOnSuccess = useCallback(
        (text: React.ReactNode, shouldReload = false) => () => {
            setNotification({ text })
            if (shouldReload) {
                onEnd()
            }
        },
        [onEnd]
    )

    const handleForceSignOutUsers = useCallback(
        (users: SiteUser[]) => {
            if (confirm('Are you sure you want to force sign out the selected user(s)?')) {
                forceSignOutUsers({ variables: { userIDs: users.map(user => user.id) } })
                    .then(
                        createOnSuccess(
                            <Text as="span">
                                Successfully force signed out following {users.length} user(s):{' '}
                                <strong>{getUsernames(users)}</strong>
                            </Text>
                        )
                    )
                    .catch(onError)
            }
        },
        [forceSignOutUsers, onError, createOnSuccess]
    )

    const handleDeleteUsers = useCallback(
        (users: SiteUser[]) => {
            if (confirm('Are you sure you want to delete the selected user(s)?')) {
                deleteUsers({ variables: { userIDs: users.map(user => user.id) } })
                    .then(
                        createOnSuccess(
                            <Text as="span">
                                Successfully deleted following {users.length} user(s):{' '}
                                <strong>{getUsernames(users)}</strong>
                            </Text>,
                            true
                        )
                    )
                    .catch(onError)
            }
        },
        [deleteUsers, onError, createOnSuccess]
    )
    const handleDeleteUsersForever = useCallback(
        (users: SiteUser[]) => {
            if (confirm('Are you sure you want to delete the selected user(s)?')) {
                deleteUsersForever({ variables: { userIDs: users.map(user => user.id) } })
                    .then(
                        createOnSuccess(
                            <Text as="span">
                                Successfully deleted forever following {users.length} user(s):{' '}
                                <strong>{getUsernames(users)}</strong>
                            </Text>,
                            true
                        )
                    )
                    .catch(onError)
            }
        },
        [deleteUsersForever, onError, createOnSuccess]
    )

    const handlePromoteToSiteAdmin = useCallback(
        ([user]: SiteUser[]) => {
            if (confirm('Are you sure you want to promote the selected user to site admin?')) {
                setUserIsSiteAdmin(user.id, true)
                    .toPromise()
                    .then(
                        createOnSuccess(
                            <Text as="span">
                                Successfully promoted user <strong>{user.username}</strong> to site admin.
                            </Text>,
                            true
                        )
                    )
                    .catch(onError)
            }
        },
        [onError, createOnSuccess]
    )

    const handleRevokeSiteAdmin = useCallback(
        ([user]: SiteUser[]) => {
            if (confirm('Are you sure you want to revoke the selected user from site admin?')) {
                setUserIsSiteAdmin(user.id, false)
                    .toPromise()
                    .then(
                        createOnSuccess(
                            <Text as="span">
                                Successfully revoked site admin from <strong>{user.username}</strong> user.
                            </Text>,
                            true
                        )
                    )
                    .catch(onError)
            }
        },
        [onError, createOnSuccess]
    )

    const handleResetUserPassword = useCallback(
        ([user]: SiteUser[]) => {
            if (confirm('Are you sure you want to reset the selected user password?')) {
                randomizeUserPassword(user.id)
                    .toPromise()
                    .then(({ resetPasswordURL }) => {
                        if (resetPasswordURL === null) {
                            createOnSuccess(
                                <Text as="span">
                                    Password was reset. The reset link was sent to the primary email of the user:{' '}
                                    <strong>{user.username}</strong>
                                </Text>
                            )()
                        } else {
                            createOnSuccess(
                                <>
                                    <Text>
                                        Password was reset. You must manually send <strong>{user.username}</strong> this
                                        reset link:
                                    </Text>
                                    <CopyableText text={resetPasswordURL} size={40} />
                                </>
                            )()
                        }
                    })
                    .catch(onError)
            }
        },
        [onError, createOnSuccess]
    )

    return {
        notification,
        handleForceSignOutUsers,
        handleDeleteUsers,
        handleDeleteUsersForever,
        handlePromoteToSiteAdmin,
        handleRevokeSiteAdmin,
        handleResetUserPassword,
        handleDismissNotification,
    }
}
