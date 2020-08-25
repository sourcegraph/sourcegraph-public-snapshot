import React, { useState, useMemo, useEffect } from 'react'
import * as H from 'history'
import { PageTitle } from '../../../components/PageTitle'
import { CampaignHeader } from '../detail/CampaignHeader'
import { CampaignCloseAlert } from './CampaignCloseAlert'
import { Scalars } from '../../../graphql-operations'
import {
    queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs,
    queryChangesets as _queryChangesets,
    fetchCampaignById as _fetchCampaignById,
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
import { BreadcrumbSetters } from '../../../components/Breadcrumbs'
import { Link } from '../../../../../shared/src/components/Link'
import { CampaignInfoByline } from '../detail/CampaignInfoByline'

export interface CampaignClosePageProps
    extends ThemeProps,
        TelemetryProps,
        PlatformContextProps,
        BreadcrumbSetters,
        ExtensionsControllerProps {
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
    closeCampaign?: typeof _closeCampaign
}

export const CampaignClosePage: React.FunctionComponent<CampaignClosePageProps> = ({
    campaignID,
    history,
    location,
    extensionsController,
    isLightTheme,
    platformContext,
    telemetryService,
    setBreadcrumb,
    fetchCampaignById = _fetchCampaignById,
    queryChangesets,
    queryExternalChangesetWithFileDiffs,
    closeCampaign,
}) => {
    const [closeChangesets, setCloseChangesets] = useState<boolean>(false)
    const campaign = useObservable(useMemo(() => fetchCampaignById(campaignID), [campaignID, fetchCampaignById]))

    useEffect(() => {
        if (campaign) {
            const subscription = setBreadcrumb({
                element: <Link to={campaign.url}>{campaign.name}</Link>,
                key: 'CampaignClosePage',
            })
            return () => subscription.unsubscribe()
        }
        return () => undefined
    }, [campaign, setBreadcrumb])

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
            <CampaignHeader name={campaign.name} namespace={campaign.namespace} className="test-campaign-close-page" />
            <CampaignInfoByline
                createdAt={campaign.createdAt}
                initialApplier={campaign.initialApplier}
                lastAppliedAt={campaign.lastAppliedAt}
                lastApplier={campaign.lastApplier}
                className="mb-3"
            />
            <CampaignCloseAlert
                campaignID={campaignID}
                campaignURL={campaign.url}
                closeChangesets={closeChangesets}
                setCloseChangesets={setCloseChangesets}
                history={history}
                closeCampaign={closeCampaign}
                viewerCanAdminister={campaign.viewerCanAdminister}
            />
            {closeChangesets && (
                <h2 className="test-campaigns-close-willclose-header">
                    Closing the campaign will close the following changesets:
                </h2>
            )}
            {!closeChangesets && <h2>The following changesets will remain open:</h2>}
            <CampaignCloseChangesetsList
                campaignID={campaignID}
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
            />
        </>
    )
}
