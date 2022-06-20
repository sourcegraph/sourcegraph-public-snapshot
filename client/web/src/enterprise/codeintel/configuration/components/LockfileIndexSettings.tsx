import React, { FunctionComponent } from 'react'

import { Alert, H3 } from '@sourcegraph/wildcard'

import { RadioButtons } from '../../../../components/RadioButtons'
import { CodeIntelligenceConfigurationPolicyFields } from '../../../../graphql-operations'
import { nullPolicy } from '../hooks/types'

// This uses the same styles as the RetentionSettings component to style radio buttons
import styles from './RetentionSettings.module.scss'

export interface LockfileIndexingSettingsProps {
    policy: CodeIntelligenceConfigurationPolicyFields
    repo?: { id: string }
    setPolicy: (
        updater: (
            policy: CodeIntelligenceConfigurationPolicyFields | undefined
        ) => CodeIntelligenceConfigurationPolicyFields
    ) => void
    allowGlobalPolicies?: boolean
}

export const LockfileIndexingSettings: FunctionComponent<React.PropsWithChildren<LockfileIndexingSettingsProps>> = ({
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
            id: 'disable-lockfile-indexing',
            label: 'Disable for this policy',
        },
        {
            id: 'enable-lockfile-indexing',
            label: 'Enable for this policy',
        },
    ]

    const onChange = (event: React.ChangeEvent<HTMLInputElement>): void => {
        const lockfileIndexingEnabled = event.target.value === 'enable-lockfile-indexing'
        updatePolicy({ lockfileIndexingEnabled })
    }

    return (
        <div className="form-group">
            <H3>Lockfile-indexing</H3>
            <div className="mb-4 form-group">
                <RadioButtons
                    nodes={radioButtons}
                    name="toggle-lockfile-indexing"
                    onChange={onChange}
                    selected={policy.lockfileIndexingEnabled ? 'enable-lockfile-indexing' : 'disable-lockfile-indexing'}
                    className={styles.radioButtons}
                />

                {!allowGlobalPolicies &&
                    repo === undefined &&
                    (policy.repositoryPatterns || []).length === 0 &&
                    policy.lockfileIndexingEnabled && (
                        <Alert variant="danger">
                            This Sourcegraph instance has disabled global policies for lockfile-indexing. Create a more
                            constrained policy targeting an explicit set of repositories to enable this policy.
                        </Alert>
                    )}
            </div>
        </div>
    )
}
