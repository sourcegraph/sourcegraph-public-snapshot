import * as H from 'history'
import React, { useState, useCallback, useMemo, useEffect, useContext } from 'react'
import { Subject } from 'rxjs'
import { withLatestFrom, map, filter } from 'rxjs/operators'

import { createHoverifier } from '@sourcegraph/codeintellify'
import { isDefined } from '@sourcegraph/common'
import { ActionItemAction } from '@sourcegraph/shared/src/actions/ActionItem'
import { HoverMerged } from '@sourcegraph/shared/src/api/client/types/hover'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { dataOrThrowErrors } from '@sourcegraph/shared/src/graphql/graphql'
import { getHoverActions } from '@sourcegraph/shared/src/hover/actions'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { property } from '@sourcegraph/shared/src/util/types'
import { RepoSpec, RevisionSpec, FileSpec, ResolvedRevisionSpec } from '@sourcegraph/shared/src/util/url'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { useConnection } from '@sourcegraph/web/src/components/FilteredConnection/hooks/useConnection'
import {
    ConnectionContainer,
    ConnectionError,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
} from '@sourcegraph/web/src/components/FilteredConnection/ui'
import { Container } from '@sourcegraph/wildcard'

import { getHover, getDocumentHighlights } from '../../../../backend/features'
import { WebHoverOverlay } from '../../../../components/shared'
import {
    ExternalChangesetFields,
    HiddenExternalChangesetFields,
    Scalars,
    BatchChangeChangesetsResult,
    BatchChangeChangesetsVariables,
} from '../../../../graphql-operations'
import { MultiSelectContext, MultiSelectContextProvider } from '../../MultiSelectContext'
import { getLSPTextDocumentPositionParameters } from '../../utils'
import {
    queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs,
    queryAllChangesetIDs as _queryAllChangesetIDs,
    CHANGESETS,
} from '../backend'

import styles from './BatchChangeChangesets.module.scss'
import { BatchChangeChangesetsHeader } from './BatchChangeChangesetsHeader'
import { ChangesetFilters, ChangesetFilterRow } from './ChangesetFilterRow'
import { ChangesetNode } from './ChangesetNode'
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
    refetchBatchChange: () => void

    /** For testing only. */
    queryExternalChangesetWithFileDiffs?: typeof _queryExternalChangesetWithFileDiffs
    /** For testing only. */
    queryAllChangesetIDs?: typeof _queryAllChangesetIDs
    /** For testing only. */
    expandByDefault?: boolean
}

/**
 * A list of a batch change's changesets.
 */
export const BatchChangeChangesets: React.FunctionComponent<Props> = props => (
    <MultiSelectContextProvider>
        <BatchChangeChangesetsImpl {...props} />
    </MultiSelectContextProvider>
)

const BATCH_COUNT = 15

const BatchChangeChangesetsImpl: React.FunctionComponent<Props> = ({
    batchChangeID,
    viewerCanAdminister,
    history,
    location,
    isLightTheme,
    extensionsController,
    platformContext,
    telemetryService,
    hideFilters = false,
    queryAllChangesetIDs = _queryAllChangesetIDs,
    queryExternalChangesetWithFileDiffs,
    expandByDefault,
    onlyArchived,
    refetchBatchChange,
}) => {
    // You might look at this destructuring statement and wonder why this isn't
    // just a single context consumer object. The reason is because making it a
    // single object makes it hard to have hooks that depend on individual
    // callbacks and objects within the context. Therefore, we'll have a nice,
    // ugly destructured set of variables here.
    const {
        selected,
        deselectAll,
        areAllVisibleSelected,
        isSelected,
        toggleSingle,
        toggleVisible,
        setVisible,
    } = useContext(MultiSelectContext)

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

    // After selecting and performing a bulk action, deselect all changesets and refetch
    // the batch change to get the actively-running bulk operations.
    const onSubmitBulkAction = useCallback(() => {
        deselectAll()
        refetchBatchChange()
    }, [deselectAll, refetchBatchChange])

    const queryArguments = useMemo(
        () => ({
            state: changesetFilters.state,
            reviewState: changesetFilters.reviewState,
            checkState: changesetFilters.checkState,
            batchChange: batchChangeID,
            onlyPublishedByThisBatchChange: null,
            search: changesetFilters.search,
            onlyArchived: !!onlyArchived,
        }),
        [changesetFilters, batchChangeID, onlyArchived]
    )

    const { connection, error, loading, fetchMore, hasNextPage } = useConnection<
        BatchChangeChangesetsResult,
        BatchChangeChangesetsVariables,
        ExternalChangesetFields | HiddenExternalChangesetFields
    >({
        query: CHANGESETS,
        variables: {
            ...queryArguments,
            first: BATCH_COUNT,
            after: null,
        },
        options: {
            useURL: true,
            fetchPolicy: 'cache-and-network',
            pollInterval: 5000,
        },
        getConnection: result => {
            const data = dataOrThrowErrors(result)

            if (!data.node) {
                throw new Error(`Batch change with ID ${batchChangeID} does not exist`)
            }
            if (data.node.__typename !== 'BatchChange') {
                throw new Error(`The given ID is a ${data.node.__typename as string}, not a BatchChange`)
            }
            return data.node.changesets
        },
    })

    useEffect(() => {
        if (connection?.nodes?.length) {
            // Available changesets are all changesets that the user can view.
            setVisible(connection.nodes.filter(node => node.__typename === 'ExternalChangeset').map(node => node.id))
        }
    }, [connection?.nodes, setVisible])

    const containerElements = useMemo(() => new Subject<HTMLElement | null>(), [])
    const nextContainerElement = useMemo(() => containerElements.next.bind(containerElements), [containerElements])

    const hoverOverlayElements = useMemo(() => new Subject<HTMLElement | null>(), [])
    const nextOverlayElement = useCallback((element: HTMLElement | null): void => hoverOverlayElements.next(element), [
        hoverOverlayElements,
    ])

    const componentRerenders = useMemo(() => new Subject<void>(), [])

    const hoverifier = useMemo(
        () =>
            createHoverifier<RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec, HoverMerged, ActionItemAction>({
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
            }),
        [containerElements, extensionsController, hoverOverlayElements, platformContext, componentRerenders]
    )
    useEffect(() => () => hoverifier.unsubscribe(), [hoverifier])

    const hoverState = useObservable(useMemo(() => hoverifier.hoverStateUpdates, [hoverifier]))
    useEffect(() => {
        componentRerenders.next()
    }, [componentRerenders, hoverState])

    const showSelectRow = viewerCanAdminister && (selected === 'all' || selected.size > 0)

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
                    onSubmit={onSubmitBulkAction}
                    queryAllChangesetIDs={queryAllChangesetIDs}
                    queryArguments={queryArguments}
                />
            )}
            <div className="list-group position-relative" ref={nextContainerElement}>
                <ConnectionContainer>
                    {error && <ConnectionError errors={[error.message]} />}
                    <ConnectionList className={styles.batchChangeChangesetsGrid}>
                        {connection?.nodes?.length ? (
                            <BatchChangeChangesetsHeader
                                allSelected={showSelectRow && areAllVisibleSelected()}
                                toggleSelectAll={toggleVisible}
                                disabled={!viewerCanAdminister}
                            />
                        ) : null}
                        {connection?.nodes?.map(node => (
                            <ChangesetNode
                                key={node.id}
                                node={node}
                                isLightTheme={isLightTheme}
                                viewerCanAdminister={viewerCanAdminister}
                                history={history}
                                location={location}
                                extensionInfo={{ extensionsController, hoverifier }}
                                expandByDefault={expandByDefault}
                                queryExternalChangesetWithFileDiffs={queryExternalChangesetWithFileDiffs}
                                selectable={{ onSelect: toggleSingle, isSelected }}
                            />
                        ))}
                    </ConnectionList>
                    {/* TODO: This is to prevent the spinner from flashing as we constantly
                    poll in the background. Once we can distinguish between "loading new data"
                    and "refetching existing data" with the Apollo cache, we should rework to
                    show the spinner whenever we are loading new data. */}
                    {loading && connection?.nodes?.length === 0 && <ConnectionLoading />}
                    {connection && (
                        <SummaryContainer centered={true}>
                            <ConnectionSummary
                                noSummaryIfAllNodesVisible={true}
                                first={BATCH_COUNT}
                                connection={connection}
                                noun="changeset"
                                pluralNoun="changesets"
                                hasNextPage={hasNextPage}
                                emptyElement={
                                    filtersSelected(changesetFilters) ? (
                                        <EmptyChangesetSearchElement />
                                    ) : onlyArchived ? (
                                        <EmptyArchivedChangesetListElement />
                                    ) : (
                                        <EmptyChangesetListElement />
                                    )
                                }
                            />
                            {hasNextPage && <ShowMoreButton onClick={fetchMore} />}
                        </SummaryContainer>
                    )}
                </ConnectionContainer>
                {hoverState?.hoverOverlayProps && (
                    <WebHoverOverlay
                        {...hoverState.hoverOverlayProps}
                        nav={url => history.push(url)}
                        hoveredTokenElement={hoverState.hoveredTokenElement}
                        telemetryService={telemetryService}
                        extensionsController={extensionsController}
                        isLightTheme={isLightTheme}
                        location={location}
                        platformContext={platformContext}
                        hoverRef={nextOverlayElement}
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
