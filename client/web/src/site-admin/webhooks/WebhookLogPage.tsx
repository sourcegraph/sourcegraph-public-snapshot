import classNames from 'classnames'
import React, { useCallback, useState } from 'react'
import { RouteComponentProps } from 'react-router'

import { Container, PageHeader } from '@sourcegraph/wildcard'

import { FilteredConnection, FilteredConnectionQueryArguments } from '../../components/FilteredConnection'
import { PageTitle } from '../../components/PageTitle'

import { queryWebhookLogs as _queryWebhookLogs, SelectedExternalService } from './backend'
import { WebhookLogNode } from './WebhookLogNode'
import styles from './WebhookLogPage.module.scss'
import { WebhookLogPageHeader } from './WebhookLogPageHeader'

export interface Props extends Pick<RouteComponentProps, 'history' | 'location'> {
    queryWebhookLogs?: typeof _queryWebhookLogs
}

export const WebhookLogPage: React.FunctionComponent<Props> = ({
    history,
    location,
    queryWebhookLogs = _queryWebhookLogs,
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
                onlyErrors
            ),
        [externalService, onlyErrors, queryWebhookLogs]
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
            <Container>
                <WebhookLogPageHeader
                    onlyErrors={onlyErrors}
                    onSetOnlyErrors={setOnlyErrors}
                    externalService={externalService}
                    onSelectExternalService={setExternalService}
                />
                <FilteredConnection
                    history={history}
                    location={location}
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

const Header: React.FunctionComponent<{}> = () => (
    <>
        <span className="d-none d-md-block" />
        <h5 className="d-none d-md-block text-uppercase text-center text-nowrap">Status code</h5>
        <h5 className="d-none d-md-block text-uppercase text-nowrap">External service</h5>
        <h5 className="d-none d-md-block text-uppercase text-center text-nowrap">Received at</h5>
    </>
)
