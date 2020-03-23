import React, { useState, useCallback, useMemo, useEffect } from 'react'
import H from 'history'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { ChangesetNode, ChangesetNodeProps } from './ChangesetNode'
import { ThemeProps } from '../../../../../../shared/src/theme'
import { FilteredConnection, FilteredConnectionQueryArgs, Connection } from '../../../../components/FilteredConnection'
import { Observable, Subject } from 'rxjs'
import { DEFAULT_CHANGESET_LIST_COUNT } from '../presentation'
import { upperFirst, lowerCase } from 'lodash'
import { queryChangesets as _queryChangesets, queryPatches } from '../backend'
import { repeatWhen, delay, withLatestFrom, map, filter } from 'rxjs/operators'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { createHoverifier, HoveredToken } from '@sourcegraph/codeintellify'
import {
    RepoSpec,
    RevSpec,
    FileSpec,
    ResolvedRevSpec,
    UIPositionSpec,
    ModeSpec,
} from '../../../../../../shared/src/util/url'
import { HoverMerged } from '../../../../../../shared/src/api/client/types/hover'
import { ActionItemAction } from '../../../../../../shared/src/actions/ActionItem'
import { getHoverActions } from '../../../../../../shared/src/hover/actions'
import { WebHoverOverlay } from '../../../../components/shared'
import { getModeFromPath } from '../../../../../../shared/src/languages'
import { getHover } from '../../../../backend/features'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { TelemetryProps } from '../../../../../../shared/src/telemetry/telemetryService'
import { propertyIsDefined } from '../../../../../../shared/src/util/types'
import { useObservable } from '../../../../../../shared/src/util/useObservable'

interface Props extends ThemeProps, PlatformContextProps, TelemetryProps, ExtensionsControllerProps {
    campaign: Pick<GQL.IPatchSet, '__typename' | 'id'> | Pick<GQL.ICampaign, '__typename' | 'id' | 'closedAt'>
    history: H.History
    location: H.Location
    campaignUpdates: Subject<void>
    changesetUpdates: Subject<void>

    /** For testing only. */
    queryChangesets?: (
        campaignID: GQL.ID,
        args: FilteredConnectionQueryArgs
    ) => Observable<Connection<GQL.IExternalChangeset | GQL.IPatch>>
}

function getLSPTextDocumentPositionParams(
    hoveredToken: HoveredToken & RepoSpec & RevSpec & FileSpec & ResolvedRevSpec
): RepoSpec & RevSpec & ResolvedRevSpec & FileSpec & UIPositionSpec & ModeSpec {
    return {
        repoName: hoveredToken.repoName,
        rev: hoveredToken.rev,
        filePath: hoveredToken.filePath,
        commitID: hoveredToken.commitID,
        position: hoveredToken,
        mode: getModeFromPath(hoveredToken.filePath || ''),
    }
}

/**
 * A list of a campaign's or campaign preview's changesets.
 */
export const CampaignChangesets: React.FunctionComponent<Props> = ({
    campaign,
    history,
    location,
    isLightTheme,
    changesetUpdates,
    campaignUpdates,
    extensionsController,
    platformContext,
    telemetryService,
    queryChangesets = _queryChangesets,
}) => {
    const [state, setState] = useState<GQL.ChangesetState | undefined>()
    const [reviewState, setReviewState] = useState<GQL.ChangesetReviewState | undefined>()
    const [checkState, setCheckState] = useState<GQL.ChangesetCheckState | undefined>()

    const queryChangesetsConnection = useCallback(
        (args: FilteredConnectionQueryArgs) => {
            const queryObservable: Observable<GQL.IPatchConnection | Connection<GQL.IExternalChangeset | GQL.IPatch>> =
                campaign.__typename === 'PatchSet'
                    ? queryPatches(campaign.id, args)
                    : queryChangesets(campaign.id, { ...args, state, reviewState, checkState })
            return queryObservable.pipe(repeatWhen(obs => obs.pipe(delay(5000))))
        },
        [campaign.id, campaign.__typename, state, reviewState, checkState, queryChangesets]
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
            createHoverifier<RepoSpec & RevSpec & FileSpec & ResolvedRevSpec, HoverMerged, ActionItemAction>({
                closeButtonClicks,
                hoverOverlayElements,
                hoverOverlayRerenders: componentRerenders.pipe(
                    withLatestFrom(hoverOverlayElements, containerElements),
                    map(([, hoverOverlayElement, relativeElement]) => ({
                        hoverOverlayElement,
                        // The root component element is guaranteed to be rendered after a componentDidUpdate
                        relativeElement: relativeElement!,
                    })),
                    // Can't reposition HoverOverlay if it wasn't rendered
                    filter(propertyIsDefined('hoverOverlayElement'))
                ),
                getHover: hoveredToken =>
                    getHover(getLSPTextDocumentPositionParams(hoveredToken), { extensionsController }),
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

    const changesetFiltersRow = (
        <div className="form-inline mb-0 mt-2">
            <label htmlFor="changeset-state-filter">State</label>
            <select
                className="form-control mx-2"
                value={state}
                onChange={e => setState((e.target.value || undefined) as GQL.ChangesetState | undefined)}
                id="changeset-state-filter"
            >
                <option value="">All</option>
                {Object.values(GQL.ChangesetState).map(state => (
                    <option value={state} key={state}>
                        {upperFirst(lowerCase(state))}
                    </option>
                ))}
            </select>
            <label htmlFor="changeset-review-state-filter">Review state</label>
            <select
                className="form-control mx-2"
                value={reviewState}
                onChange={e => setReviewState((e.target.value || undefined) as GQL.ChangesetReviewState | undefined)}
                id="changeset-review-state-filter"
            >
                <option value="">All</option>
                {Object.values(GQL.ChangesetReviewState).map(state => (
                    <option value={state} key={state}>
                        {upperFirst(lowerCase(state))}
                    </option>
                ))}
            </select>
            <label htmlFor="changeset-check-state-filter">Check state</label>
            <select
                className="form-control mx-2"
                value={checkState}
                onChange={e => setCheckState((e.target.value || undefined) as GQL.ChangesetCheckState | undefined)}
                id="changeset-check-state-filter"
            >
                <option value="">All</option>
                {Object.values(GQL.ChangesetCheckState).map(state => (
                    <option value={state} key={state}>
                        {upperFirst(lowerCase(state))}
                    </option>
                ))}
            </select>
        </div>
    )

    return (
        <>
            {campaign.__typename === 'Campaign' && changesetFiltersRow}
            <div className="list-group position-relative" ref={nextContainerElement}>
                <FilteredConnection<GQL.IExternalChangeset | GQL.IPatch, Omit<ChangesetNodeProps, 'node'>>
                    className="mt-2"
                    updates={changesetUpdates}
                    nodeComponent={ChangesetNode}
                    nodeComponentProps={{
                        isLightTheme,
                        history,
                        location,
                        campaignUpdates,
                        enablePublishing: campaign.__typename === 'Campaign' && !campaign.closedAt,
                        extensionInfo: { extensionsController, hoverifier },
                    }}
                    queryConnection={queryChangesetsConnection}
                    hideSearch={true}
                    defaultFirst={DEFAULT_CHANGESET_LIST_COUNT}
                    noun="changeset"
                    pluralNoun="changesets"
                    history={history}
                    location={location}
                    noShowLoaderOnSlowLoad={true}
                    useURLQuery={false}
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
