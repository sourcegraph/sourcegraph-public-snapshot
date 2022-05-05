import React, { useCallback, useState } from 'react'

import * as H from 'history'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { isErrorLike, asError, pluralize } from '@sourcegraph/common'
import { Button, AlertLink, CardBody, Card, Alert, Checkbox } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../components/LoaderButton'
import { Scalars } from '../../../graphql-operations'

import { closeBatchChange as _closeBatchChange } from './backend'

export interface BatchChangeCloseAlertProps {
    batchChangeID: Scalars['ID']
    batchChangeURL: string
    closeChangesets: boolean
    viewerCanAdminister: boolean
    totalCount: number
    setCloseChangesets: (newValue: boolean) => void
    history: H.History

    /** For testing only. */
    closeBatchChange?: typeof _closeBatchChange
}

export const BatchChangeCloseAlert: React.FunctionComponent<React.PropsWithChildren<BatchChangeCloseAlertProps>> = ({
    batchChangeID,
    batchChangeURL,
    closeChangesets,
    totalCount,
    setCloseChangesets,
    viewerCanAdminister,
    history,
    closeBatchChange = _closeBatchChange,
}) => {
    const onChangeCloseChangesets = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        event => {
            setCloseChangesets(event.target.checked)
        },
        [setCloseChangesets]
    )
    const onCancel = useCallback<React.MouseEventHandler>(() => {
        history.push(batchChangeURL)
    }, [history, batchChangeURL])
    const [isClosing, setIsClosing] = useState<boolean | Error>(false)
    const onClose = useCallback<React.MouseEventHandler>(async () => {
        setIsClosing(true)
        try {
            await closeBatchChange({ batchChange: batchChangeID, closeChangesets })
            history.push(batchChangeURL)
        } catch (error) {
            setIsClosing(asError(error))
        }
    }, [history, closeChangesets, closeBatchChange, batchChangeID, batchChangeURL])
    return (
        <>
            <Card className="mb-3">
                <CardBody>
                    <p>
                        <strong>
                            After closing this batch change, it will be read-only and no new batch specs can be applied.
                        </strong>
                    </p>
                    {totalCount > 0 && (
                        <>
                            <p>By default, all changesets remain untouched.</p>
                            <Checkbox
                                wrapperClassName="mb-3"
                                id="closeChangesets"
                                checked={closeChangesets}
                                onChange={onChangeCloseChangesets}
                                className="test-batches-close-changesets-checkbox"
                                disabled={isClosing === true || !viewerCanAdminister}
                                label={
                                    <>
                                        Also close {pluralize('the', totalCount, 'all')} {totalCount}{' '}
                                        {pluralize(
                                            'open changeset on the code host',
                                            totalCount,
                                            'open changesets on the code hosts'
                                        )}
                                        .
                                    </>
                                }
                            />
                        </>
                    )}
                    {!viewerCanAdminister && (
                        <Alert variant="warning">
                            You don't have permission to close this batch change. See{' '}
                            <AlertLink to="https://docs.sourcegraph.com/batch_changes/explanations/permissions_in_batch_changes">
                                Permissions in batch changes
                            </AlertLink>{' '}
                            for more information about the batch changes permission model.
                        </Alert>
                    )}
                    <div className="d-flex justify-content-end">
                        <Button
                            className="mr-2 test-batches-close-abort-btn"
                            onClick={onCancel}
                            disabled={isClosing === true || !viewerCanAdminister}
                            variant="secondary"
                        >
                            Cancel
                        </Button>
                        <LoaderButton
                            className="test-batches-confirm-close-btn"
                            onClick={onClose}
                            disabled={isClosing === true || !viewerCanAdminister}
                            variant="danger"
                            loading={isClosing === true}
                            label="Close batch change"
                            alwaysShowLabel={true}
                        />
                    </div>
                </CardBody>
            </Card>
            {isErrorLike(isClosing) && <ErrorAlert error={isClosing} />}
        </>
    )
}
