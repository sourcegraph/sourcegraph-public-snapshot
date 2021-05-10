import * as H from 'history'
import React, { useCallback } from 'react'
import { delay, repeatWhen } from 'rxjs/operators'

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
                repeatWhen(notifier => notifier.pipe(delay(5000)))
            ),
        [batchChangeID, queryBulkOperations]
    )
    return (
        <FilteredConnection<BulkOperationFields, Omit<BulkOperationNodeProps, 'node'>>
            className="mt-2"
            nodeComponent={BulkOperationNode}
            nodeComponentProps={{}}
            queryConnection={query}
            hideSearch={true}
            defaultFirst={15}
            noun="bulk operation"
            pluralNoun="bulk operations"
            history={history}
            location={location}
            useURLQuery={true}
            listComponent="div"
            // listClassName={classNames(styles.batchChangeChangesetsGridWithCheckboxes, 'mb-3')}
            // headComponent={BatchChangeChangesetsHeader}
            // headComponentProps={{
            //     allSelected,
            //     toggleSelectAll,
            //     disabled: !(viewerCanAdminister && !isSubmittingSelected),
            // }}
            // Only show the empty element, if no filters are selected.
            // emptyElement={
            //     filtersSelected(changesetFilters) ? (
            //         <EmptyChangesetSearchElement />
            //     ) : onlyArchived ? (
            //         <EmptyArchivedChangesetListElement />
            //     ) : (
            //         <EmptyChangesetListElement />
            //     )
            // }
            noSummaryIfAllNodesVisible={true}
        />
    )
}
