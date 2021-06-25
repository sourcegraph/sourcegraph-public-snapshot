import * as H from 'history'
import React, { useState, useCallback } from 'react'
import { map, tap } from 'rxjs/operators'

import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Container } from '@sourcegraph/wildcard'

import { FilteredConnection, FilteredConnectionQueryArguments } from '../../../components/FilteredConnection'
import { RepoBatchChange, RepositoryFields, Scalars } from '../../../graphql-operations'
import { queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs } from '../detail/backend'
import { ChangesetFilterRow, ChangesetFilters } from '../detail/changesets/ChangesetFilterRow'

import { queryRepoBatchChanges as _queryRepoBatchChanges } from './backend'
import { BatchChangeNode, BatchChangeNodeProps } from './BatchChangeNode'
import styles from './RepoBatchChanges.module.scss'

interface Props extends ThemeProps {
    viewerCanAdminister: boolean
    history: H.History
    location: H.Location
    repo: RepositoryFields
    hideFilters?: boolean
    onlyArchived?: boolean

    /** For testing only. */
    queryRepoBatchChanges?: typeof _queryRepoBatchChanges
    /** For testing only. */
    queryExternalChangesetWithFileDiffs?: typeof _queryExternalChangesetWithFileDiffs
}

/**
 * A list of a batch change's changesets.
 */
export const RepoBatchChanges: React.FunctionComponent<Props> = ({
    viewerCanAdminister,
    history,
    location,
    repo,
    isLightTheme,
    hideFilters = false,
    queryRepoBatchChanges = _queryRepoBatchChanges,
    queryExternalChangesetWithFileDiffs = _queryExternalChangesetWithFileDiffs,
}) => {
    // Whether all the changesets are selected, beyond the scope of what's on screen right now.
    const [allSelected, setAllSelected] = useState<boolean>(false)
    // // The overall amount of all changesets in the connection.
    // const [totalChangesetCount, setTotalChangesetCount] = useState<number>(0)
    // All changesets that are currently in view and can be selected. That currently
    // just means they are visible.
    const [availableChangesets, setAvailableChangesets] = useState<Set<Scalars['ID']>>(new Set())
    // The list of all selected changesets. This list does not reflect the selection
    // when `allSelected` is true.
    const [selectedChangesets, setSelectedChangesets] = useState<Set<Scalars['ID']>>(new Set())

    const onSelectChangeset = useCallback((id: string, selected: boolean): void => {
        if (selected) {
            setSelectedChangesets(previous => {
                const newSet = new Set(previous).add(id)
                return newSet
            })
            return
        }
        setSelectedChangesets(previous => {
            const newSet = new Set(previous)
            newSet.delete(id)
            return newSet
        })
        setAllSelected(false)
    }, [])

    /**
     * Whether the given changeset is currently selected. Returns always true, if `allSelected` is true.
     */
    const changesetSelected = useCallback((id: Scalars['ID']): boolean => allSelected || selectedChangesets.has(id), [
        allSelected,
        selectedChangesets,
    ])

    const deselectAll = useCallback((): void => {
        setSelectedChangesets(new Set())
        setAllSelected(false)
    }, [setSelectedChangesets])

    const selectAll = useCallback((): void => {
        setSelectedChangesets(availableChangesets)
    }, [availableChangesets, setSelectedChangesets])

    // True when all in the current list are selected. It ticks the header row
    // checkbox when true.
    const allSelectedCheckboxChecked = allSelected || selectedChangesets.size === availableChangesets.size

    const toggleSelectAll = useCallback((): void => {
        if (allSelectedCheckboxChecked) {
            deselectAll()
        } else {
            selectAll()
        }
    }, [allSelectedCheckboxChecked, selectAll, deselectAll])

    // const onSelectAll = useCallback(() => {
    //     setAllSelected(true)
    // }, [])

    const [changesetFilters, setChangesetFilters] = useState<ChangesetFilters>({
        checkState: null,
        state: null,
        reviewState: null,
        search: null,
    })

    const setChangesetFiltersAndDeselectAll = useCallback(
        (filters: ChangesetFilters) => {
            deselectAll()
            setChangesetFilters(filters)
        },
        [deselectAll, setChangesetFilters]
    )

    const query = useCallback(
        (args: FilteredConnectionQueryArguments) => {
            const passedArguments = {
                name: repo.name,
                repoID: repo.id,
                first: args.first ?? null,
                after: args.after ?? null,
                search: changesetFilters.search,
                // TODO:
                // state: changesetFilters.state,
                // reviewState: changesetFilters.reviewState,
                // checkState: changesetFilters.checkState,
            }
            return queryRepoBatchChanges(passedArguments).pipe(
                tap(data => {
                    if (!data) {
                        return
                    }
                    // Available changesets are all changesets that the user can view.
                    setAvailableChangesets(
                        new Set(
                            data.batchChanges.nodes.flatMap(batchChange =>
                                batchChange.changesets.nodes.map(changeset => changeset.id)
                            )
                        )
                    )
                    // TODO:
                    // Remember the totalCount.
                    // setTotalChangesetCount(data.totalCount)
                }),
                map(data => {
                    if (!data) {
                        return {
                            totalCount: 0,
                            nodes: [],
                            pageInfo: {
                                endCursor: null,
                                hasNextPage: false,
                            },
                        }
                    }
                    return data.batchChanges
                })
            )
        },
        [queryRepoBatchChanges, changesetFilters.search, repo.id, repo.name]
    )

    // TODO:
    // const showSelectRow = viewerCanAdminister && selectedChangesets.size > 0
    const showSelectRow = false

    return (
        <Container>
            {!hideFilters && !showSelectRow && (
                <ChangesetFilterRow
                    history={history}
                    location={location}
                    onFiltersChange={setChangesetFiltersAndDeselectAll}
                    searchPlaceholderText="Search changeset title"
                />
            )}
            {/* TODO: */}
            {/* {showSelectRow && queryArguments && (
                <ChangesetSelectRow
                    batchChangeID={batchChangeID}
                    selected={selectedChangesets}
                    onSubmit={deselectAll}
                    totalCount={totalChangesetCount}
                    allVisibleSelected={allSelectedCheckboxChecked}
                    allSelected={allSelected}
                    setAllSelected={onSelectAll}
                    queryArguments={queryArguments}
                />
            )} */}
            <FilteredConnection<RepoBatchChange, Omit<BatchChangeNodeProps, 'node'>, RepoBatchChangesHeaderProps>
                history={history}
                location={location}
                nodeComponent={BatchChangeNode}
                nodeComponentProps={{
                    isLightTheme,
                    viewerCanAdminister,
                    history,
                    location,
                    queryExternalChangesetWithFileDiffs,
                    isChangesetSelected: changesetSelected,
                    onSelectChangeset,
                }}
                queryConnection={query}
                hideSearch={true}
                defaultFirst={15}
                noun="batch change"
                pluralNoun="batch changes"
                listComponent="div"
                listClassName={styles.batchChangesGrid}
                className="filtered-connection__centered-summary mt-2"
                headComponent={RepoBatchChangesHeader}
                headComponentProps={{
                    allSelected: allSelectedCheckboxChecked,
                    toggleSelectAll,
                    disabled: !viewerCanAdminister,
                }}
                cursorPaging={true}
                noSummaryIfAllNodesVisible={true}
                emptyElement={
                    <div className="w-100 py-5 text-center">
                        <p>
                            <strong>No batch changes have been created</strong>
                        </p>
                    </div>
                }
            />
        </Container>
    )
}

interface RepoBatchChangesHeaderProps {
    allSelected?: boolean
    toggleSelectAll?: () => void
    disabled?: boolean
}

export const RepoBatchChangesHeader: React.FunctionComponent<RepoBatchChangesHeaderProps> = ({
    allSelected,
    toggleSelectAll,
    disabled,
}) => (
    <>
        <span className="d-none d-md-block" />
        {toggleSelectAll && (
            <input
                type="checkbox"
                className="btn ml-2"
                checked={allSelected}
                onChange={toggleSelectAll}
                disabled={!!disabled}
                data-tooltip="Click to select all changesets"
                aria-label="Click to select all changesets"
            />
        )}
        <h5 className="p-2 d-none d-md-block text-uppercase text-center text-nowrap">Status</h5>
        <h5 className="p-2 d-none d-md-block text-uppercase text-nowrap">Changeset information</h5>
        <h5 className="p-2 d-none d-md-block text-uppercase text-center text-nowrap">Check state</h5>
        <h5 className="p-2 d-none d-md-block text-uppercase text-center text-nowrap">Review state</h5>
        <h5 className="p-2 d-none d-md-block text-uppercase text-center text-nowrap">Changes</h5>
    </>
)
