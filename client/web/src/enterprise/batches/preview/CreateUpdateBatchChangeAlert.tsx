import classNames from 'classnames'
import * as H from 'history'
import React, { useCallback, useContext, useState } from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { ErrorAlert } from '../../../components/alerts'
import { BatchSpecFields } from '../../../graphql-operations'

import { createBatchChange, applyBatchChange } from './backend'
import { BatchChangePreviewContext } from './BatchChangePreviewContext'
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

    const { publicationStates } = useContext(BatchChangePreviewContext)

    const [isLoading, setIsLoading] = useState<boolean | Error>(false)

    const onApply = useCallback(async () => {
        if (!confirm(`Are you sure you want to ${batchChangeID ? 'update' : 'create'} this batch change?`)) {
            return
        }
        setIsLoading(true)
        try {
            const batchChange = batchChangeID
                ? await applyBatchChange({ batchSpec: specID, batchChange: batchChangeID, publicationStates })
                : await createBatchChange({ batchSpec: specID, publicationStates })

            if (toBeArchived > 0) {
                history.push(`${batchChange.url}?archivedCount=${toBeArchived}&archivedBy=${specID}`)
            } else {
                history.push(batchChange.url)
            }
            telemetryService.logViewEvent(`BatchChangeDetailsPageAfter${batchChangeID ? 'Create' : 'Update'}`)
        } catch (error) {
            setIsLoading(error)
        }
    }, [specID, setIsLoading, history, batchChangeID, telemetryService, toBeArchived, publicationStates])

    return (
        <>
            <div className="alert alert-info mb-3 d-block d-md-flex align-items-center body-lead">
                <div className={classNames(styles.createUpdateBatchChangeAlertCopy, 'flex-grow-1 mr-3')}>
                    {batchChange ? (
                        <>
                            This operation will update the existing batch change{' '}
                            <Link to={batchChange.url}>{batchChange.name}</Link>
                        </>
                    ) : (
                        'Review the proposed changesets below.'
                    )}{' '}
                    Click 'Apply' or run <code>src batch apply</code> against your batch spec to{' '}
                    {batchChange ? 'update' : 'create'} the batch change and perform the indicated action on each
                    changeset. Select a changeset and modify the action to customize the publication state of each or
                    all changesets.
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
                        Apply
                    </button>
                </div>
            </div>
            {isErrorLike(isLoading) && <ErrorAlert error={isLoading} />}
        </>
    )
}
