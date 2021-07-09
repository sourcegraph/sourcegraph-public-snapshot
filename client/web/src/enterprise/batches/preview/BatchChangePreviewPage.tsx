import { isEqual } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { delay, distinctUntilChanged, repeatWhen } from 'rxjs/operators'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { PageHeader } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../auth'
import { BatchChangesIcon } from '../../../batches/icons'
import { HeroPage } from '../../../components/HeroPage'
import { PageTitle } from '../../../components/PageTitle'
import { Description } from '../Description'
import { SupersedingBatchSpecAlert } from '../detail/SupersedingBatchSpecAlert'
import { MultiSelectContext, MultiSelectContextSelected } from '../MultiSelectContext'

import {
    applyBatchChange,
    createBatchChange,
    fetchBatchSpecById as _fetchBatchSpecById,
    queryAllChangesetSpecIDs,
} from './backend'
import { BatchChangePreviewStatsBar } from './BatchChangePreviewStatsBar'
import { BatchChangePreviewProps, BatchChangePreviewTabs } from './BatchChangePreviewTabs'
import { BatchSpecInfoByline } from './BatchSpecInfoByline'
import { CreateUpdateBatchChangeAlert, CreateUpdateBatchChangeAlertAction } from './CreateUpdateBatchChangeAlert'
import { MissingCredentialsAlert } from './MissingCredentialsAlert'

export type PreviewPageAuthenticatedUser = Pick<AuthenticatedUser, 'url' | 'displayName' | 'username' | 'email'>

export interface BatchChangePreviewPageProps extends BatchChangePreviewProps {
    /** Used for testing. */
    fetchBatchSpecById?: typeof _fetchBatchSpecById
}

export const BatchChangePreviewPage: React.FunctionComponent<BatchChangePreviewPageProps> = props => {
    const {
        batchSpecID: specID,
        history,
        authenticatedUser,
        telemetryService,
        fetchBatchSpecById = _fetchBatchSpecById,
    } = props

    const spec = useObservable(
        useMemo(
            () =>
                fetchBatchSpecById(specID).pipe(
                    repeatWhen(notifier => notifier.pipe(delay(5000))),
                    distinctUntilChanged((a, b) => isEqual(a, b))
                ),
            [specID, fetchBatchSpecById]
        )
    )

    const [selectedChangesetSpecs, setSelectedChangesetSpecs] = useState<MultiSelectContextSelected>(new Set())
    const onSelectAllChangesetSpecs = useCallback(() => setSelectedChangesetSpecs('all'), [setSelectedChangesetSpecs])
    const onDeselectAllChangesetSpecs = useCallback(() => setSelectedChangesetSpecs(new Set()), [
        setSelectedChangesetSpecs,
    ])
    const onSelectChangesetSpec = useCallback(
        (id: string) => {
            const updated = new Set(selectedChangesetSpecs)
            updated.add(id)

            setSelectedChangesetSpecs(updated)
        },
        [selectedChangesetSpecs, setSelectedChangesetSpecs]
    )
    const onDeselectChangesetSpec = useCallback(
        (id: string) => {
            const updated = new Set(selectedChangesetSpecs)
            updated.delete(id)

            setSelectedChangesetSpecs(updated)
        },
        [selectedChangesetSpecs, setSelectedChangesetSpecs]
    )

    useEffect(() => {
        telemetryService.logViewEvent('BatchChangeApplyPage')
    }, [telemetryService])

    const onApply = useCallback(
        async (action: CreateUpdateBatchChangeAlertAction, setIsLoading: (loadingOrError: boolean | Error) => void) => {
            if (!spec) {
                return
            }

            if (!confirm(`Are you sure you want to ${spec.id ? 'update' : 'create'} this batch change?`)) {
                return
            }
            setIsLoading(true)
            try {
                const batchChangeID = spec.appliesToBatchChange?.id
                const toBeArchived = spec.applyPreview.stats.archive

                // TODO: switch on action and use the current selections if necessary.
                const batchChange = batchChangeID
                    ? await applyBatchChange({ batchSpec: spec.id, batchChange: batchChangeID })
                    : await createBatchChange({ batchSpec: spec.id })

                if (toBeArchived > 0) {
                    history.push(`${batchChange.url}?archivedCount=${toBeArchived}&archivedBy=${spec.id}`)
                } else {
                    history.push(batchChange.url)
                }
                telemetryService.logViewEvent(`BatchChangeDetailsPageAfter${batchChangeID ? 'Create' : 'Update'}`)
            } catch (error) {
                setIsLoading(error)
            }
        },
        [spec, history, telemetryService]
    )

    if (spec === undefined) {
        return (
            <div className="text-center">
                <LoadingSpinner className="icon-inline mx-auto my-4" />
            </div>
        )
    }
    if (spec === null) {
        return <HeroPage icon={AlertCircleIcon} title="Batch spec not found" />
    }

    return (
        <MultiSelectContext.Provider
            value={{
                selected: selectedChangesetSpecs,
                onDeselectAll: onDeselectAllChangesetSpecs,
                onDeselect: onDeselectChangesetSpec,
                onSelectAll: onSelectAllChangesetSpecs,
                onSelect: onSelectChangesetSpec,
            }}
        >
            <div className="pb-5">
                <PageTitle title="Apply batch spec" />
                <PageHeader
                    path={[
                        {
                            icon: BatchChangesIcon,
                            to: '/batch-changes',
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
                <BatchChangePreviewStatsBar batchSpec={spec} />
                <CreateUpdateBatchChangeAlert
                    batchChange={spec.appliesToBatchChange}
                    showPublishUI={spec.applyPreview.stats.uiPublished > 0}
                    onApply={onApply}
                    viewerCanAdminister={spec.viewerCanAdminister}
                />
                <Description description={spec.description.description} />
                <BatchChangePreviewTabs spec={spec} {...props} />
            </div>
        </MultiSelectContext.Provider>
    )
}
