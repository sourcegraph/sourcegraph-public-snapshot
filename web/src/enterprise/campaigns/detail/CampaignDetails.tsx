import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React, { useEffect, useMemo } from 'react'
import { HeroPage } from '../../../components/HeroPage'
import { PageTitle } from '../../../components/PageTitle'
import { isEqual } from 'lodash'
import {
    fetchCampaignById as _fetchCampaignById,
    queryChangesets as _queryChangesets,
    queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs,
    queryChangesetCountsOverTime as _queryChangesetCountsOverTime,
    deleteCampaign as _deleteCampaign,
} from './backend'
import { useObservable } from '../../../../../shared/src/util/useObservable'
import * as H from 'history'
import { Subject, of, merge } from 'rxjs'
import { switchMap, distinctUntilChanged } from 'rxjs/operators'
import { ThemeProps } from '../../../../../shared/src/theme'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import { CampaignFields, Scalars } from '../../../graphql-operations'
import { CampaignDescription } from './CampaignDescription'
import { CampaignStatsCard } from './CampaignStatsCard'
import { CampaignHeader } from './CampaignHeader'
import { CampaignTabs } from './CampaignTabs'
import { CampaignDetailsActionSection } from './CampaignDetailsActionSection'

export interface CampaignDetailsProps
    extends ThemeProps,
        ExtensionsControllerProps,
        PlatformContextProps,
        TelemetryProps {
    /**
     * The campaign ID.
     */
    campaignID: Scalars['ID']
    history: H.History
    location: H.Location

    /** For testing only. */
    fetchCampaignById?: typeof _fetchCampaignById
    /** For testing only. */
    queryChangesets?: typeof _queryChangesets
    /** For testing only. */
    queryExternalChangesetWithFileDiffs?: typeof _queryExternalChangesetWithFileDiffs
    /** For testing only. */
    queryChangesetCountsOverTime?: typeof _queryChangesetCountsOverTime
    /** For testing only. */
    deleteCampaign?: typeof _deleteCampaign
}

/**
 * The area for a single campaign.
 */
export const CampaignDetails: React.FunctionComponent<CampaignDetailsProps> = ({
    campaignID,
    history,
    location,
    isLightTheme,
    extensionsController,
    platformContext,
    telemetryService,
    fetchCampaignById = _fetchCampaignById,
    queryChangesets,
    queryExternalChangesetWithFileDiffs,
    queryChangesetCountsOverTime,
    deleteCampaign,
}) => {
    /** Retrigger fetching */
    const campaignUpdates = useMemo(() => new Subject<void>(), [])

    useEffect(() => {
        telemetryService.logViewEvent(campaignID ? 'CampaignDetailsPage' : 'NewCampaignPage')
    }, [campaignID, telemetryService])

    const campaign: CampaignFields | null | undefined = useObservable(
        useMemo(
            () =>
                merge(of(undefined), campaignUpdates).pipe(
                    switchMap(() => fetchCampaignById(campaignID)),
                    distinctUntilChanged((a, b) => isEqual(a, b))
                ),
            [campaignID, campaignUpdates, fetchCampaignById]
        )
    )

    // Is loading.
    if (campaign === undefined) {
        return (
            <div className="text-center">
                <LoadingSpinner className="icon-inline mx-auto my-4" />
            </div>
        )
    }
    // Campaign was not found
    if (campaign === null) {
        return <HeroPage icon={AlertCircleIcon} title="Campaign not found" />
    }

    return (
        <>
            <PageTitle title={campaign.name} />
            <CampaignHeader
                name={campaign.name}
                namespace={campaign.namespace}
                creator={campaign.initialApplier}
                createdAt={campaign.createdAt}
                actionSection={
                    <CampaignDetailsActionSection
                        campaignID={campaign.id}
                        campaignClosed={!!campaign.closedAt}
                        deleteCampaign={deleteCampaign}
                        campaignNamespaceURL={campaign.namespace.url}
                        history={history}
                    />
                }
                className="mb-3 test-campaign-details-page"
            />
            <CampaignStatsCard closedAt={campaign.closedAt} stats={campaign.changesets.stats} className="mb-3" />
            <CampaignDescription history={history} description={campaign.description} />
            <CampaignTabs
                campaign={campaign}
                campaignUpdates={campaignUpdates}
                extensionsController={extensionsController}
                history={history}
                isLightTheme={isLightTheme}
                location={location}
                platformContext={platformContext}
                telemetryService={telemetryService}
                queryChangesets={queryChangesets}
                queryChangesetCountsOverTime={queryChangesetCountsOverTime}
                queryExternalChangesetWithFileDiffs={queryExternalChangesetWithFileDiffs}
            />
        </>
    )
}
