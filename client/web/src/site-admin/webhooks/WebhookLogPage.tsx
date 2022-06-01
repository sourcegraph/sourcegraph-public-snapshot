import React, { useCallback, useState } from 'react'

import classNames from 'classnames'
import { RouteComponentProps } from 'react-router'

import { Container, PageHeader, Typography } from '@sourcegraph/wildcard'

import { FilteredConnection, FilteredConnectionQueryArguments } from '../../components/FilteredConnection'
import { PageTitle } from '../../components/PageTitle'

import { queryWebhookLogs as _queryWebhookLogs, SelectedExternalService } from './backend'
import { WebhookLogNode } from './WebhookLogNode'
import { WebhookLogPageHeader } from './WebhookLogPageHeader'

import styles from './WebhookLogPage.module.scss'

export interface Props extends Pick<RouteComponentProps, 'history' | 'location'> {
    queryWebhookLogs?: typeof _queryWebhookLogs
}

export const WebhookLogPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
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

const Header: React.FunctionComponent<React.PropsWithChildren<{}>> = () => (
    <>
        <span className="d-none d-md-block" />
        <Typography.H5 className="d-none d-md-block text-uppercase text-center text-nowrap">Status code</Typography.H5>
        <Typography.H5 className="d-none d-md-block text-uppercase text-nowrap">External service</Typography.H5>
        <Typography.H5 className="d-none d-md-block text-uppercase text-center text-nowrap">Received at</Typography.H5>
    </>
)
