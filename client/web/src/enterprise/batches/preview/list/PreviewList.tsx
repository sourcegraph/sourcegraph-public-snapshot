import React, { useCallback, useContext, useState } from 'react'

import { mdiMagnify } from '@mdi/js'
import { tap } from 'rxjs/operators'

import { Container, Icon } from '@sourcegraph/wildcard'

import { DismissibleAlert } from '../../../../components/DismissibleAlert'
import { FilteredConnection, type FilteredConnectionQueryArguments } from '../../../../components/FilteredConnection'
import type {
    BatchSpecApplyPreviewVariables,
    ChangesetApplyPreviewFields,
    Scalars,
} from '../../../../graphql-operations'
import { MultiSelectContext } from '../../MultiSelectContext'
import { BatchChangePreviewContext } from '../BatchChangePreviewContext'
import type { PreviewPageAuthenticatedUser } from '../BatchChangePreviewPage'
import { filterPublishableIDs } from '../utils'

import {
    queryChangesetApplyPreview as _queryChangesetApplyPreview,
    type queryChangesetSpecFileDiffs,
    type queryPublishableChangesetSpecIDs as _queryPublishableChangesetSpecIDs,
} from './backend'
import { ChangesetApplyPreviewNode, type ChangesetApplyPreviewNodeProps } from './ChangesetApplyPreviewNode'
import { EmptyPreviewListElement } from './EmptyPreviewListElement'
import { PreviewFilterRow } from './PreviewFilterRow'
import { PreviewListHeader, type PreviewListHeaderProps } from './PreviewListHeader'
import { PreviewSelectRow } from './PreviewSelectRow'

import styles from './PreviewList.module.scss'

interface Props {
    batchSpecID: Scalars['ID']
    authenticatedUser: PreviewPageAuthenticatedUser

    /** For testing only. */
    queryChangesetApplyPreview?: typeof _queryChangesetApplyPreview
    /** For testing only. */
    queryChangesetSpecFileDiffs?: typeof queryChangesetSpecFileDiffs
    /** Expand changeset descriptions, for testing only. */
    expandChangesetDescriptions?: boolean
    /** For testing only. */
    queryPublishableChangesetSpecIDs?: typeof _queryPublishableChangesetSpecIDs
}

/**
 * A list of a batch spec's preview nodes.
 */
export const PreviewList: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    batchSpecID,
    authenticatedUser,

    queryChangesetApplyPreview = _queryChangesetApplyPreview,
    queryChangesetSpecFileDiffs,
    expandChangesetDescriptions,
    queryPublishableChangesetSpecIDs,
}) => {
    const { selected, areAllVisibleSelected, isSelected, toggleSingle, toggleVisible, setVisible } =
        useContext(MultiSelectContext)
    // The user can modify the desired publication states for changesets in this preview
    // list from the UI. However, these modifications are transient and are not persisted
    // to the backend (until the user applies the batch change and the publication states
    // are realized, of course). Rather, they are provided as arguments to the
    // `applyPreview` connection, and later the `applyBatchChange` mutation, in order to
    // override the original publication states computed by the reconciler on the backend.
    // `BatchChangePreviewContext` is responsible for managing these publication states,
    // as well as filter arguments to the connection query, clientside.
    const { filters, filtersChanged, setFiltersChanged, publicationStates, resolveRecalculationUpdates } =
        useContext(BatchChangePreviewContext)

    const [queryArguments, setQueryArguments] = useState<BatchSpecApplyPreviewVariables>()

    const queryChangesetApplyPreviewConnection = useCallback(
        (args: FilteredConnectionQueryArguments) => {
            const passedArguments = {
                first: args.first ?? null,
                after: args.after ?? null,
                batchSpec: batchSpecID,
                search: filters.search,
                currentState: filters.currentState,
                action: filters.action,
                publicationStates,
            }
            return queryChangesetApplyPreview(passedArguments).pipe(
                tap(data => {
                    // Store the query arguments used for the current connection.
                    setQueryArguments(passedArguments)
                    // Available changeset specs are all changesets specs that a user can
                    // modify the publication state of from the UI.
                    setVisible(filtersChanged, filterPublishableIDs(data.nodes))
                    if (filtersChanged) {
                        setFiltersChanged(false)
                    }
                    // If we re-queried on account of any publication states changing, make
                    // sure to mark the timestamp record for this recalculation event as
                    // complete so that it produces a banner.
                    resolveRecalculationUpdates()
                })
            )
        },
        [
            filtersChanged,
            setFiltersChanged,
            batchSpecID,
            filters.search,
            filters.currentState,
            filters.action,
            queryChangesetApplyPreview,
            setVisible,
            publicationStates,
            resolveRecalculationUpdates,
        ]
    )

    const showSelectRow = selected === 'all' || selected.size > 0

    return (
        <Container role="region" aria-label="preview changesets">
            {showSelectRow && queryArguments ? (
                <PreviewSelectRow
                    queryPublishableChangesetSpecIDs={queryPublishableChangesetSpecIDs}
                    queryArguments={queryArguments}
                />
            ) : (
                <PreviewFilterRow />
            )}
            <PublicationStatesUpdateAlerts />
            <FilteredConnection<
                ChangesetApplyPreviewFields,
                Omit<ChangesetApplyPreviewNodeProps, 'node'>,
                PreviewListHeaderProps
            >
                className="mt-2"
                nodeComponent={ChangesetApplyPreviewNode}
                nodeComponentProps={{
                    authenticatedUser,
                    queryChangesetSpecFileDiffs,
                    expandChangesetDescriptions,
                    selectable: { onSelect: toggleSingle, isSelected },
                }}
                queryConnection={queryChangesetApplyPreviewConnection}
                hideSearch={true}
                defaultFirst={15}
                noun="changeset"
                pluralNoun="changesets"
                useURLQuery={true}
                listClassName={styles.previewListGrid}
                headComponent={PreviewListHeader}
                headComponentProps={{
                    allSelected: showSelectRow && areAllVisibleSelected(),
                    toggleSelectAll: toggleVisible,
                }}
                cursorPaging={true}
                noSummaryIfAllNodesVisible={true}
                emptyElement={
                    filters.search || filters.currentState || filters.action ? (
                        <EmptyPreviewSearchElement />
                    ) : (
                        <EmptyPreviewListElement />
                    )
                }
            />
        </Container>
    )
}

const EmptyPreviewSearchElement: React.FunctionComponent<React.PropsWithChildren<{}>> = () => (
    <div className="text-muted row w-100">
        <div className="col-12 text-center">
            <Icon className="icon" svgPath={mdiMagnify} inline={false} aria-hidden={true} />
            <div className="pt-2">No changesets matched the search.</div>
        </div>
    </div>
)

/**
 * A list of none to many dismissible alerts, one for each time the publication state
 * actions are recalculated when the user modifies the publication states for preview
 * changesets.
 */
const PublicationStatesUpdateAlerts: React.FunctionComponent<React.PropsWithChildren<{}>> = () => {
    // `BatchChangePreviewContext` keeps a record of each time the user modifies the
    // desired publication states for changesets in the preview list from the UI.
    const { recalculationUpdates } = useContext(BatchChangePreviewContext)

    return (
        <div className="mt-2">
            {recalculationUpdates.map(([timestamp, status]) =>
                // Wait to show publication state update alerts until the connection query
                // request resolves.
                status === 'complete' ? (
                    <DismissibleAlert variant="success" key={timestamp}>
                        Publication state actions were recalculated.
                    </DismissibleAlert>
                ) : null
            )}
        </div>
    )
}
