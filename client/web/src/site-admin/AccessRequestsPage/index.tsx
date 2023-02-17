import React, { Fragment, useEffect, useCallback, useState } from 'react'

import { formatDistanceToNowStrict } from 'date-fns'

import { useLazyQuery, useMutation, useQuery } from '@sourcegraph/http-client'
import { H1, Card, Text, Button, Grid, Alert } from '@sourcegraph/wildcard'

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
import { useURLSyncedState } from '../../hooks'
import { eventLogger } from '../../tracking/eventLogger'
import { AccountCreatedAlert } from '../components/AccountCreatedAlert'
import { DropdownPagination } from '../components/DropdownPagination'

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

const LIMIT_FILTER_OPTIONS = [25, 50, 100]
const DEFAULT_FILTERS = {
    offset: '0',
    limit: LIMIT_FILTER_OPTIONS[0].toString(),
}

export const AccessRequestsPage: React.FunctionComponent = () => {
    useEffect(() => {
        eventLogger.logPageView('AccessRequestsPage')
    }, [])
    const [filters, setFilters] = useURLSyncedState(DEFAULT_FILTERS)
    const [error, setError] = useState<Error | null>(null)

    const offset = Number(filters.offset)
    const limit = Number(filters.limit)

    const {
        data,
        refetch,
        error: queryError,
    } = useQuery<PendingAccessRequestsListResult, PendingAccessRequestsListVariables>(PENDING_ACCESS_REQUESTS_LIST, {
        variables: {
            limit,
            offset,
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
            <div className="d-flex justify-content-between align-items-center mb-4 mt-2">
                <H1 className="d-flex align-items-center mb-0">Access Requests</H1>
            </div>
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
                <div className="d-flex justify-content-end mb-3">
                    <DropdownPagination
                        limit={limit}
                        offset={offset}
                        total={data?.accessRequests?.totalCount ?? 0}
                        onLimitChange={limit => setFilters({ limit: limit.toString() })}
                        onOffsetChange={offset => setFilters({ offset: offset.toString() })}
                        options={LIMIT_FILTER_OPTIONS}
                    />
                </div>
                <Grid columnCount={5}>
                    {['Email', 'Name', 'Last requested at', 'Extra Details', ''].map((value, index) => (
                        // eslint-disable-next-line react/no-array-index-key
                        <Text weight="medium" key={index} className="mb-1">
                            {value}
                        </Text>
                    ))}
                    {data?.accessRequests?.nodes.map(({ id, email, name, createdAt, additionalInfo }) => (
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
                                <Button
                                    variant="success"
                                    size="sm"
                                    className="mr-2"
                                    onClick={() => handleApprove(id, name, email)}
                                >
                                    Approve
                                </Button>
                                <Button variant="danger" size="sm" onClick={() => handleReject(id)}>
                                    Reject
                                </Button>
                            </div>
                        </Fragment>
                    ))}
                </Grid>
            </Card>
        </>
    )
}
