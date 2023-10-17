import React, { useState, useMemo, useCallback } from 'react'

import { subDays } from 'date-fns'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import { useParams } from 'react-router-dom'

import { type ErrorLike, isErrorLike } from '@sourcegraph/common'
import { PageHeader, LoadingSpinner, useObservable } from '@sourcegraph/wildcard'

import { BatchChangesIcon } from '../../../batches/icons'
import { CreatedByAndUpdatedByInfoByline } from '../../../components/Byline/CreatedByAndUpdatedByInfoByline'
import { HeroPage } from '../../../components/HeroPage'
import { PageTitle } from '../../../components/PageTitle'
import type { BatchChangeChangesetsResult, Scalars } from '../../../graphql-operations'
import {
    type queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs,
    type queryChangesets as _queryChangesets,
    fetchBatchChangeByNamespace as _fetchBatchChangeByNamespace,
} from '../detail/backend'

import type { closeBatchChange as _closeBatchChange } from './backend'
import { BatchChangeCloseAlert } from './BatchChangeCloseAlert'
import { BatchChangeCloseChangesetsList } from './BatchChangeCloseChangesetsList'

export interface BatchChangeClosePageProps {
    /**
     * The namespace ID.
     */
    namespaceID: Scalars['ID']

    /** For testing only. */
    fetchBatchChangeByNamespace?: typeof _fetchBatchChangeByNamespace
    /** For testing only. */
    queryChangesets?: typeof _queryChangesets
    /** For testing only. */
    queryExternalChangesetWithFileDiffs?: typeof _queryExternalChangesetWithFileDiffs
    /** For testing only. */
    closeBatchChange?: typeof _closeBatchChange
}

export const BatchChangeClosePage: React.FunctionComponent<React.PropsWithChildren<BatchChangeClosePageProps>> = ({
    namespaceID,
    fetchBatchChangeByNamespace = _fetchBatchChangeByNamespace,
    queryChangesets,
    queryExternalChangesetWithFileDiffs,
    closeBatchChange,
}) => {
    const { batchChangeName } = useParams()
    const [closeChangesets, setCloseChangesets] = useState<boolean>(false)
    const createdAfter = useMemo(() => subDays(new Date(), 3).toISOString(), [])
    const batchChange = useObservable(
        useMemo(
            () => fetchBatchChangeByNamespace(namespaceID, batchChangeName!, createdAfter),
            [fetchBatchChangeByNamespace, namespaceID, batchChangeName, createdAfter]
        )
    )

    const [totalCount, setTotalCount] = useState<number>()

    const onFetchChangesets = useCallback(
        (
            connection?: (BatchChangeChangesetsResult['node'] & { __typename: 'BatchChange' })['changesets'] | ErrorLike
        ) => {
            if (!connection || isErrorLike(connection)) {
                return
            }
            setTotalCount(connection.totalCount)
        },
        []
    )

    // Is loading.
    if (batchChange === undefined) {
        return (
            <div className="text-center">
                <LoadingSpinner className="mx-auto my-4" />
            </div>
        )
    }

    // Batch change not found.
    if (batchChange === null) {
        return <HeroPage icon={AlertCircleIcon} title="Batch change not found" />
    }

    return (
        <>
            <PageTitle title="Preview close" />
            <PageHeader
                path={[
                    {
                        icon: BatchChangesIcon,
                        to: '/batch-changes',
                    },
                    { to: `${batchChange.namespace.url}/batch-changes`, text: batchChange.namespace.namespaceName },
                    { text: batchChange.name },
                ]}
                byline={
                    <CreatedByAndUpdatedByInfoByline
                        createdAt={batchChange.createdAt}
                        createdBy={batchChange.creator}
                        updatedAt={batchChange.lastAppliedAt}
                        updatedBy={batchChange.lastApplier}
                    />
                }
                className="test-batch-change-close-page mb-3"
            />
            {totalCount !== undefined && (
                <BatchChangeCloseAlert
                    batchChangeID={batchChange.id}
                    batchChangeURL={batchChange.url}
                    closeChangesets={closeChangesets}
                    setCloseChangesets={setCloseChangesets}
                    closeBatchChange={closeBatchChange}
                    viewerCanAdminister={batchChange.viewerCanAdminister}
                    totalCount={totalCount}
                />
            )}
            <BatchChangeCloseChangesetsList
                batchChangeID={batchChange.id}
                viewerCanAdminister={batchChange.viewerCanAdminister}
                queryChangesets={queryChangesets}
                queryExternalChangesetWithFileDiffs={queryExternalChangesetWithFileDiffs}
                willClose={closeChangesets}
                onUpdate={onFetchChangesets}
            />
        </>
    )
}
