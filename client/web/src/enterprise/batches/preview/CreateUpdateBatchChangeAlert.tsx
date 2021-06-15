import classNames from 'classnames'
import * as H from 'history'
import React, { useCallback, useState } from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { useRedesignToggle } from '@sourcegraph/shared/src/util/useRedesignToggle'

import { ErrorAlert } from '../../../components/alerts'
import { BatchSpecFields } from '../../../graphql-operations'

import { createBatchChange, applyBatchChange } from './backend'
import styles from './CreateUpdateBatchChangeAlert.module.scss'

export interface CreateUpdateBatchChangeAlertProps extends TelemetryProps {
    specID: string
    toBeArchived: number
    batchChange: BatchSpecFields['appliesToBatchChange']
    viewerCanAdminister: boolean
    history: H.History
}

export const CreateUpdateBatchChangeAlert: React.FunctionComponent<CreateUpdateBatchChangeAlertProps> = ({
    specID,
    toBeArchived,
    batchChange,
    viewerCanAdminister,
    history,
    telemetryService,
}) => {
    const batchChangeID = batchChange?.id
    const [isLoading, setIsLoading] = useState<boolean | Error>(false)
    const [isRedesignEnabled] = useRedesignToggle()

    const onApply = useCallback(async () => {
        if (!confirm(`Are you sure you want to ${batchChangeID ? 'update' : 'create'} this batch change?`)) {
            return
        }
        setIsLoading(true)
        try {
            const batchChange = batchChangeID
                ? await applyBatchChange({ batchSpec: specID, batchChange: batchChangeID })
                : await createBatchChange({ batchSpec: specID })

            if (toBeArchived > 0) {
                history.push(`${batchChange.url}?archivedCount=${toBeArchived}&archivedBy=${specID}`)
            } else {
                history.push(batchChange.url)
            }
            telemetryService.logViewEvent(`BatchChangeDetailsPageAfter${batchChangeID ? 'Create' : 'Update'}`)
        } catch (error) {
            setIsLoading(error)
        }
    }, [specID, setIsLoading, history, batchChangeID, telemetryService, toBeArchived])

    return (
        <>
            <div
                className={classNames(
                    'alert alert-info mb-3 d-block d-md-flex align-items-center body-lead',
                    !isRedesignEnabled && 'p-3'
                )}
            >
                <div className={classNames(styles.createUpdateBatchChangeAlertCopy, 'flex-grow-1 mr-3')}>
                    {!batchChange && (
                        <>
                            Review the proposed changesets below. Click 'Apply spec' or run <code>src batch apply</code>{' '}
                            against your batch spec to create the batch change and perform the indicated action on each
                            changeset.
                        </>
                    )}
                    {batchChange && (
                        <>
                            This operation will update the existing batch change{' '}
                            <Link to={batchChange.url}>{batchChange.name}</Link>. Click 'Apply spec' or run{' '}
                            <code>src batch apply</code> against your batch spec to update the batch change and perform
                            the indicated action on each changeset.
                        </>
                    )}
                </div>
                <div className={styles.createUpdateBatchChangeAlertBtn}>
                    <button
                        type="button"
                        className={classNames(
                            'btn btn-primary test-batches-confirm-apply-btn text-nowrap',
                            isLoading === true || (!viewerCanAdminister && 'disabled')
                        )}
                        onClick={onApply}
                        disabled={isLoading === true || !viewerCanAdminister}
                        data-tooltip={
                            !viewerCanAdminister ? 'You have no permission to apply this batch change.' : undefined
                        }
                    >
                        Apply spec
                    </button>
                </div>
            </div>
            {isErrorLike(isLoading) && <ErrorAlert error={isLoading} />}
        </>
    )
}
