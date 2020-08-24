import React, { useState, useCallback } from 'react'
import * as H from 'history'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { ThemeProps } from '../../../../../shared/src/theme'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import { CampaignFields } from '../../../graphql-operations'
import { Subject } from 'rxjs'
import {
    queryChangesets as _queryChangesets,
    queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs,
    queryChangesetCountsOverTime as _queryChangesetCountsOverTime,
} from './backend'
import classNames from 'classnames'
import SourceBranchIcon from 'mdi-react/SourceBranchIcon'
import ChartLineVariantIcon from 'mdi-react/ChartLineVariantIcon'
import { CampaignBurndownChart } from './BurndownChart'
import { CampaignChangesets } from './changesets/CampaignChangesets'

type SelectedTab = 'changesets' | 'chart'

export interface CampaignTabsProps extends ExtensionsControllerProps, ThemeProps, PlatformContextProps, TelemetryProps {
    campaign: CampaignFields
    campaignUpdates: Subject<void>
    history: H.History
    location: H.Location
    /** For testing only. */
    queryChangesets?: typeof _queryChangesets
    /** For testing only. */
    queryExternalChangesetWithFileDiffs?: typeof _queryExternalChangesetWithFileDiffs
    /** For testing only. */
    queryChangesetCountsOverTime?: typeof _queryChangesetCountsOverTime
}

export const CampaignTabs: React.FunctionComponent<CampaignTabsProps> = ({
    extensionsController,
    history,
    isLightTheme,
    location,
    platformContext,
    telemetryService,
    campaign,
    campaignUpdates,
    queryChangesets,
    queryChangesetCountsOverTime,
    queryExternalChangesetWithFileDiffs,
}) => {
    const [selectedTab, setSelectedTab] = useState<SelectedTab>('changesets')
    const onSelectChangesets = useCallback<React.MouseEventHandler>(
        event => {
            event.preventDefault()
            setSelectedTab('changesets')
        },
        [setSelectedTab]
    )
    const onSelectChart = useCallback<React.MouseEventHandler>(
        event => {
            event.preventDefault()
            setSelectedTab('chart')
        },
        [setSelectedTab]
    )
    return (
        <>
            <ul className="nav nav-tabs mb-2">
                <li className="nav-item">
                    <a
                        href=""
                        onClick={onSelectChangesets}
                        className={classNames('nav-link', selectedTab === 'changesets' && 'active')}
                    >
                        <SourceBranchIcon className="icon-inline text-muted mr-1" /> Changesets
                    </a>
                </li>
                <li className="nav-item test-campaigns-chart-tab">
                    <a
                        href=""
                        onClick={onSelectChart}
                        className={classNames('nav-link', selectedTab === 'chart' && 'active')}
                    >
                        <ChartLineVariantIcon className="icon-inline text-muted mr-1" /> Burndown chart
                    </a>
                </li>
            </ul>
            {selectedTab === 'chart' && (
                <CampaignBurndownChart
                    campaignID={campaign.id}
                    queryChangesetCountsOverTime={queryChangesetCountsOverTime}
                    history={history}
                />
            )}
            {selectedTab === 'changesets' && (
                <CampaignChangesets
                    campaignID={campaign.id}
                    viewerCanAdminister={campaign.viewerCanAdminister}
                    changesetUpdates={campaignUpdates}
                    campaignUpdates={campaignUpdates}
                    history={history}
                    location={location}
                    isLightTheme={isLightTheme}
                    extensionsController={extensionsController}
                    platformContext={platformContext}
                    telemetryService={telemetryService}
                    queryChangesets={queryChangesets}
                    queryExternalChangesetWithFileDiffs={queryExternalChangesetWithFileDiffs}
                />
            )}
        </>
    )
}
