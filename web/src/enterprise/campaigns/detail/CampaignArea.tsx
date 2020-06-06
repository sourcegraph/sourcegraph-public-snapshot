import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React, { useMemo } from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { HeroPage } from '../../../components/HeroPage'
import { isEqual } from 'lodash'
import {
    fetchCampaignById,
    fetchPatchSetById,
    queryPatchesFromCampaign,
    queryPatchesFromPatchSet,
    queryChangesets,
    queryPatchFileDiffs,
} from './backend'
import { useObservable } from '../../../../../shared/src/util/useObservable'
import { Subject, of, merge, Observable } from 'rxjs'
import { switchMap, distinctUntilChanged, tap } from 'rxjs/operators'
import { ThemeProps } from '../../../../../shared/src/theme'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import { repeatUntil } from '../../../../../shared/src/util/rxjs/repeatUntil'
import { Route, Switch, RouteComponentProps } from 'react-router'
import { CampaignDetailArea } from './CampaignDetailArea'
import { UpdateCampaignPage } from './update/UpdateCampaignPage'
import { CampaignBreadcrumbs } from '../common/CampaignBreadcrumbs'

export interface MinimalCampaign
    extends Pick<
        GQL.ICampaign,
        | '__typename'
        | 'id'
        | 'name'
        | 'url'
        | 'description'
        | 'author'
        | 'changesetCountsOverTime'
        | 'hasUnpublishedPatches'
        | 'createdAt'
        | 'updatedAt'
        | 'closedAt'
        | 'viewerCanAdminister'
        | 'branch'
    > {
    patchSet: Pick<GQL.IPatchSet, 'id'> | null
    changesets: Pick<GQL.ICampaign['changesets'], 'totalCount'>
    patches: Pick<GQL.ICampaign['patches'], 'totalCount'>
    status: Pick<GQL.ICampaign['status'], 'completedCount' | 'pendingCount' | 'errors' | 'state'>
    diffStat: Pick<GQL.ICampaign['diffStat'], 'added' | 'deleted' | 'changed'>
}

export interface MinimalPatchSet extends Pick<GQL.IPatchSet, '__typename' | 'id'> {
    diffStat: Pick<GQL.IPatchSet['diffStat'], 'added' | 'deleted' | 'changed'>
    patches: Pick<GQL.IPatchSet['patches'], 'totalCount'>
}

interface Props
    extends ThemeProps,
        ExtensionsControllerProps,
        PlatformContextProps,
        TelemetryProps,
        RouteComponentProps<{ campaignID: string }> {
    /**
     * The campaign ID.
     */
    campaignID: GQL.ID
    authenticatedUser: Pick<GQL.IUser, 'id' | 'username' | 'avatarURL'>

    /** For testing only. */
    _fetchCampaignById?: typeof fetchCampaignById | ((campaign: GQL.ID) => Observable<MinimalCampaign | null>)
    /** For testing only. */
    _fetchPatchSetById?: typeof fetchPatchSetById | ((patchSet: GQL.ID) => Observable<MinimalPatchSet | null>)
    /** For testing only. */
    _queryPatchesFromCampaign?: typeof queryPatchesFromCampaign
    /** For testing only. */
    _queryPatchesFromPatchSet?: typeof queryPatchesFromPatchSet
    /** For testing only. */
    _queryPatchFileDiffs?: typeof queryPatchFileDiffs
    /** For testing only. */
    _queryChangesets?: typeof queryChangesets
    /** For testing only. */
    _noSubject?: boolean
}

/**
 * The area for a single campaign.
 */
export const CampaignArea: React.FunctionComponent<Props> = ({
    campaignID,
    match,
    _fetchCampaignById = fetchCampaignById,
    _fetchPatchSetById = fetchPatchSetById,
    _queryPatchesFromCampaign = queryPatchesFromCampaign,
    _queryPatchesFromPatchSet = queryPatchesFromPatchSet,
    _queryPatchFileDiffs = queryPatchFileDiffs,
    _queryChangesets = queryChangesets,
    _noSubject = false,
    ...props
}) => {
    /** Retrigger campaign fetching */
    const campaignUpdates = useMemo(() => new Subject<void>(), [])
    /** Retrigger changeset fetching */
    const changesetUpdates = useMemo(() => new Subject<void>(), [])

    const campaign = useObservable(
        useMemo(() => {
            // On the very first fetch, a reload of the changesets is not required.
            let isFirstCampaignFetch = true

            return merge(of(undefined), _noSubject ? new Observable<void>() : campaignUpdates).pipe(
                switchMap(() =>
                    _fetchCampaignById(campaignID).pipe(
                        // repeat fetching the campaign as long as the state is still processing
                        repeatUntil(campaign => campaign?.status?.state !== GQL.BackgroundProcessState.PROCESSING, {
                            delay: 2000,
                        })
                    )
                ),
                distinctUntilChanged((a, b) => isEqual(a, b)),
                tap(() => {
                    if (!isFirstCampaignFetch) {
                        changesetUpdates.next()
                    }
                    isFirstCampaignFetch = false
                })
            )
        }, [_fetchCampaignById, _noSubject, campaignID, campaignUpdates, changesetUpdates])
    )

    return campaign === undefined ? (
        <div className="text-center">
            <LoadingSpinner className="icon-inline mx-auto my-4" />
        </div>
    ) : campaign === null ? (
        <HeroPage icon={AlertCircleIcon} title="Campaign not found" />
    ) : (
        <div className="w-100 mt-4">
            <div className="container">
                <CampaignBreadcrumbs campaign={campaign} className="mb-2" />
            </div>
            <Switch>
                <Route path={`${match.url}/edit`}>
                    <div className="container">
                        <UpdateCampaignPage
                            {...props}
                            campaign={campaign}
                            patchsetID={new URLSearchParams(location.search).get('patchset')}
                            fetchPatchSetById={_fetchPatchSetById}
                            queryPatchFileDiffs={_queryPatchFileDiffs}
                            queryPatchesFromCampaign={queryPatchesFromCampaign}
                            queryPatchesFromPatchSet={queryPatchesFromPatchSet}
                            queryChangesets={queryChangesets}
                        />
                    </div>
                </Route>
                <Route
                    path={match.url}
                    render={({ match }) => (
                        <CampaignDetailArea
                            {...props}
                            match={match}
                            campaign={campaign}
                            fetchPatchSetById={_fetchPatchSetById}
                            queryPatchesFromCampaign={_queryPatchesFromCampaign}
                            queryPatchesFromPatchSet={_queryPatchesFromPatchSet}
                            queryPatchFileDiffs={_queryPatchFileDiffs}
                            queryChangesets={_queryChangesets}
                        />
                    )}
                />
            </Switch>
        </div>
    )
}
