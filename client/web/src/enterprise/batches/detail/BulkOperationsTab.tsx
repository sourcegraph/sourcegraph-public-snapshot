import * as H from 'history'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React, { useCallback, useState } from 'react'
import { delay, repeatWhen, tap } from 'rxjs/operators'

import { BulkOperationState } from '@sourcegraph/shared/src/graphql-operations'
import { pluralize } from '@sourcegraph/shared/src/util/strings'
import { Container } from '@sourcegraph/wildcard'

import { dismissAlert } from '../../../components/DismissibleAlert'
import { FilteredConnection, FilteredConnectionQueryArguments } from '../../../components/FilteredConnection'
import { BulkOperationFields, Scalars } from '../../../graphql-operations'

import { queryBulkOperations as _queryBulkOperations } from './backend'
import { BulkOperationNode, BulkOperationNodeProps } from './bulk-operations/BulkOperationNode'

export interface BulkOperationsTabProps {
    batchChangeID: Scalars['ID']
    history: H.History
    location: H.Location

    queryBulkOperations?: typeof _queryBulkOperations
}

export const BulkOperationsTab: React.FunctionComponent<BulkOperationsTabProps> = ({
    batchChangeID,
    history,
    location,
    queryBulkOperations = _queryBulkOperations,
}) => {
    const [totalCount, setTotalCount] = useState<number>()
    const query = useCallback(
        ({ first, after }: FilteredConnectionQueryArguments) =>
            queryBulkOperations({ batchChange: batchChangeID, after: after ?? null, first: first ?? null }).pipe(
                tap(connection => {
                    setTotalCount(connection.totalCount)
                    for (const node of connection.nodes) {
                        if (node.state !== BulkOperationState.PROCESSING) {
                            // Hide alerts for bulk operations seen already.
                            // When the user visits this tab, we want to auto-dismiss notifications
                            // for failed and completed operations.
                            dismissAlert(`bulkOperation-${node.state.toLocaleLowerCase()}-${node.id}`)
                        }
                    }
                }),
                repeatWhen(notifier => notifier.pipe(delay(2000)))
            ),
        [batchChangeID, queryBulkOperations]
    )
    return (
        <>
            <BulkOperationsListHeadComponent totalCount={totalCount} />
            <Container>
                <FilteredConnection<BulkOperationFields, Omit<BulkOperationNodeProps, 'node'>>
                    nodeComponent={BulkOperationNode}
                    nodeComponentProps={{ showErrors: true }}
                    queryConnection={query}
                    hideSearch={true}
                    defaultFirst={15}
                    noun="bulk operation"
                    pluralNoun="bulk operations"
                    history={history}
                    location={location}
                    useURLQuery={true}
                    listComponent="div"
                    emptyElement={<EmptyBulkOperationsListElement />}
                    noSummaryIfAllNodesVisible={true}
                />
            </Container>
        </>
    )
}

export const EmptyBulkOperationsListElement: React.FunctionComponent<{}> = () => (
    <div className="text-muted text-center w-100 mb-3">
        <MapSearchIcon className="icon" />
        <div className="pt-2">No bulk operations have been run on this batch change.</div>
    </div>
)

export const BulkOperationsListHeadComponent: React.FunctionComponent<{ totalCount?: number }> = ({ totalCount }) => (
    <h3>
        {totalCount} changeset {pluralize('update', totalCount ?? 0)}
    </h3>
)
