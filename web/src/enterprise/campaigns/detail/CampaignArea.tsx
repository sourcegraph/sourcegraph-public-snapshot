import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import BookmarkOutlineIcon from 'mdi-react/BookmarkOutlineIcon'
import H from 'history'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import ForumIcon from 'mdi-react/ForumIcon'
import UserGroupIcon from 'mdi-react/UserGroupIcon'
import React, { useCallback, useEffect, useMemo } from 'react'
import { RouteComponentProps } from 'react-router'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { HeroPage } from '../../../components/HeroPage'
import { InfoSidebar, InfoSidebarSection } from '../../../components/infoSidebar/InfoSidebar'
import { OverviewPagesArea } from '../../../components/overviewPagesArea/OverviewPagesArea'
import { PageTitle } from '../../../components/PageTitle'
import { WithSidebar } from '../../../components/withSidebar/WithSidebar'
import { DiffIcon } from '../../../util/octicons'
import { ThreadsIcon } from '../../threads/icons'
import { CampaignDeleteButton } from '../common/CampaignDeleteButton'
import { CampaignForceRefreshButton } from '../common/CampaignForceRefreshButton'
import { NamespaceCampaignsAreaContext } from '../namespace/NamespaceCampaignsArea'
import { CampaignActivity } from './activity/CampaignActivity'
import { CampaignOverview } from './CampaignOverview'
import { CampaignDiagnostics } from './diagnostics/CampaignDiagnostics'
import { CampaignFileDiffsList } from './fileDiffs/CampaignFileDiffsList'
import { CampaignParticipantListPage } from './participants/CampaignParticipantListPage'
import { CampaignRepositoriesList } from './repositories/CampaignRepositoriesList'
import { CampaignThreadListPage } from './threads/CampaignThreadListPage'
import { useCampaignByID } from './useCampaignByID'
import { DiagnosticsIcon } from '../../../diagnostics/icons'
import { isDefined } from '../../../../../shared/src/util/types'
import SettingsIcon from 'mdi-react/SettingsIcon'
import { CampaignManagePage } from './manage/CampaignManagePage'

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
                      campaign.isDraft
                          ? {
                                expanded: (
                                    <>
                                        <div
                                            className="badge badge-secondary w-100 d-inline-flex align-items-center py-2 px-3 h6 mb-0 font-weight-bold"
                                            // eslint-disable-next-line react/forbid-dom-props
                                            style={{ fontSize: '0.85rem' }}
                                        >
                                            <BookmarkOutlineIcon className="icon-inline mr-1 flex-0" /> Draft campaign
                                        </div>
                                    </>
                                ),
                            }
                          : undefined,
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
                  ].filter(isDefined)
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
        <>
            <WithSidebar
                sidebarPosition="right"
                sidebar={<InfoSidebar sections={sidebarSections} />}
                className="flex-1"
            >
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
                            render: () => (
                                <>
                                    <PageTitle title={`Activity - ${campaign.name}`} />
                                    <CampaignActivity {...context} className={PAGE_CLASS_NAME} />
                                </>
                            ),
                        },
                        {
                            title: 'Diagnostics',
                            icon: DiagnosticsIcon,
                            count: campaign.diagnostics.totalCount,
                            path: '/diagnostics',
                            render: () => (
                                <>
                                    <PageTitle title={`Diagnostics - ${campaign.name}`} />
                                    <CampaignDiagnostics {...context} className={PAGE_CLASS_NAME} />
                                </>
                            ),
                            condition: () => campaign.diagnostics.totalCount > 0,
                        },

                        {
                            title: 'Changes',
                            icon: DiffIcon,
                            count: campaign.repositoryComparisons.reduce(
                                (n, c) => n + (c.fileDiffs.totalCount || 0),
                                0
                            ),
                            path: '/changes',
                            render: () => (
                                <>
                                    <PageTitle title={`Changes - ${campaign.name}`} />
                                    <div className={PAGE_CLASS_NAME}>
                                        <CampaignRepositoriesList {...context} showCommits={true} />
                                        <CampaignFileDiffsList {...context} platformContext={props.platformContext} />
                                    </div>
                                </>
                            ),
                            condition: () => campaign.repositoryComparisons.length > 0,
                        },
                        {
                            title: 'Threads',
                            icon: ThreadsIcon,
                            count: campaign.threads.totalCount,
                            path: '/threads',
                            render: () => (
                                <>
                                    <PageTitle title={`Threads - ${campaign.name}`} />
                                    <CampaignThreadListPage {...context} className={PAGE_CLASS_NAME} />
                                </>
                            ),
                            navbarDividerBefore: true,
                        },
                        {
                            title: 'Participants',
                            icon: UserGroupIcon,
                            count: campaign.participants.totalCount,
                            path: '/participants',
                            render: () => (
                                <>
                                    <PageTitle title={`Participants - ${campaign.name}`} />
                                    <CampaignParticipantListPage {...context} className={PAGE_CLASS_NAME} />
                                </>
                            ),
                        },
                        {
                            title: 'Manage',
                            icon: SettingsIcon,
                            path: '/manage',
                            render: ({ match }) => (
                                <>
                                    <PageTitle title={`Manage - ${campaign.name}`} />
                                    <CampaignManagePage {...context} match={match} className="container" />
                                </>
                            ),
                            hideInNavbar: true,
                            fullPage: true,
                        },
                    ]}
                    location={props.location}
                    match={match}
                />
            </WithSidebar>
        </>
    )
}
