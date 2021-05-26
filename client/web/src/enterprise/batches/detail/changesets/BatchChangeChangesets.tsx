import classNames from 'classnames'
import * as H from 'history'
import React, { useState, useCallback, useMemo, useEffect } from 'react'
import { Subject } from 'rxjs'
import { repeatWhen, delay, withLatestFrom, map, filter, tap } from 'rxjs/operators'

import { createHoverifier } from '@sourcegraph/codeintellify'
import { ActionItemAction } from '@sourcegraph/shared/src/actions/ActionItem'
import { HoverMerged } from '@sourcegraph/shared/src/api/client/types/hover'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { getHoverActions } from '@sourcegraph/shared/src/hover/actions'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { property, isDefined } from '@sourcegraph/shared/src/util/types'
import { RepoSpec, RevisionSpec, FileSpec, ResolvedRevisionSpec } from '@sourcegraph/shared/src/util/url'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { Container } from '@sourcegraph/wildcard'

import { getHover, getDocumentHighlights } from '../../../../backend/features'
import { FilteredConnection, FilteredConnectionQueryArguments } from '../../../../components/FilteredConnection'
import { WebHoverOverlay } from '../../../../components/shared'
import { AllChangesetIDsVariables, ChangesetFields, Scalars } from '../../../../graphql-operations'
import { getLSPTextDocumentPositionParameters } from '../../utils'
import {
    queryChangesets as _queryChangesets,
    queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs,
} from '../backend'

import styles from './BatchChangeChangesets.module.scss'
import { BatchChangeChangesetsHeader, BatchChangeChangesetsHeaderProps } from './BatchChangeChangesetsHeader'
import { ChangesetFilters, ChangesetFilterRow } from './ChangesetFilterRow'
import { ChangesetNodeProps, ChangesetNode } from './ChangesetNode'
import { ChangesetSelectRow } from './ChangesetSelectRow'
import { EmptyArchivedChangesetListElement } from './EmptyArchivedChangesetListElement'
import { EmptyChangesetListElement } from './EmptyChangesetListElement'
import { EmptyChangesetSearchElement } from './EmptyChangesetSearchElement'

interface Props extends ThemeProps, PlatformContextProps, TelemetryProps, ExtensionsControllerProps {
    batchChangeID: Scalars['ID']
    viewerCanAdminister: boolean
    history: H.History
    location: H.Location

    hideFilters?: boolean
    onlyArchived?: boolean

    /** For testing only. */
    queryChangesets?: typeof _queryChangesets
    /** For testing only. */
    queryExternalChangesetWithFileDiffs?: typeof _queryExternalChangesetWithFileDiffs
    /** For testing only. */
    expandByDefault?: boolean
}

/**
 * A list of a batch change's changesets.
 */
export const BatchChangeChangesets: React.FunctionComponent<Props> = ({
    batchChangeID,
    viewerCanAdminister,
    history,
    location,
    isLightTheme,
    extensionsController,
    platformContext,
    telemetryService,
    hideFilters = false,
    queryChangesets = _queryChangesets,
    queryExternalChangesetWithFileDiffs,
    expandByDefault,
    onlyArchived,
}) => {
    // Whether all the changesets are selected, beyond the scope of what's on screen right now.
    const [allSelected, setAllSelected] = useState<boolean>(false)
    // The overall amount of all changesets in the connection.
    const [totalChangesetCount, setTotalChangesetCount] = useState<number>(0)
    // All changesets that are currently in view and can be selected. That currently
    // just means they are visible.
    const [availableChangesets, setAvailableChangesets] = useState<Set<Scalars['ID']>>(new Set())
    // The list of all selected changesets. This list does not reflect the selection
    // when `allSelected` is true.
    const [selectedChangesets, setSelectedChangesets] = useState<Set<Scalars['ID']>>(new Set())

    const onSelect = useCallback((id: string, selected: boolean): void => {
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

    const onSelectAll = useCallback(() => {
        setAllSelected(true)
    }, [])

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

    const [queryArguments, setQueryArguments] = useState<Omit<AllChangesetIDsVariables, 'after'>>()

    const queryChangesetsConnection = useCallback(
        (args: FilteredConnectionQueryArguments) => {
            const passedArguments = {
                state: changesetFilters.state,
                reviewState: changesetFilters.reviewState,
                checkState: changesetFilters.checkState,
                first: args.first ?? null,
                after: args.after ?? null,
                batchChange: batchChangeID,
                onlyPublishedByThisBatchChange: null,
                search: changesetFilters.search,
                onlyArchived: !!onlyArchived,
            }
            return queryChangesets(passedArguments)
                .pipe(
                    tap(data => {
                        // Store the query arguments used for the current connection.
                        setQueryArguments(passedArguments)
                        // Available changesets are all changesets that the user
                        // can view.
                        setAvailableChangesets(
                            new Set(
                                data.nodes.filter(node => node.__typename === 'ExternalChangeset').map(node => node.id)
                            )
                        )
                        // Remember the totalCount.
                        setTotalChangesetCount(data.totalCount)
                    })
                )
                .pipe(repeatWhen(notifier => notifier.pipe(delay(5000))))
        },
        [
            batchChangeID,
            changesetFilters.state,
            changesetFilters.reviewState,
            changesetFilters.checkState,
            changesetFilters.search,
            queryChangesets,
            onlyArchived,
        ]
    )

    const containerElements = useMemo(() => new Subject<HTMLElement | null>(), [])
    const nextContainerElement = useMemo(() => containerElements.next.bind(containerElements), [containerElements])

    const hoverOverlayElements = useMemo(() => new Subject<HTMLElement | null>(), [])
    const nextOverlayElement = useCallback((element: HTMLElement | null): void => hoverOverlayElements.next(element), [
        hoverOverlayElements,
    ])

    const closeButtonClicks = useMemo(() => new Subject<MouseEvent>(), [])
    const nextCloseButtonClick = useCallback((event: MouseEvent): void => closeButtonClicks.next(event), [
        closeButtonClicks,
    ])

    const componentRerenders = useMemo(() => new Subject<void>(), [])

    const hoverifier = useMemo(
        () =>
            createHoverifier<RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec, HoverMerged, ActionItemAction>({
                closeButtonClicks,
                hoverOverlayElements,
                hoverOverlayRerenders: componentRerenders.pipe(
                    withLatestFrom(hoverOverlayElements, containerElements),
                    map(([, hoverOverlayElement, relativeElement]) => ({
                        hoverOverlayElement,
                        // The root component element is guaranteed to be rendered after a componentDidUpdate
                        // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
                        relativeElement: relativeElement!,
                    })),
                    // Can't reposition HoverOverlay if it wasn't rendered
                    filter(property('hoverOverlayElement', isDefined))
                ),
                getHover: hoveredToken =>
                    getHover(getLSPTextDocumentPositionParameters(hoveredToken), { extensionsController }),
                getDocumentHighlights: hoveredToken =>
                    getDocumentHighlights(getLSPTextDocumentPositionParameters(hoveredToken), { extensionsController }),
                getActions: context => getHoverActions({ extensionsController, platformContext }, context),
                pinningEnabled: true,
            }),
        [
            closeButtonClicks,
            containerElements,
            extensionsController,
            hoverOverlayElements,
            platformContext,
            componentRerenders,
        ]
    )
    useEffect(() => () => hoverifier.unsubscribe(), [hoverifier])

    const hoverState = useObservable(useMemo(() => hoverifier.hoverStateUpdates, [hoverifier]))
    useEffect(() => {
        componentRerenders.next()
    }, [componentRerenders, hoverState])

    const showSelectRow = viewerCanAdminister && selectedChangesets.size > 0

    return (
        <Container>
            {!hideFilters && !showSelectRow && (
                <ChangesetFilterRow
                    history={history}
                    location={location}
                    onFiltersChange={setChangesetFiltersAndDeselectAll}
                />
            )}
            {showSelectRow && queryArguments && (
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
            )}
            <div className="list-group position-relative" ref={nextContainerElement}>
                <FilteredConnection<ChangesetFields, Omit<ChangesetNodeProps, 'node'>, BatchChangeChangesetsHeaderProps>
                    nodeComponent={ChangesetNode}
                    nodeComponentProps={{
                        isLightTheme,
                        viewerCanAdminister,
                        history,
                        location,
                        extensionInfo: { extensionsController, hoverifier },
                        expandByDefault,
                        queryExternalChangesetWithFileDiffs,
                        onSelect,
                        isSelected: changesetSelected,
                    }}
                    queryConnection={queryChangesetsConnection}
                    hideSearch={true}
                    defaultFirst={15}
                    noun="changeset"
                    pluralNoun="changesets"
                    history={history}
                    location={location}
                    useURLQuery={true}
                    listComponent="div"
                    listClassName={styles.batchChangeChangesetsGrid}
                    headComponent={BatchChangeChangesetsHeader}
                    headComponentProps={{
                        allSelected: allSelectedCheckboxChecked,
                        toggleSelectAll,
                        disabled: !viewerCanAdminister,
                    }}
                    // Only show the empty element, if no filters are selected.
                    emptyElement={
                        filtersSelected(changesetFilters) ? (
                            <EmptyChangesetSearchElement />
                        ) : onlyArchived ? (
                            <EmptyArchivedChangesetListElement />
                        ) : (
                            <EmptyChangesetListElement />
                        )
                    }
                    noSummaryIfAllNodesVisible={true}
                />
                {hoverState?.hoverOverlayProps && (
                    <WebHoverOverlay
                        {...hoverState.hoverOverlayProps}
                        telemetryService={telemetryService}
                        extensionsController={extensionsController}
                        isLightTheme={isLightTheme}
                        location={location}
                        platformContext={platformContext}
                        hoverRef={nextOverlayElement}
                        onCloseButtonClick={nextCloseButtonClick}
                    />
                )}
            </div>
        </Container>
    )
}

/**
 * Returns true, if any filter is selected.
 */
function filtersSelected(filters: ChangesetFilters): boolean {
    return filters.checkState !== null || filters.state !== null || filters.reviewState !== null || !!filters.search
}
