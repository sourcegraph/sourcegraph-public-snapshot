import React, { useCallback } from 'react'
import { UpdateGraphVariables } from '../../../graphql-operations'
import { VALID_USERNAME_REGEXP } from '../../../user'

export type GraphFormValue = Pick<UpdateGraphVariables['input'], 'name' | 'description' | 'spec'>

const VALID_NAME_REGEXP = VALID_USERNAME_REGEXP // same format as for username (graphs_name_valid_chars PostgreSQL constraint)
const NAME_MAX_LENGTH = 255 // graphs_name_max_length PostgreSQL constraint

interface Props {
    value: GraphFormValue
    onChange: (newValue: GraphFormValue) => void
}

export const GraphFormFields: React.FunctionComponent<Props> = ({ value, onChange }) => {
    const onNameChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        event => onChange({ ...value, name: event.target.value }),
        [onChange, value]
    )
    const onDescriptionChange = useCallback<React.ChangeEventHandler<HTMLTextAreaElement>>(
        event => onChange({ ...value, description: event.target.value }),
        [onChange, value]
    )
    const onSpecChange = useCallback<React.ChangeEventHandler<HTMLTextAreaElement>>(
        event => onChange({ ...value, spec: event.target.value }),
        [onChange, value]
    )

    return (
        <>
            <div className="form-group">
                <label htmlFor="GraphFormFields__name">Name</label>
                <input
                    id="GraphFormFields__name"
                    type="text"
                    className="form-control"
                    value={value.name}
                    onChange={onNameChange}
                    required={true}
                    spellCheck={false}
                    pattern={VALID_NAME_REGEXP}
                    maxLength={NAME_MAX_LENGTH}
                    autoCapitalize="off"
                />
            </div>
            <div className="form-group">
                <label htmlFor="GraphFormFields__description">Description</label>
                <textarea
                    id="GraphFormFields__description"
                    className="form-control"
                    value={value.description || ''}
                    onChange={onDescriptionChange}
                    rows={3}
                />
            </div>
            <div className="form-group">
                <label htmlFor="GraphFormFields__spec">Repositories</label>
                <textarea
                    id="GraphFormFields__spec"
                    className="form-control"
                    value={value.spec}
                    onChange={onSpecChange}
                    rows={5}
                />
                <small className="form-text text-muted">List repositories by name (one per line).</small>
            </div>
        </>
    )
}
