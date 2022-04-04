import React, { useCallback, useContext, useState } from 'react'

import * as H from 'history'
import MagnifyIcon from 'mdi-react/MagnifyIcon'
import { tap } from 'rxjs/operators'

import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Container } from '@sourcegraph/wildcard'

import { DismissibleAlert } from '../../../../components/DismissibleAlert'
import { FilteredConnection, FilteredConnectionQueryArguments } from '../../../../components/FilteredConnection'
import { BatchSpecApplyPreviewVariables, ChangesetApplyPreviewFields, Scalars } from '../../../../graphql-operations'
import { MultiSelectContext } from '../../MultiSelectContext'
import { BatchChangePreviewContext } from '../BatchChangePreviewContext'
import { PreviewPageAuthenticatedUser } from '../BatchChangePreviewPage'
import { filterPublishableIDs } from '../utils'

import {
    queryChangesetApplyPreview as _queryChangesetApplyPreview,
    queryChangesetSpecFileDiffs,
    queryPublishableChangesetSpecIDs as _queryPublishableChangesetSpecIDs,
} from './backend'
import { ChangesetApplyPreviewNode, ChangesetApplyPreviewNodeProps } from './ChangesetApplyPreviewNode'
import { EmptyPreviewListElement } from './EmptyPreviewListElement'
import { PreviewFilterRow } from './PreviewFilterRow'
import { PreviewListHeader, PreviewListHeaderProps } from './PreviewListHeader'
import { PreviewSelectRow } from './PreviewSelectRow'

import styles from './PreviewList.module.scss'

interface Props extends ThemeProps {
    batchSpecID: Scalars['ID']
    history: H.History
    location: H.Location
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
export const PreviewList: React.FunctionComponent<Props> = ({
    batchSpecID,
    history,
    location,
    authenticatedUser,
    isLightTheme,

    queryChangesetApplyPreview = _queryChangesetApplyPreview,
    queryChangesetSpecFileDiffs,
    expandChangesetDescriptions,
    queryPublishableChangesetSpecIDs,
}) => {
    const { selected, areAllVisibleSelected, isSelected, toggleSingle, toggleVisible, setVisible } = useContext(
        MultiSelectContext
    )
    const { filters, publicationStates, addRecalculationUpdate } = useContext(BatchChangePreviewContext)

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
                    setVisible(filterPublishableIDs(data.nodes))
                })
            )
        },
        [
            batchSpecID,
            filters.search,
            filters.currentState,
            filters.action,
            queryChangesetApplyPreview,
            setVisible,
            publicationStates,
        ]
    )

    // Every subsequent query after the first will have its success time recorded
    const [isInitialQuery, setIsInitialQuery] = useState(true)
    const onUpdate = useCallback(() => {
        if (isInitialQuery) {
            setIsInitialQuery(false)
        } else {
            addRecalculationUpdate(new Date())
        }
    }, [addRecalculationUpdate, isInitialQuery])

    const showSelectRow = selected === 'all' || selected.size > 0

    return (
        <Container>
            {showSelectRow && queryArguments ? (
                <PreviewSelectRow
                    queryPublishableChangesetSpecIDs={queryPublishableChangesetSpecIDs}
                    queryArguments={queryArguments}
                />
            ) : (
                <PreviewFilterRow history={history} location={location} />
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
                    isLightTheme,
                    history,
                    location,
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
                history={history}
                location={location}
                useURLQuery={true}
                listComponent="div"
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
                onUpdate={onUpdate}
            />
        </Container>
    )
}

const EmptyPreviewSearchElement: React.FunctionComponent<{}> = () => (
    <div className="text-muted row w-100">
        <div className="col-12 text-center">
            <MagnifyIcon className="icon" />
            <div className="pt-2">No changesets matched the search.</div>
        </div>
    </div>
)

/**
 * A list of none to many dismissible alerts, one for each time the publication state
 * actions are recalculated when the user modifies the publication states for preview
 * changesets.
 */
const PublicationStatesUpdateAlerts: React.FunctionComponent<{}> = () => {
    const { recalculationUpdates } = useContext(BatchChangePreviewContext)

    return (
        <div className="mt-2">
            {recalculationUpdates.map(timestamp => (
                <DismissibleAlert variant="success" key={timestamp}>
                    Publication state actions were recalculated.
                </DismissibleAlert>
            ))}
        </div>
    )
}
