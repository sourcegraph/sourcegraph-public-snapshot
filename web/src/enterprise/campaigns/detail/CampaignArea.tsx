import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React, { useEffect } from 'react'
import { RouteComponentProps } from 'react-router'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { HeroPage } from '../../../components/HeroPage'
import { OverviewPagesArea } from '../../../components/overviewPagesArea/OverviewPagesArea'
import { NamespaceCampaignsAreaContext } from '../namespace/NamespaceCampaignsArea'
import { CampaignOverview } from './CampaignOverview'
import { CampaignThreadsListPage } from './threads/CampaignThreadsListPage'
import { useCampaignByID } from './useCampaignByID'

export interface CampaignAreaContext
    extends Pick<NamespaceCampaignsAreaContext, Exclude<keyof NamespaceCampaignsAreaContext, 'namespace'>> {
    /** The campaign ID. */
    campaignID: GQL.ID

    /** The campaign, queried from the GraphQL API. */
    campaign: GQL.ICampaign

    /** Called to refresh the campaign. */
    onCampaignUpdate: () => void

    location: H.Location
    history: H.History
}

interface Props
    extends Pick<CampaignAreaContext, Exclude<keyof CampaignAreaContext, 'campaign' | 'onCampaignUpdate'>>,
        RouteComponentProps<never> {
    header: React.ReactFragment
}

const LOADING = 'loading' as const

/**
 * The area for a single campaign.
 */
export const CampaignArea: React.FunctionComponent<Props> = ({
    header,
    campaignID,
    setBreadcrumbItem,
    match,
    ...props
}) => {
    const [campaignOrError, onCampaignUpdate] = useCampaignByID(campaignID)

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
        onCampaignUpdate,
        setBreadcrumbItem,
    }

    return (
        <>
            <style>{`.user-area-header, .org-header { display: none; } .org-area > .container, .user-area > .container { margin: unset; margin-top: unset !important; width: unset; padding: unset; } /* TODO!(sqs): hack */`}</style>
            <OverviewPagesArea<CampaignAreaContext>
                context={context}
                header={header}
                overviewComponent={CampaignOverview}
                pages={[{ title: 'Threads', path: '', render: () => <CampaignThreadsListPage {...context} /> }]}
                location={props.location}
                match={match}
            />
        </>
    )
}
