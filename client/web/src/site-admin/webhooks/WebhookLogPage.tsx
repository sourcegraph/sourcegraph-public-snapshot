import classNames from 'classnames'
import React, { useState } from 'react'

import { dataOrThrowErrors } from '@sourcegraph/shared/src/graphql/graphql'
import {
    useConnection,
    UseConnectionResult,
} from '@sourcegraph/web/src/components/FilteredConnection/hooks/useConnection'
import {
    ConnectionContainer,
    ConnectionError,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
} from '@sourcegraph/web/src/components/FilteredConnection/ui'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import { PageTitle } from '../../components/PageTitle'
import {
    Scalars,
    ServiceWebhookLogsResult,
    ServiceWebhookLogsVariables,
    WebhookLogFields,
    WebhookLogsResult,
    WebhookLogsVariables,
} from '../../graphql-operations'

import { EXTERNAL_SERVICE_WEBHOOK_LOGS, GLOBAL_WEBHOOK_LOGS, SelectedExternalService } from './backend'
import { WebhookLogNode } from './WebhookLogNode'
import styles from './WebhookLogPage.module.scss'
import { WebhookLogPageHeader } from './WebhookLogPageHeader'

export const WebhookLogPage: React.FunctionComponent<{}> = () => {
    const [onlyErrors, setOnlyErrors] = useState(false)
    const [externalService, setExternalService] = useState<SelectedExternalService>('all')

    return (
        <>
            <PageTitle title="Incoming webhook logs" />
            <PageHeader
                headingElement="h2"
                path={[{ text: 'Incoming webhook logs' }]}
                description="Use these logs of received webhooks to debug integrations"
                className="mb-3"
            />
            <Container>
                <WebhookLogPageHeader
                    onlyErrors={onlyErrors}
                    onSetOnlyErrors={setOnlyErrors}
                    externalService={externalService}
                    onSelectExternalService={setExternalService}
                />
                {externalService === 'all' || externalService === 'unmatched' ? (
                    <GlobalWebhookLogs onlyErrors={onlyErrors} onlyUnmatched={externalService === 'unmatched'} />
                ) : (
                    <ExternalServiceWebhookLogs externalService={externalService} onlyErrors={onlyErrors} />
                )}
            </Container>
        </>
    )
}

const Header: React.FunctionComponent<{}> = () => (
    <>
        <span className="d-none d-md-block" />
        <h5 className="d-none d-md-block text-uppercase text-center text-nowrap">Status code</h5>
        <h5 className="d-none d-md-block text-uppercase text-nowrap">External service</h5>
        <h5 className="d-none d-md-block text-uppercase text-center text-nowrap">Received at</h5>
    </>
)

const PAGE_SIZE = 5

const GlobalWebhookLogs: React.FunctionComponent<{ onlyErrors: boolean; onlyUnmatched: boolean }> = ({
    onlyErrors,
    onlyUnmatched,
}) => {
    const result = useConnection<WebhookLogsResult, WebhookLogsVariables, WebhookLogFields>({
        query: GLOBAL_WEBHOOK_LOGS,
        variables: {
            first: PAGE_SIZE,
            after: null,
            onlyErrors,
            onlyUnmatched,
        },
        options: {
            // The caching seems to break around the message fields: the fields
            // inherited from the common WebhookLogMessage type aren't
            // persisted. Since we always want the freshest webhooks anyway,
            // we'll just bypass the cache altogether.
            fetchPolicy: 'no-cache',
        },
        getConnection: result => dataOrThrowErrors(result).webhookLogs,
    })

    return <Connection result={result} />
}

const ExternalServiceWebhookLogs: React.FunctionComponent<{ externalService: Scalars['ID']; onlyErrors: boolean }> = ({
    externalService,
    onlyErrors,
}) => {
    const result = useConnection<ServiceWebhookLogsResult, ServiceWebhookLogsVariables, WebhookLogFields>({
        query: EXTERNAL_SERVICE_WEBHOOK_LOGS,
        variables: {
            first: PAGE_SIZE,
            after: null,
            id: externalService,
            onlyErrors,
        },
        options: {
            // The caching seems to break around the message fields: the fields
            // inherited from the common WebhookLogMessage type aren't
            // persisted. Since we always want the freshest webhooks anyway,
            // we'll just bypass the cache altogether.
            fetchPolicy: 'no-cache',
        },
        getConnection: result => {
            const data = dataOrThrowErrors(result)

            if (data.node?.__typename !== 'ExternalService') {
                throw new Error('unexpected non-ExternalService when querying external services')
            }

            return data.node.webhookLogs
        },
    })

    return <Connection result={result} />
}

const Connection: React.FunctionComponent<{ result: UseConnectionResult<WebhookLogFields> }> = ({
    result: { connection, error, loading, fetchMore, hasNextPage },
}) => (
    <ConnectionContainer>
        {error && <ConnectionError errors={[error.message]} />}
        <ConnectionList className={classNames('mt-3', styles.logs)}>
            <Header />
            {connection?.nodes?.map(node => (
                <WebhookLogNode key={node.id} node={node} />
            ))}
        </ConnectionList>
        {loading && <ConnectionLoading />}
        {connection && (
            <SummaryContainer centered={true}>
                <ConnectionSummary
                    noSummaryIfAllNodesVisible={true}
                    first={PAGE_SIZE}
                    connection={connection}
                    noun="webhook log"
                    pluralNoun="webhook logs"
                    hasNextPage={hasNextPage}
                />
                {hasNextPage && <ShowMoreButton onClick={fetchMore} />}
            </SummaryContainer>
        )}
    </ConnectionContainer>
)
