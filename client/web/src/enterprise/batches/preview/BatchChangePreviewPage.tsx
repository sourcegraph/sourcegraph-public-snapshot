import React, { useEffect } from 'react'

import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import { useHistory, useLocation } from 'react-router'

import { useQuery } from '@sourcegraph/http-client'
import { PageHeader, LoadingSpinner } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../auth'
import { BatchChangesIcon } from '../../../batches/icons'
import { HeroPage } from '../../../components/HeroPage'
import { PageTitle } from '../../../components/PageTitle'
import { BatchSpecByIDResult, BatchSpecByIDVariables } from '../../../graphql-operations'
import { Description } from '../Description'
import { SupersedingBatchSpecAlert } from '../detail/SupersedingBatchSpecAlert'
import { MultiSelectContextProvider } from '../MultiSelectContext'

import { BATCH_SPEC_BY_ID, queryApplyPreviewStats as _queryApplyPreviewStats } from './backend'
import { BatchChangePreviewContextProvider } from './BatchChangePreviewContext'
import { BatchChangePreviewStatsBar } from './BatchChangePreviewStatsBar'
import { BatchChangePreviewProps, BatchChangePreviewTabs } from './BatchChangePreviewTabs'
import { BatchSpecInfoByline } from './BatchSpecInfoByline'
import { CreateUpdateBatchChangeAlert } from './CreateUpdateBatchChangeAlert'
import { PreviewList } from './list/PreviewList'
import { MissingCredentialsAlert } from './MissingCredentialsAlert'

export type PreviewPageAuthenticatedUser = Pick<AuthenticatedUser, 'url' | 'displayName' | 'username' | 'email'>

export interface BatchChangePreviewPageProps extends BatchChangePreviewProps {
    /** Used for testing. */
    queryApplyPreviewStats?: typeof _queryApplyPreviewStats
}

export const BatchChangePreviewPage: React.FunctionComponent<
    React.PropsWithChildren<BatchChangePreviewPageProps>
> = props => {
    const history = useHistory()

    const { batchSpecID: specID, authenticatedUser, telemetryService, queryApplyPreviewStats } = props

    const { data, loading } = useQuery<BatchSpecByIDResult, BatchSpecByIDVariables>(BATCH_SPEC_BY_ID, {
        variables: {
            batchSpec: specID,
        },
        fetchPolicy: 'cache-and-network',
        pollInterval: 5000,
    })

    useEffect(() => {
        telemetryService.logViewEvent('BatchChangeApplyPage')
    }, [telemetryService])

    if (loading) {
        return (
            <div className="text-center">
                <LoadingSpinner className="mx-auto my-4" />
            </div>
        )
    }
    if (data?.node?.__typename !== 'BatchSpec') {
        return <HeroPage icon={AlertCircleIcon} title="Batch spec not found" />
    }
    const spec = data.node

    return (
        <MultiSelectContextProvider>
            <BatchChangePreviewContextProvider>
                <div className="pb-5">
                    <PageTitle title="Apply batch spec" />
                    <PageHeader
                        path={[
                            {
                                icon: BatchChangesIcon,
                                to: '/batch-changes',
                                ariaLabel: 'Batch changes',
                            },
                            { to: `${spec.namespace.url}/batch-changes`, text: spec.namespace.namespaceName },
                            { text: spec.description.name },
                        ]}
                        byline={<BatchSpecInfoByline createdAt={spec.createdAt} creator={spec.creator} />}
                        headingElement="h2"
                        className="test-batch-change-apply-page mb-3"
                    />
                    <MissingCredentialsAlert
                        authenticatedUser={authenticatedUser}
                        viewerBatchChangesCodeHosts={spec.viewerBatchChangesCodeHosts}
                    />
                    <SupersedingBatchSpecAlert spec={spec.supersedingBatchSpec} />
                    <BatchChangePreviewStatsBar
                        batchSpec={spec.id}
                        diffStat={spec.diffStat!}
                        queryApplyPreviewStats={queryApplyPreviewStats}
                    />
                    <CreateUpdateBatchChangeAlert
                        history={history}
                        specID={spec.id}
                        toBeArchived={spec.applyPreview.stats.archive}
                        batchChange={spec.appliesToBatchChange}
                        viewerCanAdminister={spec.viewerCanAdminister}
                        telemetryService={telemetryService}
                    />
                    <Description description={spec.description.description} />
                    <BatchChangePreviewTabs spec={spec} {...props} />
                </div>
            </BatchChangePreviewContextProvider>
        </MultiSelectContextProvider>
    )
}

/**
 * This is the "new" preview page, as used in SSBC. It will eventually replace the
 * current one, but until we are ready to flip the feature flag, we need to keep
 * both around.
 */
export const NewBatchChangePreviewPage: React.FunctionComponent<
    React.PropsWithChildren<BatchChangePreviewPageProps>
> = props => {
    const history = useHistory()
    const location = useLocation()

    const {
        batchSpecID: specID,
        isLightTheme,
        expandChangesetDescriptions,
        queryChangesetApplyPreview,
        queryChangesetSpecFileDiffs,
        authenticatedUser,
        telemetryService,
        queryApplyPreviewStats,
    } = props

    const { data, loading, error } = useQuery<BatchSpecByIDResult, BatchSpecByIDVariables>(BATCH_SPEC_BY_ID, {
        variables: {
            batchSpec: specID,
        },
        fetchPolicy: 'cache-and-network',
        pollInterval: 5000,
    })

    useEffect(() => {
        telemetryService.logViewEvent('BatchChangeApplyPage')
    }, [telemetryService])

    // If we're loading and haven't received any data yet
    if (loading && !data) {
        return (
            <div className="text-center">
                <LoadingSpinner className="mx-auto my-4" />
            </div>
        )
    }
    // If we received an error before we successfully received any data
    if (error && !data) {
        throw new Error(error.message)
    }
    // If there weren't any errors and we just didn't receive any data
    if (data?.node?.__typename !== 'BatchSpec') {
        return <HeroPage icon={AlertCircleIcon} title="Batch spec not found" />
    }

    const spec = data.node

    return (
        <MultiSelectContextProvider>
            <BatchChangePreviewContextProvider>
                <div className="pb-5">
                    <MissingCredentialsAlert
                        authenticatedUser={authenticatedUser}
                        viewerBatchChangesCodeHosts={spec.viewerBatchChangesCodeHosts}
                    />
                    <BatchChangePreviewStatsBar
                        batchSpec={spec.id}
                        diffStat={spec.diffStat!}
                        queryApplyPreviewStats={queryApplyPreviewStats}
                    />
                    <CreateUpdateBatchChangeAlert
                        history={history}
                        specID={spec.id}
                        toBeArchived={spec.applyPreview.stats.archive}
                        batchChange={spec.appliesToBatchChange}
                        viewerCanAdminister={spec.viewerCanAdminister}
                        telemetryService={telemetryService}
                    />
                    <PreviewList
                        batchSpecID={specID}
                        history={history}
                        location={location}
                        authenticatedUser={authenticatedUser}
                        isLightTheme={isLightTheme}
                        queryChangesetApplyPreview={queryChangesetApplyPreview}
                        queryChangesetSpecFileDiffs={queryChangesetSpecFileDiffs}
                        expandChangesetDescriptions={expandChangesetDescriptions}
                    />
                </div>
            </BatchChangePreviewContextProvider>
        </MultiSelectContextProvider>
    )
}
