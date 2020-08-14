import React, { useState, useCallback, useMemo, useEffect } from 'react'
import * as H from 'history'
import { ChangesetNodeProps, ChangesetNode } from './ChangesetNode'
import { ThemeProps } from '../../../../../../shared/src/theme'
import { FilteredConnection, FilteredConnectionQueryArgs } from '../../../../components/FilteredConnection'
import { Subject, merge, of } from 'rxjs'
import { upperFirst, lowerCase } from 'lodash'
import {
    queryChangesets as _queryChangesets,
    queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs,
} from '../backend'
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
import {
    ChangesetFields,
    ChangesetExternalState,
    ChangesetReviewState,
    ChangesetCheckState,
    Scalars,
} from '../../../../graphql-operations'
import { isValidChangesetExternalState, isValidChangesetReviewState, isValidChangesetCheckState } from '../../utils'
import classNames from 'classnames'
import { CampaignChangesetsHeader } from './CampaignChangesetsHeader'

interface Props extends ThemeProps, PlatformContextProps, TelemetryProps, ExtensionsControllerProps {
    campaignID: Scalars['ID']
    viewerCanAdminister: boolean
    history: H.History
    location: H.Location
    campaignUpdates: Subject<void>
    changesetUpdates: Subject<void>
    /** When true, only open changesets will be listed. */
    onlyOpen?: boolean
    hideFilters?: boolean

    /** For testing only. */
    queryChangesets?: typeof _queryChangesets
    /** For testing only. */
    queryExternalChangesetWithFileDiffs?: typeof _queryExternalChangesetWithFileDiffs
}

interface ChangesetFilters {
    externalState: ChangesetExternalState | null
    reviewState: ChangesetReviewState | null
    checkState: ChangesetCheckState | null
}

/**
 * A list of a campaign's changesets.
 */
export const CampaignChangesets: React.FunctionComponent<Props> = ({
    campaignID,
    viewerCanAdminister,
    history,
    location,
    isLightTheme,
    changesetUpdates,
    campaignUpdates,
    extensionsController,
    platformContext,
    telemetryService,
    onlyOpen = false,
    hideFilters = false,
    queryChangesets = _queryChangesets,
    queryExternalChangesetWithFileDiffs,
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
                        externalState: changesetFilters.externalState,
                        reviewState: changesetFilters.reviewState,
                        checkState: changesetFilters.checkState,
                        ...(onlyOpen ? { externalState: ChangesetExternalState.OPEN } : {}),
                        first: args.first ?? null,
                        campaign: campaignID,
                    }).pipe(repeatWhen(notifier => notifier.pipe(delay(5000))))
                )
            ),
        [
            campaignID,
            changesetFilters.externalState,
            changesetFilters.reviewState,
            changesetFilters.checkState,
            queryChangesets,
            changesetUpdates,
            onlyOpen,
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
            <div className="list-group position-relative" ref={nextContainerElement}>
                <FilteredConnection<ChangesetFields, Omit<ChangesetNodeProps, 'node'>>
                    className="mt-2"
                    nodeComponent={ChangesetNode}
                    nodeComponentProps={{
                        isLightTheme,
                        viewerCanAdminister,
                        history,
                        location,
                        campaignUpdates,
                        extensionInfo: { extensionsController, hoverifier },
                        queryExternalChangesetWithFileDiffs,
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
                    listClassName="campaign-changesets__grid mb-3"
                    headComponent={CampaignChangesetsHeader}
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
    const [externalState, setExternalState] = useState<ChangesetExternalState | undefined>(() => {
        const value = searchParameters.get('external_state')
        return value && isValidChangesetExternalState(value) ? value : undefined
    })
    const [reviewState, setReviewState] = useState<ChangesetReviewState | undefined>(() => {
        const value = searchParameters.get('review_state')
        return value && isValidChangesetReviewState(value) ? value : undefined
    })
    const [checkState, setCheckState] = useState<ChangesetCheckState | undefined>(() => {
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
        <div className="form-inline m-0">
            <ChangesetFilter<ChangesetExternalState>
                values={Object.values(ChangesetExternalState)}
                label="State"
                selected={externalState}
                onChange={setExternalState}
                className="mr-2"
            />
            <ChangesetFilter<ChangesetReviewState>
                values={Object.values(ChangesetReviewState)}
                label="Review state"
                selected={reviewState}
                onChange={setReviewState}
                className="mr-2"
            />
            <ChangesetFilter<ChangesetCheckState>
                values={Object.values(ChangesetCheckState)}
                label="Check state"
                selected={checkState}
                onChange={setCheckState}
            />
        </div>
    )
}

interface ChangesetFilterProps<T extends string> {
    label: string
    values: T[]
    selected: T | undefined
    onChange: (value: T | undefined) => void
    className?: string
}

export const ChangesetFilter = <T extends string>({
    label,
    values,
    selected,
    onChange,
    className,
}: ChangesetFilterProps<T>): React.ReactElement<ChangesetFilterProps<T>> => (
    <>
        <select
            className={classNames('form-control', className)}
            value={selected}
            onChange={event => onChange((event.target.value ?? undefined) as T | undefined)}
        >
            <option value="">{label}</option>
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
