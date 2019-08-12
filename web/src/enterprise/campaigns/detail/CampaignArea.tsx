import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import ForumIcon from 'mdi-react/ForumIcon'
import React, { useCallback, useEffect, useMemo } from 'react'
import { RouteComponentProps } from 'react-router'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { HeroPage } from '../../../components/HeroPage'
import { InfoSidebar, InfoSidebarSection } from '../../../components/infoSidebar/InfoSidebar'
import { OverviewPagesArea } from '../../../components/overviewPagesArea/OverviewPagesArea'
import { WithSidebar } from '../../../components/withSidebar/WithSidebar'
import { DiagnosticListByResource } from '../../../diagnostics/list/byResource/DiagnosticListByResource'
import { DiffIcon } from '../../../util/octicons'
import { DiagnosticsIcon } from '../../checks/icons'
import { RulesIcon } from '../../rules/icons'
import { RulesList } from '../../rules/list/RulesList'
import { ThreadsIcon } from '../../threads/icons'
import { CampaignDeleteButton } from '../common/CampaignDeleteButton'
import { CampaignForceRefreshButton } from '../common/CampaignForceRefreshButton'
import { NamespaceCampaignsAreaContext } from '../namespace/NamespaceCampaignsArea'
import { CampaignActivity } from './activity/CampaignActivity'
import { CampaignOverview } from './CampaignOverview'
import { CampaignDiagnostics } from './diagnostics/CampaignDiagnostics'
import { CampaignFileDiffsList } from './fileDiffs/CampaignFileDiffsList'
import { CampaignRepositoriesList } from './repositories/CampaignRepositoriesList'
import { CampaignThreadListPage } from './threads/CampaignThreadListPage'
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
                        icon: ForumIcon,
                        count: campaign.comments.totalCount - 1,
                        path: '',
                        exact: true,
                        render: () => <CampaignActivity {...context} className={PAGE_CLASS_NAME} />,
                    },
                    {
                        title: 'Diagnostics',
                        icon: DiagnosticsIcon,
                        count: campaign.diagnostics.totalCount,
                        path: '/diagnostics',
                        render: () => <CampaignDiagnostics {...context} className={PAGE_CLASS_NAME} />,
                        condition: () => campaign.diagnostics.totalCount > 0,
                    },

                    {
                        title: 'Changes',
                        icon: DiffIcon,
                        count: campaign.repositoryComparisons.reduce((n, c) => n + (c.fileDiffs.totalCount || 0), 0),
                        path: '/changes',
                        render: () => (
                            <div className={PAGE_CLASS_NAME}>
                                <CampaignRepositoriesList {...context} showCommits={true} />
                                <CampaignFileDiffsList {...context} platformContext={props.platformContext} />
                            </div>
                        ),
                        condition: () => campaign.repositoryComparisons.length > 0,
                    },
                    {
                        title: 'Threads',
                        icon: ThreadsIcon,
                        count: campaign.threads.totalCount,
                        path: '/threads',
                        render: () => <CampaignThreadListPage {...context} className={PAGE_CLASS_NAME} />,
                        navbarDividerBefore: true,
                    },
                    {
                        title: 'Rules',
                        icon: RulesIcon,
                        count: campaign.rules.totalCount,
                        path: '/rules',
                        render: ({ match }) => (
                            <RulesList {...context} container={campaign} match={match} className={PAGE_CLASS_NAME} />
                        ),
                    },
                ]}
                location={props.location}
                match={match}
            />
        </WithSidebar>
    )
}
