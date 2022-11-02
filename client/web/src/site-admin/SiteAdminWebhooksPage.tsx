import React, { useCallback, useEffect } from 'react'

import { mdiMapSearch, mdiPlus } from '@mdi/js'
import { RouteComponentProps } from 'react-router'
import { WebhookFields } from 'src/graphql-operations'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ButtonLink, Container, H5, Icon, PageHeader } from '@sourcegraph/wildcard'

import { FilteredConnection } from '../components/FilteredConnection'
import { PageTitle } from '../components/PageTitle'

import { queryWebhooks } from './backend'
import { WebhookNode } from './WebhookNode'

import styles from './SiteAdminWebhooksPage.module.scss'

interface Props extends RouteComponentProps<{}>, TelemetryProps {}

export const SiteAdminWebhooksPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    history,
    location,
    telemetryService,
}) => {
    useEffect(() => {
        telemetryService.logPageView('SiteAdminWebhooks')
    }, [telemetryService])

    const queryConnection = useCallback(() => queryWebhooks(), [])

    return (
        <div className="site-admin-webhooks-page">
            <PageTitle title="Webhook receivers" />
            <PageHeader
                path={[{ text: 'Webhook receivers' }]}
                headingElement="h2"
                description="All configured webhooks receivers"
                className="mb-3"
                actions={
                    <>
                        <ButtonLink className="test-create-webhook" variant="primary">
                            <Icon aria-hidden={true} svgPath={mdiPlus} /> Add webhook
                        </ButtonLink>
                    </>
                }
            />

            <Container className="mb-3">
                <FilteredConnection<WebhookFields>
                    className="mb-0"
                    noun="webhook"
                    pluralNoun="webhooks"
                    queryConnection={queryConnection}
                    nodeComponent={WebhookNode}
                    nodeComponentProps={{}}
                    hideSearch={true}
                    listComponent="div"
                    listClassName={styles.specsGrid}
                    withCenteredSummary={true}
                    noSummaryIfAllNodesVisible={true}
                    history={history}
                    location={location}
                    headComponent={Header}
                    emptyElement={<EmptyList />}
                />
            </Container>
        </div>
    )
}

const Header: React.FunctionComponent<React.PropsWithChildren<{}>> = () => (
    <>
        <H5 className="p-2 d-none d-md-block text-uppercase text-left text-nowrap">Receiver</H5>
        <H5 className="d-none d-md-block text-uppercase text-center text-nowrap">Actions</H5>
    </>
)

const EmptyList: React.FunctionComponent<React.PropsWithChildren<{}>> = () => (
    <div className="text-muted text-center mb-3 w-100">
        <Icon className="icon" svgPath={mdiMapSearch} inline={false} aria-hidden={true} />
        <div className="pt-2">No webhooks have been created so far.</div>
    </div>
)
