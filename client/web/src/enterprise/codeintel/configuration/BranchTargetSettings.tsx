import { debounce } from 'lodash'
import React, { FunctionComponent, useState, useMemo } from 'react'

import { CodeIntelligenceConfigurationPolicyFields, GitObjectType } from '../../../graphql-operations'

import { GitObjectPreview } from './GitObjectPreview'
import { RepositoryPreview } from './RepositoryPreview'

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

    const [repositoryPatterns, setRepositoryPatterns] = useState<string[]>([])

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

            {!repoId && (
                <>
                    {repositoryPatterns.length === 0 ? (
                        <>
                            <p>Something special here</p>

                            <button
                                className="btn btn-primary"
                                onClick={() => setRepositoryPatterns(repositoryPatterns.concat(['']))}
                            >
                                add first repository pattern
                            </button>
                        </>
                    ) : (
                        <>
                            {repositoryPatterns.map((p, index) => (
                                <div className="form-group" key={index}>
                                    <label htmlFor="repo-pattern">Repository pattern #{index + 1}</label>
                                    <input
                                        id={`repo-pattern-${index}`}
                                        type="text"
                                        className="form-control text-monospace"
                                        value={repositoryPatterns[index]}
                                        onChange={({ target }) =>
                                            setRepositoryPatterns(
                                                repositoryPatterns.map((p, j) => (index === j ? target.value : p))
                                            )
                                        }
                                        disabled={disabled}
                                        required={true}
                                    />

                                    <button
                                        className="btn btn-danger"
                                        onClick={() =>
                                            setRepositoryPatterns(repositoryPatterns.filter((_, j) => index !== j))
                                        }
                                    >
                                        remove me
                                    </button>

                                    <RepositoryPreview pattern={p} />
                                </div>
                            ))}

                            <button
                                className="btn btn-primary"
                                onClick={() => setRepositoryPatterns(repositoryPatterns.concat(['']))}
                            >
                                add new repository pattern
                            </button>
                        </>
                    )}
                </>
            )}
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
