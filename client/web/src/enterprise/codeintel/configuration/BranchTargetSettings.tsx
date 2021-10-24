import { debounce } from 'lodash'
import React, { FunctionComponent, useState, useMemo } from 'react'

import { CodeIntelligenceConfigurationPolicyFields, GitObjectType } from '../../../graphql-operations'

import { GitObjectPreview } from './GitObjectPreview'

const DEBOUNCED_WAIT = 250
export interface BranchTargetSettingsProps {
    repoId?: string
    policy: CodeIntelligenceConfigurationPolicyFields
    setPolicy: (policy: CodeIntelligenceConfigurationPolicyFields) => void
    disabled: boolean
}

export const BranchTargetSettings: FunctionComponent<BranchTargetSettingsProps> = ({
    repoId,
    policy,
    setPolicy,
    disabled = false,
}) => {
    const [pattern, setPattern] = useState(policy.pattern)
    const debouncedSetPattern = useMemo(() => debounce(value => setPattern(value), DEBOUNCED_WAIT), [])

    return (
        <>
            <div className="form-group">
                <label htmlFor="name">Name</label>
                <input
                    id="name"
                    type="text"
                    className="form-control"
                    value={policy.name}
                    onChange={({ target: { value } }) => setPolicy({ ...policy, name: value })}
                    disabled={disabled}
                    required={true}
                />
                <small className="form-text text-muted">Required.</small>
            </div>
            <div className="form-group">
                <label htmlFor="type">Type</label>
                <select
                    id="type"
                    className="form-control"
                    value={policy.type}
                    onChange={({ target: { value } }) =>
                        setPolicy({
                            ...policy,
                            type: value as GitObjectType,
                            ...(value !== GitObjectType.GIT_TREE
                                ? {
                                      retainIntermediateCommits: false,
                                      indexIntermediateCommits: false,
                                  }
                                : {}),
                        })
                    }
                    disabled={disabled}
                >
                    <option value="">Select Git object type</option>
                    <option value={GitObjectType.GIT_COMMIT}>HEAD</option>
                    <option value={GitObjectType.GIT_TAG}>Tag</option>
                    <option value={GitObjectType.GIT_TREE}>Branch</option>
                </select>
                <small className="form-text text-muted">Required.</small>
            </div>
            {policy.type !== GitObjectType.GIT_COMMIT && (
                <div className="form-group">
                    <label htmlFor="pattern">Pattern</label>
                    <input
                        id="pattern"
                        type="text"
                        className="form-control text-monospace"
                        value={policy.pattern}
                        onChange={({ target: { value } }) => {
                            setPolicy({ ...policy, pattern: value })
                            debouncedSetPattern(value)
                        }}
                        disabled={disabled}
                        required={true}
                    />
                    <small className="form-text text-muted">Required.</small>
                </div>
            )}
            {repoId && <GitObjectPreview repoId={repoId} type={policy.type} pattern={pattern} />}
        </>
    )
}
