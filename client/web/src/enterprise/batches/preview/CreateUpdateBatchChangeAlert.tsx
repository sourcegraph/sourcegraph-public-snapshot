import classNames from 'classnames'
import * as H from 'history'
import React, { useContext, useState, useCallback, useMemo } from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { TelemetryProps, TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { ErrorAlert } from '../../../components/alerts'
import { BatchSpecFields, ChangesetSpecPublicationStateInput } from '../../../graphql-operations'
import { Action, DropdownButton } from '../DropdownButton'
import { MultiSelectContext } from '../MultiSelectContext'

import { applyBatchChange, createBatchChange, queryPublishableChangesetSpecs } from './backend'
import { BatchChangePreviewContext } from './BatchChangePreviewContext'
import styles from './CreateUpdateBatchChangeAlert.module.scss'

export enum CreateUpdateBatchChangeAlertAction {
    Apply,
    PublishAll,
    PublishSelected,
    DraftAll,
    DraftSelected,
}

export interface CreateUpdateBatchChangeAlertProps extends TelemetryProps {
    specID: string
    toBeArchived: number
    batchChange: BatchSpecFields['appliesToBatchChange']
    showPublishUI: boolean
    viewerCanAdminister: boolean
    history: H.History
    telemetryService: TelemetryService
}

export const CreateUpdateBatchChangeAlert: React.FunctionComponent<CreateUpdateBatchChangeAlertProps> = ({
    specID,
    toBeArchived,
    batchChange,
    showPublishUI,
    viewerCanAdminister,
    history,
    telemetryService,
}) => {
    const [isLoading, setIsLoading] = useState<boolean | Error>(false)

    const { filters, pagination } = useContext(BatchChangePreviewContext)
    const { selected } = useContext(MultiSelectContext)

    const batchChangeID = batchChange?.id
    const onApply = useCallback(
        async (action: CreateUpdateBatchChangeAlertAction, setIsLoading: (loadingOrError: boolean | Error) => void) => {
            if (!confirm(`Are you sure you want to ${batchChangeID ? 'update' : 'create'} this batch change?`)) {
                return
            }
            setIsLoading(true)
            try {
                let publicationStates: ChangesetSpecPublicationStateInput[] | null = null
                if (action !== CreateUpdateBatchChangeAlertAction.Apply) {
                    const ids =
                        selected === 'all'
                            ? await queryPublishableChangesetSpecs({
                                  batchSpec: specID,
                                  ...filters,
                                  ...pagination,
                              }).toPromise()
                            : [...selected]

                    const state =
                        action === CreateUpdateBatchChangeAlertAction.DraftAll ||
                        action === CreateUpdateBatchChangeAlertAction.DraftSelected
                            ? 'draft'
                            : true

                    publicationStates = ids.map(id => ({
                        changesetSpec: id,
                        publicationState: state,
                    }))
                }

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
        },
        [batchChangeID, filters, history, pagination, selected, specID, telemetryService, toBeArchived]
    )

    // The available actions are relatively complex. For batch specs without any
    // UI-controlled changesets, we'll only have one Apply action. For batch
    // specs with UI-controlled changesets, we'll provide options based on
    // whether there are changesets selected or not.
    const actions = useMemo((): Action[] => {
        let actions: Action[] = [
            {
                type: 'apply',
                buttonLabel: 'Apply',
                dropdownTitle: 'Apply',
                dropdownDescription:
                    'Apply the proposed changes without publishing any changesets without an explicit published field.',
                isAvailable: () => true,
                onTrigger: async onDone => {
                    await onApply(CreateUpdateBatchChangeAlertAction.Apply, setIsLoading)
                    onDone()
                },
            },
        ]

        if (showPublishUI) {
            const hasSelectedChangesets = selected === 'all' || selected.size > 0

            if (hasSelectedChangesets) {
                actions = [
                    {
                        type: 'publish-selected',
                        buttonLabel: 'Publish selected',
                        dropdownTitle: 'Publish selected',
                        dropdownDescription: 'Apply the proposed changes, publishing the selected changesets.',
                        isAvailable: () => true,
                        onTrigger: async onDone => {
                            await onApply(CreateUpdateBatchChangeAlertAction.PublishSelected, setIsLoading)
                            onDone()
                        },
                    },
                    {
                        type: 'draft-selected',
                        buttonLabel: 'Publish selected as draft',
                        dropdownTitle: 'Publish selected as draft',
                        dropdownDescription:
                            'Apply the proposed changes, publishing all selected changesets as drafts.',
                        isAvailable: () => true,
                        onTrigger: async onDone => {
                            await onApply(CreateUpdateBatchChangeAlertAction.DraftSelected, setIsLoading)
                            onDone()
                        },
                    },
                ]
            } else {
                actions.push(
                    {
                        type: 'publish-all',
                        buttonLabel: 'Publish all',
                        dropdownTitle: 'Publish all',
                        dropdownDescription: 'Apply the proposed changes, publishing all changesets.',
                        isAvailable: () => true,
                        onTrigger: async onDone => {
                            await onApply(CreateUpdateBatchChangeAlertAction.PublishAll, setIsLoading)
                            onDone()
                        },
                    },
                    {
                        type: 'draft-all',
                        buttonLabel: 'Publish all as draft',
                        dropdownTitle: 'Publish all as draft',
                        dropdownDescription: 'Apply the proposed changes, publishing all changesets as drafts.',
                        isAvailable: () => true,
                        onTrigger: async onDone => {
                            await onApply(CreateUpdateBatchChangeAlertAction.DraftAll, setIsLoading)
                            onDone()
                        },
                    }
                )
            }
        }

        return actions
    }, [onApply, selected, showPublishUI])

    // We'll track the button label so we can update the text in the banner.
    const [label, setLabel] = useState<string | undefined>(undefined)

    // Now we have the label, let's construct the call to action for the user.
    const callToAction = useMemo(() => (label !== undefined ? `Click '${label}'` : 'Select an action'), [label])

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
                            Review the proposed changesets below. {callToAction} to create the batch change and perform
                            the indicated action on each changeset.
                        </>
                    )}
                    {batchChange && (
                        <>
                            This operation will update the existing batch change{' '}
                            <Link to={batchChange.url}>{batchChange.name}</Link>. {callToAction} to update the batch
                            change and perform the indicated action on each changeset.
                        </>
                    )}
                </div>
                <div className={styles.createUpdateBatchChangeAlertBtn}>
                    <DropdownButton
                        actions={actions}
                        defaultAction={actions[0].type === 'apply' ? 0 : undefined}
                        disabled={!viewerCanAdminister}
                        onLabel={setLabel}
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
