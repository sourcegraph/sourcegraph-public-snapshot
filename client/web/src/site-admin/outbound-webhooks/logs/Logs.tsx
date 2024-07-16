import { useMemo, type FC } from 'react'

import { mdiAlertCircle, mdiMapSearch } from '@mdi/js'
import classNames from 'classnames'
import { useLocation, useNavigate } from 'react-router-dom'

import { Button, H3, Icon } from '@sourcegraph/wildcard'

import {
    ConnectionContainer,
    ConnectionError,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
} from '../../../components/FilteredConnection/ui'
import { SiteAdminWebhookPageHeader } from '../../SiteAdminWebhookPage'
import { WebhookLogNode } from '../../webhooks/WebhookLogNode'

import { useOutboundWebhookLogsConnection } from './backend'

import styles from './Logs.module.scss'

const ONLY_ERRORS_PARAM = 'only_errors'

export interface LogsProps {
    id: string
}

export const Logs: FC<LogsProps> = ({ id }) => {
    const navigate = useNavigate()
    const location = useLocation()
    const params = useMemo(() => new URLSearchParams(location.search), [location.search])
    const onlyErrors = useMemo(() => params.has(ONLY_ERRORS_PARAM), [params])

    const { loading, hasNextPage, fetchMore, connection, error } = useOutboundWebhookLogsConnection(id, onlyErrors)

    return (
        <>
            <Header
                onlyErrors={onlyErrors}
                onSetOnlyErrors={onlyErrors => {
                    if (onlyErrors) {
                        params.set(ONLY_ERRORS_PARAM, 'true')
                        navigate({ search: params.toString() })
                    } else {
                        params.delete(ONLY_ERRORS_PARAM)
                        navigate({ search: params.toString() })
                    }
                }}
            />
            <ConnectionContainer>
                {error && <ConnectionError errors={[error.message]} />}
                {loading && !connection && <ConnectionLoading />}
                <ConnectionList className={styles.logs}>
                    <SiteAdminWebhookPageHeader middleColumnLabel="Event type" timeLabel="Sent at" />
                    {connection?.nodes?.map(node => {
                        // We're going to thunk OutboundWebhookLogFields into
                        // something close enough to WebhookLogFields for display
                        // purposes.
                        const display = {
                            id: node.id,
                            receivedAt: node.sentAt,
                            statusCode: node.statusCode,
                            externalService: null,
                            eventType: node.job.eventType,
                            request: node.request,
                            response: node.response ?? undefined,
                            error: node.error ?? undefined,
                        }

                        return <WebhookLogNode key={node.id} node={display} doNotShowExternalService={true} />
                    })}
                </ConnectionList>
                {connection && (
                    <SummaryContainer>
                        <ConnectionSummary
                            noSummaryIfAllNodesVisible={false}
                            centered={true}
                            connection={connection}
                            noun="webhook log"
                            pluralNoun="webhook logs"
                            hasNextPage={hasNextPage}
                            emptyElement={<EmptyList onlyErrors={onlyErrors} />}
                        />
                        {hasNextPage && <ShowMoreButton centered={true} onClick={fetchMore} />}
                    </SummaryContainer>
                )}
            </ConnectionContainer>
        </>
    )
}

const EmptyList: FC<{ onlyErrors: boolean }> = ({ onlyErrors }) => (
    <div className="text-muted text-center mb-3 w-100">
        <Icon className="icon" svgPath={mdiMapSearch} inline={false} aria-hidden={true} />
        <div className="pt-2">
            {onlyErrors
                ? 'No errors have been received from this webhook recently.'
                : 'No requests have been sent to this webhook recently.'}
        </div>
    </div>
)

interface HeaderProps {
    onlyErrors: boolean
    onSetOnlyErrors: (onlyErrors: boolean) => void
}

const Header: FC<HeaderProps> = ({ onlyErrors, onSetOnlyErrors }) => (
    <div className={styles.header}>
        <H3>Logs</H3>
        <Button variant="danger" onClick={() => onSetOnlyErrors(!onlyErrors)} outline={!onlyErrors}>
            <Icon
                className={classNames(styles.icon, onlyErrors && styles.enabled)}
                aria-hidden={true}
                svgPath={mdiAlertCircle}
            />
            <span className="ml-1">Show errors</span>
        </Button>
    </div>
)
