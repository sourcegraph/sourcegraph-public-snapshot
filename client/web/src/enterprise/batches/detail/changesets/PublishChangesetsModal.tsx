import React, { useCallback, useState } from 'react'

import { asError, isErrorLike } from '@sourcegraph/common'
import { Button, Checkbox, Modal, H3, Text, ErrorAlert, Form } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../../components/LoaderButton'
import type { Scalars } from '../../../../graphql-operations'
import { publishChangesets as _publishChangesets } from '../backend'

export interface PublishChangesetsModalProps {
    onCancel: () => void
    afterCreate: () => void
    batchChangeID: Scalars['ID']
    changesetIDs: Scalars['ID'][]

    /** For testing only. */
    publishChangesets?: typeof _publishChangesets
}

export const PublishChangesetsModal: React.FunctionComponent<React.PropsWithChildren<PublishChangesetsModalProps>> = ({
    onCancel,
    afterCreate,
    batchChangeID,
    changesetIDs,
    publishChangesets: publishChangesets = _publishChangesets,
}) => {
    const [isLoading, setIsLoading] = useState<boolean | Error>(false)
    const [draft, setDraft] = useState(false)

    const onSubmit = useCallback<React.FormEventHandler>(async () => {
        setIsLoading(true)
        try {
            await publishChangesets(batchChangeID, changesetIDs, draft)
            afterCreate()
        } catch (error) {
            setIsLoading(asError(error))
        }
    }, [changesetIDs, publishChangesets, batchChangeID, draft, afterCreate])

    const onToggleDraft = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setDraft(event.target.checked)
    }, [])

    return (
        <Modal onDismiss={onCancel} aria-labelledby={MODAL_LABEL_ID}>
            <H3 id={MODAL_LABEL_ID}>Publish changesets</H3>
            <Text className="mb-4">
                Are you sure you want to publish all the selected changesets to the code hosts?
            </Text>
            <Form>
                <div className="form-group">
                    <Checkbox
                        id={CHECKBOX_ID}
                        checked={draft}
                        onChange={onToggleDraft}
                        disabled={isLoading === true}
                        label="Publish as draft."
                    />
                </div>
            </Form>
            {isErrorLike(isLoading) && <ErrorAlert error={isLoading} />}
            <div className="d-flex justify-content-end">
                <Button
                    disabled={isLoading === true}
                    className="mr-2"
                    onClick={onCancel}
                    outline={true}
                    variant="secondary"
                >
                    Cancel
                </Button>
                <LoaderButton
                    onClick={onSubmit}
                    disabled={isLoading === true}
                    variant="primary"
                    loading={isLoading === true}
                    alwaysShowLabel={true}
                    label="Publish"
                />
            </div>
        </Modal>
    )
}

const MODAL_LABEL_ID = 'publish-changesets-modal-title'
const CHECKBOX_ID = 'publish-changesets-modal-draft-check'
