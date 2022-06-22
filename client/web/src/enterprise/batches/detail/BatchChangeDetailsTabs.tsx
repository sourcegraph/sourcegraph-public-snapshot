import React, { useMemo } from 'react'

import * as H from 'history'
import ArchiveIcon from 'mdi-react/ArchiveIcon'
import ChartLineVariantIcon from 'mdi-react/ChartLineVariantIcon'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'
import MonitorStarIcon from 'mdi-react/MonitorStarIcon'
import SourceBranchIcon from 'mdi-react/SourceBranchIcon'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { BatchSpecSource } from '@sourcegraph/shared/src/schema'
import { Settings, SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Badge, Container, Icon } from '@sourcegraph/wildcard'

import { isBatchChangesExecutionEnabled } from '../../../batches'
import { BatchSpecState, BatchChangeFields } from '../../../graphql-operations'
import {
    BatchChangeTab,
    BatchChangeTabList,
    BatchChangeTabPanel,
    BatchChangeTabPanels,
    BatchChangeTabs,
} from '../BatchChangeTabs'
import { BatchSpec, BatchSpecDownloadButton, BatchSpecMeta } from '../BatchSpec'
import { BatchChangeBatchSpecList } from '../BatchSpecsPage'

import {
    queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs,
    queryChangesetCountsOverTime as _queryChangesetCountsOverTime,
    queryAllChangesetIDs as _queryAllChangesetIDs,
} from './backend'
import { BatchChangeBurndownChart } from './BatchChangeBurndownChart'
import { BulkOperationsTab } from './BulkOperationsTab'
import { BatchChangeChangesets } from './changesets/BatchChangeChangesets'

import styles from './BatchChangeDetailsTabs.module.scss'

export enum TabName {
    Changesets = 'changesets',
    Chart = 'chart',
    // Non-SSBC
    Spec = 'spec',
    // SSBC-only
    Executions = 'executions',
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
    /** The name of the tab that should be initially open */
    initialTab?: TabName

    /** For testing only. */
    queryExternalChangesetWithFileDiffs?: typeof _queryExternalChangesetWithFileDiffs
    /** For testing only. */
    queryChangesetCountsOverTime?: typeof _queryChangesetCountsOverTime
    /** For testing only. */
    queryAllChangesetIDs?: typeof _queryAllChangesetIDs
}

interface BatchChangeDetailsTabsProps extends BatchChangeDetailsProps, SettingsCascadeProps<Settings> {
    batchChange: BatchChangeFields
    refetchBatchChange: () => void
}

export const BatchChangeDetailsTabs: React.FunctionComponent<React.PropsWithChildren<BatchChangeDetailsTabsProps>> = ({
    batchChange,
    extensionsController,
    history,
    isLightTheme,
    location,
    platformContext,
    settingsCascade,
    initialTab = TabName.Changesets,
    queryChangesetCountsOverTime,
    queryExternalChangesetWithFileDiffs,
    queryAllChangesetIDs,
    refetchBatchChange,
    telemetryService,
}) => {
    const isExecutionEnabled = isBatchChangesExecutionEnabled(settingsCascade)

    const executingCount = useMemo(
        () =>
            batchChange.batchSpecs.nodes.filter(
                node => node.state === BatchSpecState.PROCESSING || node.state === BatchSpecState.QUEUED
            ).length,
        [batchChange.batchSpecs.nodes]
    )

    const isBatchSpecLocallyCreated = batchChange.currentSpec.source === BatchSpecSource.LOCAL
    const shouldDisplayOldUI = !isExecutionEnabled || isBatchSpecLocallyCreated

    return (
        <BatchChangeTabs history={history} location={location} initialTab={initialTab}>
            <BatchChangeTabList>
                <BatchChangeTab index={0} name={TabName.Changesets}>
                    <span>
                        <Icon aria-hidden={true} className="text-muted mr-2" as={SourceBranchIcon} />
                        <span className="text-content" data-tab-content="Changesets">
                            Changesets
                        </span>
                        <Badge variant="secondary" pill={true} className="ml-2">
                            {batchChange.changesetsStats.total - batchChange.changesetsStats.archived}
                        </Badge>
                    </span>
                </BatchChangeTab>
                <BatchChangeTab index={1} name={TabName.Chart}>
                    <span>
                        <Icon aria-hidden={true} className="text-muted mr-2" as={ChartLineVariantIcon} />
                        <span className="text-content" data-tab-content="Burndown chart">
                            Burndown chart
                        </span>
                    </span>
                </BatchChangeTab>
                {shouldDisplayOldUI ? (
                    <BatchChangeTab index={2} name={TabName.Spec}>
                        <span>
                            <Icon aria-hidden={true} className="text-muted mr-2" as={FileDocumentIcon} />
                            <span className="text-content" data-tab-content="Spec">
                                Spec
                            </span>
                        </span>
                    </BatchChangeTab>
                ) : (
                    <BatchChangeTab index={2} name={TabName.Executions} customPath="/executions">
                        <span>
                            <Icon aria-hidden={true} className="text-muted mr-2" as={FileDocumentIcon} />
                            <span className="text-content" data-tab-content="Executions">
                                Executions
                            </span>
                            {executingCount > 0 && (
                                <Badge variant="warning" pill={true} className="ml-2">
                                    {executingCount} {batchChange.batchSpecs.pageInfo.hasNextPage && <>+</>}
                                </Badge>
                            )}
                        </span>
                    </BatchChangeTab>
                )}
                <BatchChangeTab index={3} name={TabName.Archived}>
                    <span>
                        <Icon aria-hidden={true} className="text-muted mr-2" as={ArchiveIcon} />
                        <span className="text-content" data-tab-content="Archived">
                            Archived
                        </span>
                        <Badge variant="secondary" pill={true} className="ml-2">
                            {batchChange.changesetsStats.archived}
                        </Badge>
                    </span>
                </BatchChangeTab>
                <BatchChangeTab index={4} name={TabName.BulkOperations}>
                    <span>
                        <Icon aria-hidden={true} className="text-muted mr-2" as={MonitorStarIcon} />
                        <span className="text-content" data-tab-content="Bulk operations">
                            Bulk operations
                        </span>
                        <Badge variant="secondary" pill={true} className="ml-2">
                            {batchChange.bulkOperations.totalCount}
                        </Badge>
                    </span>
                </BatchChangeTab>
            </BatchChangeTabList>
            <BatchChangeTabPanels>
                <BatchChangeTabPanel>
                    <BatchChangeChangesets
                        batchChangeID={batchChange.id}
                        batchChangeState={batchChange.state}
                        viewerCanAdminister={batchChange.viewerCanAdminister}
                        refetchBatchChange={refetchBatchChange}
                        history={history}
                        location={location}
                        isLightTheme={isLightTheme}
                        extensionsController={extensionsController}
                        platformContext={platformContext}
                        telemetryService={telemetryService}
                        queryExternalChangesetWithFileDiffs={queryExternalChangesetWithFileDiffs}
                        queryAllChangesetIDs={queryAllChangesetIDs}
                        onlyArchived={false}
                        settingsCascade={settingsCascade}
                        isExecutionEnabled={isExecutionEnabled}
                    />
                </BatchChangeTabPanel>
                <BatchChangeTabPanel>
                    <BatchChangeBurndownChart
                        batchChangeID={batchChange.id}
                        queryChangesetCountsOverTime={queryChangesetCountsOverTime}
                        history={history}
                    />
                </BatchChangeTabPanel>
                <BatchChangeTabPanel>
                    {shouldDisplayOldUI ? (
                        <>
                            <div className="d-flex flex-wrap justify-content-between align-items-baseline mb-2 test-batches-spec">
                                <BatchSpecMeta
                                    createdAt={batchChange.createdAt}
                                    lastApplier={batchChange.lastApplier}
                                    lastAppliedAt={batchChange.lastAppliedAt}
                                />
                                <BatchSpecDownloadButton
                                    name={batchChange.name}
                                    isLightTheme={isLightTheme}
                                    originalInput={batchChange.currentSpec.originalInput}
                                />
                            </div>
                            <Container>
                                <BatchSpec
                                    name={batchChange.name}
                                    originalInput={batchChange.currentSpec.originalInput}
                                    isLightTheme={isLightTheme}
                                    className={styles.batchSpec}
                                />
                            </Container>
                        </>
                    ) : (
                        <Container>
                            <BatchChangeBatchSpecList
                                history={history}
                                location={location}
                                batchChangeID={batchChange.id}
                                currentSpecID={batchChange.currentSpec.id}
                                isLightTheme={isLightTheme}
                            />
                        </Container>
                    )}
                </BatchChangeTabPanel>
                <BatchChangeTabPanel>
                    <BatchChangeChangesets
                        batchChangeID={batchChange.id}
                        batchChangeState={batchChange.state}
                        viewerCanAdminister={batchChange.viewerCanAdminister}
                        history={history}
                        location={location}
                        isLightTheme={isLightTheme}
                        extensionsController={extensionsController}
                        platformContext={platformContext}
                        telemetryService={telemetryService}
                        queryExternalChangesetWithFileDiffs={queryExternalChangesetWithFileDiffs}
                        onlyArchived={true}
                        refetchBatchChange={refetchBatchChange}
                        settingsCascade={settingsCascade}
                        isExecutionEnabled={isExecutionEnabled}
                    />
                </BatchChangeTabPanel>
                <BatchChangeTabPanel>
                    <BulkOperationsTab batchChangeID={batchChange.id} />
                </BatchChangeTabPanel>
            </BatchChangeTabPanels>
        </BatchChangeTabs>
    )
}
