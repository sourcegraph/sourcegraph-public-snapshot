import React, { useCallback, useMemo, useEffect } from 'react'

import * as H from 'history'
import { Subject } from 'rxjs'
import { repeatWhen, withLatestFrom, filter, map, delay } from 'rxjs/operators'

import { HoverMerged } from '@sourcegraph/client-api'
import { createHoverifier } from '@sourcegraph/codeintellify'
import { ErrorLike, isDefined, property } from '@sourcegraph/common'
import { ActionItemAction } from '@sourcegraph/shared/src/actions/ActionItem'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { getHoverActions } from '@sourcegraph/shared/src/hover/actions'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { RepoSpec, RevisionSpec, FileSpec, ResolvedRevisionSpec } from '@sourcegraph/shared/src/util/url'
import { Container, useObservable } from '@sourcegraph/wildcard'

import { getHover, getDocumentHighlights } from '../../../backend/features'
import { FilteredConnectionQueryArguments, FilteredConnection } from '../../../components/FilteredConnection'
import { WebHoverOverlay } from '../../../components/shared'
import { Scalars, ChangesetFields, BatchChangeChangesetsResult } from '../../../graphql-operations'
import {
    queryChangesets as _queryChangesets,
    queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs,
} from '../detail/backend'
import { getLSPTextDocumentPositionParameters } from '../utils'

import {
    BatchChangeCloseHeaderWillCloseChangesets,
    BatchChangeCloseHeaderWillKeepChangesets,
} from './BatchChangeCloseHeader'
import { ChangesetCloseNodeProps, ChangesetCloseNode } from './ChangesetCloseNode'
import { CloseChangesetsListEmptyElement } from './CloseChangesetsListEmptyElement'

import styles from './BatchChangeCloseChangesetsList.module.scss'

interface Props
    extends ThemeProps,
        PlatformContextProps,
        TelemetryProps,
        ExtensionsControllerProps,
        SettingsCascadeProps {
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
export const BatchChangeCloseChangesetsList: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
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
    settingsCascade,
}) => {
    const queryChangesetsConnection = useCallback(
        (args: FilteredConnectionQueryArguments) =>
            queryChangesets({
                state: null,
                onlyClosable: true,
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
                    withCenteredSummary={true}
                />
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
                        settingsCascade={settingsCascade}
                    />
                )}
            </Container>
        </div>
    )
}
