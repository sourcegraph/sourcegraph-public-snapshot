import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React, { useState, useEffect, useMemo } from 'react'
import { HeroPage } from '../../../components/HeroPage'
import { PageTitle } from '../../../components/PageTitle'
import { isEqual } from 'lodash'
import { fetchCampaignById } from './backend'
import { useError } from '../../../../../shared/src/util/useObservable'
import * as H from 'history'
import { CampaignBurndownChart } from './BurndownChart'
import { Subject, of, merge } from 'rxjs'
import { switchMap, distinctUntilChanged, repeatWhen, delay } from 'rxjs/operators'
import { ThemeProps } from '../../../../../shared/src/theme'
import { CampaignActionsBar } from './CampaignActionsBar'
import { CampaignChangesets } from './changesets/CampaignChangesets'
import { pluralize } from '../../../../../shared/src/util/strings'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import { CampaignFields, Scalars } from '../../../graphql-operations'
import { CampaignInfoCard } from './CampaignInfoCard'

interface Props extends ThemeProps, ExtensionsControllerProps, PlatformContextProps, TelemetryProps {
    /**
     * The campaign ID.
     */
    campaignID: Scalars['ID']
    history: H.History
    location: H.Location

    /** For testing only. */
    _fetchCampaignById?: typeof fetchCampaignById
}

/**
 * The area for a single campaign.
 */
export const CampaignDetails: React.FunctionComponent<Props> = ({
    campaignID,
    history,
    location,
    isLightTheme,
    extensionsController,
    platformContext,
    telemetryService,
    _fetchCampaignById = fetchCampaignById,
}) => {
    // For errors during fetching
    const triggerError = useError()

    /** Retrigger campaign fetching */
    const campaignUpdates = useMemo(() => new Subject<void>(), [])
    /** Retrigger changeset fetching */
    const changesetUpdates = useMemo(() => new Subject<void>(), [])

    const [campaign, setCampaign] = useState<CampaignFields | null>()

    useEffect(() => {
        telemetryService.logViewEvent(campaignID ? 'CampaignDetailsPage' : 'NewCampaignPage')
    }, [campaignID, telemetryService])

    useEffect(() => {
        if (!campaignID) {
            return
        }
        // on the very first fetch, a reload of the changesets is not required
        let isFirstCampaignFetch = true

        // Fetch campaign if ID was given
        const subscription = merge(of(undefined), campaignUpdates)
            .pipe(
                switchMap(() =>
                    _fetchCampaignById(campaignID).pipe(repeatWhen(observer => observer.pipe(delay(5000))))
                ),
                distinctUntilChanged((a, b) => isEqual(a, b))
            )
            .subscribe({
                next: fetchedCampaign => {
                    setCampaign(fetchedCampaign)
                    if (!isFirstCampaignFetch) {
                        changesetUpdates.next()
                    }
                    isFirstCampaignFetch = false
                },
                error: triggerError,
            })
        return () => subscription.unsubscribe()
    }, [campaignID, triggerError, changesetUpdates, campaignUpdates, _fetchCampaignById])

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

    const totalChangesetCount = campaign.changesets.totalCount

    return (
        <>
            <PageTitle title={campaign.name} />
            <CampaignActionsBar campaign={campaign} />
            <CampaignInfoCard
                history={history}
                author={campaign.initialApplier}
                createdAt={campaign.createdAt}
                description={campaign.description}
            />
            {totalChangesetCount > 0 && (
                <>
                    <h3 className="mt-4 mb-2">Progress</h3>
                    <CampaignBurndownChart
                        changesetCountsOverTime={campaign.changesetCountsOverTime}
                        history={history}
                    />
                    <h3 className="mt-4 d-flex align-items-end mb-0">
                        {totalChangesetCount} {pluralize('Changeset', totalChangesetCount)}
                    </h3>
                    <CampaignChangesets
                        campaignID={campaign.id}
                        viewerCanAdminister={campaign.viewerCanAdminister}
                        changesetUpdates={changesetUpdates}
                        campaignUpdates={campaignUpdates}
                        history={history}
                        location={location}
                        isLightTheme={isLightTheme}
                        extensionsController={extensionsController}
                        platformContext={platformContext}
                        telemetryService={telemetryService}
                    />
                </>
            )}
        </>
    )
}
