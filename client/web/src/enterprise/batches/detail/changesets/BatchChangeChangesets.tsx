import * as H from 'history'
import React, { useState, useCallback, useMemo, useEffect } from 'react'
import { Subject } from 'rxjs'
import { repeatWhen, delay, withLatestFrom, map, filter } from 'rxjs/operators'

import { createHoverifier } from '@sourcegraph/codeintellify'
import { ActionItemAction } from '@sourcegraph/shared/src/actions/ActionItem'
import { HoverMerged } from '@sourcegraph/shared/src/api/client/types/hover'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { getHoverActions } from '@sourcegraph/shared/src/hover/actions'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { asError } from '@sourcegraph/shared/src/util/errors'
import { pluralize } from '@sourcegraph/shared/src/util/strings'
import { property, isDefined } from '@sourcegraph/shared/src/util/types'
import { RepoSpec, RevisionSpec, FileSpec, ResolvedRevisionSpec } from '@sourcegraph/shared/src/util/url'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { getHover, getDocumentHighlights } from '../../../../backend/features'
import { FilteredConnection, FilteredConnectionQueryArguments } from '../../../../components/FilteredConnection'
import { WebHoverOverlay } from '../../../../components/shared'
import { ChangesetFields, Scalars } from '../../../../graphql-operations'
import { getLSPTextDocumentPositionParameters } from '../../utils'
import {
    detachChangesets,
    queryChangesets as _queryChangesets,
    queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs,
} from '../backend'

import { BatchChangeChangesetsHeader, BatchChangeChangesetsHeaderWithCheckboxes } from './BatchChangeChangesetsHeader'
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

    enableSelect?: boolean

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
    enableSelect,
}) => {
    const [selectedChangesets, setSelectedChangesets] = useState<Set<string>>(new Set())
    const onSelect = useCallback(
        (id: string, selected: boolean): void => {
            if (selected) {
                setSelectedChangesets(previous => new Set(previous).add(id))
            } else {
                setSelectedChangesets(previous => {
                    const newSet = new Set(previous)
                    newSet.delete(id)
                    return newSet
                })
            }
        },
        [setSelectedChangesets]
    )

    const deselectAll = useCallback((): void => setSelectedChangesets(new Set()), [setSelectedChangesets])
    const changesetSelected = useCallback((id: string): boolean => selectedChangesets.has(id), [selectedChangesets])

    const [isSubmittingSelected, setIsSubmittingSelected] = useState<boolean | Error>(false)
    const onSubmitSelected = useCallback(async () => {
        if (
            !confirm(
                `Are you sure you want to detach ${selectedChangesets.size} ${pluralize(
                    'changeset',
                    selectedChangesets.size
                )}?`
            )
        ) {
            return
        }
        setIsSubmittingSelected(true)
        try {
            await detachChangesets(batchChangeID, [...selectedChangesets])
            deselectAll()
            telemetryService.logViewEvent('BatchChangeDetailsPageDetachArchivedChangesets')
        } catch (error) {
            setIsSubmittingSelected(asError(error))
        }
    }, [batchChangeID, selectedChangesets, setIsSubmittingSelected, deselectAll, telemetryService])

    // TODO: We need to deselect all of setChangesetFilters is called
    const [changesetFilters, setChangesetFilters] = useState<ChangesetFilters>({
        checkState: null,
        state: null,
        reviewState: null,
        search: null,
    })
    const queryChangesetsConnection = useCallback(
        (args: FilteredConnectionQueryArguments) =>
            queryChangesets({
                state: changesetFilters.state,
                reviewState: changesetFilters.reviewState,
                checkState: changesetFilters.checkState,
                first: args.first ?? null,
                after: args.after ?? null,
                batchChange: batchChangeID,
                onlyPublishedByThisBatchChange: null,
                search: changesetFilters.search,
                onlyArchived: !!onlyArchived,
            }).pipe(repeatWhen(notifier => notifier.pipe(delay(5000)))),
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

    return (
        <>
            {!hideFilters && (
                <ChangesetFilterRow history={history} location={location} onFiltersChange={setChangesetFilters} />
            )}
            {viewerCanAdminister && enableSelect && (
                <ChangesetSelectRow
                    selected={selectedChangesets}
                    onSubmit={onSubmitSelected}
                    // deselectAll={deselectAll}
                    isSubmitting={isSubmittingSelected}
                />
            )}
            <div className="list-group position-relative" ref={nextContainerElement}>
                <FilteredConnection<ChangesetFields, Omit<ChangesetNodeProps, 'node'>>
                    className="mt-2"
                    nodeComponent={ChangesetNode}
                    nodeComponentProps={{
                        isLightTheme,
                        viewerCanAdminister,
                        history,
                        location,
                        extensionInfo: { extensionsController, hoverifier },
                        expandByDefault,
                        queryExternalChangesetWithFileDiffs,
                        enableSelect,
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
                    listClassName={
                        enableSelect
                            ? 'batch-change-changesets__grid--with-checkboxes mb-3'
                            : 'batch-change-changesets__grid mb-3'
                    }
                    headComponent={
                        enableSelect ? BatchChangeChangesetsHeaderWithCheckboxes : BatchChangeChangesetsHeader
                    }
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
        </>
    )
}

/**
 * Returns true, if any filter is selected.
 */
function filtersSelected(filters: ChangesetFilters): boolean {
    return filters.checkState !== null || filters.state !== null || filters.reviewState !== null || !!filters.search
}
