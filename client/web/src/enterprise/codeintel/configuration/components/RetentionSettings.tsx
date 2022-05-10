import React, { FunctionComponent } from 'react'

import { Toggle } from '@sourcegraph/branded/src/components/Toggle'
import { Typography } from '@sourcegraph/wildcard'

import { RadioButtons } from '../../../../components/RadioButtons'
import { CodeIntelligenceConfigurationPolicyFields, GitObjectType } from '../../../../graphql-operations'
import { nullPolicy } from '../hooks/types'

import { DurationSelect } from './DurationSelect'

import styles from './RetentionSettings.module.scss'

export interface RetentionSettingsProps {
    policy: CodeIntelligenceConfigurationPolicyFields
    setPolicy: (
        updater: (
            policy: CodeIntelligenceConfigurationPolicyFields | undefined
        ) => CodeIntelligenceConfigurationPolicyFields
    ) => void
}

export const RetentionSettings: FunctionComponent<React.PropsWithChildren<RetentionSettingsProps>> = ({
    policy,
    setPolicy,
}) => {
    const updatePolicy = <K extends keyof CodeIntelligenceConfigurationPolicyFields>(
        updates: { [P in K]: CodeIntelligenceConfigurationPolicyFields[P] }
    ): void => {
        setPolicy(policy => ({ ...(policy || nullPolicy), ...updates }))
    }

    const radioButtons = [
        {
            id: 'disable-retention',
            label: 'Disable for this policy',
        },
        {
            id: 'enable-retention',
            label: 'Enable for this policy, keep data for specific duration',
        },
    ]

    const onChange = (event: React.ChangeEvent<HTMLInputElement>): void => {
        const retentionEnabled = event.target.value === 'enable-retention'
        updatePolicy({ retentionEnabled })
    }

    return (
        <div className="form-group">
            <Typography.H3>Retention</Typography.H3>

            <div className="form-group">
                <RadioButtons
                    nodes={radioButtons}
                    name="toggle-retention"
                    onChange={onChange}
                    selected={policy.retentionEnabled ? 'enable-retention' : 'disable-retention'}
                    className={styles.radioButtons}
                />

                <label className="ml-4" htmlFor="retention-duration">
                    Duration
                </label>
                <DurationSelect
                    id="retention-duration"
                    value={policy.retentionDurationHours ? `${policy.retentionDurationHours}` : null}
                    onChange={retentionDurationHours => updatePolicy({ retentionDurationHours })}
                    disabled={!policy.retentionEnabled}
                    className="ml-4"
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
