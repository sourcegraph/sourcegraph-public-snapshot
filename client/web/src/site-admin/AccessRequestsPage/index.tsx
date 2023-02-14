import React, { Fragment, useEffect, useCallback } from 'react'

import { mdiAccount, mdiPlus } from '@mdi/js'
import { formatDistanceToNowStrict } from 'date-fns'

import { useMutation, useQuery } from '@sourcegraph/http-client'
import { H1, Card, Text, Icon, Button, Link, Grid } from '@sourcegraph/wildcard'

import {
    PendingAccessRequestsListResult,
    PendingAccessRequestsListVariables,
    RejectAccessRequestResult,
    RejectAccessRequestVariables,
} from '../../graphql-operations'
import { useURLSyncedState } from '../../hooks'
import { eventLogger } from '../../tracking/eventLogger'
import { DropdownPagination } from '../components/DropdownPagination'

import { PENDING_ACCESS_REQUESTS_LIST, REJECT_ACCESS_REQUEST } from './queries'

import styles from './index.module.scss'

const DEFAULT_FILTERS = {
    offset: '0',
    limit: '25',
}

export const AccessRequestsPage: React.FunctionComponent = () => {
    useEffect(() => {
        eventLogger.logPageView('AccessRequestsPage')
    }, [])

    const [filters, setFilters] = useURLSyncedState(DEFAULT_FILTERS)

    const offset = Number(filters.offset)
    const limit = Number(filters.limit)

    const { data, refetch } = useQuery<PendingAccessRequestsListResult, PendingAccessRequestsListVariables>(
        PENDING_ACCESS_REQUESTS_LIST,
        {
            variables: {
                limit,
                offset,
            },
        }
    )

    const [rejectAccessRequest] = useMutation<RejectAccessRequestResult, RejectAccessRequestVariables>(
        REJECT_ACCESS_REQUEST
    )

    const handleReject = useCallback(
        (id: string) => {
            if (confirm('Are you sure you want to delete the selected access request?')) {
                rejectAccessRequest({
                    variables: {
                        id,
                    },
                })
                    .then(() => refetch())
                    // eslint-disable-next-line no-console
                    .catch(error => console.error(error))
            }
        },
        [refetch, rejectAccessRequest]
    )

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
                    Access Requests
                </H1>
                <div>
                    <Button to="/site-admin/users/new" variant="primary" as={Link}>
                        <Icon svgPath={mdiPlus} aria-label="create user" className="mr-1" />
                        Create User
                    </Button>
                </div>
            </div>
            <DropdownPagination
                limit={limit}
                offset={offset}
                total={data?.accessRequests?.totalCount ?? 0}
                onLimitChange={limit => setFilters({ limit: limit.toString() })}
                onOffsetChange={offset => setFilters({ offset: offset.toString() })}
                options={[4, 8, 16]}
            />
            <Card className="p-3">
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
                                <Button variant="success" size="sm" className="mr-2">
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
            <Text className="font-italic text-center mt-2">
                All events are generated from entries in the event logs table and are updated every 24 hours.
            </Text>
        </>
    )
}
