import React from 'react'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { Button, Modal, H3, Text } from '@sourcegraph/wildcard'

import {
    ConnectionContainer,
    ConnectionError,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
} from '../../../components/FilteredConnection/ui'
import type { ExecutorSecretAccessLogFields, Scalars } from '../../../graphql-operations'
import { PersonLink } from '../../../person/PersonLink'

import { useExecutorSecretAccessLogsConnection } from './backend'

export interface SecretAccessLogsModalProps {
    onCancel: () => void
    secretID: Scalars['ID']
}

export const SecretAccessLogsModal: React.FunctionComponent<React.PropsWithChildren<SecretAccessLogsModalProps>> = ({
    onCancel,
    secretID,
}) => {
    const labelId = 'secretAccessLogs'

    const { loading, hasNextPage, fetchMore, connection, error } = useExecutorSecretAccessLogsConnection(secretID)

    return (
        <Modal onDismiss={onCancel} aria-labelledby={labelId}>
            <H3 id={labelId}>Access logs</H3>
            <Text>All events when the value of this secret was read.</Text>
            <ConnectionContainer className="mb-3">
                {error && <ConnectionError errors={[error.message]} />}
                {loading && !connection && <ConnectionLoading />}
                <ConnectionList as="ul" className="list-group" aria-label="Executor secret access logs">
                    {connection?.nodes?.map(node => (
                        <ExecutorSecretAccessLogNode key={node.id} node={node} />
                    ))}
                </ConnectionList>
                {connection && (
                    <SummaryContainer className="mt-2">
                        <ConnectionSummary
                            noSummaryIfAllNodesVisible={true}
                            first={15}
                            centered={true}
                            connection={connection}
                            noun="access log"
                            pluralNoun="access logs"
                            hasNextPage={hasNextPage}
                        />
                        {hasNextPage && <ShowMoreButton centered={true} onClick={fetchMore} />}
                    </SummaryContainer>
                )}
            </ConnectionContainer>
            <div className="d-flex justify-content-end">
                <Button onClick={onCancel} outline={true} variant="secondary">
                    Close
                </Button>
            </div>
        </Modal>
    )
}

interface ExecutorSecretNodeProps {
    node: ExecutorSecretAccessLogFields
}

const ExecutorSecretAccessLogNode: React.FunctionComponent<React.PropsWithChildren<ExecutorSecretNodeProps>> = ({
    node,
}) => (
    <li className="list-group-item">
        <div className="d-flex justify-content-between align-items-center flex-wrap mb-0">
            <PersonLink
                person={{
                    // empty strings are fine here, as they are only used when `user` is not null
                    displayName: (node.user?.displayName || node.user?.username) ?? '',
                    email: node.user?.email ?? '',
                    user: node.user,
                }}
            />
            <Timestamp date={node.createdAt} />
        </div>
    </li>
)
