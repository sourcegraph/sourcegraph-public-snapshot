import React, { useMemo, useCallback } from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PageTitle } from '../../../components/PageTitle'
import { UserAvatar } from '../../../user/UserAvatar'
import { Timestamp } from '../../../components/time/Timestamp'
import {
    queryPatchesFromCampaign,
    queryPatchesFromPatchSet,
    queryChangesets,
    queryPatchFileDiffs,
    fetchPatchSetById,
} from './backend'
import * as H from 'history'
import { CampaignBurndownChart } from './BurndownChart'
import { AddChangesetForm } from './AddChangesetForm'
import { Subject, NEVER, Observable } from 'rxjs'
import { renderMarkdown } from '../../../../../shared/src/util/markdown'
import { Markdown } from '../../../../../shared/src/components/Markdown'
import { ThemeProps } from '../../../../../shared/src/theme'
import { CampaignStatus } from './CampaignStatus'
import { CampaignChangesets } from './changesets/CampaignChangesets'
import { CampaignDiffStat } from './CampaignDiffStat'
import { pluralize } from '../../../../../shared/src/util/strings'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import { CampaignPatches } from './patches/CampaignPatches'
import { PatchSetPatches } from './patches/PatchSetPatches'
import { MinimalCampaign, MinimalPatchSet } from './CampaignArea'
import { useObservable } from '../../../../../shared/src/util/useObservable'
import { CampaignDescription } from './CampaignDescription'
import { CampaignInfoBar } from './CampaignInfoBar'
import { Timeline } from '../../../components/timeline/Timeline'
import { Link, Switch, Route, RouteComponentProps } from 'react-router-dom'
import { CampaignChangesetsEditButton } from './changesets/CampaignChangesetsEditButton'
import { CampaignChangesetsAddExistingButton } from './changesets/CampaignChangesetsAddExistingButton'
import { CampaignChangesets2 } from './changesets/CampaignChangesets2'
import { CampaignsIcon } from '../icons'
import { ChangesetStateIcon } from './changesets/ChangesetStateIcon'
import { changesetStateIcons } from './changesets/presentation'
import SourcePullIcon from 'mdi-react/SourcePullIcon'
import ChartPpfIcon from 'mdi-react/ChartPpfIcon'
import SourceMergeIcon from 'mdi-react/SourceMergeIcon'
import ProgressCheckIcon from 'mdi-react/ProgressCheckIcon'
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

    fetchPatchSetById: typeof fetchPatchSetById | ((patchSet: GQL.ID) => Observable<MinimalPatchSet | null>)
    queryPatchesFromCampaign: typeof queryPatchesFromCampaign
    queryPatchesFromPatchSet: typeof queryPatchesFromPatchSet
    queryPatchFileDiffs: typeof queryPatchFileDiffs
    queryChangesets: typeof queryChangesets
    _noSubject?: boolean
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
