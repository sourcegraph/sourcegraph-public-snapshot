import * as H from 'history'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React, { useCallback } from 'react'
import { delay, repeatWhen, tap } from 'rxjs/operators'

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
    const query = useCallback(
        ({ first, after }: FilteredConnectionQueryArguments) =>
            queryBulkOperations({ batchChange: batchChangeID, after: after ?? null, first: first ?? null }).pipe(
                tap(connection => {
                    for (const node of connection.nodes) {
                        // Hide alerts for bulk operations seen already.
                        dismissAlert(`bulkOperation-${node.state.toLocaleLowerCase()}-${node.id}`)
                    }
                }),
                repeatWhen(notifier => notifier.pipe(delay(2000)))
            ),
        [batchChangeID, queryBulkOperations]
    )
    return (
        <FilteredConnection<BulkOperationFields, Omit<BulkOperationNodeProps, 'node'>>
            className="mt-2"
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
            listClassName="mb-3"
            emptyElement={<EmptyBulkOperationsListElement />}
            noSummaryIfAllNodesVisible={true}
            headComponent={BulkOperationsListHeadComponent}
        />
    )
}

export const EmptyBulkOperationsListElement: React.FunctionComponent<{}> = () => (
    <div className="text-muted my-4 pt-4 text-center">
        <MapSearchIcon className="icon" />
        <div className="pt-2">No bulk operations have been run on this batch change.</div>
    </div>
)

export const BulkOperationsListHeadComponent: React.FunctionComponent<{ totalCount?: number | null }> = ({
    totalCount,
}) => <h3 className="mt-4">{totalCount} changeset updates</h3>
