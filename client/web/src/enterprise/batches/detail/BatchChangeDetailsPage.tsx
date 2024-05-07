import React, { useEffect, useMemo } from 'react'

import { subDays, startOfDay } from 'date-fns'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import { useParams } from 'react-router-dom'

import { useQuery } from '@sourcegraph/http-client'
import type { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import type { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import type { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import type { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { PageHeader, LoadingSpinner, Alert, ErrorMessage } from '@sourcegraph/wildcard'

import { BatchChangesIcon } from '../../../batches/icons'
import { CreatedByAndUpdatedByInfoByline } from '../../../components/Byline/CreatedByAndUpdatedByInfoByline'
import { HeroPage } from '../../../components/HeroPage'
import { PageTitle } from '../../../components/PageTitle'
import type { BatchChangeByNamespaceResult, BatchChangeByNamespaceVariables } from '../../../graphql-operations'
import { Description } from '../Description'
import { MissingCredentialsAlert } from '../MissingCredentialsAlert'

import { ActiveExecutionNotice } from './ActiveExecutionNotice'
import { type deleteBatchChange as _deleteBatchChange, BATCH_CHANGE_BY_NAMESPACE } from './backend'
import { BatchChangeDetailsActionSection } from './BatchChangeDetailsActionSection'
import { type BatchChangeDetailsProps, BatchChangeDetailsTabs, type TabName } from './BatchChangeDetailsTabs'
import { BatchChangeStatsCard } from './BatchChangeStatsCard'
import { BulkOperationsAlerts } from './BulkOperationsAlerts'
import { ChangesetsArchivedNotice } from './ChangesetsArchivedNotice'
import { ClosedNotice } from './ClosedNotice'
import { SupersedingBatchSpecAlert } from './SupersedingBatchSpecAlert'
import { UnpublishedNotice } from './UnpublishedNotice'
import { WebhookAlert } from './WebhookAlert'

export interface BatchChangeDetailsPageProps extends BatchChangeDetailsProps, SettingsCascadeProps<Settings> {
    /** The namespace ID. */
    namespaceID: Scalars['ID']
    /** The name of the tab that should be initially open */
    initialTab?: TabName
    /** For testing only. */
    deleteBatchChange?: typeof _deleteBatchChange

    authenticatedUser: Pick<AuthenticatedUser, 'url'>
}

/**
 * The area for a single batch change.
 */
export const BatchChangeDetailsPage: React.FunctionComponent<
    React.PropsWithChildren<BatchChangeDetailsPageProps>
> = props => {
    const { batchChangeName } = useParams()
    const { namespaceID, telemetryService, telemetryRecorder, authenticatedUser, deleteBatchChange } = props

    useEffect(() => {
        telemetryService.logViewEvent('BatchChangeDetailsPage')
        telemetryRecorder.recordEvent('batchChange.details', 'view')
    }, [telemetryService, telemetryRecorder])

    // Query bulk operations created after this time.
    const createdAfter = useMemo(() => subDays(startOfDay(new Date()), 3).toISOString(), [])

    const { data, error, loading, refetch } = useQuery<BatchChangeByNamespaceResult, BatchChangeByNamespaceVariables>(
        BATCH_CHANGE_BY_NAMESPACE,
        {
            variables: { namespaceID, batchChange: batchChangeName!, createdAfter },
            // Cache this data but always re-request it in the background when we revisit
            // this page to pick up newer changes.
            fetchPolicy: 'cache-and-network',
            // We continuously poll for changes to the batch change, in case the bulk
            // operations, diff stats, or changeset stats are updated, or in case someone
            // applied a new batch spec in the meantime. This isn't the most effective use
            // of network bandwidth since many of these fields aren't changing and most of
            // the time there will be no changes at all, but it's also the easiest way to
            // keep this in sync for now at the cost of a bit of excess network resources.
            pollInterval: 5000,
            // For subsequent requests while this page is open, make additional network
            // requests; this is necessary for `refetch` to actually use the network. (see
            // https://github.com/apollographql/apollo-client/issues/5515)
            nextFetchPolicy: 'cache-and-network',
        }
    )

    // If we're loading and haven't received any data yet
    if (loading && !data) {
        return (
            <div className="text-center">
                <LoadingSpinner className="mx-auto my-4" />
            </div>
        )
    }
    // If we received an error before we had received any data
    if (error && !data) {
        throw new Error(error.message)
    }
    // If there weren't any errors and we just didn't receive any data
    if (!data?.batchChange) {
        return <HeroPage icon={AlertCircleIcon} title="Batch change not found" />
    }

    const { batchChange } = data

    return (
        <>
            <PageTitle title={batchChange.name} />
            {/* If we received an error after we already had data, we keep the
                data on the page but also surface the error with an alert. */}
            {error && (
                <Alert variant="danger">
                    <ErrorMessage error={error.message} />
                </Alert>
            )}
            <PageHeader
                byline={
                    <CreatedByAndUpdatedByInfoByline
                        createdAt={batchChange.createdAt}
                        createdBy={batchChange.creator}
                        updatedAt={batchChange.lastAppliedAt}
                        updatedBy={batchChange.lastApplier}
                    />
                }
                actions={
                    batchChange.viewerCanAdminister ? (
                        <BatchChangeDetailsActionSection
                            batchChangeID={batchChange.id}
                            batchChangeClosed={!!batchChange.closedAt}
                            deleteBatchChange={deleteBatchChange}
                            batchChangeNamespaceURL={batchChange.namespace.url}
                            batchChangeURL={batchChange.url}
                            settingsCascade={props.settingsCascade}
                            telemetryRecorder={telemetryRecorder}
                        />
                    ) : null
                }
                className="test-batch-change-details-page mb-3"
            >
                <PageHeader.Heading as="h2" styleAs="h1">
                    <PageHeader.Breadcrumb icon={BatchChangesIcon} to="/batch-changes" aria-label="Batch Changes" />
                    <PageHeader.Breadcrumb to={`${batchChange.namespace.url}/batch-changes`}>
                        {batchChange.namespace.namespaceName}
                    </PageHeader.Breadcrumb>
                    <PageHeader.Breadcrumb>{batchChange.name}</PageHeader.Breadcrumb>
                </PageHeader.Heading>
            </PageHeader>
            <BulkOperationsAlerts bulkOperations={batchChange.activeBulkOperations} />
            {batchChange.viewerCanAdminister && (
                <MissingCredentialsAlert
                    authenticatedUser={authenticatedUser}
                    viewerBatchChangesCodeHosts={batchChange.currentSpec.viewerBatchChangesCodeHosts}
                />
            )}
            {batchChange.viewerCanAdminister && (
                <SupersedingBatchSpecAlert spec={batchChange.currentSpec.supersedingBatchSpec} />
            )}
            <ActiveExecutionNotice
                batchSpecs={batchChange.batchSpecs.nodes}
                batchChangeURL={batchChange.url}
                className="mb-3"
            />
            <ClosedNotice closedAt={batchChange.closedAt} className="mb-3" />
            {batchChange.closedAt === null && batchChange.viewerCanAdminister && (
                <UnpublishedNotice
                    unpublished={batchChange.changesetsStats.unpublished}
                    total={batchChange.changesetsStats.total}
                    className="mb-3"
                />
            )}
            <ChangesetsArchivedNotice />
            <WebhookAlert batchChange={batchChange} />
            <BatchChangeStatsCard batchChange={batchChange} className="mb-3" />
            <Description description={batchChange.description} />
            <BatchChangeDetailsTabs batchChange={batchChange} refetchBatchChange={refetch} {...props} />
        </>
    )
}
