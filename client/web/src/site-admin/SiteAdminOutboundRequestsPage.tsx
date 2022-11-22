import React, { useCallback, useEffect, useState } from 'react'

import classNames from 'classnames'
import { RouteComponentProps } from 'react-router'
import { of } from 'rxjs'
import { map } from 'rxjs/operators'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { useQuery } from '@sourcegraph/http-client/src'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container, H3, LoadingSpinner, PageHeader, Text } from '@sourcegraph/wildcard'

import {
    FilteredConnection,
    FilteredConnectionFilter,
    FilteredConnectionQueryArguments,
} from '../components/FilteredConnection'
import { PageTitle } from '../components/PageTitle'
import { Timestamp } from '../components/time/Timestamp'
import { OutboundRequestsResult, OutboundRequestsVariables } from '../graphql-operations'

import { OUTBOUND_REQUESTS, OUTBOUND_REQUESTS_PAGE_POLL_INTERVAL } from './backend'

import styles from './SiteAdminOutboundRequestsPage.module.scss'

export interface SiteAdminOutboundRequestsPageProps extends RouteComponentProps, TelemetryProps {
    now?: () => Date
}

export type OutboundRequest = OutboundRequestsResult['outboundRequests'][0]

export const SiteAdminOutboundRequestsPage: React.FunctionComponent<
    React.PropsWithChildren<SiteAdminOutboundRequestsPageProps>
> = ({ history, location, telemetryService }) => {
    const [items, setItems] = useState<OutboundRequest[]>([])
    // const [previousData, setPreviousData] = useState<OutboundRequestsResult | null>(null)
    useEffect(() => {
        telemetryService.logPageView('SiteAdminOutboundRequests')
    }, [telemetryService])

    const lastKey = items[items.length - 1]?.key ?? null
    const { data, loading, error, stopPolling, refetch, startPolling } = useQuery<
        OutboundRequestsResult,
        OutboundRequestsVariables
    >(OUTBOUND_REQUESTS, {
        variables: { lastKey },
        pollInterval: OUTBOUND_REQUESTS_PAGE_POLL_INTERVAL,
    })

    if (data?.outboundRequests?.length && (!lastKey || data?.outboundRequests[0].key > lastKey)) {
        const newItems = items
            .concat(...data.outboundRequests)
            .slice(Math.max(items.length + data.outboundRequests.length - 50, 0))
        stopPolling()
        setItems(newItems)
        refetch({ lastKey: newItems[newItems.length - 1]?.key ?? null })
            .then(() => {})
            .catch(() => {})
        startPolling(OUTBOUND_REQUESTS_PAGE_POLL_INTERVAL)
    }

    const queryOutboundRequests = useCallback(
        (args: FilteredConnectionQueryArguments & { failed?: boolean }) =>
            of(items.reverse()).pipe(
                map(items => {
                    const filtered = items?.filter(
                        request =>
                            (!args.query || request.url.includes(args.query)) &&
                            (args.failed !== false || request.statusCode < 400) &&
                            (args.failed !== true || request.statusCode >= 400)
                    )
                    return {
                        nodes: filtered ?? [],
                        totalCount: (filtered ?? []).length,
                    }
                })
            ),
        [items]
    )

    const filters: FilteredConnectionFilter[] = [
        {
            id: 'filters',
            label: 'Filter by success',
            type: 'select',
            values: [
                {
                    label: 'All',
                    value: 'all',
                    tooltip: 'Show all requests',
                    args: {},
                },
                {
                    label: 'Failed',
                    value: 'failed',
                    tooltip: 'Show only failed requests',
                    args: { failed: true },
                },
                {
                    label: 'Successful',
                    value: 'successful',
                    tooltip: 'Show only successful requests',
                    args: { failed: false },
                },
            ],
        },
    ]

    return (
        <div className="site-admin-migrations-page">
            <PageTitle title="Outbound requests - Admin" />
            <PageHeader
                path={[{ text: 'Outbound requests' }]}
                headingElement="h2"
                description={
                    <>
                        This is the log of recent external requests sent by the Sourcegraph instance. Handy for seeing
                        what's happening between Sourcegraph and other services.
                        <br />
                        <strong>The list updates every five seconds.</strong>
                    </>
                }
                className="mb-3"
            />

            <Container className="mb-3">
                {error && !loading && <ErrorAlert error={error} />}
                {loading && !error && <LoadingSpinner />}
                <FilteredConnection<OutboundRequest>
                    className="mb-0"
                    listComponent="div"
                    listClassName={classNames('list-group mb-3', styles.requestsGrid)}
                    noun="request"
                    pluralNoun="requests"
                    queryConnection={queryOutboundRequests}
                    nodeComponent={MigrationNode}
                    filters={filters}
                    history={history}
                    location={location}
                />
            </Container>
        </div>
    )
}

const MigrationNode: React.FunctionComponent<{ node: React.PropsWithChildren<OutboundRequest> }> = ({ node }) => {
    const roundedSecond = Math.round((node.duration + Number.EPSILON) * 100) / 100
    return (
        <React.Fragment key={node.key}>
            <span className={styles.separator} />
            <div className={classNames('d-flex flex-column', styles.progress)}>
                <Text className="">
                    <Timestamp date={node.startedAt} noAbout={true} />
                </Text>
                <Text>{node.statusCode}</Text>
            </div>
            <div className={classNames('d-flex flex-column', styles.information)}>
                <div>
                    <H3>
                        {node.url} <span className="text-muted">Req headers</span>{' '}
                        <span className="text-muted">Req body</span> <span className="text-muted">Resp headers</span>{' '}
                        <span className="text-muted">Error message</span>{' '}
                        <span className="text-muted">Creation stack trace</span>{' '}
                        <span className="text-muted">Call stack trace</span>
                    </H3>

                    <Text className="m-0">
                        <span className="text-muted">Method: </span> <strong>{node.method}</strong>.{' '}
                        <span className="text-muted">Took </span>{' '}
                        <strong>
                            {roundedSecond} second{roundedSecond === 1 ? '' : 's'} to complete.
                        </strong>
                        <span className="text-muted">.</span>
                    </Text>

                    <Text className="m-0">
                        <span className="text-muted">Began running at</span>
                        {node.startedAt}
                    </Text>
                </div>
            </div>
        </React.Fragment>
    )
}
