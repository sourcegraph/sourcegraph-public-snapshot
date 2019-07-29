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
import { CampaignRepositoriesList } from './repositories/CampaignRepositoriesList'
import { CampaignThreadsListPage } from './threads/CampaignThreadsListPage'
import { useCampaignByID } from './useCampaignByID'

export interface CampaignAreaContext
    extends Pick<NamespaceCampaignsAreaContext, Exclude<keyof NamespaceCampaignsAreaContext, 'namespace'>> {
    /** The campaign, queried from the GraphQL API. */
    campaign: GQL.ICampaign

    /** Called to refresh the campaign. */
    onCampaignUpdate: () => void

    location: H.Location
    history: H.History
}

interface Props
    extends Pick<CampaignAreaContext, Exclude<keyof CampaignAreaContext, 'campaign' | 'onCampaignUpdate'>>,
        RouteComponentProps<{}> {
    /** The campaign ID. */
    campaignID: GQL.ID

    header: React.ReactFragment
}

const LOADING = 'loading' as const

const PAGE_CLASS_NAME = 'container mt-4'

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
    const [campaign, onCampaignUpdate] = useCampaignByID(campaignID)

    useEffect(() => {
        if (setBreadcrumbItem) {
            setBreadcrumbItem(
                campaign !== LOADING && campaign !== null && !isErrorLike(campaign)
                    ? { text: campaign.name, to: campaign.url }
                    : undefined
            )
        }
        return () => {
            if (setBreadcrumbItem) {
                setBreadcrumbItem(undefined)
            }
        }
    }, [campaign, setBreadcrumbItem])

    if (campaign === LOADING) {
        return <LoadingSpinner className="icon-inline mx-auto my-4" />
    }
    if (campaign === null) {
        return <HeroPage icon={AlertCircleIcon} title="Campaign not found" />
    }
    if (isErrorLike(campaign)) {
        return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={campaign.message} />
    }

    const context: CampaignAreaContext = {
        ...props,
        campaign,
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
                pages={[
                    {
                        title: 'Threads',
                        path: '',
                        render: () => <CampaignThreadsListPage {...context} className={PAGE_CLASS_NAME} />,
                    },
                    {
                        title: 'Commits',
                        path: '/commits',
                        render: () => <CampaignRepositoriesList {...context} className={PAGE_CLASS_NAME} />,
                    },
                    // { title: 'Changes', path: '/changes', render: () => <CampaignChangesListPage {...context} /> },
                ]}
                location={props.location}
                match={match}
            />
        </>
    )
}
