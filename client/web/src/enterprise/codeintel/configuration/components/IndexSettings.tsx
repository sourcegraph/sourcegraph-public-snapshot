import React, { FunctionComponent } from 'react'

import { Toggle } from '@sourcegraph/branded/src/components/Toggle'

import { CodeIntelligenceConfigurationPolicyFields, GitObjectType } from '../../../../graphql-operations'
import { nullPolicy } from '../hooks/types'

import { DurationSelect } from './DurationSelect'

export interface IndexingSettingsProps {
    policy: CodeIntelligenceConfigurationPolicyFields
    repo?: { id: string }
    setPolicy: (
        updater: (
            policy: CodeIntelligenceConfigurationPolicyFields | undefined
        ) => CodeIntelligenceConfigurationPolicyFields
    ) => void
    allowGlobalPolicies?: boolean
}

export const IndexingSettings: FunctionComponent<IndexingSettingsProps> = ({
    policy,
    repo,
    setPolicy,
    allowGlobalPolicies = window.context?.codeIntelAutoIndexingAllowGlobalPolicies,
}) => {
    const updatePolicy = <K extends keyof CodeIntelligenceConfigurationPolicyFields>(
        updates: { [P in K]: CodeIntelligenceConfigurationPolicyFields[P] }
    ): void => {
        setPolicy(policy => ({ ...(policy || nullPolicy), ...updates }))
    }

    return (
        <div className="form-group">
            <h3>Auto-indexing</h3>
            <div className="form-group">
                <Toggle
                    id="indexing-enabled"
                    title="Enabled"
                    value={policy.indexingEnabled}
                    onToggle={indexingEnabled => updatePolicy({ indexingEnabled })}
                />
                <label htmlFor="indexing-enabled" className="ml-2">
                    Enabled / disabled
                </label>
            </div>

            {!allowGlobalPolicies &&
                repo === undefined &&
                (policy.repositoryPatterns || []).length === 0 &&
                policy.indexingEnabled && (
                    <div className="alert alert-danger">
                        This Sourcegraph instance has disabled global policies for auto-indexing. Create a more
                        constrained policy targeting an explicit set of repositories to enable this policy.
                    </div>
                )}

            <div className="mb-4">
                <label htmlFor="index-commit-max-age">Commit max age</label>
                <DurationSelect
                    id="index-commit-max-age"
                    value={policy.indexCommitMaxAgeHours ? `${policy.indexCommitMaxAgeHours}` : null}
                    onChange={indexCommitMaxAgeHours => updatePolicy({ indexCommitMaxAgeHours })}
                    disabled={!policy.indexingEnabled}
                />
            </div>
            {policy.type === GitObjectType.GIT_TREE && (
                <div className="form-group">
                    <Toggle
                        id="index-intermediate-commits"
                        title="Enabled"
                        value={policy.indexIntermediateCommits}
                        onToggle={indexIntermediateCommits => updatePolicy({ indexIntermediateCommits })}
                        disabled={!policy.indexingEnabled}
                    />
                    <label htmlFor="index-intermediate-commits" className="ml-2">
                        Index intermediate commits
                    </label>
                </div>
            )}
        </div>
    )
}
