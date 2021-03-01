import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React, { useEffect, useMemo } from 'react'
import { HeroPage } from '../../../components/HeroPage'
import { PageTitle } from '../../../components/PageTitle'
import { isEqual } from 'lodash'
import {
    fetchBatchChangeByNamespace as _fetchCampaignByNamespace,
    queryChangesets as _queryChangesets,
    queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs,
    queryChangesetCountsOverTime as _queryChangesetCountsOverTime,
    deleteCampaign as _deleteCampaign,
} from './backend'
import { useObservable } from '../../../../../shared/src/util/useObservable'
import * as H from 'history'
import { delay, distinctUntilChanged, repeatWhen } from 'rxjs/operators'
import { ThemeProps } from '../../../../../shared/src/theme'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import { BatchChangeFields, Scalars } from '../../../graphql-operations'
import { Description } from '../Description'
import { CampaignStatsCard } from './CampaignStatsCard'
import { CampaignTabs } from './CampaignTabs'
import { CampaignDetailsActionSection } from './CampaignDetailsActionSection'
import { CampaignInfoByline } from './CampaignInfoByline'
import { UnpublishedNotice } from './UnpublishedNotice'
import { SupersedingCampaignSpecAlert } from './SupersedingCampaignSpecAlert'
import { CampaignsIcon } from '../icons'
import { PageHeader } from '../../../components/PageHeader'
import { ClosedNotice } from './ClosedNotice'

export interface CampaignDetailsPageProps
    extends ThemeProps,
        ExtensionsControllerProps,
        PlatformContextProps,
        TelemetryProps {
    /**
     * The namespace ID.
     */
    namespaceID: Scalars['ID']
    /**
     * The batch change name.
     */
    batchChangeName: BatchChangeFields['name']
    history: H.History
    location: H.Location

    /** For testing only. */
    fetchBatchChangeByNamespace?: typeof _fetchCampaignByNamespace
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
export const CampaignDetailsPage: React.FunctionComponent<CampaignDetailsPageProps> = ({
    namespaceID,
    batchChangeName,
    history,
    location,
    isLightTheme,
    extensionsController,
    platformContext,
    telemetryService,
    fetchBatchChangeByNamespace: fetchBatchChangeByNamespace = _fetchCampaignByNamespace,
    queryChangesets,
    queryExternalChangesetWithFileDiffs,
    queryChangesetCountsOverTime,
    deleteCampaign,
}) => {
    useEffect(() => {
        telemetryService.logViewEvent('CampaignDetailsPagePage')
    }, [telemetryService])

    const batchChange: BatchChangeFields | null | undefined = useObservable(
        useMemo(
            () =>
                fetchBatchChangeByNamespace(namespaceID, batchChangeName).pipe(
                    repeatWhen(notifier => notifier.pipe(delay(5000))),
                    distinctUntilChanged((a, b) => isEqual(a, b))
                ),
            [namespaceID, batchChangeName, fetchBatchChangeByNamespace]
        )
    )

    // Is loading.
    if (batchChange === undefined) {
        return (
            <div className="text-center">
                <LoadingSpinner className="icon-inline mx-auto my-4" />
            </div>
        )
    }
    // Campaign was not found
    if (batchChange === null) {
        return <HeroPage icon={AlertCircleIcon} title="Campaign not found" />
    }

    return (
        <>
            <PageTitle title={batchChange.name} />
            <PageHeader
                path={[
                    {
                        icon: CampaignsIcon,
                        to: '/campaigns',
                    },
                    { to: `${batchChange.namespace.url}/campaigns`, text: batchChange.namespace.namespaceName },
                    { text: batchChange.name },
                ]}
                byline={
                    <CampaignInfoByline
                        createdAt={batchChange.createdAt}
                        initialApplier={batchChange.initialApplier}
                        lastAppliedAt={batchChange.lastAppliedAt}
                        lastApplier={batchChange.lastApplier}
                    />
                }
                actions={
                    <CampaignDetailsActionSection
                        campaignID={batchChange.id}
                        campaignClosed={!!batchChange.closedAt}
                        deleteCampaign={deleteCampaign}
                        campaignNamespaceURL={batchChange.namespace.url}
                        history={history}
                    />
                }
                className="test-campaign-details-page mb-3"
            />
            <SupersedingCampaignSpecAlert spec={batchChange.currentSpec.supersedingCampaignSpec} />
            <ClosedNotice closedAt={batchChange.closedAt} className="mb-3" />
            <UnpublishedNotice
                unpublished={batchChange.changesetsStats.unpublished}
                total={batchChange.changesetsStats.total}
                className="mb-3"
            />
            <CampaignStatsCard
                closedAt={batchChange.closedAt}
                stats={batchChange.changesetsStats}
                diff={batchChange.diffStat}
                className="mb-3"
            />
            <Description history={history} description={batchChange.description} />
            <CampaignTabs
                batchChange={batchChange}
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
