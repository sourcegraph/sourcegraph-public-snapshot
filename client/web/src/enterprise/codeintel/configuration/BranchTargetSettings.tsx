import classNames from 'classnames'
import { debounce } from 'lodash'
import React, { FunctionComponent, useMemo, useState } from 'react'

import { Button } from '@sourcegraph/wildcard'
import { Tooltip } from '@sourcegraph/branded/src/components/tooltip/Tooltip'

import { CodeIntelligenceConfigurationPolicyFields, GitObjectType } from '../../../graphql-operations'

import { GitObjectPreview } from './GitObjectPreview'
import styles from './BranchTargetSettings.module.scss'
import TrashIcon from 'mdi-react/TrashIcon'

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
                        <div>
                            This configuration policy applies to all repositories. To restrict the set of repositories
                            to which this configuration applies,{' '}
                            <span
                                className={styles.addFirstRepositoryPattern}
                                onClick={() => setRepositoryPatterns(repositoryPatterns.concat(['']))}
                            >
                                add a repository pattern
                            </span>
                            .
                        </div>
                    ) : (
                        <div className={styles.grid}>
                            {repositoryPatterns.map((p, index) => (
                                <React.Fragment key={index}>
                                    <div className={classNames(styles.name, 'form-group d-flex flex-column')}>
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
                                    </div>

                                    <span className={classNames(styles.button, 'd-none d-md-inline')}>
                                        <Button
                                            onClick={() =>
                                                setRepositoryPatterns(repositoryPatterns.filter((_, j) => index !== j))
                                            }
                                            className="p-0 m-0 pt-1"
                                            disabled={disabled}
                                        >
                                            <Tooltip />
                                            <TrashIcon
                                                className="icon-inline text-danger"
                                                data-tooltip="Delete the repository pattern"
                                            />
                                        </Button>
                                    </span>

                                    {/* <RepositoryPreview pattern={p} /> */}
                                </React.Fragment>
                            ))}

                            <button
                                className="btn btn-primary"
                                onClick={() => setRepositoryPatterns(repositoryPatterns.concat(['']))}
                            >
                                add new repository pattern
                            </button>
                        </div>
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
