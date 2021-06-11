import * as H from 'history'
import React, { useCallback, useMemo, useEffect } from 'react'
import { Subject } from 'rxjs'
import { repeatWhen, withLatestFrom, filter, map, delay } from 'rxjs/operators'

import { createHoverifier } from '@sourcegraph/codeintellify'
import { ActionItemAction } from '@sourcegraph/shared/src/actions/ActionItem'
import { HoverMerged } from '@sourcegraph/shared/src/api/client/types/hover'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { ChangesetState } from '@sourcegraph/shared/src/graphql-operations'
import { getHoverActions } from '@sourcegraph/shared/src/hover/actions'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { ErrorLike } from '@sourcegraph/shared/src/util/errors'
import { isDefined, property } from '@sourcegraph/shared/src/util/types'
import { RepoSpec, RevisionSpec, FileSpec, ResolvedRevisionSpec } from '@sourcegraph/shared/src/util/url'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { Container } from '@sourcegraph/wildcard'

import { getHover, getDocumentHighlights } from '../../../backend/features'
import { FilteredConnectionQueryArguments, FilteredConnection } from '../../../components/FilteredConnection'
import { WebHoverOverlay } from '../../../components/shared'
import { Scalars, ChangesetFields, BatchChangeChangesetsResult } from '../../../graphql-operations'
import {
    queryChangesets as _queryChangesets,
    queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs,
} from '../detail/backend'
import { getLSPTextDocumentPositionParameters } from '../utils'

import styles from './BatchChangeCloseChangesetsList.module.scss'
import {
    BatchChangeCloseHeaderWillCloseChangesets,
    BatchChangeCloseHeaderWillKeepChangesets,
} from './BatchChangeCloseHeader'
import { ChangesetCloseNodeProps, ChangesetCloseNode } from './ChangesetCloseNode'
import { CloseChangesetsListEmptyElement } from './CloseChangesetsListEmptyElement'

interface Props extends ThemeProps, PlatformContextProps, TelemetryProps, ExtensionsControllerProps {
    batchChangeID: Scalars['ID']
    viewerCanAdminister: boolean
    history: H.History
    location: H.Location
    willClose: boolean
    onUpdate?: (
        connection?: (BatchChangeChangesetsResult['node'] & { __typename: 'BatchChange' })['changesets'] | ErrorLike
    ) => void

    /** For testing only. */
    queryChangesets?: typeof _queryChangesets
    /** For testing only. */
    queryExternalChangesetWithFileDiffs?: typeof _queryExternalChangesetWithFileDiffs
}

/**
 * A list of a batch change's changesets that may be closed.
 */
export const BatchChangeCloseChangesetsList: React.FunctionComponent<Props> = ({
    batchChangeID,
    viewerCanAdminister,
    history,
    location,
    isLightTheme,
    extensionsController,
    platformContext,
    telemetryService,
    willClose,
    onUpdate,
    queryChangesets = _queryChangesets,
    queryExternalChangesetWithFileDiffs,
}) => {
    const queryChangesetsConnection = useCallback(
        (args: FilteredConnectionQueryArguments) =>
            queryChangesets({
                // TODO: This doesn't account for draft changesets. Ideally, this would
                // use the delta API and apply an empty batch spec, but then changesets
                // would currently be lost.
                state: ChangesetState.OPEN,
                checkState: null,
                reviewState: null,
                first: args.first ?? null,
                after: args.after ?? null,
                batchChange: batchChangeID,
                onlyPublishedByThisBatchChange: true,
                search: null,
                onlyArchived: false,
            }).pipe(repeatWhen(notifier => notifier.pipe(delay(5000)))),
        [batchChangeID, queryChangesets]
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
        <div className="list-group position-relative" ref={nextContainerElement}>
            <Container>
                <FilteredConnection<
                    ChangesetFields,
                    Omit<ChangesetCloseNodeProps, 'node'>,
                    {},
                    (BatchChangeChangesetsResult['node'] & { __typename: 'BatchChange' })['changesets']
                >
                    nodeComponent={ChangesetCloseNode}
                    nodeComponentProps={{
                        isLightTheme,
                        viewerCanAdminister,
                        history,
                        location,
                        extensionInfo: { extensionsController, hoverifier },
                        queryExternalChangesetWithFileDiffs,
                        willClose,
                    }}
                    queryConnection={queryChangesetsConnection}
                    hideSearch={true}
                    defaultFirst={15}
                    noun="open changeset"
                    pluralNoun="open changesets"
                    history={history}
                    location={location}
                    useURLQuery={true}
                    listComponent="div"
                    listClassName={styles.batchChangeCloseChangesetsListGrid}
                    headComponent={
                        willClose ? BatchChangeCloseHeaderWillCloseChangesets : BatchChangeCloseHeaderWillKeepChangesets
                    }
                    noSummaryIfAllNodesVisible={true}
                    onUpdate={onUpdate}
                    emptyElement={<CloseChangesetsListEmptyElement />}
                    className="filtered-connection__centered-summary"
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
            </Container>
        </div>
    )
}
