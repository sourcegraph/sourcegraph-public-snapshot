import { FunctionComponent } from 'react'

import { GitObjectType } from '@sourcegraph/shared/src/graphql-operations'

import { CodeIntelligenceConfigurationPolicyFields } from '../../../../graphql-operations'
import { nullPolicy } from '../hooks/types'

import { GitTypeSelector } from './GitTypeSelector'
import { ObjectsMatchingGitPattern } from './ObjectsMatchingGitPattern'
import { ReposMatchingPatternList } from './ReposMatchingPatternList'

export interface BranchTargetSettingsProps {
    repoId?: string
    policy: CodeIntelligenceConfigurationPolicyFields
    setPolicy: (
        updater: (
            policy: CodeIntelligenceConfigurationPolicyFields | undefined
        ) => CodeIntelligenceConfigurationPolicyFields
    ) => void
    disabled: boolean
}

export const BranchTargetSettings: FunctionComponent<React.PropsWithChildren<BranchTargetSettingsProps>> = ({
    repoId,
    policy,
    setPolicy,
    disabled = false,
}) => {
    const updatePolicy = <K extends keyof CodeIntelligenceConfigurationPolicyFields>(
        updates: { [P in K]: CodeIntelligenceConfigurationPolicyFields[P] }
    ): void => {
        setPolicy(policy => ({ ...(policy || nullPolicy), ...updates }))
    }

    return (
        <>
            <div className="form-group">
                <label htmlFor="name">Name</label>
                <input
                    id="name"
                    type="text"
                    className="form-control"
                    value={policy.name}
                    onChange={({ target: { value: name } }) => updatePolicy({ name })}
                    disabled={disabled}
                    required={true}
                />
                <small className="form-text text-muted">Required.</small>
            </div>

            {repoId || policy.repository ? (
                <div className="mb-3">
                    This configuration policy applies only to {policy.repository?.name || 'the current repository'}.
                </div>
            ) : (
                <ReposMatchingPatternList
                    repositoryPatterns={policy.repositoryPatterns}
                    setRepositoryPatterns={updater =>
                        updatePolicy({
                            repositoryPatterns: updater((policy || nullPolicy).repositoryPatterns),
                        })
                    }
                    disabled={disabled}
                />
            )}

            <GitTypeSelector
                type={policy.type}
                setType={type =>
                    type === GitObjectType.GIT_TREE
                        ? updatePolicy({ type })
                        : updatePolicy({ type, retainIntermediateCommits: false, indexIntermediateCommits: false })
                }
                disabled={disabled}
            />

            <ObjectsMatchingGitPattern
                repoId={repoId || policy.repository?.id}
                type={policy.type}
                pattern={policy.pattern}
                setPattern={pattern => updatePolicy({ ...(policy || nullPolicy), pattern })}
                disabled={disabled}
            />
        </>
    )
}
