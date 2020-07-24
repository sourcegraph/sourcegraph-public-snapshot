import React from 'react'
import { PageTitle } from '../../../components/PageTitle'
import { queryChangesets } from './backend'
import * as H from 'history'
import { ThemeProps } from '../../../../../shared/src/theme'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import { MinimalCampaign } from './CampaignArea'
import { RouteComponentProps } from 'react-router-dom'
import SourcePullIcon from 'mdi-react/SourcePullIcon'
import ChartPpfIcon from 'mdi-react/ChartPpfIcon'
import { CampaignPreamble } from './preamble/CampaignPreamble'
import { CampaignChangesetListPage } from './changesets/CampaignChangesetListPage'
import { CampaignBurndownPage } from './burndown/CampaignBurndownChartSection'
import { OverviewPagesArea } from '../../../components/overviewPagesArea/OverviewPagesArea'

export type CampaignUIMode = 'viewing' | 'deleting' | 'closing'

interface Props
    extends ThemeProps,
        ExtensionsControllerProps,
        PlatformContextProps,
        TelemetryProps,
        RouteComponentProps<{}> {
    campaign: MinimalCampaign
    history: H.History
    location: H.Location

    queryChangesets: typeof queryChangesets
}

/**
 * The area for a single campaign.
 */
export const CampaignDetailArea: React.FunctionComponent<Props> = ({ campaign, history, match, ...props }) => (
    <>
        <PageTitle title={campaign.name} />
        <div className="container mb-5">
            <CampaignPreamble campaign={campaign} history={history} />
        </div>
        <OverviewPagesArea
            context={{ campaign, history, ...props }}
            pages={[
                {
                    title: 'Changesets',
                    icon: SourcePullIcon,
                    count: campaign.changesets.totalCount,
                    path: '',
                    exact: true,
                    render: () => (
                        <div className="container mt-3">
                            <CampaignChangesetListPage campaign={campaign} history={history} {...props} />
                        </div>
                    ),
                },
                {
                    title: 'Burndown chart',
                    icon: ChartPpfIcon,
                    path: '/burndown',
                    exact: true,
                    render: () => (
                        <div className="container mt-3">
                            <CampaignBurndownPage campaign={campaign} history={history} />
                        </div>
                    ),
                },
            ]}
            location={props.location}
            match={match}
            className="mb-3"
        />
    </>
)
