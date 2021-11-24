import React, { FunctionComponent } from 'react'

import { Toggle } from '@sourcegraph/branded/src/components/Toggle'

import { CodeIntelligenceConfigurationPolicyFields, GitObjectType } from '../../../../graphql-operations'
import { nullPolicy } from '../hooks/types'

import { DurationSelect } from './DurationSelect'

export interface RetentionSettingsProps {
    policy: CodeIntelligenceConfigurationPolicyFields
    setPolicy: (
        updater: (
            policy: CodeIntelligenceConfigurationPolicyFields | undefined
        ) => CodeIntelligenceConfigurationPolicyFields
    ) => void
}

export const RetentionSettings: FunctionComponent<RetentionSettingsProps> = ({ policy, setPolicy }) => {
    const updatePolicy = <K extends keyof CodeIntelligenceConfigurationPolicyFields>(
        updates: { [P in K]: CodeIntelligenceConfigurationPolicyFields[P] }
    ): void => {
        setPolicy(policy => ({ ...(policy || nullPolicy), ...updates }))
    }

    return (
        <div className="form-group">
            <h3>Retention</h3>

            <div className="form-group">
                <Toggle
                    id="retention-enabled"
                    title="Enabled"
                    value={policy.retentionEnabled}
                    onToggle={retentionEnabled => updatePolicy({ retentionEnabled })}
                    disabled={policy.protected}
                />
                <label htmlFor="retention-enabled" className="ml-2">
                    Enabled / disabled
                </label>
            </div>

            <div className="form-group">
                <label htmlFor="retention-duration">Duration</label>

                <DurationSelect
                    id="retention-duration"
                    value={policy.retentionDurationHours ? `${policy.retentionDurationHours}` : null}
                    onChange={retentionDurationHours => updatePolicy({ retentionDurationHours })}
                    disabled={!policy.retentionEnabled}
                />
            </div>

            {policy.type === GitObjectType.GIT_TREE && (
                <div className="form-group">
                    <Toggle
                        id="retain-intermediate-commits"
                        title="Enabled"
                        value={policy.retainIntermediateCommits}
                        onToggle={retainIntermediateCommits => updatePolicy({ retainIntermediateCommits })}
                        disabled={policy.protected || !policy.retentionEnabled}
                    />
                    <label htmlFor="retain-intermediate-commits" className="ml-2">
                        Retain intermediate commits
                    </label>
                </div>
            )}
        </div>
    )
}
