import React, { useCallback, useContext, useState } from 'react'

import classNames from 'classnames'
import * as H from 'history'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { isErrorLike } from '@sourcegraph/common'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Alert, Link } from '@sourcegraph/wildcard'

import { BatchSpecFields } from '../../../graphql-operations'
import { MultiSelectContext } from '../MultiSelectContext'

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

export const CreateUpdateBatchChangeAlert: React.FunctionComponent<
    React.PropsWithChildren<CreateUpdateBatchChangeAlertProps>
> = ({ specID, toBeArchived, batchChange, viewerCanAdminister, history, telemetryService }) => {
    const batchChangeID = batchChange?.id

    // `BatchChangePreviewContext` is responsible for managing the overrideable
    // publication states for preview changesets on the clientside.
    const { publicationStates } = useContext(BatchChangePreviewContext)
    const { selected } = useContext(MultiSelectContext)

    const [isLoading, setIsLoading] = useState<boolean | Error>(false)

    const canApply = selected !== 'all' && selected.size === 0 && !isLoading && viewerCanAdminister

    // Returns a tooltip error message appropriate for the given circumstances preventing
    // the user from applying the preview.
    const disabledTooltip = useCallback((): string | undefined => {
        if (canApply) {
            return undefined
        }
        if (!viewerCanAdminister) {
            return "You don't have permission to apply this batch change."
        }
        if (selected === 'all' || selected.size > 0) {
            return 'You have selected changesets. Choose the action to be performed on apply or deselect to continue.'
        }
        return undefined
    }, [canApply, selected, viewerCanAdminister])

    const onApply = useCallback(async () => {
        if (!canApply) {
            return
        }
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
    }, [canApply, specID, setIsLoading, history, batchChangeID, telemetryService, toBeArchived, publicationStates])

    return (
        <>
            <Alert className="mb-3 d-block d-md-flex align-items-center body-lead" variant="info">
                <div className={classNames(styles.createUpdateBatchChangeAlertCopy, 'flex-grow-1 mr-3')}>
                    {batchChange ? (
                        <>
                            This operation will update the existing batch change{' '}
                            <Link to={batchChange.url}>{batchChange.name}</Link>.
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
                    <Button
                        variant="primary"
                        className={classNames(
                            'test-batches-confirm-apply-btn text-nowrap',
                            isLoading === true || (!viewerCanAdminister && 'disabled')
                        )}
                        onClick={onApply}
                        disabled={!canApply}
                        data-tooltip={disabledTooltip()}
                    >
                        Apply
                    </Button>
                </div>
            </Alert>
            {isErrorLike(isLoading) && <ErrorAlert error={isLoading} />}
        </>
    )
}
