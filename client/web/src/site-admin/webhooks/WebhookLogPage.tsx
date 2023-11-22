import React, { useCallback, useState } from 'react'

import classNames from 'classnames'

import { Alert, Container, PageHeader, H5, Link } from '@sourcegraph/wildcard'

import { FilteredConnection, type FilteredConnectionQueryArguments } from '../../components/FilteredConnection'
import { PageTitle } from '../../components/PageTitle'

import { queryWebhookLogs as _queryWebhookLogs, type SelectedExternalService } from './backend'
import { WebhookLogNode } from './WebhookLogNode'
import { WebhookLogPageHeader } from './WebhookLogPageHeader'

import styles from './WebhookLogPage.module.scss'

export interface Props {
    queryWebhookLogs?: typeof _queryWebhookLogs
    webhookID?: string
}

export const WebhookLogPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    queryWebhookLogs = _queryWebhookLogs,
    webhookID,
}) => {
    const [onlyErrors, setOnlyErrors] = useState(false)
    const [externalService, setExternalService] = useState<SelectedExternalService>('all')

    const query = useCallback(
        ({ first, after }: FilteredConnectionQueryArguments) =>
            queryWebhookLogs(
                {
                    first: first ?? null,
                    after: after ?? null,
                },
                externalService,
                onlyErrors,
                webhookID
            ),
        [externalService, onlyErrors, queryWebhookLogs, webhookID]
    )

    return (
        <>
            <PageTitle title="Incoming webhook logs" />
            <PageHeader
                headingElement="h2"
                path={[{ text: 'Incoming webhook logs' }]}
                description="Use these logs of received webhooks to debug integrations"
                className="mb-3"
            />
            <Alert variant="warning">
                This webhooks page has been deprecated, please see our{' '}
                <Link to="/site-admin/webhooks/incoming">new webhooks page</Link>.
            </Alert>
            <Container>
                <WebhookLogPageHeader
                    onlyErrors={onlyErrors}
                    onSetOnlyErrors={setOnlyErrors}
                    externalService={externalService}
                    onSelectExternalService={setExternalService}
                />
                <FilteredConnection
                    queryConnection={query}
                    nodeComponent={WebhookLogNode}
                    noun="webhook log"
                    pluralNoun="webhook logs"
                    hideSearch={true}
                    headComponent={Header}
                    listClassName={classNames('mt-3', styles.logs)}
                    emptyElement={<div className="m-4 w-100 text-center">No webhook logs found</div>}
                />
            </Container>
        </>
    )
}

const Header: React.FunctionComponent<React.PropsWithChildren<{}>> = () => (
    <>
        <span className="d-none d-md-block" />
        <H5 className="d-none d-md-block text-uppercase text-center text-nowrap">Status code</H5>
        <H5 className="d-none d-md-block text-uppercase text-nowrap">External service</H5>
        <H5 className="d-none d-md-block text-uppercase text-center text-nowrap">Received at</H5>
    </>
)
