import classNames from 'classnames'
import React, { useState } from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { ErrorAlert } from '../../../components/alerts'
import { BatchSpecFields } from '../../../graphql-operations'
import { Action, DropdownButton } from '../DropdownButton'

import styles from './CreateUpdateBatchChangeAlert.module.scss'

export enum CreateUpdateBatchChangeAlertAction {
    Apply,
    PublishAll,
    PublishSelected,
    DraftAll,
    DraftSelected,
}

export interface CreateUpdateBatchChangeAlertProps extends TelemetryProps {
    batchChange: BatchSpecFields['appliesToBatchChange']
    showPublishUI: boolean
    onApply: (
        action: CreateUpdateBatchChangeAlertAction,
        setIsLoading: (loadingOrError: boolean | Error) => void
    ) => Promise<void>
    viewerCanAdminister: boolean
}

export const CreateUpdateBatchChangeAlert: React.FunctionComponent<CreateUpdateBatchChangeAlertProps> = ({
    batchChange,
    showPublishUI,
    onApply,
    viewerCanAdminister,
}) => {
    const [isLoading, setIsLoading] = useState<boolean | Error>(false)

    const actions: Action[] = [
        {
            type: 'apply',
            buttonLabel: 'Apply',
            dropdownTitle: 'Apply',
            dropdownDescription:
                'Apply the proposed changes without publishing any changesets without an explicit published field.',
            isAvailable: () => true,
            onTrigger: async (onDone, onCancel) => {
                await onApply(CreateUpdateBatchChangeAlertAction.Apply, setIsLoading)
                onDone()
            },
        },
        {
            type: 'publish-all',
            buttonLabel: 'Publish all',
            dropdownTitle: 'Publish all',
            dropdownDescription: 'Apply the proposed changes, publishing all changesets.',
            isAvailable: () => showPublishUI,
            onTrigger: async (onDone, onCancel) => {
                await onApply(CreateUpdateBatchChangeAlertAction.PublishAll, setIsLoading)
                onDone()
            },
        },
        {
            type: 'publish-selected',
            buttonLabel: 'Publish selected',
            dropdownTitle: 'Publish selected',
            dropdownDescription: 'Apply the proposed changes, publishing the selected changesets.',
            isAvailable: () => showPublishUI,
            onTrigger: async (onDone, onCancel) => {
                await onApply(CreateUpdateBatchChangeAlertAction.PublishSelected, setIsLoading)
                onDone()
            },
        },
        {
            type: 'draft-all',
            buttonLabel: 'Publish all as draft',
            dropdownTitle: 'Publish all as draft',
            dropdownDescription: 'Apply the proposed changes, publishing all changesets as drafts.',
            isAvailable: () => showPublishUI,
            onTrigger: async (onDone, onCancel) => {
                await onApply(CreateUpdateBatchChangeAlertAction.DraftAll, setIsLoading)
                onDone()
            },
        },
        {
            type: 'draft-selected',
            buttonLabel: 'Publish selected as draft',
            dropdownTitle: 'Publish selected as draft',
            dropdownDescription: 'Apply the proposed changes, publishing all selected changesets as drafts.',
            isAvailable: () => showPublishUI,
            onTrigger: async (onDone, onCancel) => {
                await onApply(CreateUpdateBatchChangeAlertAction.DraftSelected, setIsLoading)
                onDone()
            },
        },
    ]

    return (
        <>
            <div
                className={classNames(
                    'alert alert-info mb-3 d-block d-md-flex align-items-center body-lead',
                    styles.createUpdateBatchChangeAlert
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
                    <DropdownButton
                        actions={actions}
                        disabled={!viewerCanAdminister}
                        tooltip={
                            !viewerCanAdminister ? 'You do not have permission to apply this batch change.' : undefined
                        }
                    />
                </div>
            </div>
            {isErrorLike(isLoading) && <ErrorAlert error={isLoading} />}
        </>
    )
}
