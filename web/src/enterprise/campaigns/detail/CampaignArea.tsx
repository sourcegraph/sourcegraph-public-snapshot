import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React, { useCallback, useEffect, useMemo } from 'react'
import { RouteComponentProps } from 'react-router'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { HeroPage } from '../../../components/HeroPage'
import { InfoSidebar, InfoSidebarSection } from '../../../components/infoSidebar/InfoSidebar'
import { OverviewPagesArea } from '../../../components/overviewPagesArea/OverviewPagesArea'
import { WithSidebar } from '../../../components/withSidebar/WithSidebar'
import { CampaignDeleteButton } from '../common/CampaignDeleteButton'
import { CampaignForceRefreshButton } from '../common/CampaignForceRefreshButton'
import { NamespaceCampaignsAreaContext } from '../namespace/NamespaceCampaignsArea'
import { CampaignActivity } from './activity/CampaignActivity'
import { CampaignOverview } from './CampaignOverview'
import { CampaignFileDiffsList } from './fileDiffs/CampaignFileDiffsList'
import { CampaignRepositoriesList } from './repositories/CampaignRepositoriesList'
import { CampaignRulesListOLD } from './rules/CampaignRulesListOLD'
import { CampaignThreadsListPage } from './threads/CampaignThreadsListPage'
import { useCampaignByID } from './useCampaignByID'
import { RulesList } from '../../rules/list/RulesList'

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

const PAGE_CLASS_NAME = 'container my-5'

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

    const onCampaignDelete = useCallback(() => {
        if (campaign !== LOADING && campaign !== null && !isErrorLike(campaign)) {
            props.history.push(props.campaignsURL)
        }
    }, [campaign, props.campaignsURL, props.history])

    const sidebarSections = useMemo<InfoSidebarSection[]>(
        () =>
            campaign !== LOADING && campaign !== null && !isErrorLike(campaign)
                ? [
                      {
                          expanded: (
                              <CampaignForceRefreshButton
                                  {...props}
                                  campaign={campaign}
                                  onComplete={onCampaignUpdate}
                                  buttonClassName="btn-link"
                                  className="btn-sm px-0 text-decoration-none"
                              />
                          ),
                      },
                      {
                          expanded: (
                              <CampaignDeleteButton
                                  {...props}
                                  campaign={campaign}
                                  onDelete={onCampaignDelete}
                                  buttonClassName="btn-link"
                                  className="btn-sm px-0 text-decoration-none"
                              />
                          ),
                      },
                  ]
                : [],
        [campaign, onCampaignDelete, onCampaignUpdate, props]
    )

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
                        render: () => <RulesList {...context} container={campaign} className={PAGE_CLASS_NAME} />,
                    },
                    {
                        title: 'RulesOLD',
                        path: '/rulesOLD',
                        render: () => <CampaignRulesListOLD {...context} className={PAGE_CLASS_NAME} />,
                    },
                    {
                        title: 'Threads',
                        path: '/threads',
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
                    {
                        title: 'Impact',
                        path: '/impact',
                        render: () => 'TODO!(sqs)',
                    },
                ]}
                location={props.location}
                match={match}
            />
        </WithSidebar>
    )
}
