import React, { FunctionComponent } from 'react'

import { Toggle } from '@sourcegraph/branded/src/components/Toggle'
import { Container } from '@sourcegraph/wildcard'

import { CodeIntelligenceConfigurationPolicyFields, GitObjectType } from '../../../graphql-operations'

import { DurationSelect } from './DurationSelect'

export interface RetentionSettingsProps {
    policy: CodeIntelligenceConfigurationPolicyFields
    setPolicy: (policy: CodeIntelligenceConfigurationPolicyFields) => void
}

export const RetentionSettings: FunctionComponent<RetentionSettingsProps> = ({ policy, setPolicy }) => (
    <Container className="mt-2">
        <h3>Retention</h3>

        <div className="form-group">
            <Toggle
                id="retention-enabled"
                title="Enabled"
                value={policy.retentionEnabled}
                onToggle={value => setPolicy({ ...policy, retentionEnabled: value })}
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
                onChange={value => setPolicy({ ...policy, retentionDurationHours: value })}
                disabled={!policy.retentionEnabled}
            />
        </div>

        {policy.type === GitObjectType.GIT_TREE && (
            <div className="form-group">
                <Toggle
                    id="retain-intermediate-commits"
                    title="Enabled"
                    value={policy.retainIntermediateCommits}
                    onToggle={value => setPolicy({ ...policy, retainIntermediateCommits: value })}
                    disabled={!policy.retentionEnabled}
                />
                <label htmlFor="retain-intermediate-commits" className="ml-2">
                    Retain intermediate commits
                </label>
            </div>
        )}
    </Container>
)
