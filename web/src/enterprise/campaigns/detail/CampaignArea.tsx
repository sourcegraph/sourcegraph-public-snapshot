import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React, { useEffect } from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { ErrorBoundary } from '../../../components/ErrorBoundary'
import { HeroPage } from '../../../components/HeroPage'
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
            <div className="d-flex align-items-center justify-content-between border-top border-bottom py-3 my-3">
                <div className="d-flex align-items-center">
                    <div className="badge border border-success text-success font-size-base py-2 px-3 mr-3">Active</div>
                    Last action 11 minutes ago
                </div>
                <div>
                    <Link className="btn btn-secondary mr-2" to="#editTODO!(sqs)">
                        Delete
                    </Link>
                </div>
            </div>
            <h2>{campaignOrError.name}</h2>
            <div className="flex-1 d-flex flex-column overflow-auto">
                <ErrorBoundary location={props.location}>
                    <CampaignThreadsListPage {...context} />
                </ErrorBoundary>
            </div>
        </div>
    )
}
