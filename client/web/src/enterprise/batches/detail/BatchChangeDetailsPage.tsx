import * as H from 'history'
import { isEqual } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React, { useEffect, useMemo } from 'react'
import { delay, distinctUntilChanged, repeatWhen } from 'rxjs/operators'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { BatchChangesIcon } from '../../../batches/icons'
import { HeroPage } from '../../../components/HeroPage'
import { PageHeader } from '../../../components/PageHeader'
import { PageTitle } from '../../../components/PageTitle'
import { BatchChangeFields, Scalars } from '../../../graphql-operations'
import { Description } from '../Description'

import {
    fetchBatchChangeByNamespace as _fetchBatchChangeByNamespace,
    queryChangesets as _queryChangesets,
    queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs,
    queryChangesetCountsOverTime as _queryChangesetCountsOverTime,
    deleteBatchChange as _deleteBatchChange,
} from './backend'
import { BatchChangeDetailsActionSection } from './BatchChangeDetailsActionSection'
import { BatchChangeInfoByline } from './BatchChangeInfoByline'
import { BatchChangeStatsCard } from './BatchChangeStatsCard'
import { BatchChangeTabs } from './BatchChangeTabs'
import { ChangesetsArchivedNotice } from './ChangesetsArchivedNotice'
import { ClosedNotice } from './ClosedNotice'
import { SupersedingBatchSpecAlert } from './SupersedingBatchSpecAlert'
import { UnpublishedNotice } from './UnpublishedNotice'

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
}) => {
    useEffect(() => {
        telemetryService.logViewEvent('BatchChangeDetailsPage')
    }, [telemetryService])

    const batchChange: BatchChangeFields | null | undefined = useObservable(
        useMemo(
            () =>
                fetchBatchChangeByNamespace(namespaceID, batchChangeName).pipe(
                    repeatWhen(notifier => notifier.pipe(delay(5000))),
                    distinctUntilChanged((a, b) => isEqual(a, b))
                ),
            [namespaceID, batchChangeName, fetchBatchChangeByNamespace]
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
            <Description history={history} description={batchChange.description} />
            <BatchChangeTabs
                batchChange={batchChange}
                changesetsCount={batchChange.changesetsStats.total - batchChange.changesetsStats.archived}
                archivedCount={batchChange.changesetsStats.archived}
                extensionsController={extensionsController}
                history={history}
                isLightTheme={isLightTheme}
                location={location}
                platformContext={platformContext}
                telemetryService={telemetryService}
                queryChangesets={queryChangesets}
                queryChangesetCountsOverTime={queryChangesetCountsOverTime}
                queryExternalChangesetWithFileDiffs={queryExternalChangesetWithFileDiffs}
            />
        </>
    )
}
