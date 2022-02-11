import React, { FunctionComponent } from 'react'

import { GitObjectType } from '../../../../graphql-operations'

export interface GitTypeSelectorProps {
    type: GitObjectType
    setType: (type: GitObjectType) => void
    disabled: boolean
}

export const GitTypeSelector: FunctionComponent<GitTypeSelectorProps> = ({ type, setType, disabled }) => (
    <div className="form-group">
        <label htmlFor="type">Type</label>
        <select
            id="type"
            className="form-control"
            value={type}
            onChange={({ target: { value } }) => setType(value as GitObjectType)}
            disabled={disabled}
        >
            <option value="">Select Git object type</option>
            <option value={GitObjectType.GIT_COMMIT}>HEAD</option>
            <option value={GitObjectType.GIT_TAG}>Tag</option>
            <option value={GitObjectType.GIT_TREE}>Branch</option>
        </select>
        <small className="form-text text-muted">Required.</small>
    </div>
)
