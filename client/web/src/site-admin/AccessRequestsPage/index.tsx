import React, { useEffect, useCallback, useState } from 'react'

import { mdiAccount, mdiCancel, mdiCheck } from '@mdi/js'
import { formatDistanceToNowStrict } from 'date-fns'
import { capitalize } from 'lodash'

import { useLazyQuery, useMutation, useQuery } from '@sourcegraph/http-client'
import { Card, Text, Alert, PageSwitcher, Link, Select } from '@sourcegraph/wildcard'

import { usePageSwitcherPagination } from '../../components/FilteredConnection/hooks/usePageSwitcherPagination'
import {
    PendingAccessRequestsListResult,
    PendingAccessRequestsListVariables,
    RejectAccessRequestResult,
    RejectAccessRequestVariables,
    ApproveAccessRequestResult,
    ApproveAccessRequestVariables,
    DoesUsernameExistResult,
    DoesUsernameExistVariables,
    AccessRequestCreateUserResult,
    AccessRequestCreateUserVariables,
    HasLicenseSeatsResult,
    HasLicenseSeatsVariables,
    AccessRequestStatus,
} from '../../graphql-operations'
import { useURLSyncedString } from '../../hooks/useURLSyncedString'
import { eventLogger } from '../../tracking/eventLogger'
import { AccountCreatedAlert } from '../components/AccountCreatedAlert'
import { SiteAdminPageTitle } from '../components/SiteAdminPageTitle'
import { IColumn, Table } from '../UserManagement/components/Table'

import {
    APPROVE_ACCESS_REQUEST,
    ACCESS_REQUEST_CREATE_USER,
    DOES_USERNAME_EXIST,
    PENDING_ACCESS_REQUESTS_LIST,
    REJECT_ACCESS_REQUEST,
    HAS_LICENSE_SEATS,
} from './queries'
import { pluralize } from '@sourcegraph/common'

/**
 * Converts a name to a username by removing all non-alphanumeric characters and converting to lowercase.
 * @param name user's name / full name
 * @param randomize whether to add a random suffix to the username to avoid collisions
 * @returns username
 */
function toUsername(name: string, randomize?: boolean): string {
    // Remove all non-alphanumeric characters from the name and convert to lowercase.
    const username = name.replace(/[^\dA-Za-z]/g, '').toLowerCase()
    if (!randomize) {
        return username
    }
    // Add a random 5-character suffix to the username to avoid collisions.
    return username + '-' + Math.random().toString(36).slice(2, 7)
}

/**
 * A react hook that returns a function that generates a username for a user with the given name.
 * It checks if the username already exists and if so, it adds a random suffix to the username.
 */
function useGenerateUsername(): (name: string) => Promise<string> {
    const [doesUsernameExist] = useLazyQuery<DoesUsernameExistResult, DoesUsernameExistVariables>(
        DOES_USERNAME_EXIST,
        {}
    )

    return useCallback(
        async (name: string) => {
            let username = toUsername(name)
            while (
                await doesUsernameExist({
                    variables: {
                        username,
                    },
                }).then(({ data }) => !!data?.user)
            ) {
                username = toUsername(name, true)
            }
            return username
        },
        [doesUsernameExist]
    )
}

function useHasRemainingSeats(): boolean {
    const { data } = useQuery<HasLicenseSeatsResult, HasLicenseSeatsVariables>(HAS_LICENSE_SEATS, {})

    const licenseSeatsCount = data?.site?.productSubscription?.license?.userCount
    const usersCount = data?.site?.users?.totalCount
    return typeof licenseSeatsCount !== 'number' || typeof usersCount !== 'number' || licenseSeatsCount > usersCount
}

const FIRST_COUNT = 25
export const AccessRequestsPage: React.FunctionComponent = () => {
    useEffect(() => {
        eventLogger.logPageView('AccessRequestsPage')
    }, [])
    const [error, setError] = useState<Error | null>(null)

    const [status, setStatus] = useURLSyncedString('status', AccessRequestStatus.PENDING)

    const {
        connection,
        error: queryError,
        loading,
        refetch,
        ...paginationArgs
    } = usePageSwitcherPagination<
        PendingAccessRequestsListResult,
        PendingAccessRequestsListVariables,
        PendingAccessRequestsListResult['accessRequests']['nodes'][0]
    >({
        query: PENDING_ACCESS_REQUESTS_LIST,
        variables: {
            status: status as AccessRequestStatus,
            first: FIRST_COUNT,
        },
        getConnection: result => result.data?.accessRequests,
        options: {
            fetchPolicy: 'cache-first',
            pageSize: FIRST_COUNT,
            useURL: true,
        },
    })

    const [rejectAccessRequest] = useMutation<RejectAccessRequestResult, RejectAccessRequestVariables>(
        REJECT_ACCESS_REQUEST
    )

    const handleReject = useCallback(
        (id: string) => {
            if (!confirm('Are you sure you want to reject the selected access request?')) {
                return
            }
            eventLogger.log('AccessRequestRejected', { id })
            rejectAccessRequest({
                variables: {
                    id,
                },
            })
                .then(() => refetch())
                .catch(error => {
                    setError(error)
                    // eslint-disable-next-line no-console
                    console.error(error)
                })
        },
        [refetch, rejectAccessRequest]
    )

    const [lastApprovedUser, setLastApprovedUser] = useState<{
        email: string
        resetPasswordURL?: string | null
        username: string
    }>()

    const [createUser] = useMutation<AccessRequestCreateUserResult, AccessRequestCreateUserVariables>(
        ACCESS_REQUEST_CREATE_USER
    )

    const [approveAccessRequest] = useMutation<ApproveAccessRequestResult, ApproveAccessRequestVariables>(
        APPROVE_ACCESS_REQUEST
    )

    const generateUsername = useGenerateUsername()

    const handleApprove = useCallback(
        (id: string, name: string, email: string): void => {
            if (!confirm('Are you sure you want to approve the selected access request?')) {
                return
            }
            eventLogger.log('AccessRequestApproved', { id })
            async function createUserAndApproveRequest(): Promise<void> {
                const username = await generateUsername(name)
                const { data } = await createUser({
                    variables: {
                        email,
                        username,
                    },
                })

                if (!data) {
                    throw new Error('No data returned from approveAccessRequest mutation')
                }

                await approveAccessRequest({
                    variables: {
                        id,
                    },
                })

                setLastApprovedUser({
                    username,
                    email,
                    resetPasswordURL: data?.createUser.resetPasswordURL,
                })
                await refetch()
            }
            createUserAndApproveRequest().catch(error => {
                setError(error)
                // eslint-disable-next-line no-console
                console.error(error)
            })
        },
        [generateUsername, createUser, approveAccessRequest, refetch]
    )

    const hasRemainingSeats = useHasRemainingSeats()

    return (
        <>
            <SiteAdminPageTitle icon={mdiAccount}>
                <span>Users</span>
                <span>Account requests</span>
            </SiteAdminPageTitle>
            {!hasRemainingSeats && (
                <Alert variant="danger">
                    No licenses remaining. To approve requests,{' '}
                    <Link to="https://about.sourcegraph.com/pricing" target="_blank" rel="noopener">
                        purchase additional licenses
                    </Link>{' '}
                    or <Link to="/site-admin/users">remove inactive users</Link>.
                </Alert>
            )}
            <Card className="p-3">
                {[queryError, error].filter(Boolean).map((err, index) => (
                    <Alert variant="danger" key={index}>
                        {err?.message}
                    </Alert>
                ))}
                {lastApprovedUser && (
                    <AccountCreatedAlert
                        email={lastApprovedUser.email}
                        username={lastApprovedUser.username}
                        resetPasswordURL={lastApprovedUser.resetPasswordURL}
                    />
                )}
                <div className="d-flex align-items-start justify-content-between">
                    <AccessRequestStatusPicker status={status as AccessRequestStatus} onChange={setStatus} />
                    <div className="d-flex justify-content-end mt-4">
                        <PageSwitcher
                            totalCount={connection?.totalCount ?? null}
                            totalLabel={pluralize('account request', connection?.totalCount || 0)}
                            {...paginationArgs}
                        />
                    </div>
                </div>
                {!!connection?.nodes.length && (
                    <>
                        <Table<AccessRequestItem>
                            columns={TableColumns}
                            getRowId={node => node.id}
                            data={connection.nodes}
                            actions={[
                                {
                                    key: 'Reject',
                                    label: 'Reject',
                                    icon: mdiCancel,
                                    bulk: true,
                                    onClick: ([item]) => handleReject(item.id),
                                    // Can reject only if pending
                                    condition: () => status === AccessRequestStatus.PENDING,
                                },
                                {
                                    key: 'Approve',
                                    label: 'Approve',
                                    icon: mdiCheck,
                                    onClick: ([item]) => handleApprove(item.id, item.name, item.email),
                                    // Can approve if there are remaining seats and not already approved
                                    // We allow to approve previously rejected requests too
                                    condition: () => hasRemainingSeats && status !== AccessRequestStatus.APPROVED,
                                },
                            ]}
                        />
                    </>
                )}
                {!loading && connection?.nodes.length === 0 && (
                    <div>
                        <Alert variant="info">No {capitalize(status)} requests</Alert>
                        <Text>
                            Users can request access to Sourcegraph via the login page. View the documentation to learn
                            more about{' '}
                            <Link to="/help/admin/auth#how-to-control-user-sign-up">controlling sign up requests</Link>.
                        </Text>
                    </div>
                )}
            </Card>
        </>
    )
}

type AccessRequestItem = PendingAccessRequestsListResult['accessRequests']['nodes'][0]

const TableColumns: IColumn<AccessRequestItem>[] = [
    {
        key: 'Email',
        header: 'Email',
        render: (node: AccessRequestItem) => <Text className="mb-0 mr-2 d-flex align-items-center">{node.email}</Text>,
    },
    {
        key: 'Name',
        header: 'Name',
        render: (node: AccessRequestItem) => (
            <Text className="mb-0 mr-2 d-flex align-items-center text-nowrap">{node.name}</Text>
        ),
    },
    {
        key: 'Status',
        header: 'Status',
        align: 'right',
        render: (node: AccessRequestItem) => (
            <Text className="mb-0 mr-2 d-flex align-items-center">{capitalize(node.status)}</Text>
        ),
    },
    {
        key: 'Created at',
        header: 'Created at',
        align: 'right',
        render: (node: AccessRequestItem) => (
            <Text className="mb-0 mr-2 d-flex align-items-center text-nowrap">
                {formatDistanceToNowStrict(new Date(node.createdAt), { addSuffix: true })}
            </Text>
        ),
    },
    {
        key: 'Notes',
        header: 'Notes',
        align: 'right',
        render: (node: AccessRequestItem) => (
            <Text className="text-muted my-2" size="small">
                {node.additionalInfo}
            </Text>
        ),
    },
]

const AccessRequestStatusPicker: React.FunctionComponent<{
    status: AccessRequestStatus
    onChange: (value: AccessRequestStatus) => void
}> = ({ status, onChange }) => {
    const handleStatusChange = useCallback(
        (event: React.ChangeEvent<HTMLSelectElement>) => {
            onChange(event.target.value as AccessRequestStatus)
        },
        [onChange]
    )

    return (
        <Select id="access-request-status-filter" value={status} label="Status" onChange={handleStatusChange}>
            {Object.entries(AccessRequestStatus).map(([key, value]) => (
                <option key={key} value={value}>
                    {capitalize(value)}
                </option>
            ))}
        </Select>
    )
}
