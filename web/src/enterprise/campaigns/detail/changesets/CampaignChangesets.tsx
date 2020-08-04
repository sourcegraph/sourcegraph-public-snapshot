import React, { useState, useCallback, useMemo, useEffect } from 'react'
import * as H from 'history'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { ChangesetNodeProps, ChangesetNode } from './ChangesetNode'
import { ThemeProps } from '../../../../../../shared/src/theme'
import { FilteredConnection, FilteredConnectionQueryArgs } from '../../../../components/FilteredConnection'
import { Subject, merge, of } from 'rxjs'
import { upperFirst, lowerCase } from 'lodash'
import { queryChangesets as _queryChangesets } from '../backend'
import { repeatWhen, delay, withLatestFrom, map, filter, switchMap } from 'rxjs/operators'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { createHoverifier, HoveredToken } from '@sourcegraph/codeintellify'
import {
    RepoSpec,
    RevisionSpec,
    FileSpec,
    ResolvedRevisionSpec,
    UIPositionSpec,
    ModeSpec,
} from '../../../../../../shared/src/util/url'
import { HoverMerged } from '../../../../../../shared/src/api/client/types/hover'
import { ActionItemAction } from '../../../../../../shared/src/actions/ActionItem'
import { getHoverActions } from '../../../../../../shared/src/hover/actions'
import { WebHoverOverlay } from '../../../../components/shared'
import { getModeFromPath } from '../../../../../../shared/src/languages'
import { getHover, getDocumentHighlights } from '../../../../backend/features'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { TelemetryProps } from '../../../../../../shared/src/telemetry/telemetryService'
import { property, isDefined } from '../../../../../../shared/src/util/types'
import { useObservable } from '../../../../../../shared/src/util/useObservable'
import { ChangesetFields } from '../../../../graphql-operations'
import { isValidChangesetExternalState, isValidChangesetReviewState, isValidChangesetCheckState } from '../../utils'

interface Props extends ThemeProps, PlatformContextProps, TelemetryProps, ExtensionsControllerProps {
    campaign: Pick<GQL.ICampaign, 'id' | 'closedAt' | 'viewerCanAdminister'>
    history: H.History
    location: H.Location
    campaignUpdates: Subject<void>
    changesetUpdates: Subject<void>

    /** For testing only. */
    queryChangesets?: typeof _queryChangesets
}

interface ChangesetFilters {
    externalState: GQL.ChangesetExternalState | null
    reviewState: GQL.ChangesetReviewState | null
    checkState: GQL.ChangesetCheckState | null
}

/**
 * A list of a campaign's changesets.
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
    const [changesetFilters, setChangesetFilters] = useState<ChangesetFilters>({
        checkState: null,
        externalState: null,
        reviewState: null,
    })
    const queryChangesetsConnection = useCallback(
        (args: FilteredConnectionQueryArgs) =>
            merge(of(undefined), changesetUpdates).pipe(
                switchMap(() =>
                    queryChangesets({
                        ...changesetFilters,
                        first: args.first ?? null,
                        campaign: campaign.id,
                    }).pipe(repeatWhen(notifier => notifier.pipe(delay(5000))))
                )
            ),
        [campaign.id, changesetFilters, queryChangesets, changesetUpdates]
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
            <ChangesetFilterRow history={history} location={location} onFiltersChange={setChangesetFilters} />
            <div className="list-group position-relative" ref={nextContainerElement}>
                <FilteredConnection<ChangesetFields, Omit<ChangesetNodeProps, 'node'>>
                    className="mt-2"
                    nodeComponent={ChangesetNode}
                    nodeComponentProps={{
                        isLightTheme,
                        viewerCanAdminister: campaign.viewerCanAdminister,
                        history,
                        location,
                        campaignUpdates,
                        extensionInfo: { extensionsController, hoverifier },
                    }}
                    queryConnection={queryChangesetsConnection}
                    hideSearch={true}
                    defaultFirst={15}
                    noun="changeset"
                    pluralNoun="changesets"
                    history={history}
                    location={location}
                    useURLQuery={true}
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

interface ChangesetFilterRowProps {
    history: H.History
    location: H.Location
    onFiltersChange: (newFilters: ChangesetFilters) => void
}

const ChangesetFilterRow: React.FunctionComponent<ChangesetFilterRowProps> = ({
    history,
    location,
    onFiltersChange,
}) => {
    const searchParameters = new URLSearchParams(location.search)
    const [externalState, setExternalState] = useState<GQL.ChangesetExternalState | undefined>(() => {
        const value = searchParameters.get('external_state')
        return value && isValidChangesetExternalState(value) ? value : undefined
    })
    const [reviewState, setReviewState] = useState<GQL.ChangesetReviewState | undefined>(() => {
        const value = searchParameters.get('review_state')
        return value && isValidChangesetReviewState(value) ? value : undefined
    })
    const [checkState, setCheckState] = useState<GQL.ChangesetCheckState | undefined>(() => {
        const value = searchParameters.get('check_state')
        return value && isValidChangesetCheckState(value) ? value : undefined
    })
    useEffect(() => {
        const searchParameters = new URLSearchParams(location.search)
        if (externalState) {
            searchParameters.set('external_state', externalState)
        } else {
            searchParameters.delete('external_state')
        }
        if (reviewState) {
            searchParameters.set('review_state', reviewState)
        } else {
            searchParameters.delete('review_state')
        }
        if (checkState) {
            searchParameters.set('check_state', checkState)
        } else {
            searchParameters.delete('check_state')
        }
        history.replace({ ...location, search: searchParameters.toString() })
        // Update the filters in the parent component.
        onFiltersChange({
            externalState: externalState || null,
            reviewState: reviewState || null,
            checkState: checkState || null,
        })
        // We cannot depend on the history, since it's modified by this hook and that would cause an infinite render loop.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [externalState, reviewState, checkState])
    return (
        <div className="form-inline mb-0 mt-2">
            <ChangesetFilter<GQL.ChangesetExternalState>
                values={Object.values(GQL.ChangesetExternalState)}
                label="State"
                htmlID="changeset-state-filter"
                selected={externalState}
                onChange={setExternalState}
            />
            <ChangesetFilter<GQL.ChangesetReviewState>
                values={Object.values(GQL.ChangesetReviewState)}
                label="Review state"
                htmlID="changeset-review-state-filter"
                selected={reviewState}
                onChange={setReviewState}
            />
            <ChangesetFilter<GQL.ChangesetCheckState>
                values={Object.values(GQL.ChangesetCheckState)}
                label="Check state"
                htmlID="changeset-check-state-filter"
                selected={checkState}
                onChange={setCheckState}
            />
        </div>
    )
}

interface ChangesetFilterProps<T extends string> {
    label: string
    htmlID: string
    values: T[]
    selected: T | undefined
    onChange: (value: T | undefined) => void
}

export const ChangesetFilter = <T extends string>({
    htmlID,
    label,
    values,
    selected,
    onChange,
}: ChangesetFilterProps<T>): React.ReactElement<ChangesetFilterProps<T>> => (
    <>
        <label htmlFor={htmlID}>{label}</label>
        <select
            className="form-control mx-2"
            value={selected}
            onChange={event => onChange((event.target.value ?? undefined) as T | undefined)}
            id={htmlID}
        >
            <option value="">All</option>
            {values.map(state => (
                <option value={state} key={state}>
                    {upperFirst(lowerCase(state))}
                </option>
            ))}
        </select>
    </>
)

function getLSPTextDocumentPositionParameters(
    hoveredToken: HoveredToken & RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec
): RepoSpec & RevisionSpec & ResolvedRevisionSpec & FileSpec & UIPositionSpec & ModeSpec {
    return {
        repoName: hoveredToken.repoName,
        revision: hoveredToken.revision,
        filePath: hoveredToken.filePath,
        commitID: hoveredToken.commitID,
        position: hoveredToken,
        mode: getModeFromPath(hoveredToken.filePath || ''),
    }
}
