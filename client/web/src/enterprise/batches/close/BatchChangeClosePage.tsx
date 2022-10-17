import React, { useState, useMemo, useCallback } from 'react'

import { subDays } from 'date-fns'
import * as H from 'history'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'

import { ErrorLike, isErrorLike } from '@sourcegraph/common'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { PageHeader, LoadingSpinner, useObservable } from '@sourcegraph/wildcard'

import { BatchChangesIcon } from '../../../batches/icons'
import { HeroPage } from '../../../components/HeroPage'
import { PageTitle } from '../../../components/PageTitle'
import { BatchChangeChangesetsResult, BatchChangeFields, Scalars } from '../../../graphql-operations'
import {
    queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs,
    queryChangesets as _queryChangesets,
    fetchBatchChangeByNamespace as _fetchBatchChangeByNamespace,
} from '../detail/backend'
import { BatchChangeInfoByline } from '../detail/BatchChangeInfoByline'

import { closeBatchChange as _closeBatchChange } from './backend'
import { BatchChangeCloseAlert } from './BatchChangeCloseAlert'
import { BatchChangeCloseChangesetsList } from './BatchChangeCloseChangesetsList'

export interface BatchChangeClosePageProps
    extends ThemeProps,
        TelemetryProps,
        PlatformContextProps,
        ExtensionsControllerProps,
        SettingsCascadeProps {
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
    closeBatchChange?: typeof _closeBatchChange
}

export const BatchChangeClosePage: React.FunctionComponent<React.PropsWithChildren<BatchChangeClosePageProps>> = ({
    namespaceID,
    batchChangeName,
    history,
    location,
    extensionsController,
    isLightTheme,
    platformContext,
    telemetryService,
    fetchBatchChangeByNamespace = _fetchBatchChangeByNamespace,
    queryChangesets,
    queryExternalChangesetWithFileDiffs,
    closeBatchChange,
    settingsCascade,
}) => {
    const [closeChangesets, setCloseChangesets] = useState<boolean>(false)
    const createdAfter = useMemo(() => subDays(new Date(), 3).toISOString(), [])
    const batchChange = useObservable(
        useMemo(() => fetchBatchChangeByNamespace(namespaceID, batchChangeName, createdAfter), [
            fetchBatchChangeByNamespace,
            namespaceID,
            batchChangeName,
            createdAfter,
        ])
    )

    const [totalCount, setTotalCount] = useState<number>()

    const onFetchChangesets = useCallback(
        (
            connection?: (BatchChangeChangesetsResult['node'] & { __typename: 'BatchChange' })['changesets'] | ErrorLike
        ) => {
            if (!connection || isErrorLike(connection)) {
                return
            }
            setTotalCount(connection.totalCount)
        },
        []
    )

    // Is loading.
    if (batchChange === undefined) {
        return (
            <div className="text-center">
                <LoadingSpinner className="mx-auto my-4" />
            </div>
        )
    }

    // Batch change not found.
    if (batchChange === null) {
        return <HeroPage icon={AlertCircleIcon} title="Batch change not found" />
    }

    return (
        <>
            <PageTitle title="Preview close" />
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
                        creator={batchChange.creator}
                        lastAppliedAt={batchChange.lastAppliedAt}
                        lastApplier={batchChange.lastApplier}
                    />
                }
                className="test-batch-change-close-page mb-3"
            />
            {totalCount !== undefined && (
                <BatchChangeCloseAlert
                    batchChangeID={batchChange.id}
                    batchChangeURL={batchChange.url}
                    closeChangesets={closeChangesets}
                    setCloseChangesets={setCloseChangesets}
                    history={history}
                    closeBatchChange={closeBatchChange}
                    viewerCanAdminister={batchChange.viewerCanAdminister}
                    totalCount={totalCount}
                />
            )}
            <BatchChangeCloseChangesetsList
                batchChangeID={batchChange.id}
                history={history}
                location={location}
                viewerCanAdminister={batchChange.viewerCanAdminister}
                extensionsController={extensionsController}
                isLightTheme={isLightTheme}
                platformContext={platformContext}
                telemetryService={telemetryService}
                queryChangesets={queryChangesets}
                queryExternalChangesetWithFileDiffs={queryExternalChangesetWithFileDiffs}
                willClose={closeChangesets}
                onUpdate={onFetchChangesets}
                settingsCascade={settingsCascade}
            />
        </>
    )
}
