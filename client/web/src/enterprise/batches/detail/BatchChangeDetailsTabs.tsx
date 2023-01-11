import React, { useCallback, useMemo, useState } from 'react'

import { mdiSourceBranch, mdiChartLineVariant, mdiFileDocument, mdiArchive, mdiMonitorStar } from '@mdi/js'
import * as H from 'history'
import { useHistory, useLocation } from 'react-router'

import { Settings, SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Badge, Container, Icon, Tab, TabPanel, TabPanels } from '@sourcegraph/wildcard'

import { isBatchChangesExecutionEnabled } from '../../../batches'
import { resetFilteredConnectionURLQuery } from '../../../components/FilteredConnection'
import { BatchSpecState, BatchChangeFields, BatchSpecSource } from '../../../graphql-operations'
import { BatchChangeTabList, BatchChangeTabs } from '../BatchChangeTabs'
import { BatchSpecDownloadButton, BatchSpecMeta } from '../BatchSpec'
import { BatchSpecInfo } from '../BatchSpecNode'
import { BatchChangeBatchSpecList } from '../BatchSpecsPage'

import {
    queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs,
    queryAllChangesetIDs as _queryAllChangesetIDs,
} from './backend'
import { BatchChangeBurndownChart } from './BatchChangeBurndownChart'
import { BulkOperationsTab } from './BulkOperationsTab'
import { BatchChangeChangesets } from './changesets/BatchChangeChangesets'

export enum TabName {
    Changesets = 'changesets',
    Chart = 'chart',
    // Non-SSBC or SSBC with viewerCanAdminister=false
    Spec = 'spec',
    // SSBC-only
    Executions = 'executions',
    Archived = 'archived',
    BulkOperations = 'bulkoperations',
}

const getTabIndex = (tabName: string, shouldDisplayExecutionsTab: boolean): number =>
    (
        [
            TabName.Changesets,
            TabName.Chart,
            shouldDisplayExecutionsTab ? TabName.Executions : TabName.Spec,
            TabName.Archived,
            TabName.BulkOperations,
        ] as string[]
    ).indexOf(tabName)

const getTabName = (tabIndex: number, shouldDisplayExecutionsTab: boolean): TabName =>
    [
        TabName.Changesets,
        TabName.Chart,
        shouldDisplayExecutionsTab ? TabName.Executions : TabName.Spec,
        TabName.Archived,
        TabName.BulkOperations,
    ][tabIndex]

/** `BatchChangeDetailsPage` and `BatchChangeDetailsTabs` share all these props */
export interface BatchChangeDetailsProps extends ThemeProps, TelemetryProps {
    history: H.History
    location: H.Location
    /** The name of the tab that should be initially open */
    initialTab?: TabName

    /** For testing only. */
    queryExternalChangesetWithFileDiffs?: typeof _queryExternalChangesetWithFileDiffs
    /** For testing only. */
    queryAllChangesetIDs?: typeof _queryAllChangesetIDs
}

interface BatchChangeDetailsTabsProps extends BatchChangeDetailsProps, SettingsCascadeProps<Settings> {
    batchChange: BatchChangeFields
    refetchBatchChange: () => void
}

export const BatchChangeDetailsTabs: React.FunctionComponent<React.PropsWithChildren<BatchChangeDetailsTabsProps>> = ({
    batchChange,
    isLightTheme,
    settingsCascade,
    initialTab = TabName.Changesets,
    queryExternalChangesetWithFileDiffs,
    queryAllChangesetIDs,
    refetchBatchChange,
}) => {
    const isExecutionEnabled = isBatchChangesExecutionEnabled(settingsCascade)

    const pendingExecutionsCount = useMemo(
        () =>
            batchChange.batchSpecs.nodes.filter(
                node => node.state === BatchSpecState.PROCESSING || node.state === BatchSpecState.QUEUED
            ).length,
        [batchChange.batchSpecs.nodes]
    )

    const isBatchSpecLocallyCreated = batchChange.currentSpec.source === BatchSpecSource.LOCAL
    const shouldDisplayExecutionsTab =
        isExecutionEnabled && !isBatchSpecLocallyCreated && batchChange.viewerCanAdminister

    // We track the current tab in a URL parameter so that tabs are easy to navigate to
    // and share.
    const history = useHistory()
    const location = useLocation()
    const initialURLTab = new URLSearchParams(location.search).get('tab')
    const defaultTabIndex = getTabIndex(initialURLTab || initialTab, shouldDisplayExecutionsTab) || 0

    // The executions tab uses an additional custom short URL, "/executions".
    const [customShortPath, setCustomShortPath] = useState(
        initialTab === TabName.Executions ? '/executions' : undefined
    )

    const onTabChange = useCallback(
        (index: number) => {
            const urlParameters = new URLSearchParams(location.search)
            resetFilteredConnectionURLQuery(urlParameters)

            const newTabName = getTabName(index, shouldDisplayExecutionsTab)

            // The executions tab uses a custom short URL.
            if (newTabName === TabName.Executions) {
                if (location.pathname.includes('/executions')) {
                    return
                }
                // Remember our current custom short path, so that it's easy to remove
                // when we navigate to a different tab.
                setCustomShortPath('/executions')
                history.replace(location.pathname + '/executions')
            } else {
                // The first tab is the default, so it's not necessary to set it in the URL.
                if (index === 0) {
                    urlParameters.delete('tab')
                } else {
                    urlParameters.set('tab', getTabName(index, shouldDisplayExecutionsTab))
                }
                // Make sure to unset the custom short path, if we were previously on a
                // tab that had one.
                const newLocation = customShortPath
                    ? { ...location, pathname: location.pathname.replace(customShortPath, '') }
                    : location
                setCustomShortPath(undefined)

                history.replace({ ...newLocation, search: urlParameters.toString() })
            }
        },
        [history, location, shouldDisplayExecutionsTab, customShortPath]
    )

    const changesetCount = batchChange.changesetsStats.total - batchChange.changesetsStats.archived
    const executionsCount = `${pendingExecutionsCount}${batchChange.batchSpecs.pageInfo.hasNextPage ? '+' : ''}`

    return (
        <BatchChangeTabs defaultIndex={defaultTabIndex} onChange={onTabChange}>
            <BatchChangeTabList>
                <Tab aria-label={`Changesets (${changesetCount} total)`}>
                    <span>
                        <Icon aria-hidden={true} className="text-muted mr-2" svgPath={mdiSourceBranch} />
                        <span className="text-content" data-tab-content="Changesets">
                            Changesets
                        </span>
                        <Badge variant="secondary" pill={true} className="ml-2">
                            {changesetCount}
                        </Badge>
                    </span>
                </Tab>
                <Tab>
                    <span>
                        <Icon aria-hidden={true} className="text-muted mr-2" svgPath={mdiChartLineVariant} />
                        <span className="text-content" data-tab-content="Burndown chart">
                            Burndown chart
                        </span>
                    </span>
                </Tab>
                {shouldDisplayExecutionsTab ? (
                    <Tab
                        aria-label={`Executions${
                            pendingExecutionsCount > 0 ? ' (' + executionsCount + 'total active)' : ''
                        }`}
                    >
                        <span>
                            <Icon aria-hidden={true} className="text-muted mr-2" svgPath={mdiFileDocument} />
                            <span className="text-content" data-tab-content="Executions">
                                Executions
                            </span>
                            {pendingExecutionsCount > 0 && (
                                <Badge variant="warning" pill={true} className="ml-2">
                                    {executionsCount}
                                </Badge>
                            )}
                        </span>
                    </Tab>
                ) : (
                    <Tab>
                        <span>
                            <Icon aria-hidden={true} className="text-muted mr-2" svgPath={mdiFileDocument} />
                            <span className="text-content" data-tab-content="Spec">
                                Spec
                            </span>
                        </span>
                    </Tab>
                )}
                <Tab aria-label={`Archived (${batchChange.changesetsStats.archived} total)`}>
                    <span>
                        <Icon aria-hidden={true} className="text-muted mr-2" svgPath={mdiArchive} />
                        <span className="text-content" data-tab-content="Archived">
                            Archived
                        </span>
                        <Badge variant="secondary" pill={true} className="ml-2">
                            {batchChange.changesetsStats.archived}
                        </Badge>
                    </span>
                </Tab>
                <Tab aria-label={`Bulk operations (${batchChange.bulkOperations.totalCount} total)`}>
                    <span>
                        <Icon aria-hidden={true} className="text-muted mr-2" svgPath={mdiMonitorStar} />
                        <span className="text-content" data-tab-content="Bulk operations">
                            Bulk operations
                        </span>
                        <Badge variant="secondary" pill={true} className="ml-2">
                            {batchChange.bulkOperations.totalCount}
                        </Badge>
                    </span>
                </Tab>
            </BatchChangeTabList>
            <TabPanels>
                <TabPanel>
                    <BatchChangeChangesets
                        batchChangeID={batchChange.id}
                        batchChangeState={batchChange.state}
                        viewerCanAdminister={batchChange.viewerCanAdminister}
                        refetchBatchChange={refetchBatchChange}
                        history={history}
                        location={location}
                        queryExternalChangesetWithFileDiffs={queryExternalChangesetWithFileDiffs}
                        queryAllChangesetIDs={queryAllChangesetIDs}
                        onlyArchived={false}
                        isExecutionEnabled={isExecutionEnabled}
                    />
                </TabPanel>
                <TabPanel>
                    <BatchChangeBurndownChart batchChangeID={batchChange.id} history={history} />
                </TabPanel>
                <TabPanel>
                    {shouldDisplayExecutionsTab ? (
                        <Container>
                            <BatchChangeBatchSpecList
                                history={history}
                                location={location}
                                batchChangeID={batchChange.id}
                                currentSpecID={batchChange.currentSpec.id}
                                isLightTheme={isLightTheme}
                            />
                        </Container>
                    ) : (
                        <>
                            <div className="d-flex flex-wrap justify-content-between align-items-baseline mb-2 test-batches-spec">
                                <BatchSpecMeta
                                    createdAt={batchChange.createdAt}
                                    lastApplier={batchChange.lastApplier}
                                    lastAppliedAt={batchChange.lastAppliedAt}
                                />
                                <BatchSpecDownloadButton
                                    name={batchChange.name}
                                    originalInput={batchChange.currentSpec.originalInput}
                                />
                            </div>
                            <BatchSpecInfo spec={batchChange.currentSpec} isLightTheme={isLightTheme} />
                        </>
                    )}
                </TabPanel>
                <TabPanel>
                    <BatchChangeChangesets
                        batchChangeID={batchChange.id}
                        batchChangeState={batchChange.state}
                        viewerCanAdminister={batchChange.viewerCanAdminister}
                        history={history}
                        location={location}
                        queryExternalChangesetWithFileDiffs={queryExternalChangesetWithFileDiffs}
                        onlyArchived={true}
                        refetchBatchChange={refetchBatchChange}
                        isExecutionEnabled={isExecutionEnabled}
                    />
                </TabPanel>
                <TabPanel>
                    <BulkOperationsTab batchChangeID={batchChange.id} />
                </TabPanel>
            </TabPanels>
        </BatchChangeTabs>
    )
}
