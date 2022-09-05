import { FunctionComponent } from 'react'

import { Select } from '@sourcegraph/wildcard'

import { GitObjectType } from '../../../../graphql-operations'

export interface GitTypeSelectorProps {
    type: GitObjectType
    setType: (type: GitObjectType) => void
    disabled: boolean
}

export const GitTypeSelector: FunctionComponent<React.PropsWithChildren<GitTypeSelectorProps>> = ({
    type,
    setType,
    disabled,
}) => (
    <>
        <Select
            id="type"
            label="Type"
            value={type}
            onChange={({ target: { value } }) => setType(value as GitObjectType)}
            disabled={disabled}
            message="Required."
        >
            <option value="">Select Git object type</option>
            <option value={GitObjectType.GIT_COMMIT}>HEAD</option>
            <option value={GitObjectType.GIT_TAG}>Tag</option>
            <option value={GitObjectType.GIT_TREE}>Branch</option>
        </Select>
    </>
)
