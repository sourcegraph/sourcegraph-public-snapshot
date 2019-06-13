import PencilIcon from 'mdi-react/PencilIcon'
import React, { useCallback, useState } from 'react'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { Label } from '../../../components/Label'
import { UpdateLabelForm } from './EditLabelForm'
import { LabelDeleteButton } from './LabelDeleteButton'

interface Props extends ExtensionsControllerProps {
    label: GQL.ILabel

    /** Called when the label is updated. */
    onLabelUpdate: () => void
}

/**
 * A row in the list of labels.
 */
export const LabelRow: React.FunctionComponent<Props> = ({ label, onLabelUpdate, ...props }) => {
    const [isEditing, setIsEditing] = useState(false)
    const toggleIsEditing = useCallback(() => setIsEditing(!isEditing), [isEditing])

    return isEditing ? (
        <UpdateLabelForm label={label} onLabelUpdate={onLabelUpdate} onDismiss={toggleIsEditing} />
    ) : (
        <div className="d-flex align-items-center flex-wrap">
            <div className="flex-1">
                <Label label={label} className="h5 mb-0" />
            </div>
            <p className="mb-0 flex-1">{label.description}</p>
            <div className="text-right flex-0">
                <button type="button" className="btn btn-link text-decoration-none" onClick={toggleIsEditing}>
                    <PencilIcon className="icon-inline" /> Edit
                </button>
                <LabelDeleteButton {...props} label={label} onDelete={onLabelUpdate} />
            </div>
        </div>
    )
}
