import React, { FunctionComponent, useState } from 'react'

import { CodeIntelligenceConfigurationPolicyFields, GitObjectType } from '../../../graphql-operations'

import { GitObjectPreview } from './GitObjectPreview'

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
                />
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
                    {repoId && <option value={GitObjectType.GIT_COMMIT}>Commit</option>}
                    <option value={GitObjectType.GIT_TAG}>Tag</option>
                    <option value={GitObjectType.GIT_TREE}>Branch</option>
                </select>
            </div>
            <div className="form-group">
                <label htmlFor="pattern">Pattern</label>
                <input
                    id="pattern"
                    type="text"
                    className="form-control text-monospace"
                    value={policy.pattern}
                    onChange={({ target: { value } }) => {
                        setPolicy({ ...policy, pattern: value })
                        setPattern(value)
                    }}
                    disabled={disabled}
                />
            </div>

            {repoId && <GitObjectPreview repoId={repoId} type={policy.type} pattern={pattern} />}
        </>
    )
}
