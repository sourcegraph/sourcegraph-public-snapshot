import { subDays } from 'date-fns'
import * as H from 'history'
import { isEqual, values } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import ArchiveIcon from 'mdi-react/ArchiveIcon'
import ChartLineVariantIcon from 'mdi-react/ChartLineVariantIcon'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'
import MonitorStarIcon from 'mdi-react/MonitorStarIcon'
import SourceBranchIcon from 'mdi-react/SourceBranchIcon'
import React, { useEffect, useMemo } from 'react'
import { delay, distinctUntilChanged, repeatWhen } from 'rxjs/operators'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { PageHeader } from '@sourcegraph/wildcard'

import { BatchChangesIcon } from '../../../batches/icons'
import { HeroPage } from '../../../components/HeroPage'
import { PageTitle } from '../../../components/PageTitle'
import { BatchChangeFields, Scalars } from '../../../graphql-operations'
import {
    BatchChangeTab,
    BatchChangeTabPanel,
    BatchChangeTabPanels,
    BatchChangeTabs,
    BatchChangeTabList,
} from '../BatchChangeTabs'
import { BatchSpec, BatchSpecDownloadLink, BatchSpecMeta } from '../BatchSpec'
import { Description } from '../Description'

import {
    fetchBatchChangeByNamespace as _fetchBatchChangeByNamespace,
    queryChangesets as _queryChangesets,
    queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs,
    queryChangesetCountsOverTime as _queryChangesetCountsOverTime,
    deleteBatchChange as _deleteBatchChange,
    queryBulkOperations as _queryBulkOperations,
} from './backend'
import { BatchChangeBurndownChart } from './BatchChangeBurndownChart'
import { BatchChangeDetailsActionSection } from './BatchChangeDetailsActionSection'
import { BatchChangeInfoByline } from './BatchChangeInfoByline'
import { BatchChangeStatsCard } from './BatchChangeStatsCard'
import { BulkOperationsAlerts } from './BulkOperationsAlerts'
import { BulkOperationsTab } from './BulkOperationsTab'
import { BatchChangeChangesets } from './changesets/BatchChangeChangesets'
import { ChangesetsArchivedNotice } from './ChangesetsArchivedNotice'
import { ClosedNotice } from './ClosedNotice'
import { SupersedingBatchSpecAlert } from './SupersedingBatchSpecAlert'
import { UnpublishedNotice } from './UnpublishedNotice'

export const TAB_NAMES = {
    changesets: 'changesets',
    chart: 'chart',
    spec: 'spec',
    archived: 'archived',
    bulkOperations: 'bulkoperations',
} as const

const TAB_NAME_VALUES = values(TAB_NAMES) as [string, string, ...string[]]

export interface BatchChangeDetailsPageProps
    extends ThemeProps,
        ExtensionsControllerProps,
        PlatformContextProps,
        TelemetryProps {
    /**
     * The namespace ID.
     */
    namespaceID: Scalars['ID']
    /**
     * The batch change name.
     */
    batchChangeName: BatchChangeFields['name']
    history: H.History
    location: H.Location

    /** For testing only. */
    fetchBatchChangeByNamespace?: typeof _fetchBatchChangeByNamespace
    /** For testing only. */
    queryChangesets?: typeof _queryChangesets
    /** For testing only. */
    queryExternalChangesetWithFileDiffs?: typeof _queryExternalChangesetWithFileDiffs
    /** For testing only. */
    queryChangesetCountsOverTime?: typeof _queryChangesetCountsOverTime
    /** For testing only. */
    deleteBatchChange?: typeof _deleteBatchChange
    /** For testing only. */
    queryBulkOperations?: typeof _queryBulkOperations
}

/**
 * The area for a single batch change.
 */
export const BatchChangeDetailsPage: React.FunctionComponent<BatchChangeDetailsPageProps> = ({
    namespaceID,
    batchChangeName,
    history,
    location,
    isLightTheme,
    extensionsController,
    platformContext,
    telemetryService,
    fetchBatchChangeByNamespace: fetchBatchChangeByNamespace = _fetchBatchChangeByNamespace,
    queryChangesets,
    queryExternalChangesetWithFileDiffs,
    queryChangesetCountsOverTime,
    deleteBatchChange,
    queryBulkOperations,
}) => {
    useEffect(() => {
        telemetryService.logViewEvent('BatchChangeDetailsPage')
    }, [telemetryService])

    const createdAfter = useMemo(() => subDays(new Date(), 3).toISOString(), [])
    const batchChange: BatchChangeFields | null | undefined = useObservable(
        useMemo(
            () =>
                fetchBatchChangeByNamespace(namespaceID, batchChangeName, createdAfter).pipe(
                    repeatWhen(notifier => notifier.pipe(delay(5000))),
                    distinctUntilChanged((a, b) => isEqual(a, b))
                ),
            [fetchBatchChangeByNamespace, namespaceID, batchChangeName, createdAfter]
        )
    )

    // Is loading.
    if (batchChange === undefined) {
        return (
            <div className="text-center">
                <LoadingSpinner className="icon-inline mx-auto my-4" />
            </div>
        )
    }
    // Batch change was not found.
    if (batchChange === null) {
        return <HeroPage icon={AlertCircleIcon} title="Batch change not found" />
    }

    return (
        <>
            <PageTitle title={batchChange.name} />
            <PageHeader
                path={[
                    {
                        icon: BatchChangesIcon,
                        to: '/batch-changes',
                    },
                    { to: `${batchChange.namespace.url}/batch-changes`, text: batchChange.namespace.namespaceName },
                    { text: batchChange.name },
                ]}
                byline={
                    <BatchChangeInfoByline
                        createdAt={batchChange.createdAt}
                        initialApplier={batchChange.initialApplier}
                        lastAppliedAt={batchChange.lastAppliedAt}
                        lastApplier={batchChange.lastApplier}
                    />
                }
                actions={
                    <BatchChangeDetailsActionSection
                        batchChangeID={batchChange.id}
                        batchChangeClosed={!!batchChange.closedAt}
                        deleteBatchChange={deleteBatchChange}
                        batchChangeNamespaceURL={batchChange.namespace.url}
                        history={history}
                    />
                }
                className="test-batch-change-details-page mb-3"
            />
            <BulkOperationsAlerts location={location} bulkOperations={batchChange.activeBulkOperations} />
            <SupersedingBatchSpecAlert spec={batchChange.currentSpec.supersedingBatchSpec} />
            <ClosedNotice closedAt={batchChange.closedAt} className="mb-3" />
            <UnpublishedNotice
                unpublished={batchChange.changesetsStats.unpublished}
                total={batchChange.changesetsStats.total}
                className="mb-3"
            />
            <ChangesetsArchivedNotice history={history} location={location} />
            <BatchChangeStatsCard
                closedAt={batchChange.closedAt}
                stats={batchChange.changesetsStats}
                diff={batchChange.diffStat}
                className="mb-3"
            />
            <Description description={batchChange.description} />
            <BatchChangeTabs history={history} location={location} tabNames={TAB_NAME_VALUES}>
                    <BatchChangeTab index={0}>
                <BatchChangeTabList>
                        <SourceBranchIcon className="icon-inline text-muted mr-1" />
                        Changesets{' '}
                        <span className="badge badge-pill badge-secondary ml-1">
                            {batchChange.changesetsStats.total - batchChange.changesetsStats.archived}
                        </span>
                    </BatchChangeTab>
                    <BatchChangeTab index={1}>
                        <ChartLineVariantIcon className="icon-inline text-muted mr-1" /> Burndown chart
                    </BatchChangeTab>
                    <BatchChangeTab index={2}>
                        <FileDocumentIcon className="icon-inline text-muted mr-1" /> Spec
                    </BatchChangeTab>
                    <BatchChangeTab index={3}>
                        <ArchiveIcon className="icon-inline text-muted mr-1" /> Archived{' '}
                        <span className="badge badge-pill badge-secondary ml-1">
                            {batchChange.changesetsStats.archived}
                        </span>
                    </BatchChangeTab>
                    <BatchChangeTab index={4}>
                        <MonitorStarIcon className="icon-inline text-muted mr-1" /> Bulk operations{' '}
                        <span className="badge badge-pill badge-secondary ml-1">
                            {batchChange.bulkOperations.totalCount}
                        </span>
                    </BatchChangeTab>
                </BatchChangeTabList>
                <BatchChangeTabPanels>
                    <BatchChangeTabPanel>
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
                    <BatchChangeTabPanel>
                        <BatchChangeBurndownChart
                            batchChangeID={batchChange.id}
                            queryChangesetCountsOverTime={queryChangesetCountsOverTime}
                            history={history}
                        />
                    </BatchChangeTabPanel>
                    <BatchChangeTabPanel>
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
                    <BatchChangeTabPanel>
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
                    <BatchChangeTabPanel>
                        <BulkOperationsTab
                            batchChangeID={batchChange.id}
                            history={history}
                            location={location}
                            queryBulkOperations={queryBulkOperations}
                        />
                    </BatchChangeTabPanel>
                </BatchChangeTabPanels>
            </BatchChangeTabs>
        </>
    )
}
