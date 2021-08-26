import React, { FunctionComponent } from 'react'

import { Toggle } from '@sourcegraph/branded/src/components/Toggle'
import { Container } from '@sourcegraph/wildcard'

import { CodeIntelligenceConfigurationPolicyFields, GitObjectType } from '../../../graphql-operations'

import { DurationSelect } from './DurationSelect'

export interface IndexingSettingsProps {
    policy: CodeIntelligenceConfigurationPolicyFields
    setPolicy: (policy: CodeIntelligenceConfigurationPolicyFields) => void
}

export const IndexingSettings: FunctionComponent<IndexingSettingsProps> = ({ policy, setPolicy }) => (
    <Container className="mt-2">
        <h3>Auto-indexing</h3>

        <div className="form-group">
            <Toggle
                id="indexing-enabled"
                title="Enabled"
                value={policy.indexingEnabled}
                onToggle={value => setPolicy({ ...policy, indexingEnabled: value })}
            />
            <label htmlFor="indexing-enabled" className="ml-2">
                Enabled / disabled
            </label>
        </div>

        <div className="form-group">
            <label htmlFor="index-commit-max-age">Commit max age</label>
            <DurationSelect
                id="index-commit-max-age"
                value={`${policy.indexCommitMaxAgeHours}`}
                disabled={!policy.indexingEnabled}
                onChange={value => setPolicy({ ...policy, indexCommitMaxAgeHours: value })}
            />
        </div>

        {policy.type === GitObjectType.GIT_TREE && (
            <div className="form-group">
                <Toggle
                    id="index-intermediate-commits"
                    title="Enabled"
                    value={policy.indexIntermediateCommits}
                    onToggle={value => setPolicy({ ...policy, indexIntermediateCommits: value })}
                    disabled={!policy.indexingEnabled}
                />
                <label htmlFor="index-intermediate-commits" className="ml-2">
                    Index intermediate commits
                </label>
            </div>
        )}
    </Container>
)
