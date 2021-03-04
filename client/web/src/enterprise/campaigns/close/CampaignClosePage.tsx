import React, { useState, useMemo, useCallback } from 'react'
import * as H from 'history'
import { PageTitle } from '../../../components/PageTitle'
import { CampaignCloseAlert } from './CampaignCloseAlert'
import { BatchChangeChangesetsResult, BatchChangeFields, Scalars } from '../../../graphql-operations'
import {
    queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs,
    queryChangesets as _queryChangesets,
    fetchBatchChangeByNamespace as _fetchCampaignByNamespace,
} from '../detail/backend'
import { ThemeProps } from '../../../../../shared/src/theme'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import { closeCampaign as _closeCampaign } from './backend'
import { CampaignCloseChangesetsList } from './CampaignCloseChangesetsList'
import { useObservable } from '../../../../../shared/src/util/useObservable'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { HeroPage } from '../../../components/HeroPage'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import { CampaignInfoByline } from '../detail/CampaignInfoByline'
import { ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { CampaignsIcon } from '../icons'
import { PageHeader } from '../../../components/PageHeader'

export interface CampaignClosePageProps
    extends ThemeProps,
        TelemetryProps,
        PlatformContextProps,
        ExtensionsControllerProps {
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
    closeCampaign?: typeof _closeCampaign
}

export const CampaignClosePage: React.FunctionComponent<CampaignClosePageProps> = ({
    namespaceID,
    batchChangeName: campaignName,
    history,
    location,
    extensionsController,
    isLightTheme,
    platformContext,
    telemetryService,
    fetchBatchChangeByNamespace: fetchCampaignByNamespace = _fetchCampaignByNamespace,
    queryChangesets,
    queryExternalChangesetWithFileDiffs,
    closeCampaign,
}) => {
    const [closeChangesets, setCloseChangesets] = useState<boolean>(false)
    const campaign = useObservable(
        useMemo(() => fetchCampaignByNamespace(namespaceID, campaignName), [
            namespaceID,
            campaignName,
            fetchCampaignByNamespace,
        ])
    )

    const [totalCount, setTotalCount] = useState<number>()

    const onFetchChangesets = useCallback(
        (
            connection?: (BatchChangeChangesetsResult['node'] & { __typename: 'BatchChange' })['changesets'] | ErrorLike
        ) => {
            if (!connection || isErrorLike(connection)) {
                return
            }
            setTotalCount(connection.totalCount)
        },
        []
    )

    // Is loading.
    if (campaign === undefined) {
        return (
            <div className="text-center">
                <LoadingSpinner className="icon-inline mx-auto my-4" />
            </div>
        )
    }

    // Campaign not found.
    if (campaign === null) {
        return <HeroPage icon={AlertCircleIcon} title="Campaign not found" />
    }

    return (
        <>
            <PageTitle title="Preview close" />
            <PageHeader
                path={[
                    {
                        icon: CampaignsIcon,
                        to: '/campaigns',
                    },
                    { to: `${campaign.namespace.url}/campaigns`, text: campaign.namespace.namespaceName },
                    { text: campaign.name },
                ]}
                byline={
                    <CampaignInfoByline
                        createdAt={campaign.createdAt}
                        initialApplier={campaign.initialApplier}
                        lastAppliedAt={campaign.lastAppliedAt}
                        lastApplier={campaign.lastApplier}
                    />
                }
                className="test-campaign-close-page mb-3"
            />
            {totalCount !== undefined && (
                <CampaignCloseAlert
                    campaignID={campaign.id}
                    campaignURL={campaign.url}
                    closeChangesets={closeChangesets}
                    setCloseChangesets={setCloseChangesets}
                    history={history}
                    closeCampaign={closeCampaign}
                    viewerCanAdminister={campaign.viewerCanAdminister}
                    totalCount={totalCount}
                />
            )}
            <CampaignCloseChangesetsList
                campaignID={campaign.id}
                history={history}
                location={location}
                viewerCanAdminister={campaign.viewerCanAdminister}
                extensionsController={extensionsController}
                isLightTheme={isLightTheme}
                platformContext={platformContext}
                telemetryService={telemetryService}
                queryChangesets={queryChangesets}
                queryExternalChangesetWithFileDiffs={queryExternalChangesetWithFileDiffs}
                willClose={closeChangesets}
                onUpdate={onFetchChangesets}
            />
        </>
    )
}
