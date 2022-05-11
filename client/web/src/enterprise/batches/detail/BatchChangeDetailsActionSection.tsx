import React, { useCallback, useState } from 'react'

import * as H from 'history'
import DeleteIcon from 'mdi-react/DeleteIcon'
import InformationIcon from 'mdi-react/InformationIcon'
import PencilIcon from 'mdi-react/PencilIcon'

import { isErrorLike, asError } from '@sourcegraph/common'
import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { Button, Link, Icon } from '@sourcegraph/wildcard'

import { isBatchChangesExecutionEnabled } from '../../../batches'
import { Scalars } from '../../../graphql-operations'

import { deleteBatchChange as _deleteBatchChange } from './backend'

export interface BatchChangeDetailsActionSectionProps extends SettingsCascadeProps<Settings> {
    batchChangeID: Scalars['ID']
    batchChangeClosed: boolean
    batchChangeNamespaceURL: string
    batchChangeURL: string
    history: H.History

    /** For testing only. */
    deleteBatchChange?: typeof _deleteBatchChange
}

export const BatchChangeDetailsActionSection: React.FunctionComponent<
    React.PropsWithChildren<BatchChangeDetailsActionSectionProps>
> = ({
    batchChangeID,
    batchChangeClosed,
    batchChangeNamespaceURL,
    batchChangeURL,
    history,
    settingsCascade,
    deleteBatchChange = _deleteBatchChange,
}) => {
    const showEditButton = isBatchChangesExecutionEnabled(settingsCascade)

    const [isDeleting, setIsDeleting] = useState<boolean | Error>(false)
    const onDeleteBatchChange = useCallback(async () => {
        if (!confirm('Do you really want to delete this batch change?')) {
            return
        }
        setIsDeleting(true)
        try {
            await deleteBatchChange(batchChangeID)
            history.push(batchChangeNamespaceURL + '/batch-changes')
        } catch (error) {
            setIsDeleting(asError(error))
        }
    }, [batchChangeID, deleteBatchChange, history, batchChangeNamespaceURL])
    if (batchChangeClosed) {
        return (
            <Button
                className="test-batches-delete-btn"
                onClick={onDeleteBatchChange}
                data-tooltip="Deleting this batch change is a final action."
                disabled={isDeleting === true}
                outline={true}
                variant="danger"
            >
                {isErrorLike(isDeleting) && <Icon data-tooltip={isDeleting} as={InformationIcon} />}
                <Icon as={DeleteIcon} /> Delete
            </Button>
        )
    }
    return (
        <div className="d-flex">
            {showEditButton && (
                <Button to={`${batchChangeURL}/edit`} className="mr-2" variant="secondary" as={Link}>
                    <Icon as={PencilIcon} /> Edit
                </Button>
            )}
            <Button
                to={`${batchChangeURL}/close`}
                className="test-batches-close-btn"
                data-tooltip="View a preview of all changes that will happen when you close this batch change."
                variant="danger"
                outline={true}
                as={Link}
            >
                <Icon as={DeleteIcon} /> Close
            </Button>
        </div>
    )
}
