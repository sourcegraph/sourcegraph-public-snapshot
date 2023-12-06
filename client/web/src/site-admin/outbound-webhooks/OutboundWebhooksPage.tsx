import { type FC, useEffect } from 'react'

import { mdiAlertCircle, mdiWebhook, mdiMapSearch, mdiPencil, mdiPlus } from '@mdi/js'

import { pluralize } from '@sourcegraph/common'
import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ButtonLink, Container, H3, Icon, Link, PageHeader, Tooltip } from '@sourcegraph/wildcard'

import {
    ConnectionContainer,
    ConnectionError,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
} from '../../components/FilteredConnection/ui'
import { PageTitle } from '../../components/PageTitle'
import type { OutboundWebhookFieldsWithStats } from '../../graphql-operations'

import { useOutboundWebhooksConnection } from './backend'
import { DeleteButton } from './delete/DeleteButton'

import styles from './OutboundWebhooksPage.module.scss'

export interface OutboundWebhooksPageProps extends TelemetryProps, TelemetryV2Props {}

export const OutboundWebhooksPage: FC<OutboundWebhooksPageProps> = ({ telemetryService, telemetryRecorder }) => {
    useEffect(() => {
        telemetryService.logPageView('OutboundWebhooksPage')
        telemetryRecorder.recordEvent('OutboundWebhooksPage', 'viewed')
    }, [telemetryService, telemetryRecorder])

    const { loading, hasNextPage, fetchMore, refetchAll, connection, error } = useOutboundWebhooksConnection()

    return (
        <div>
            <PageTitle title="Outgoing webhooks" />
            <PageHeader
                path={[{ icon: mdiWebhook }, { to: '/site-admin/webhooks/outgoing', text: 'Outgoing webhooks' }]}
                headingElement="h2"
                description="All configured outgoing webhooks"
                className="mb-3"
                actions={
                    <ButtonLink to="/site-admin/webhooks/outgoing/create" variant="primary">
                        <Icon aria-hidden={true} svgPath={mdiPlus} /> Create webhook
                    </ButtonLink>
                }
            />

            <Container>
                <ConnectionContainer>
                    {error && <ConnectionError errors={[error.message]} />}
                    {loading && !connection && <ConnectionLoading />}
                    <ConnectionList className={styles.grid} aria-label="Outgoing webhooks">
                        {connection?.nodes?.map(node => (
                            <OutboundWebhookNode key={node.id} node={node} onDelete={refetchAll} />
                        ))}
                    </ConnectionList>
                    {connection && (
                        <SummaryContainer centered={true}>
                            <ConnectionSummary
                                noSummaryIfAllNodesVisible={false}
                                first={connection.totalCount ?? 0}
                                centered={true}
                                connection={connection}
                                noun="webhook"
                                pluralNoun="webhooks"
                                hasNextPage={hasNextPage}
                                emptyElement={<EmptyList />}
                            />
                            {hasNextPage && <ShowMoreButton centered={true} onClick={fetchMore} />}
                        </SummaryContainer>
                    )}
                </ConnectionContainer>
            </Container>
        </div>
    )
}

const OutboundWebhookNode: FC<{
    node: OutboundWebhookFieldsWithStats
    onDelete: () => void
}> = ({ node, onDelete }) => {
    const edit = `/site-admin/webhooks/outgoing/${node.id}`

    return (
        <li className={styles.node}>
            <span className={styles.separator} />
            <div className={styles.url}>
                <H3 className="mb-0">
                    <Link to={edit}>{node.url}</Link>
                </H3>
                <small className="text-muted">
                    {node.stats.total} recent {pluralize('request', node.stats.total)} sent,{' '}
                    {node.stats.errored > 0
                        ? `including ${node.stats.errored} ${pluralize('error', node.stats.errored)}`
                        : 'with no errors'}
                </small>
            </div>
            <RecentErrorIcon count={node.stats.errored} link={`${edit}?only_errors=true#logs`} />
            <div className={styles.buttons}>
                <ButtonLink to={edit} variant="secondary" className="mr-2">
                    <Icon aria-hidden={true} svgPath={mdiPencil} /> Edit
                </ButtonLink>
                <DeleteButton id={node.id} onDeleted={onDelete} />
            </div>
        </li>
    )
}

const EmptyList: FC<React.PropsWithChildren<{}>> = () => (
    <div className="text-muted text-center mb-3 w-100">
        <Icon className="icon" svgPath={mdiMapSearch} inline={false} aria-hidden={true} />
        <div className="pt-2">No webhooks have been created so far.</div>
    </div>
)

interface RecentErrorIconProps {
    count: number
    link: string
}

const RecentErrorIcon: FC<RecentErrorIconProps> = ({ count, link }) => {
    if (count === 0) {
        return null
    }

    const label = `${count} recent ${pluralize('error', count)}`

    return (
        <div className={styles.error}>
            <Link to={link}>
                <Tooltip content={label}>
                    <Icon aria-label={label} color="var(--danger)" svgPath={mdiAlertCircle} />
                </Tooltip>
            </Link>
        </div>
    )
}
