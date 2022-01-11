import * as H from 'history'
import DeleteIcon from 'mdi-react/DeleteIcon'
import InformationIcon from 'mdi-react/InformationIcon'
import PencilIcon from 'mdi-react/PencilIcon'
import React, { useCallback, useState } from 'react'

import { isErrorLike, asError } from '@sourcegraph/common'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { Button } from '@sourcegraph/wildcard'

import { isBatchChangesExecutionEnabled } from '../../../batches'
import { Scalars } from '../../../graphql-operations'
import { Settings } from '../../../schema/settings.schema'

import { deleteBatchChange as _deleteBatchChange } from './backend'

export interface BatchChangeDetailsActionSectionProps extends SettingsCascadeProps<Settings> {
    batchChangeID: Scalars['ID']
    batchChangeClosed: boolean
    batchChangeNamespaceURL: string
    history: H.History

    /** For testing only. */
    deleteBatchChange?: typeof _deleteBatchChange
}

export const BatchChangeDetailsActionSection: React.FunctionComponent<BatchChangeDetailsActionSectionProps> = ({
    batchChangeID,
    batchChangeClosed,
    batchChangeNamespaceURL,
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
                {isErrorLike(isDeleting) && <InformationIcon className="icon-inline" data-tooltip={isDeleting} />}
                <DeleteIcon className="icon-inline" /> Delete
            </Button>
        )
    }
    return (
        <div className="d-flex">
            {showEditButton && (
                <Link to={`${location.pathname}/edit`} className="mr-2 btn btn-secondary">
                    <PencilIcon className="icon-inline" /> Edit
                </Link>
            )}
            <Link
                to={`${location.pathname}/close`}
                className="btn btn-outline-danger test-batches-close-btn"
                data-tooltip="View a preview of all changes that will happen when you close this batch change."
            >
                <DeleteIcon className="icon-inline" /> Close
            </Link>
        </div>
    )
}
