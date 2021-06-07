import * as H from 'history'
import ArchiveIcon from 'mdi-react/ArchiveIcon'
import ChartLineVariantIcon from 'mdi-react/ChartLineVariantIcon'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'
import MonitorStarIcon from 'mdi-react/MonitorStarIcon'
import SourceBranchIcon from 'mdi-react/SourceBranchIcon'
import React from 'react'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { BatchChangeFields } from '../../../graphql-operations'
import {
    BatchChangeTab,
    BatchChangeTabList,
    BatchChangeTabPanel,
    BatchChangeTabPanels,
    BatchChangeTabs,
} from '../BatchChangeTabs'
import { BatchSpec, BatchSpecDownloadLink, BatchSpecMeta } from '../BatchSpec'

import {
    queryChangesets as _queryChangesets,
    queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs,
    queryChangesetCountsOverTime as _queryChangesetCountsOverTime,
    queryBulkOperations as _queryBulkOperations,
} from './backend'
import { BatchChangeBurndownChart } from './BatchChangeBurndownChart'
import { BulkOperationsTab } from './BulkOperationsTab'
import { BatchChangeChangesets } from './changesets/BatchChangeChangesets'

export enum TabName {
    Changesets = 'changesets',
    Chart = 'chart',
    Spec = 'spec',
    Archived = 'archived',
    BulkOperations = 'bulkoperations',
}

/** `BatchChangeDetailsPage` and `BatchChangeDetailsTabs` share all these props */
export interface BatchChangeDetailsProps
    extends ThemeProps,
        ExtensionsControllerProps,
        PlatformContextProps,
        TelemetryProps {
    history: H.History
    location: H.Location

    /** For testing only. */
    queryChangesets?: typeof _queryChangesets
    /** For testing only. */
    queryExternalChangesetWithFileDiffs?: typeof _queryExternalChangesetWithFileDiffs
    /** For testing only. */
    queryChangesetCountsOverTime?: typeof _queryChangesetCountsOverTime
    /** For testing only. */
    queryBulkOperations?: typeof _queryBulkOperations
}

interface BatchChangeDetailsTabsProps extends BatchChangeDetailsProps {
    batchChange: BatchChangeFields
}

export const BatchChangeDetailsTabs: React.FunctionComponent<BatchChangeDetailsTabsProps> = ({
    batchChange,
    extensionsController,
    history,
    isLightTheme,
    location,
    platformContext,
    queryBulkOperations,
    queryChangesetCountsOverTime,
    queryChangesets,
    queryExternalChangesetWithFileDiffs,
    telemetryService,
}) => (
    <BatchChangeTabs history={history} location={location}>
        <BatchChangeTabList>
            <BatchChangeTab index={0} name={TabName.Changesets}>
                <span>
                    <SourceBranchIcon className="icon-inline text-muted mr-1" />
                    <span className="text-content" data-tab-content="Changesets">
                        Changesets
                    </span>{' '}
                    <span className="badge badge-pill badge-secondary ml-1">
                        {batchChange.changesetsStats.total - batchChange.changesetsStats.archived}
                    </span>
                </span>
            </BatchChangeTab>
            <BatchChangeTab index={1} name={TabName.Chart}>
                <span>
                    <ChartLineVariantIcon className="icon-inline text-muted mr-1" />{' '}
                    <span className="text-content" data-tab-content="Burndown chart">
                        Burndown chart
                    </span>
                </span>
            </BatchChangeTab>
            <BatchChangeTab index={2} name={TabName.Spec}>
                <span>
                    <FileDocumentIcon className="icon-inline text-muted mr-1" />{' '}
                    <span className="text-content" data-tab-content="Spec">
                        Spec
                    </span>
                </span>
            </BatchChangeTab>
            <BatchChangeTab index={3} name={TabName.Archived}>
                <span>
                    <ArchiveIcon className="icon-inline text-muted mr-1" />{' '}
                    <span className="text-content" data-tab-content="Archived">
                        Archived
                    </span>{' '}
                    <span className="badge badge-pill badge-secondary ml-1">
                        {batchChange.changesetsStats.archived}
                    </span>
                </span>
            </BatchChangeTab>
            <BatchChangeTab index={4} name={TabName.BulkOperations}>
                <span>
                    <MonitorStarIcon className="icon-inline text-muted mr-1" />{' '}
                    <span className="text-content" data-tab-content="Bulk operations">
                        Bulk operations
                    </span>{' '}
                    <span className="badge badge-pill badge-secondary ml-1">
                        {batchChange.bulkOperations.totalCount}
                    </span>
                </span>
            </BatchChangeTab>
        </BatchChangeTabList>
        <BatchChangeTabPanels>
            <BatchChangeTabPanel index={0}>
                <BatchChangeChangesets
                    batchChangeID={batchChange.id}
                    viewerCanAdminister={batchChange.viewerCanAdminister}
                    history={history}
                    location={location}
                    isLightTheme={isLightTheme}
                    extensionsController={extensionsController}
                    platformContext={platformContext}
                    telemetryService={telemetryService}
                    queryChangesets={queryChangesets}
                    queryExternalChangesetWithFileDiffs={queryExternalChangesetWithFileDiffs}
                    onlyArchived={false}
                />
            </BatchChangeTabPanel>
            <BatchChangeTabPanel index={1}>
                <BatchChangeBurndownChart
                    batchChangeID={batchChange.id}
                    queryChangesetCountsOverTime={queryChangesetCountsOverTime}
                    history={history}
                />
            </BatchChangeTabPanel>
            <BatchChangeTabPanel index={2}>
                <div className="d-flex flex-wrap justify-content-between align-items-baseline mb-2 test-batches-spec">
                    <BatchSpecMeta
                        createdAt={batchChange.createdAt}
                        lastApplier={batchChange.lastApplier}
                        lastAppliedAt={batchChange.lastAppliedAt}
                    />
                    <BatchSpecDownloadLink
                        name={batchChange.name}
                        originalInput={batchChange.currentSpec.originalInput}
                    />
                </div>
                <BatchSpec originalInput={batchChange.currentSpec.originalInput} />
            </BatchChangeTabPanel>
            <BatchChangeTabPanel index={3}>
                <BatchChangeChangesets
                    batchChangeID={batchChange.id}
                    viewerCanAdminister={batchChange.viewerCanAdminister}
                    history={history}
                    location={location}
                    isLightTheme={isLightTheme}
                    extensionsController={extensionsController}
                    platformContext={platformContext}
                    telemetryService={telemetryService}
                    queryChangesets={queryChangesets}
                    queryExternalChangesetWithFileDiffs={queryExternalChangesetWithFileDiffs}
                    onlyArchived={true}
                />
            </BatchChangeTabPanel>
            <BatchChangeTabPanel index={4}>
                <BulkOperationsTab
                    batchChangeID={batchChange.id}
                    history={history}
                    location={location}
                    queryBulkOperations={queryBulkOperations}
                />
            </BatchChangeTabPanel>
        </BatchChangeTabPanels>
    </BatchChangeTabs>
)
