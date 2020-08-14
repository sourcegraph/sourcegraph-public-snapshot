import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React, { useEffect, useMemo, useState, useCallback } from 'react'
import { HeroPage } from '../../../components/HeroPage'
import { PageTitle } from '../../../components/PageTitle'
import { isEqual } from 'lodash'
import {
    fetchCampaignById as _fetchCampaignById,
    queryChangesets as _queryChangesets,
    queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs,
} from './backend'
import { useObservable } from '../../../../../shared/src/util/useObservable'
import * as H from 'history'
import { Subject, of, merge } from 'rxjs'
import { switchMap, distinctUntilChanged } from 'rxjs/operators'
import { ThemeProps } from '../../../../../shared/src/theme'
import { CampaignChangesets } from './changesets/CampaignChangesets'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import { CampaignFields, Scalars } from '../../../graphql-operations'
import { CampaignDescription } from './CampaignDescription'
import { CampaignStatsCard } from './CampaignStatsCard'
import { CampaignHeader } from './CampaignHeader'
import SourceBranchIcon from 'mdi-react/SourceBranchIcon'
import ChartLineVariantIcon from 'mdi-react/ChartLineVariantIcon'
import { CampaignBurndownChart } from './BurndownChart'
import classNames from 'classnames'

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
                className="mb-3"
            />
            <CampaignStatsCard closedAt={campaign.closedAt} stats={campaign.changesets.stats} className="mb-3" />
            <CampaignDescription history={history} description={campaign.description} />
            <CampaignTabs
                campaignID={campaignID}
                campaign={campaign}
                campaignUpdates={campaignUpdates}
                extensionsController={extensionsController}
                history={history}
                isLightTheme={isLightTheme}
                location={location}
                platformContext={platformContext}
                telemetryService={telemetryService}
                fetchCampaignById={fetchCampaignById}
                queryChangesets={queryChangesets}
                queryExternalChangesetWithFileDiffs={queryExternalChangesetWithFileDiffs}
            />
        </>
    )
}

type SelectedTab = 'changesets' | 'chart'

const CampaignTabs: React.FunctionComponent<
    CampaignDetailsProps & { campaign: CampaignFields; campaignUpdates: Subject<void> }
> = ({
    extensionsController,
    history,
    isLightTheme,
    location,
    platformContext,
    telemetryService,
    campaign,
    campaignUpdates,
    queryChangesets,
    queryExternalChangesetWithFileDiffs,
}) => {
    const [selectedTab, setSelectedTab] = useState<SelectedTab>('changesets')
    const onSelectChangesets = useCallback<React.MouseEventHandler>(
        event => {
            event.preventDefault()
            setSelectedTab('changesets')
        },
        [setSelectedTab]
    )
    const onSelectChart = useCallback<React.MouseEventHandler>(
        event => {
            event.preventDefault()
            setSelectedTab('chart')
        },
        [setSelectedTab]
    )
    return (
        <>
            <ul className="nav nav-tabs mb-2">
                <li className="nav-item">
                    <a
                        href=""
                        onClick={onSelectChangesets}
                        className={classNames('nav-link', selectedTab === 'changesets' && 'active')}
                    >
                        <SourceBranchIcon className="icon-inline text-muted mr-1" /> Changesets
                    </a>
                </li>
                <li className="nav-item">
                    <a
                        href=""
                        onClick={onSelectChart}
                        className={classNames('nav-link', selectedTab === 'chart' && 'active')}
                    >
                        <ChartLineVariantIcon className="icon-inline text-muted mr-1" /> Burndown chart
                    </a>
                </li>
            </ul>
            {selectedTab === 'chart' && (
                <CampaignBurndownChart changesetCountsOverTime={campaign.changesetCountsOverTime} history={history} />
            )}
            {selectedTab === 'changesets' && (
                <CampaignChangesets
                    campaignID={campaign.id}
                    viewerCanAdminister={campaign.viewerCanAdminister}
                    changesetUpdates={campaignUpdates}
                    campaignUpdates={campaignUpdates}
                    history={history}
                    location={location}
                    isLightTheme={isLightTheme}
                    extensionsController={extensionsController}
                    platformContext={platformContext}
                    telemetryService={telemetryService}
                    queryChangesets={queryChangesets}
                    queryExternalChangesetWithFileDiffs={queryExternalChangesetWithFileDiffs}
                />
            )}
        </>
    )
}
