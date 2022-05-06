import React, { useEffect } from 'react'

import MapSearchIcon from 'mdi-react/MapSearchIcon'

import { dataOrThrowErrors } from '@sourcegraph/http-client'
import { BulkOperationState } from '@sourcegraph/shared/src/graphql-operations'
import { Container } from '@sourcegraph/wildcard'

import { dismissAlert } from '../../../components/DismissibleAlert'
import { useConnection, UseConnectionResult } from '../../../components/FilteredConnection/hooks/useConnection'
import {
    ConnectionContainer,
    ConnectionError,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
} from '../../../components/FilteredConnection/ui'
import {
    BatchChangeBulkOperationsResult,
    BatchChangeBulkOperationsVariables,
    BulkOperationFields,
    Scalars,
} from '../../../graphql-operations'

import { BULK_OPERATIONS } from './backend'
import { BulkOperationNode } from './bulk-operations/BulkOperationNode'

export interface BulkOperationsTabProps {
    batchChangeID: Scalars['ID']
}

export const BulkOperationsTab: React.FunctionComponent<React.PropsWithChildren<BulkOperationsTabProps>> = ({
    batchChangeID,
}) => {
    const { connection, error, loading, fetchMore, hasNextPage } = useBulkOperationsListConnection(batchChangeID)

    return (
        <Container>
            <ConnectionContainer>
                {error && <ConnectionError errors={[error.message]} />}
                <ConnectionList className="list-group list-group-flush">
                    {connection?.nodes?.map(node => (
                        <BulkOperationNode key={node.id} node={node} />
                    ))}
                </ConnectionList>
                {loading && <ConnectionLoading />}
                {connection && (
                    <SummaryContainer centered={true}>
                        <ConnectionSummary
                            noSummaryIfAllNodesVisible={true}
                            first={BATCH_COUNT}
                            connection={connection}
                            noun="bulk operation"
                            pluralNoun="bulk operations"
                            hasNextPage={hasNextPage}
                            emptyElement={<EmptyBulkOperationsListElement />}
                        />
                        {hasNextPage && <ShowMoreButton onClick={fetchMore} />}
                    </SummaryContainer>
                )}
            </ConnectionContainer>
        </Container>
    )
}

const EmptyBulkOperationsListElement: React.FunctionComponent<React.PropsWithChildren<{}>> = () => (
    <div className="text-muted text-center mb-3 w-100">
        <MapSearchIcon className="icon" />
        <div className="pt-2">No bulk operations have been run on this batch change.</div>
    </div>
)

const BATCH_COUNT = 15

const useBulkOperationsListConnection = (batchChangeID: Scalars['ID']): UseConnectionResult<BulkOperationFields> => {
    const { connection, startPolling, stopPolling, ...rest } = useConnection<
        BatchChangeBulkOperationsResult,
        BatchChangeBulkOperationsVariables,
        BulkOperationFields
    >({
        query: BULK_OPERATIONS,
        variables: {
            batchChange: batchChangeID,
            after: null,
            first: BATCH_COUNT,
        },
        options: {
            useURL: true,
            fetchPolicy: 'cache-and-network',
        },
        getConnection: result => {
            const data = dataOrThrowErrors(result)

            if (!data.node) {
                throw new Error(`Batch change with ID ${batchChangeID} does not exist`)
            }
            if (data.node.__typename !== 'BatchChange') {
                throw new Error(`The given ID is a ${data.node.__typename as string}, not a BatchChange`)
            }
            return data.node.bulkOperations
        },
    })

    useEffect(() => {
        if (connection?.nodes?.length) {
            // Filter to bulk operations that are done running.
            const finishedNodes = connection.nodes.filter(node => node.state !== BulkOperationState.PROCESSING)

            // If any operations are still actively running, poll for updates.
            if (finishedNodes.length < connection.nodes.length) {
                startPolling(2000)
            } else {
                stopPolling()
            }

            // Automatically dismiss alerts for bulk operations once they have been viewed.
            for (const node of finishedNodes) {
                dismissAlert(`bulkOperation-${node.state.toLocaleLowerCase()}-${node.id}`)
            }
        }
    }, [connection, startPolling, stopPolling])

    return { connection, startPolling, stopPolling, ...rest }
}
