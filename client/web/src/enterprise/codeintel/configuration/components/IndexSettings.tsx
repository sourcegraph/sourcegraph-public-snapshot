import React, { FunctionComponent } from 'react'

import { Toggle } from '@sourcegraph/branded/src/components/Toggle'
import { Alert, Typography } from '@sourcegraph/wildcard'

import { RadioButtons } from '../../../../components/RadioButtons'
import { CodeIntelligenceConfigurationPolicyFields, GitObjectType } from '../../../../graphql-operations'
import { nullPolicy } from '../hooks/types'

import { DurationSelect } from './DurationSelect'

// This uses the same styles as the RetentionSettings component to style radio buttons
import styles from './RetentionSettings.module.scss'

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

export const IndexingSettings: FunctionComponent<React.PropsWithChildren<IndexingSettingsProps>> = ({
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

    const radioButtons = [
        {
            id: 'disable-indexing',
            label: 'Disable for this policy',
        },
        {
            id: 'enable-indexing',
            label: 'Enable for this policy, for commits with a maximum age',
        },
    ]

    const onChange = (event: React.ChangeEvent<HTMLInputElement>): void => {
        const indexingEnabled = event.target.value === 'enable-indexing'
        updatePolicy({ indexingEnabled })
    }

    return (
        <div className="form-group">
            <Typography.H3>Auto-indexing</Typography.H3>
            <div className="mb-4 form-group">
                <RadioButtons
                    nodes={radioButtons}
                    name="toggle-indexing"
                    onChange={onChange}
                    selected={policy.indexingEnabled ? 'enable-indexing' : 'disable-indexing'}
                    className={styles.radioButtons}
                />

                {!allowGlobalPolicies &&
                    repo === undefined &&
                    (policy.repositoryPatterns || []).length === 0 &&
                    policy.indexingEnabled && (
                        <Alert variant="danger">
                            This Sourcegraph instance has disabled global policies for auto-indexing. Create a more
                            constrained policy targeting an explicit set of repositories to enable this policy.
                        </Alert>
                    )}

                <label className="ml-4" htmlFor="index-commit-max-age">
                    Commit max age
                </label>
                <DurationSelect
                    id="index-commit-max-age"
                    value={policy.indexCommitMaxAgeHours ? `${policy.indexCommitMaxAgeHours}` : null}
                    onChange={indexCommitMaxAgeHours => updatePolicy({ indexCommitMaxAgeHours })}
                    disabled={!policy.indexingEnabled}
                    className="ml-4"
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
