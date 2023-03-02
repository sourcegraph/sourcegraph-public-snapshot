import React, { Fragment, useEffect, useCallback, useState } from 'react'

import { mdiAccount } from '@mdi/js'
import { formatDistanceToNowStrict } from 'date-fns'

import { useLazyQuery, useMutation } from '@sourcegraph/http-client'
import { Card, Text, Button, Grid, Alert, PageSwitcher, Link } from '@sourcegraph/wildcard'

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
} from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'
import { AccountCreatedAlert } from '../components/AccountCreatedAlert'
import { SiteAdminPageTitle } from '../components/SiteAdminPageTitle'

import {
    APPROVE_ACCESS_REQUEST,
    ACCESS_REQUEST_CREATE_USER,
    DOES_USERNAME_EXIST,
    PENDING_ACCESS_REQUESTS_LIST,
    REJECT_ACCESS_REQUEST,
} from './queries'

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

const FIRST_COUNT = 25
export const AccessRequestsPage: React.FunctionComponent = () => {
    useEffect(() => {
        eventLogger.logPageView('AccessRequestsPage')
    }, [])
    const [error, setError] = useState<Error | null>(null)

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

    return (
        <>
            <SiteAdminPageTitle icon={mdiAccount}>
                <span>Users</span>
                <span>Account requests</span>
            </SiteAdminPageTitle>
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
                <div className="d-flex justify-content-end">
                    <PageSwitcher
                        totalCount={connection?.totalCount ?? null}
                        totalLabel={connection?.totalCount === 1 ? 'account request' : 'account requests'}
                        {...paginationArgs}
                    />
                </div>
                <AccessRequestsList onApprove={handleApprove} onReject={handleReject} items={connection?.nodes || []} />
                {!loading && connection?.nodes.length === 0 && (
                    <div>
                        <Alert variant="info">No pending requests</Alert>
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

interface AccessRequestsListProps {
    onApprove: (id: string, name: string, email: string) => void
    onReject: (id: string) => void
    items: PendingAccessRequestsListResult['accessRequests']['nodes']
}

const AccessRequestsList: React.FunctionComponent<AccessRequestsListProps> = ({ onApprove, onReject, items }) => {
    if (items.length === 0) {
        return null
    }
    return (
        <Grid columnCount={5}>
            {['Email', 'Name', 'Created at', 'Notes', ''].map((value, index) => (
                // eslint-disable-next-line react/no-array-index-key
                <Text weight="medium" key={index} className="mb-1">
                    {value}
                </Text>
            ))}
            {items.map(({ id, email, name, createdAt, additionalInfo }) => (
                <Fragment key={email}>
                    <Text className="mb-0 d-flex align-items-center">{email}</Text>
                    <Text className="mb-0 d-flex align-items-center">{name}</Text>
                    <Text className="mb-0 d-flex align-items-center">
                        {formatDistanceToNowStrict(new Date(createdAt), { addSuffix: true })}
                    </Text>
                    <Text className="text-muted mb-0 d-flex align-items-center" size="small">
                        {additionalInfo}
                    </Text>
                    <div className="d-flex justify-content-end align-items-start">
                        <Button variant="link" onClick={() => onReject(id)}>
                            Reject
                        </Button>
                        <Button variant="success" className="ml-2" onClick={() => onApprove(id, name, email)}>
                            Approve
                        </Button>
                    </div>
                </Fragment>
            ))}
        </Grid>
    )
}
