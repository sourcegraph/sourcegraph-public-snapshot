import React, { useCallback, useEffect } from 'react'

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
    FilteredConnectionQueryArguments
} from '../components/FilteredConnection'
import { PageTitle } from '../components/PageTitle'
import { OutboundRequestsResult, OutboundRequestsVariables } from '../graphql-operations'

import { OUTBOUND_REQUESTS, OUTBOUND_REQUESTS_PAGE_POLL_INTERVAL } from './backend'

import styles from './SiteAdminOutboundRequestsPage.module.scss'

export interface SiteAdminOutboundRequestsPageProps extends RouteComponentProps, TelemetryProps {
    now?: () => Date
}

export const SiteAdminOutboundRequestsPage: React.FunctionComponent<
    React.PropsWithChildren<SiteAdminOutboundRequestsPageProps>
> = ({ history, location, telemetryService }) => {
    useEffect(() => {
        telemetryService.logPageView('SiteAdminOutboundRequests')
    }, [telemetryService])

    const { data, loading, error } = useQuery<OutboundRequestsResult, OutboundRequestsVariables>(OUTBOUND_REQUESTS, {
        pollInterval: OUTBOUND_REQUESTS_PAGE_POLL_INTERVAL,
    })

    const queryOutboundRequests = useCallback(
        (args: FilteredConnectionQueryArguments & {failed?: boolean}) =>
            of(data).pipe(
                map(data => {
                    const all = data?.outboundRequests;
                    const filtered = all?.filter(
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
        [data]
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
                <FilteredConnection<OutboundRequestsResult['outboundRequests'][number]>
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

const MigrationNode: React.FunctionComponent<{
    node: React.PropsWithChildren<OutboundRequestsResult['outboundRequests'][number]>
}> = ({ node: outboundRequest }) => (
    <React.Fragment key={`${outboundRequest.startedAt} ${outboundRequest.duration}`}>
        <span className={styles.separator} />

        <div className={classNames('d-flex flex-column', styles.information)}>
            <div>
                <H3>{outboundRequest.url}</H3>

                <Text className="m-0">
                    <span className="text-muted">Method: </span> <strong>{outboundRequest.method}</strong>{' '}
                    <span className="text-muted">in </span> <strong>{outboundRequest.duration}</strong>
                    <span className="text-muted">.</span>
                </Text>

                <Text className="m-0">
                    <span className="text-muted">Began running at</span>
                    {outboundRequest.startedAt}
                </Text>
            </div>
        </div>
    </React.Fragment>
)
