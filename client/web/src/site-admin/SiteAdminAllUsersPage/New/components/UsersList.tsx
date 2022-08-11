import React, { useState, useCallback } from 'react'

import { mdiLogoutVariant, mdiArchive, mdiDelete, mdiLockReset, mdiClipboardMinus, mdiClipboardPlus } from '@mdi/js'
import classNames from 'classnames'
import { format as formatDate } from 'date-fns'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { useMutation, useQuery } from '@sourcegraph/http-client'
import { H2, LoadingSpinner, Text, Input, Button, Alert } from '@sourcegraph/wildcard'

import { CopyableText } from '../../../../components/CopyableText'
import {
    SiteUsersLastActivePeriod,
    SiteUserOrderBy,
    UsersManagementUsersListResult,
    UsersManagementUsersListVariables,
} from '../../../../graphql-operations'
import { eventLogger } from '../../../../tracking/eventLogger'
import { HorizontalSelect } from '../../../analytics/components/HorizontalSelect'
import { randomizeUserPassword, setUserIsSiteAdmin } from '../../../backend'
import { DELETE_USERS, DELETE_USERS_FOREVER, FORCE_SIGN_OUT_USERS, USERS_MANAGEMENT_USERS_LIST } from '../queries'

import { Table } from './Table'

import styles from '../index.module.scss'

type SiteUser = UsersManagementUsersListResult['site']['users']['nodes'][0]

export const UsersList: React.FunctionComponent = () => {
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

    const {
        handleDeleteUsers,
        handleDeleteUsersForever,
        handleForceSignOutUsers,
        handleRevokeSiteAdmin,
        handlePromoteToSiteAdmin,
        handleResetUserPassword,
        notification,
    } = useUserListActions(refresh)

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
            {notification && (
                <Alert className="mt-2" variant={notification.isError ? 'danger' : 'success'}>
                    {notification.text}
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
                                bulk: true,
                                condition: users => users.some(user => !user.deletedAt),
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
                                key: 'delete',
                                label: 'Delete forever',
                                icon: mdiDelete,
                                iconColor: 'danger',
                                labelColor: 'danger',
                                onClick: handleDeleteUsersForever,
                                bulk: true,
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
                                key: 'reset-password',
                                label: 'Reset password',
                                icon: mdiLockReset,
                                onClick: handleResetUserPassword,
                                condition: ([user]) => !user?.deletedAt,
                            },
                        ]}
                        columns={[
                            {
                                key: SiteUserOrderBy.USERNAME,
                                accessor: 'username',
                                header: 'User',
                                sortable: true,
                                render: RenderUsernameAndEmail,
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
                        ]}
                        note={
                            <Text as="span">
                                Note: Events is the count of all billable events which equate to billable usage.
                                {/* TODO: Add link to billable events in bottom note */}
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

function RenderUsernameAndEmail({ username, email }: SiteUser): JSX.Element {
    return (
        <div className="d-flex flex-column p-2">
            <Text className={classNames(styles.linkColor, 'mb-0')}>{username}</Text>
            <Text className="mb-0">{email}</Text>
        </div>
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
    handleResetUserPassword: ActionHandler
}

function useUserListActions(refetch: () => void): UseUserListActionReturnType {
    const [forceSignOutUsers] = useMutation(FORCE_SIGN_OUT_USERS)
    const [deleteUsers] = useMutation(DELETE_USERS)
    const [deleteUsersForever] = useMutation(DELETE_USERS_FOREVER)

    const [notification, setNotification] = useState<UseUserListActionReturnType['notification']>()

    const onError = useCallback((error: any) => {
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
    }, [])

    const onSuccess = useCallback(
        (text: React.ReactNode) => () => {
            setNotification({ text })
            refetch()
        },
        [refetch]
    )

    const handleForceSignOutUsers = useCallback(
        (users: SiteUser[]) => {
            if (confirm('Are you sure you want to force sign out the selected user(s)?')) {
                forceSignOutUsers({ variables: { userIDs: users.map(user => user.id) } })
                    .then(onSuccess(`Successfully force signed out the ${users.length} user(s).`))
                    .catch(onError)
            }
        },
        [forceSignOutUsers, onError, onSuccess]
    )

    const handleDeleteUsers = useCallback(
        (users: SiteUser[]) => {
            if (confirm('Are you sure you want to delete the selected user(s)?')) {
                deleteUsers({ variables: { userIDs: users.map(user => user.id) } })
                    .then(onSuccess(`Successfully deleted the ${users.length} user(s).`))
                    .catch(onError)
            }
        },
        [deleteUsers, onError, onSuccess]
    )
    const handleDeleteUsersForever = useCallback(
        (users: SiteUser[]) => {
            if (confirm('Are you sure you want to delete the selected user(s)?')) {
                deleteUsersForever({ variables: { userIDs: users.map(user => user.id) } })
                    .then(onSuccess(`Successfully deleted forever the ${users.length} user(s).`))
                    .catch(onError)
            }
        },
        [deleteUsersForever, onError, onSuccess]
    )

    const handlePromoteToSiteAdmin = useCallback(
        ([user]: SiteUser[]) => {
            if (confirm('Are you sure you want to promote the selected user to site admin?')) {
                setUserIsSiteAdmin(user.id, true)
                    .toPromise()
                    .then(onSuccess(`Successfully promoted to site admin user ${user.username}.`))
                    .catch(onError)
            }
        },
        [onError, onSuccess]
    )

    const handleRevokeSiteAdmin = useCallback(
        ([user]: SiteUser[]) => {
            if (confirm('Are you sure you want to revoke the selected user from site admin?')) {
                setUserIsSiteAdmin(user.id, false)
                    .toPromise()
                    .then(onSuccess(`Successfully revoked site admin from user ${user.username}.`))
                    .catch(onError)
            }
        },
        [onError, onSuccess]
    )

    const handleResetUserPassword = useCallback(
        ([user]: SiteUser[]) => {
            if (confirm('Are you sure you want to reset the selected user password?')) {
                console.log('Reset user password', user)
                randomizeUserPassword(user.id)
                    .toPromise()
                    .then(({ resetPasswordURL }) => {
                        if (resetPasswordURL === null) {
                            onSuccess(
                                `Password was reset. The reset link was sent to the primary email of the user: ${user.email}`
                            )
                        } else {
                            onSuccess(
                                <>
                                    <Text>
                                        Password was reset. You must manually send <strong>{user.username}</strong> this
                                        reset link:
                                    </Text>
                                    <CopyableText text={resetPasswordURL} size={40} />
                                </>
                            )
                        }
                    })
                    .catch(onError)
            }
        },
        [onError, onSuccess]
    )

    return {
        notification,
        handleForceSignOutUsers,
        handleDeleteUsers,
        handleDeleteUsersForever,
        handlePromoteToSiteAdmin,
        handleRevokeSiteAdmin,
        handleResetUserPassword,
    }
}
