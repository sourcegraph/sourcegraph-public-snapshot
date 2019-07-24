import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React, { useEffect } from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { ErrorBoundary } from '../../../components/ErrorBoundary'
import { HeroPage } from '../../../components/HeroPage'
import { ThreadlikeArea } from '../../threadlike/ThreadlikeArea'
import { NamespaceCampaignsAreaContext } from '../namespace/NamespaceCampaignsArea'
import { CampaignThreadsListPage } from './threads/CampaignThreadsListPage'
import { useCampaignByID } from './useCampaignByID'

export interface CampaignAreaContext
    extends Pick<NamespaceCampaignsAreaContext, Exclude<keyof NamespaceCampaignsAreaContext, 'namespace'>> {
    /** The campaign ID. */
    campaignID: GQL.ID

    /** The campaign, queried from the GraphQL API. */
    campaign: GQL.ICampaign

    location: H.Location
    history: H.History
}

interface Props extends Pick<CampaignAreaContext, Exclude<keyof CampaignAreaContext, 'campaign'>> {}

const LOADING = 'loading' as const

/**
 * The area for a single campaign.
 */
export const CampaignArea: React.FunctionComponent<Props> = ({ campaignID, setBreadcrumbItem, ...props }) => {
    const campaignOrError = useCampaignByID(campaignID)

    useEffect(() => {
        if (setBreadcrumbItem) {
            setBreadcrumbItem(
                campaignOrError !== LOADING && campaignOrError !== null && !isErrorLike(campaignOrError)
                    ? { text: campaignOrError.name, to: campaignOrError.url }
                    : undefined
            )
        }
        return () => {
            if (setBreadcrumbItem) {
                setBreadcrumbItem(undefined)
            }
        }
    }, [campaignOrError, setBreadcrumbItem])

    if (campaignOrError === LOADING) {
        return <LoadingSpinner className="icon-inline mx-auto my-4" />
    }
    if (campaignOrError === null) {
        return <HeroPage icon={AlertCircleIcon} title="Campaign not found" />
    }
    if (isErrorLike(campaignOrError)) {
        return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={campaignOrError.message} />
    }

    const context: CampaignAreaContext = {
        ...props,
        campaignID,
        campaign: campaignOrError,
        setBreadcrumbItem,
    }

    return (
        <div className="campaign-area flex-1">
            <style>{`.user-area-header, .org-header { display: none; } /* TODO!(sqs): hack */`}</style>
            <ThreadlikeArea
                {...context}
                overviewComponent={CampaignOverview}
                pages={[{ title: 'Threads', path: '', render: () => <CampaignThreadsListPage {...context} /> }]}
            />
        </div>
    )
}
