import * as H from 'history'
import React, { useMemo, useState } from 'react'
import { useObservable } from '../../../../../shared/src/util/useObservable'
import { PageTitle } from '../../../components/PageTitle'
import {
    fetchCampaignSpecById as _fetchCampaignSpecById,
    queryChangesetSpecs,
    queryChangesetSpecFileDiffs,
} from './backend'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { CampaignHeader } from '../detail/CampaignHeader'
import { ChangesetSpecList } from './ChangesetSpecList'
import { ThemeProps } from '../../../../../shared/src/theme'
import { CreateUpdateCampaignAlert } from './CreateUpdateCampaignAlert'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import { HeroPage } from '../../../components/HeroPage'
import { CampaignDescription } from '../detail/CampaignDescription'

export interface CampaignApplyPageProps extends ThemeProps {
    specID: string
    history: H.History
    location: H.Location

    /** Used for testing. */
    fetchCampaignSpecById?: typeof _fetchCampaignSpecById
    /** Used for testing. */
    queryChangesetSpecs?: typeof queryChangesetSpecs
    /** Used for testing. */
    queryChangesetSpecFileDiffs?: typeof queryChangesetSpecFileDiffs
}

export const CampaignApplyPage: React.FunctionComponent<CampaignApplyPageProps> = ({
    specID,
    history,
    location,
    isLightTheme,
    fetchCampaignSpecById = _fetchCampaignSpecById,
    queryChangesetSpecs,
    queryChangesetSpecFileDiffs,
}) => {
    const [isLoading, setIsLoading] = useState<boolean | Error>(false)
    const spec = useObservable(useMemo(() => fetchCampaignSpecById(specID), [specID, fetchCampaignSpecById]))

    if (spec === undefined) {
        return (
            <div className="text-center">
                <LoadingSpinner className="icon-inline mx-auto my-4" />
            </div>
        )
    }
    if (spec === null) {
        return <HeroPage icon={AlertCircleIcon} title="Campaign spec not found" />
    }

    return (
        <>
            <PageTitle title="Apply campaign spec" />
            <CampaignHeader
                name={spec.description.name}
                namespace={spec.namespace}
                createdAt={spec.createdAt}
                creator={spec.creator}
                verb="Uploaded"
                className="mb-3 test-campaign-apply-page"
            />
            <CreateUpdateCampaignAlert
                history={history}
                specID={spec.id}
                campaign={spec.appliesToCampaign}
                isLoading={isLoading}
                setIsLoading={setIsLoading}
                viewerCanAdminister={spec.viewerCanAdminister}
            />
            <CampaignDescription history={history} description={spec.description.description} className="mb-3" />
            <ChangesetSpecList
                campaignSpecID={specID}
                history={history}
                location={location}
                isLightTheme={isLightTheme}
                queryChangesetSpecs={queryChangesetSpecs}
                queryChangesetSpecFileDiffs={queryChangesetSpecFileDiffs}
            />
            <CreateUpdateCampaignAlert
                history={history}
                specID={spec.id}
                campaign={spec.appliesToCampaign}
                isLoading={isLoading}
                setIsLoading={setIsLoading}
                viewerCanAdminister={spec.viewerCanAdminister}
            />
        </>
    )
}
