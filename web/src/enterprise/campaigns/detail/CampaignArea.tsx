import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React, { useEffect, useMemo } from 'react'
import { RouteComponentProps } from 'react-router'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { HeroPage } from '../../../components/HeroPage'
import { InfoSidebar, InfoSidebarSection } from '../../../components/infoSidebar/InfoSidebar'
import { OverviewPagesArea } from '../../../components/overviewPagesArea/OverviewPagesArea'
import { WithSidebar } from '../../../components/withSidebar/WithSidebar'
import { NamespaceCampaignsAreaContext } from '../namespace/NamespaceCampaignsArea'
import { CampaignActivity } from './activity/CampaignActivity'
import { CampaignOverview } from './CampaignOverview'
import { CampaignFileDiffsList } from './fileDiffs/CampaignFileDiffsList'
import { CampaignRepositoriesList } from './repositories/CampaignRepositoriesList'
import { CampaignRulesList } from './rules/CampaignRulesList'
import { CampaignThreadsListPage } from './threads/CampaignThreadsListPage'
import { useCampaignByID } from './useCampaignByID'

export interface CampaignAreaContext
    extends Pick<NamespaceCampaignsAreaContext, Exclude<keyof NamespaceCampaignsAreaContext, 'namespace'>> {
    /** The campaign, queried from the GraphQL API. */
    campaign: GQL.ICampaign

    /** Called to refresh the campaign. */
    onCampaignUpdate: (update?: Partial<GQL.ICampaign>) => void

    location: H.Location
    history: H.History
}

interface Props
    extends Pick<CampaignAreaContext, Exclude<keyof CampaignAreaContext, 'campaign' | 'onCampaignUpdate'>>,
        RouteComponentProps<{}>,
        PlatformContextProps {
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

    const sidebarSections = useMemo<InfoSidebarSection[]>(() => {
        const a = 1
        return []
    }, [])

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
        <WithSidebar sidebarPosition="right" sidebar={<InfoSidebar sections={sidebarSections} />} className="flex-1">
            <OverviewPagesArea<CampaignAreaContext>
                context={context}
                header={header}
                overviewComponent={CampaignOverview}
                pages={[
                    {
                        title: 'Activity',
                        path: '/',
                        exact: true,
                        render: () => <CampaignActivity {...context} className={PAGE_CLASS_NAME} />,
                    },
                    {
                        title: 'Rules',
                        path: '/rules',
                        render: () => <CampaignRulesList {...context} className={PAGE_CLASS_NAME} />,
                    },
                    {
                        title: 'Changesets',
                        path: '/changesets',
                        render: () => <CampaignThreadsListPage {...context} className={PAGE_CLASS_NAME} />,
                    },
                    {
                        title: 'Commits',
                        path: '/commits',
                        render: () => (
                            <CampaignRepositoriesList {...context} showCommits={true} className={PAGE_CLASS_NAME} />
                        ),
                    },
                    {
                        title: 'Changes',
                        path: '/changes',
                        render: () => (
                            <CampaignFileDiffsList
                                {...context}
                                platformContext={props.platformContext}
                                className={PAGE_CLASS_NAME}
                            />
                        ),
                    },
                ]}
                location={props.location}
                match={match}
            />
        </WithSidebar>
    )
}
